package k8sUtils

import (

	"log"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"io/ioutil"
	"github.com/guipal/init-acid/utils"
	"path/filepath"
	"strconv"
	"strings"
)

func NewK8sClient(kubeconfig *string) *kubernetes.Clientset{
	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	return clientset
}

func CreateK8sNamespace(clientset *kubernetes.Clientset, k8sNamespace string) {

	_, err := clientset.CoreV1().Namespaces().Get(k8sNamespace,metav1.GetOptions{})
	if errors.IsNotFound(err) {
		nsSpec := &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: k8sNamespace}}
		_,err=clientset.CoreV1().Namespaces().Create(nsSpec)
		if err != nil {
			log.Fatal("Problem creating namespace: ",err)
		}
	}

}

func CreateImagePullSecret(clientset *kubernetes.Clientset, k8sNamespace, dockerRepo, dockerUser, dockerPassword string) {

	// Get Pod by name
	_,err:=clientset.CoreV1().Secrets(k8sNamespace).Get("acidRegistry",metav1.GetOptions{})

	if errors.IsNotFound(err) {
		secretString :="{\"auths\":{\""+dockerRepo+"\":{\"username\":\""+dockerUser+"\",\"password\":\""+dockerPassword+"\",\"email\":\"acid@test.com\"}}}"
		secret := map[string]string{".dockerconfigjson": secretString}
		secretSpec := &v1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "acidregistry"},StringData:secret,Type:v1.SecretTypeDockerConfigJson}
		_,err =clientset.CoreV1().Secrets(k8sNamespace).Create(secretSpec)
		if err != nil {
			log.Fatal("Problem creating secret: ",err)
		}
	}else{
		log.Fatal("Problem getting secret: ",err)
	}


	// --docker-server=${dockerRepo}  --docker-username=$DOCKER_USER --docker-password=$DOCKER_PASS"

}

func CreateDeployment(clientset *kubernetes.Clientset, instanceName, k8sNamespace, dockerRepo,imageVersion,configGitURL,configBranchName string){
	deploymentsClient := clientset.AppsV1().Deployments(k8sNamespace)

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: instanceName,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": instanceName,
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": instanceName,
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  instanceName,
							Image: dockerRepo+"/acid_master:"+imageVersion,
							ImagePullPolicy: v1.PullAlways,
							Env: []v1.EnvVar{
								{
									Name: "JENKINS_CONFIG_GIT_URL",
									Value: configGitURL,
								},
								{
									Name: "JENKINS_CONFIG_GIT_BRANCH",
									Value: configBranchName,
								},
							},
							Ports: []v1.ContainerPort{
								{
									Protocol:      v1.ProtocolTCP,
									ContainerPort: 8080,
								},
								{
									Protocol:      v1.ProtocolTCP,
									ContainerPort: 50000,
								},
							},
							VolumeMounts: []v1.VolumeMount{
								{
									MountPath: "/opt/jenkins/ssh",
									ReadOnly: true,
									Name: "ssh-keys",
								},

							},
						},
					},
					Volumes: []v1.Volume{
						{
							Name: "ssh-keys",
							VolumeSource: v1.VolumeSource{
								Secret: &v1.SecretVolumeSource{
									SecretName: "ssh-key-secret",
								},
							},
						},
					},
					ImagePullSecrets: []v1.LocalObjectReference{
						{
							Name: "acidregistry",
						},

					},
				},
			},
		},
	}

	// Create Deployment
	fmt.Println("Creating deployment...")
	result, err := deploymentsClient.Create(deployment)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Created deployment %q.\n", result.GetObjectMeta().GetName())
}

func CreateService(clientset *kubernetes.Clientset, instanceName, k8sNamespace, dockerRepo,imageVersion,configGitURL,configBranchName string){
	servicesClient := clientset.CoreV1().Services(k8sNamespace)

	service := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: instanceName,
		},
		Spec: v1.ServiceSpec{
			Selector: map[string]string{
				"app":instanceName,
			},
			Ports: []v1.ServicePort{
				{
					Name:"http",
					Port: 8080,
				},
				{
					Name:"slave",
					Port: 50000,
				},
			},
			Type: v1.ServiceTypeLoadBalancer,
		},
	}

	// Create Deployment
	fmt.Println("Creating service...")
	result, err := servicesClient.Create(service)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Created service %q.\n", result.GetObjectMeta().GetName())
}

func CreateIngress(clientset *kubernetes.Clientset, instanceName, k8sNamespace,k8sCluster, dockerRepo,imageVersion,configGitURL,configBranchName string){
	ingressClient := clientset.ExtensionsV1beta1().Ingresses(k8sNamespace)

	ingress := &v1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name: instanceName,
			Labels: map[string]string{
				"name": instanceName,
			},
			Annotations: map[string]string{
				"ingress.kubernetes.io/proxy-body-size": "8m",
			},
		},
		Spec: v1beta1.IngressSpec{
			Rules: []v1beta1.IngressRule{
				{
					Host: instanceName+"-"+k8sCluster,
					IngressRuleValue: v1beta1.IngressRuleValue{
						HTTP: &v1beta1.HTTPIngressRuleValue{
							Paths: []v1beta1.HTTPIngressPath{
								{
									Path:"/",
									Backend: v1beta1.IngressBackend{
										ServiceName: instanceName,
										ServicePort: intstr.IntOrString{
											IntVal: 8080,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// Create Deployment
	fmt.Println("Creating service...")
	result, err := ingressClient.Create(ingress)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Created ingress %q.\n", result.GetObjectMeta().GetName())
}

func CreateSSHSecret(clientset *kubernetes.Clientset, k8sNamespace, dockerRepo, dockerUser, dockerPassword,instance,config,privateKey,publicKey,knownHosts string) {

	// Get Pod by name
	_,err:=clientset.CoreV1().Secrets(k8sNamespace).Get("ssh-key-secret",metav1.GetOptions{})

	if errors.IsNotFound(err) {
		//secretString :="{\"config\":{\""+config+"\":{\"id_rsa."+instance+"\":\""+privateKey+"\",\"id_rsa."+instance+".pub\":\""+publicKey+"\",\"known_hosts\":\""+knownHosts+"\"}}"
		secret := map[string]string{"config":config,
									"id_rsa."+instance: privateKey,
									"id_rsa."+instance+".pub": publicKey,
									"known_hosts."+instance: knownHosts,
			}
		secretSpec := &v1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "ssh-key-secret"},StringData:secret,Type:v1.SecretTypeOpaque}
		_,err =clientset.CoreV1().Secrets(k8sNamespace).Create(secretSpec)
		if err != nil {
			log.Fatal("Problem creating secret: ",err)
		}
	}else{
		log.Fatal("Problem getting secret: ",err)
	}

}

func GetSSHData(path,host string,port int,instance,user,password string) (config,privateKey,publicKey,knownHosts string){
	privateKeyName:="id_rsa."+instance
	publicKeyName:="id_rsa."+instance+".pub"
	knownHostsName:="known_hosts."+instance
	if path == ""{
		path=""


		 bitSize := 4096

		 privateKeySeed, err := utils.GeneratePrivateKey(bitSize)
		 check(err)

		 publicKeyBytes, err := utils.GeneratePublicKey(&privateKeySeed.PublicKey)
		 check(err)

		 privateKeyBytes := utils.EncodePrivateKeyToPEM(privateKeySeed)

		 err = utils.WriteKeyToFile(privateKeyBytes, filepath.Join(path,privateKeyName))
		 check(err)

		 err = utils.WriteKeyToFile([]byte(publicKeyBytes), filepath.Join(path,publicKeyName))
		 check(err)

		d1 := []byte("Host "+host+"\n  User git\n  Port "+strconv.Itoa(port)+"\n  IdentityFile ~/.ssh/"+privateKeyName+"\n  UserKnownHostsFile ~/.ssh/"+knownHostsName)
		err = ioutil.WriteFile(filepath.Join(path,"config"), d1, 0644)
		check(err)

		d1, err = utils.GetHostKey(host,strconv.Itoa(port),user,password)
		check(err)
		hostKey := append([]byte(host+",* "),d1...)
		err = ioutil.WriteFile(filepath.Join(path,knownHostsName), hostKey, 0644)
		check(err)



	}

	config = fileContentToString(filepath.Join(path,"config"))
	privateKey = fileContentToString(filepath.Join(path,privateKeyName))
	publicKey = fileContentToString(filepath.Join(path,publicKeyName))
	knownHosts = fileContentToString(filepath.Join(path,knownHostsName))

	 return
}

func fileContentToString(path string) string{
	dat, err := ioutil.ReadFile(path)
	check(err)
	return strings.TrimSpace(string(dat))

}


func check(e error) {
	if e != nil {
		panic(e)
	}
}

func int32Ptr(i int32) *int32 { return &i }
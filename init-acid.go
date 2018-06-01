package main



import (
"flag"
"fmt"
"os"
	"os/exec"
	"io"
	"log"
	"time"
	"github.com/guipal/init-acid/incubator"
	"strings"
	"text/template"
	//"k8s.io/apimachinery/pkg/api/errors"

	"path/filepath"

	"bufio"
	"golang.org/x/crypto/ssh/terminal"
	"syscall"

	"github.com/guipal/init-acid/k8sUtils"
	"github.com/guipal/init-acid/githubUtils"
)



func main() {

	var instance,  k8scluster, k8snamespace, dockeruser, dockerpass,gituser,gitpass, dockerepo, imageversion,configrepo,configbranch,path,host,gitProject  string
	var verbose,test bool
	var port int
	var kubeconfig *string

	flag.StringVar(&instance, "instanceName", "acid-setup", "")
	//flag.StringVar(&kubeconfig, "kubeconfigFile", "~/.kube/config", "")
	flag.StringVar(&k8scluster, "k8sClusterURL", "localhost", "")
	flag.StringVar(&k8snamespace, "k8sNamespace", "acid-setup", "Changeset TAG")
	flag.StringVar(&dockeruser, "dockerUser", "", "")
	flag.StringVar(&dockerpass, "dockerPassword", "", "")
	flag.StringVar(&gituser, "gitUser", "", "")
	flag.StringVar(&gitpass, "gitPassword", "", "")
	flag.StringVar(&dockerepo, "dockerRepo", "docker.io/acid3stripes", "Docker repo to download base image without / at the end")
	flag.StringVar(&imageversion, "imageVersion", "latest", "")
	flag.StringVar(&configrepo, "configRepo", "", "")
	flag.StringVar(&configbranch, "configBranch", "master", "")
	flag.StringVar(&path, "sshDirectory", "", "")
	flag.StringVar(&host, "gitHost", "github.com", "")
	flag.StringVar(&gitProject, "gitProject", "", "Github organization")
	flag.IntVar(&port, "gitPort", 22, "")
	flag.BoolVar(&verbose, "verbose", false, "Verbose otuput")
	flag.BoolVar(&test, "test", false, "Test otuput")

	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
		}
		flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "OPTIONS:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	//args := flag.Args()

	if ! test {

		createTmpDir()
		os.Chdir("tmp")
		defer removeTmpDir()
		defer deleteDockerInstance(instance,dockerepo,verbose)

		startDockerInstance(instance,dockerepo,verbose)
		time.Sleep(20 * time.Second)
		password:= showInitialPassword(instance,dockerepo,verbose)
		fmt.Println("Go to http://localhost:10000 and use: ",password)
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Press ENTER whenever you are done with your configuration ")
		reader.ReadString('\n')
		pauseDocker(instance,dockerepo,verbose)
		copyHome(instance,dockerepo,verbose)
		resumeDocker(instance,dockerepo,verbose)
		downloadPluginList(strings.TrimSpace(password))
		gituser,gitpass=credentials("Credentials for git account")
		githubUtils.CreateRepo(gituser,gitpass,configrepo,gitProject)
		config,privateKey,publicKey,knownHosts  := k8sUtils.GetSSHData(path,host,port,instance,"guipal","e99e7dd5ce7d93e44c7702333fdea0c7c4749c80")
		githubUtils.CreateRepoKey(gituser,gitpass,configrepo,gitProject,"ssh-key-"+instance,publicKey)
		initRepository(host+":"+gitProject+"/"+configrepo,verbose)
		dockeruser,dockerpass=credentials("Credentials for docker registry")
		k8sclient:=k8sUtils.NewK8sClient(kubeconfig)
		k8sUtils.CreateK8sNamespace(k8sclient,k8snamespace)
		k8sUtils.CreateImagePullSecret(k8sclient,k8snamespace,dockerepo,dockeruser,dockerpass)
		k8sUtils.CreateDeployment(k8sclient,instance,k8snamespace,dockerepo,imageversion,host+":"+gitProject+"/"+configrepo,configbranch)
		k8sUtils.CreateService(k8sclient,instance,k8snamespace,dockerepo,imageversion,host+":"+gitProject+"/"+configrepo,configbranch)
		k8sUtils.CreateIngress(k8sclient,instance,k8snamespace,k8scluster,dockerepo,imageversion,host+":"+gitProject+"/"+configrepo,configbranch)
		k8sUtils.CreateSSHSecret(k8sclient,k8snamespace,dockerepo,dockeruser,dockerpass,instance,config,privateKey,publicKey,knownHosts)
	}else{
		config,privateKey,publicKey,knownHosts  := k8sUtils.GetSSHData(path,host,port,instance,gituser,gitpass)
		githubUtils.CreateRepoKey(gituser,gitpass,configrepo,gitProject,"ssh-key-"+instance,publicKey)
		dockeruser,dockerpass=credentials("Credentials for docker registry")
		k8sclient:=k8sUtils.NewK8sClient(kubeconfig)
		k8sUtils.CreateK8sNamespace(k8sclient,k8snamespace)
		k8sUtils.CreateImagePullSecret(k8sclient,k8snamespace,dockerepo,dockeruser,dockerpass)
		k8sUtils.CreateDeployment(k8sclient,instance,k8snamespace,dockerepo,imageversion,host+":"+gitProject+"/"+configrepo,configbranch)
		k8sUtils.CreateService(k8sclient,instance,k8snamespace,dockerepo,imageversion,host+":"+gitProject+"/"+configrepo,configbranch)
		k8sUtils.CreateIngress(k8sclient,instance,k8snamespace,k8scluster,dockerepo,imageversion,host+":"+gitProject+"/"+configrepo,configbranch)
		k8sUtils.CreateSSHSecret(k8sclient,k8snamespace,dockerepo,dockeruser,dockerpass,instance,config,privateKey,publicKey,knownHosts)
	}

}

func removeTmpDir() {
	os.Chdir("../")
	os.RemoveAll("tmp")
}

func startDockerInstance(instance,dockerRepo string,verbose bool){
	cmd := "docker"
	args := []string{"run", "-d","-p","10000:8080","--name", instance, dockerRepo+"/acid_master"}
	executeCommand(cmd,args,verbose)

}

func deleteDockerInstance(instance,dockerRepo string,verbose bool){
	cmd := "docker"
	args := []string{"rm", "-f", instance}
	executeCommand(cmd,args,verbose)
}

func waitDocker(instance,dockerRepo string,verbose bool){
	cmd := "docker"
	args := []string{"wait", instance}
	executeCommand(cmd,args,verbose)
}

func showInitialPassword(instance,dockerRepo string,verbose bool) string {
	cmd := "docker"
	args := []string{"exec", instance, "cat", "/var/jenkins_home/secrets/initialAdminPassword"}
	var (
		cmdOut []byte
		err    error
	)
	if cmdOut, err = exec.Command(cmd, args...).Output(); err != nil {
		fmt.Fprintln(os.Stderr, "There was an error running %s: ", cmd,err)
		os.Exit(1)
	}
	return string(cmdOut)
}

func copyHome(instance,dockerRepo string,verbose bool) {
	cmd := "docker"
	args := []string{"cp", instance+":/var/jenkins_home/","./"}
	executeCommand(cmd,args,verbose)
}

func startDocker(instance,dockerRepo string,verbose bool) {
	cmd := "docker"
	args := []string{"start", instance}
	executeCommand(cmd,args,verbose)
}

func pauseDocker(instance,dockerRepo string,verbose bool) {
	cmd := "docker"
	args := []string{"pause", instance}
	executeCommand(cmd,args,verbose)
}

func resumeDocker(instance,dockerRepo string,verbose bool) {
	cmd := "docker"
	args := []string{"unpause", instance}
	executeCommand(cmd,args,verbose)
}

func downloadPluginList(password string) {

	jenkins:=incubator.Jenkins{
		User:"admin",Password:password,Uri:"http://localhost:10000",
	}

	fmt.Println(jenkins)

	incubator.GetPlugins(jenkins,false, func(content []byte, filename string) {
		storeResults(content,filename)
	})
	addLinetoFile("kubernetes","jenkins_home/plugins.txt")
	addLinetoFile("scm-sync-configuration","jenkins_home/plugins.txt")


}

func storeResults(result []byte,fileName string) {
	var path string

	path = "jenkins_home/" + fileName;

	f, _ := os.Create(path)
	defer f.Close()
	_, err := f.Write(result)
	f.Sync()
	if err != nil {
		fmt.Println(err)
	}

}

func addLinetoFile(text,filename string){
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	fmt.Fprintln(f,text)

	//if _, err = f.WriteString(text); err != nil {
	//	panic(err)
	//}
}

func initRepository(repoURL string,verbose bool){

	os.Chdir("jenkins_home")
	defer os.Chdir("../")

	createSyncScmConfigFile(repoURL,verbose)

	cmd := "git"
	args := []string{"init"}
	executeCommand(cmd,args,verbose)

	cmd = "git"
	args = []string{"add", "config.xml","scm-sync-configuration.xml", "plugins.txt", "nodes", "users", "secrets", "jobs"}
	executeCommand(cmd,args,verbose)

	cmd = "git"
	args = []string{"commit", "-m","Initial commit"}
	executeCommand(cmd,args,verbose)

	cmd = "git"
	args = []string{"remote", "add", "origin", "git@"+repoURL}
	executeCommand(cmd,args,verbose)

	cmd = "git"
	args = []string{"push", "-u" , "origin", "master"}
	executeCommand(cmd,args,verbose)

}

func createSyncScmConfigFile(repoURL string, verbose bool){
	type repo struct {
		//exported field since it begins
		//with a capital letter
		JenkinsCongigGitURL string
	}

	f, err := os.Create("scm-sync-configuration.xml")
	if err != nil {
		panic(err)
	}

	defer f.Close()

		//define an instance
		repoinfo := repo{repoURL}

		//create a new template with some name
		tmpl := template.New("test")
		configTpl :="<?xml version='1.1' encoding='UTF-8'?><hudson.plugins.scm__sync__configuration.ScmSyncConfigurationPlugin version='1'><scm class='hudson.plugins.scm_sync_configuration.scms.ScmSyncGitSCM'/><scmRepositoryUrl>scm:git:{{.JenkinsCongigGitURL}}</scmRepositoryUrl><noUserCommitMessage>true</noUserCommitMessage><displayStatus>true</displayStatus><commitMessagePattern>[message]</commitMessagePattern><manualSynchronizationIncludes/></hudson.plugins.scm__sync__configuration.ScmSyncConfigurationPlugin>"
		//parse some content and generate a template
		tmpl, err = tmpl.Parse(configTpl)
		if err != nil {
			log.Fatal("Parse: ", err)
			return
		}

		//merge template 'tmpl' with content of 's'
		err1 := tmpl.Execute(f, repoinfo)
		if err1 != nil {
			log.Fatal("Execute: ", err1)
			return
		}


}



func executeCommand(cmd string, args []string,verbose bool) {
	command := exec.Command(cmd, args...)
	command.Stdin = os.Stdin
	writer := io.MultiWriter(os.Stdout)
	command.Stdout = writer
	command.Stderr = os.Stderr
	if verbose {
		command.Stdout = os.Stdout
	}
	if err := command.Run(); err != nil {
		//fmt.Println("Not able to verify access")
		log.Println(err)
	}
}

func createTmpDir() {
	if exists, _ := exists("tmp"); exists {
	} else {
		os.Mkdir("tmp", 0777)
	}
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

func credentials(prompt string) (string, string) {
	fmt.Println(prompt)
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter Username: ")
	username, _ := reader.ReadString('\n')

	fmt.Print("Enter Password: ")
	bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		log.Println(err)
	}
	password := string(bytePassword)

	return strings.TrimSpace(username), strings.TrimSpace(password)
}

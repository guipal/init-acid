package utils

import (
"crypto/rand"
"crypto/rsa"
"crypto/x509"
"encoding/pem"
"golang.org/x/crypto/ssh"
"io/ioutil"
"log"
"net"
	"time"
)

func main() {
	savePrivateFileTo := "./id_rsa_test"
	savePublicFileTo := "./id_rsa_test.pub"
	bitSize := 4096

	privateKey, err := GeneratePrivateKey(bitSize)
	if err != nil {
		log.Fatal(err.Error())
	}

	publicKeyBytes, err := GeneratePublicKey(&privateKey.PublicKey)
	if err != nil {
		log.Fatal(err.Error())
	}

	privateKeyBytes := EncodePrivateKeyToPEM(privateKey)

	err = WriteKeyToFile(privateKeyBytes, savePrivateFileTo)
	if err != nil {
		log.Fatal(err.Error())
	}

	err = WriteKeyToFile([]byte(publicKeyBytes), savePublicFileTo)
	if err != nil {
		log.Fatal(err.Error())
	}
}

// generatePrivateKey creates a RSA Private Key of specified byte size
func GeneratePrivateKey(bitSize int) (*rsa.PrivateKey, error) {
	// Private Key generation
	privateKey, err := rsa.GenerateKey(rand.Reader, bitSize)
	if err != nil {
		return nil, err
	}

	// Validate Private Key
	err = privateKey.Validate()
	if err != nil {
		return nil, err
	}

	log.Println("Private Key generated")
	return privateKey, nil
}

// encodePrivateKeyToPEM encodes Private Key from RSA to PEM format
func EncodePrivateKeyToPEM(privateKey *rsa.PrivateKey) []byte {
	// Get ASN.1 DER format
	privDER := x509.MarshalPKCS1PrivateKey(privateKey)

	// pem.Block
	privBlock := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   privDER,
	}

	// Private key in PEM format
	privatePEM := pem.EncodeToMemory(&privBlock)

	return privatePEM
}

// generatePublicKey take a rsa.PublicKey and return bytes suitable for writing to .pub file
// returns in the format "ssh-rsa ..."
func GeneratePublicKey(privatekey *rsa.PublicKey) ([]byte, error) {
	publicRsaKey, err := ssh.NewPublicKey(privatekey)
	if err != nil {
		return nil, err
	}

	pubKeyBytes := ssh.MarshalAuthorizedKey(publicRsaKey)

	log.Println("Public key generated")
	return pubKeyBytes, nil
}

// writePemToFile writes keys to a file
func WriteKeyToFile(keyBytes []byte, saveFileTo string) error {
	err := ioutil.WriteFile(saveFileTo, keyBytes, 0600)
	if err != nil {
		return err
	}

	log.Printf("Key saved to: %s", saveFileTo)
	return nil
}

func GetHostKey(host,port,user,password string) (hostKey []byte, err error) {
	d := net.Dialer{Timeout: 5*time.Second}
	conn, err := d.Dial("tcp", host+":"+port)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	var key ssh.PublicKey

	config := ssh.ClientConfig{
		HostKeyAlgorithms: []string{ssh.KeyAlgoRSA},
		HostKeyCallback:   hostKeyCallback(&key),
		User:user,
		Auth: []ssh.AuthMethod{ssh.Password(password)},

	}
	sshconn, _, _, err := ssh.NewClientConn(conn, host, &config)

	if err == nil {
		sshconn.Close()
	}

	publicKeyBytes := ssh.MarshalAuthorizedKey(key)


	return publicKeyBytes, nil
}

func hostKeyCallback(publicKey *ssh.PublicKey) func(hostname string, remote net.Addr, key ssh.PublicKey) error {
	return func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		*publicKey = key
		return nil
	}
}



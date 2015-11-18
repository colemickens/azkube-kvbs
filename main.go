package main

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/colemickens/azkvbs/Godeps/_workspace/src/github.com/Azure/go-autorest/autorest"
	"github.com/colemickens/azkvbs/Godeps/_workspace/src/github.com/Azure/go-autorest/autorest/azure"
)

var cloudConfigPath string

var client *autorest.Client
var config configStruct

const vaultAPIVersion = "2015-06-01"
const secretURLTemplate = "https://{vault-name}.vault.azure.net/secrets/{secret-name}/{secret-version}?api-version={api-version}"

type configStruct struct {
	PrivateKeyPath  string `json:"privateKeyPath"`
	CertificatePath string `json:"certificatePath"`
	ApplicationID   string `json:"applicationId"`
	TenantID        string `json:"tenantId"`
	SubscriptionID  string `json:"subscriptionId"`
	VaultName       string `json:"vaultName"`
}

type secret struct {
	Value string `json:"value"`
}

func parseRsaPrivateKey(path string) (*rsa.PrivateKey, error) {
	privateKeyData, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalln("failed", err)
	}

	block, _ := pem.Decode(privateKeyData)
	if block == nil {
		panic("failed to decode a pem block from private key pem")
	}

	privatePkcs1Key, errPkcs1 := x509.ParsePKCS1PrivateKey(block.Bytes)
	if errPkcs1 == nil {
		return privatePkcs1Key, nil
	}

	privatePkcs8Key, errPkcs8 := x509.ParsePKCS8PrivateKey(block.Bytes)
	if errPkcs8 == nil {
		privatePkcs8RsaKey, ok := privatePkcs8Key.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("Pkcs8 contained non-RSA key. Expected RSA key.")
		}
		return privatePkcs8RsaKey, nil
	}

	return nil, fmt.Errorf("Failed to parse private key as Pkcs#1 or Pkcs#8. (%s). (%s).", errPkcs1, errPkcs8)
}

func init() {
	flag.StringVar(
		&cloudConfigPath,
		"cloudConfigPath",
		"/etc/kubernetes/azure-config.json",
		"path to the azure cloud config file used by kubernetes and this bootstrap tool")
	flag.Parse()

	configFile, err := os.Open(cloudConfigPath)
	if err != nil {
		panic(err)
	}

	decoder := json.NewDecoder(configFile)
	err = decoder.Decode(&config)
	if err != nil {
		panic(err)
	}

	log.Println("loading certificate... ")
	certificateData, err := ioutil.ReadFile(config.CertificatePath)
	if err != nil {
		log.Fatalln("failed", err)
	}

	log.Println("decoding certificate pem... ")
	block, _ := pem.Decode(certificateData)
	if block == nil {
		panic("failed to decode a pem block from certificate pem")
	}

	log.Println("parsing certificate... ")
	certificate, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		panic(err)
	}

	log.Println("parsing RSA key out of private key path")
	privateKey, err := parseRsaPrivateKey(config.PrivateKeyPath)
	if err != nil {
		panic(err)
	}

	log.Println("retrieve oauth token... ")
	spt, err := azure.NewServicePrincipalTokenFromRsaKey(
		config.ApplicationID,
		certificate,
		privateKey,
		config.TenantID,
		azure.AzureResourceManagerScope)
	if err != nil {
		log.Fatalln("failed", err)
		panic(err)
	}

	client = &autorest.Client{}
	client.Authorizer = spt
}

func getSecret(secretName string) (*string, error) {
	var p map[string]interface{}
	p = map[string]interface{}{
		"vault-name":     config.VaultName,
		"secret-name":    secretName,
		"secret-version": "",
	}
	q := map[string]interface{}{
		"api-version": vaultAPIVersion,
	}

	req, err := autorest.Prepare(&http.Request{},
		autorest.AsGet(),
		autorest.WithBaseURL(secretURLTemplate),
		autorest.WithPathParameters(p),
		autorest.WithQueryParameters(q))
	
	if err != nil {
		panic(err)
	}

	resp, err := client.Send(req, http.StatusOK)
	if err != nil {
		return nil, err
	}

	var secret secret

	err = autorest.Respond(
		resp,
		autorest.ByUnmarshallingJSON(&secret))
	if err != nil {
		return nil, err
	}

	secretValue, err := base64.StdEncoding.DecodeString(secret.Value)
	if err != nil {
		return nil, err
	}

	secretValueString := string(secretValue)

	return &secretValueString, nil
}

func main() {
	minionSecrets := map[string]string{
		"minion-proxy-kubeconfig":   "/etc/kubernetes/minion-proxy-kubeconfig",
		"minion-kubelet-kubeconfig": "/etc/kubernetes/minion-kubelet-kubeconfig",
	}

	masterSecrets := map[string]string{
		"ca-crt":                               "/etc/kubernetes/ca.crt",
		"apiserver-crt":                        "/etc/kubernetes/apiserver.crt",
		"apiserver-key":                        "/etc/kubernetes/apiserver.key",
		"master-proxy-kubeconfig":              "/etc/kubernetes/master-proxy-kubeconfig",
		"master-kubelet-kubeconfig":            "/etc/kubernetes/master-kubelet-kubeconfig",
		"master-scheduler-kubeconfig":          "/etc/kubernetes/master-scheduler-kubeconfig",
		"master-controller-manager-kubeconfig": "/etc/kubernetes/master-controller-manager-kubeconfig",
	}

	log.Println("starting up")

	machineType := os.Args[1]

	var secrets map[string]string
	switch machineType {
	case "master":
		secrets = masterSecrets
	case "minion":
		secrets = minionSecrets
	default:
		log.Fatalln("don't know machine type")
	}

	for secretName, secretPath := range secrets {
		secretValue, err := getSecret(secretName)
		if err != nil {
			// TODO(colemickens): retry?
			panic(err)
		}

		err = ioutil.WriteFile(secretPath, []byte(*secretValue), 0644)
		if err != nil {
			// TODO(colemickens): retry?
			panic(err)
		}
	}

	log.Println("done")
}

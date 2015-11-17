package main

import (
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/colemickens/azkvbs/Godeps/_workspace/src/github.com/Azure/go-autorest/autorest"
	"github.com/colemickens/azkvbs/Godeps/_workspace/src/github.com/Azure/go-autorest/autorest/azure"
)

const resourceGroupURLTemplate = "https://management.azure.com/subscriptions/{subscription-id}/resourcegroups"
const apiVersion = "2015-01-01"

var (
	certificatePath string
	applicationID   string
	tenantID        string
	subscriptionID  string
)

func init() {
	flag.StringVar(&certificatePath, "certificatePath", "", "path to pk12/pfx certificate")
	flag.StringVar(&applicationID, "applicationId", "", "application id")
	flag.StringVar(&tenantID, "tenantId", "", "tenant id")
	flag.StringVar(&subscriptionID, "subscriptionId", "", "subscription id")
	flag.Parse()

	log.Println("Using these settings:")
	log.Println("* certificatePath:", certificatePath)
	log.Println("* applicationID:", applicationID)
	log.Println("* tenantID:", tenantID)
	log.Println("* subscriptionID:", subscriptionID)

	if strings.Trim(certificatePath, " ") == "" ||
		strings.Trim(applicationID, " ") == "" ||
		strings.Trim(tenantID, " ") == "" ||
		strings.Trim(subscriptionID, " ") == "" {
		log.Fatalln("Bad usage. Please specify all four parameters")
	}
}

func main() {
	log.Println("loading certificate... ")
	certData, err := ioutil.ReadFile(certificatePath)
	if err != nil {
		log.Fatalln("failed", err)
	}

	log.Println("retrieve oauth token... ")
	spt, err := azure.NewServicePrincipalTokenFromCertificate(
		applicationID,
		certData,
		"",
		tenantID,
		azure.AzureResourceManagerScope)
	if err != nil {
		log.Fatalln("failed", err)
		panic(err)
	}

	client := &autorest.Client{}
	client.Authorizer = spt

	log.Println("querying the list of resource groups... ")
	groupsAsString, err := getResourceGroups(client)
	if err != nil {
		log.Fatalln("failed", err)
	}

	log.Println("")
	log.Println("Groups:", *groupsAsString)
}

func getResourceGroups(client *autorest.Client) (*string, error) {
	var p map[string]interface{}
	var req *http.Request
	p = map[string]interface{}{
		"subscription-id": subscriptionID,
	}
	q := map[string]interface{}{
		"api-version": apiVersion,
	}

	req, _ = autorest.Prepare(&http.Request{},
		autorest.AsGet(),
		autorest.WithBaseURL(resourceGroupURLTemplate),
		autorest.WithPathParameters(p),
		autorest.WithQueryParameters(q))

	resp, err := client.Send(req, http.StatusOK)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	contentsString := string(contents)

	return &contentsString, nil
}

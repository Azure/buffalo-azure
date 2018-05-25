// Copyright © 2018 Microsoft Corporation and contributors
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2017-05-10/resources"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/marstr/guid"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// clientID is used to identify this application during a Device Auth flow.
// We have deliberately spoofed as the Azure CLI 2.0 at least temporarily.
const deviceClientID = "04b07795-8ddb-461a-bbee-02f9e1bf7b46"

var provisionConfig = viper.New()

// These constants define a parameter which gives control over the Azure Resource Group that should be used to hold
// the created assets.
const (
	ResoureGroupName       = "resource-group"
	ResourceGroupShorthand = "g"
	resourceGroupUsage     = "The name of the Resource Group that should hold the resources created."
)

// These constants define a parameter which allows control over the Azure Region that should be used when creating a
// resource group. If the specified resource group already exists, its location is used and this parameter is discarded.
const (
	LocationName      = "location"
	LocationShorthand = "l"
	LocationDefault   = "centralus"
	locationUsage     = "The Azure Region that should be used when creating a resource group."
)

// These constants define a parameter that allows control of the type of database to be provisioned. This is largely an
// escape hatch to use if Buffalo-Azure is incorrectly identifying the flavor of database to use after reading your
// application.
//
// Supported flavors:
//  - None
//  - Postgres
const (
	DatabaseName      = "database"
	DatabaseShorthand = "d"
	databaseUsage     = "The type of database to provision."
)

// These constants define a parameter which allows control over the particular Azure cloud which should be used for
// deployment.
// Some examples of Azure environments by name include:
// - AzurePublicCloud (most popular)
// - AzureChinaCloud
// - AzureGermanCloud
// - AzureUSGovernmentCloud
const (
	EnvironmentName      = "environment"
	EnvironmentShorthand = "e"
	EnvironmentDefault   = "AzurePublicCloud"
	environmentUsage     = "The Azure environment that will be targeted for deployment."
)

var environment azure.Environment

// These constants define a parameter which will control which container image is used to
const (
	// ImageName is the full parameter name of the argument that controls which container image will be used
	// when the Web App for Containers is provisioned.
	ImageName = "image"

	// ImageShorthand is the abbreviated means of using ImageName.
	ImageShorthand = "i"

	// ImageDefault is the container image that will be deployed if you didn't specify one.
	ImageDefault = "appsvc/sample-hello-world:latest"

	imageUsage = "The container image that defines this project."
)

// These constants define a parameter that allows control of the Azure Resource Management (ARM) template that should be
// used to provision infrastructure. This tool is not designed to deploy arbitrary ARM templates, rather this parameter
// is intended to give you the flexibility to lock to a known version of the gobuffalo quick start template, or tweak
// that template a little for your own usage.
//
// To prevent live-site incidents, a local copy of the default template is stored in this executable. If this parameter
// is NOT specified, this program will attempt to acquire the remote version of the default-template. Should that fail,
// the locally cached copy will be used. If the parameter is specified, this program will attempt to acquire the remote
// version. If that operation fails, the program does NOT use the cached template, and terminates with a non-zero exit
// status.
const (
	// TemplateName is the full parameter name of the argument providing a URL where the ARM template to bue used can
	// be found.
	TemplateName = "rm-template"

	// TemplateShorthand is the abbreviated means of using TemplateName.
	TemplateShorthand = "t"

	// TemplateDefault
	TemplateDefault = "https://invalidtemplate.gobuffalo.io"
	templateUsage   = "The Azure Resource Management template used to "
)

// These constants define a parameter that Azure subscription to own the resources created.
//
// This can also be specified with the environment variable AZURE_SUBSCRIPTION_ID or AZ_SUBSCRIPTION_ID.
const (
	SubscriptionName      = "subscription"
	SubscriptionShorthand = "s"
	subscriptionUsage     = "The ID (in UUID format) of the Azure subscription which should host the provisioned resources."
)

// These constants define a parameter which allows specification of a Service Principal for authentication.
// This should always be used in tandem with `--client-secret`.
//
// This can also be specified with the environment variable AZURE_CLIENT_ID or AZ_CLIENT_ID.
//
// To learn more about getting started with Service Principals you can look here:
// - Using the Azure CLI 2.0: [https://docs.microsoft.com/en-us/cli/azure/create-an-azure-service-principal-azure-cli](https://docs.microsoft.com/en-us/cli/azure/create-an-azure-service-principal-azure-cli?toc=%2Fazure%2Fazure-resource-manager%2Ftoc.json&view=azure-cli-latest)
// - Using Azure PowerShell: [https://docs.microsoft.com/en-us/azure/azure-resource-manager/resource-group-authenticate-service-principal](https://docs.microsoft.com/en-us/azure/azure-resource-manager/resource-group-authenticate-service-principal?view=azure-cli-latest)
// - Using the Azure Portal: [https://docs.microsoft.com/en-us/azure/azure-resource-manager/resource-group-create-service-principal-portal](https://docs.microsoft.com/en-us/azure/azure-resource-manager/resource-group-create-service-principal-portal?view=azure-cli-latest)
const (
	ClientIDName  = "client-id"
	clientIDUsage = "The Application ID of the App Registration being used to authenticate."
)

// These constants define a parameter which allows specification of a Service Principal for authentication.
// This should always be used in tandem with `--client-id`.
//
// This can also be specified with the environment variable AZURE_CLIENT_SECRET or AZ_CLIENT_SECRET.
const (
	ClientSecretName  = "client-secret"
	clientSecretUsage = "The Key associated with the App Registration being used to authenticate."
)

// These constants define a parameter which provides the organization that should be used during authentication.
// Providing the tenant-id explicitly can help speed up execution, but by default this program will traverse all tenants
// available to the authenticated identity (service principal or user), and find the one containing the subscription
// provided. This traversal may involve several HTTP requests, and is therefore somewhat latent.
//
// This can also be specified with the environment variable AZURE_TENANT_ID or AZ_TENANT_ID.
const (
	TenantIDName = "tenant-id"
	tenantUsage  = "The ID (in form of a UUID) of the organization that the identity being used belongs to. "
)

const (
	DeviceAuthName  = "use-device-auth"
	deviceAuthUsage = "Ignore --client-id and --client-secret, interactively authenticate instead."
)

const (
	VerboseName      = "verbose"
	VerboseShortname = "v"
	verboseUsage     = "Print out status information as this program executes."
)

var status *log.Logger

// provisionCmd represents the provision command
var provisionCmd = &cobra.Command{
	Aliases: []string{"p"},
	Use:     "provision",
	Short:   "Create the infrastructure necessary to run a buffalo app on Azure.",
	Run: func(cmd *cobra.Command, args []string) {
		exitStatus := 1
		defer func() {
			os.Exit(exitStatus)
		}()

		// Authenticate and setup clients
		subscriptionID := provisionConfig.GetString(SubscriptionName)
		tenantID := provisionConfig.GetString(TenantIDName)
		clientID := provisionConfig.GetString(ClientIDName)
		clientSecret := provisionConfig.GetString(ClientSecretName)
		status.Print("subscription selected: ", subscriptionID)
		status.Print("tenant selected: ", tenantID)

		ctx, cancel := context.WithTimeout(context.Background(), 45*time.Minute)
		defer cancel()

		auth, err := getAuthorizer(ctx, clientID, clientSecret, tenantID)
		if err != nil {
			fmt.Fprintln(os.Stderr, "unable to authenticate: ", err)
			return
		}

		groups := resources.NewGroupsClient(subscriptionID)
		groups.Authorizer = auth
		userAgentBuilder := bytes.NewBufferString("buffalo-azure")
		if version != "" {
			userAgentBuilder.WriteRune('/')
			userAgentBuilder.WriteString(version)
		}
		groups.AddToUserAgent(userAgentBuilder.String())

		// Assert the presence of the specified Resource Group
		rgName := provisionConfig.GetString(ResoureGroupName)
		created, err := insertResourceGroup(ctx, groups, rgName, provisionConfig.GetString(LocationName))
		if err != nil {
			fmt.Fprintf(os.Stderr, "unable to fetch or create resource group %s: %v\n", rgName, err)
			return
		}
		if created {
			status.Println("created resource group: ", rgName)
		} else {
			status.Println("found resource group: ", rgName)
		}

		// Provision the necessary assets.
		deployments := resources.NewDeploymentsClient(subscriptionID)
		deployments.Authorizer = auth

		status.Print("Done.")
		exitStatus = 0
	},
	Args: func(cmd *cobra.Command, args []string) error {
		if provisionConfig.GetString(SubscriptionName) == "" {
			return fmt.Errorf("no value found for %q", SubscriptionName)
		}

		hasClientID := provisionConfig.GetString(ClientIDName) != ""
		hasClientSecret := provisionConfig.GetString(ClientSecretName) != ""

		if (hasClientID || hasClientSecret) && !(hasClientID && hasClientSecret) {
			return errors.New("--client-id and --client-secret must be speficied together or not at all")
		}

		var err error
		environment, err = azure.EnvironmentFromName(provisionConfig.GetString(EnvironmentName))
		if err != nil {
			return err
		}

		statusWriter := ioutil.Discard
		if provisionConfig.GetBool(VerboseName) {
			statusWriter = os.Stdout
		}
		status = log.New(statusWriter, "", 0)

		return nil
	},
}

// insertResourceGroup checks for a Resource Groups's existence, if it is not found it creates that resource group. If
// that resource group exists, it leaves it alone.
func insertResourceGroup(ctx context.Context, groups resources.GroupsClient, name string, location string) (bool, error) {
	existenceResp, err := groups.CheckExistence(ctx, name)
	if err != nil {
		return false, err
	}

	switch existenceResp.StatusCode {
	case http.StatusNoContent:
		return false, nil
	case http.StatusNotFound:
		createResp, err := groups.CreateOrUpdate(ctx, name, resources.Group{
			Location: &location,
		})
		if err != nil {
			return false, err
		}

		if createResp.StatusCode == http.StatusCreated {
			return true, nil
		} else if createResp.StatusCode == http.StatusOK {
			return false, nil
		} else {
			return false, fmt.Errorf("unexpected status code %d during resource group creation", createResp.StatusCode)
		}
	default:
		return false, fmt.Errorf("unexpected status code %d during resource group existence check", existenceResp.StatusCode)
	}
}

func getAuthorizer(ctx context.Context, clientID, clientSecret, tenantID string) (autorest.Authorizer, error) {

	config, err := adal.NewOAuthConfig(environment.ActiveDirectoryEndpoint, tenantID)
	if err != nil {
		return nil, err
	}

	var intermediate *adal.Token

	if provisionConfig.GetBool(DeviceAuthName) {
		client := &http.Client{}
		code, err := adal.InitiateDeviceAuth(
			client,
			*config,
			deviceClientID,
			environment.ResourceManagerEndpoint)
		if err != nil {
			return nil, err
		}
		fmt.Println(*code.Message)
		token, err := adal.WaitForUserCompletion(client, code)
		if err != nil {
			return nil, err
		}
		intermediate = token
	} else {
		auth, err := adal.NewServicePrincipalToken(
			*config,
			clientID,
			clientSecret,
			environment.ResourceManagerEndpoint)
		if err != nil {
			return nil, err
		}
		status.Println("service principal token created for client: ", clientID)
		t := auth.Token()
		intermediate = &t
	}

	// TODO: If tenant ID wasn't provided, use common tenant above, then iterate over all available tenants then subscriptions to automatically decide the correct one.

	return autorest.NewBearerAuthorizer(intermediate), nil
}

func getDatabaseFlavor(buffaloRoot string) string {
	return "postgres" // TODO: parse buffalo app for the database they're using.
}

func getTenant(subscription guid.GUID) (string, error) {
	return "dynamically discovered tenant", nil // TODO: traverse all tenants this identity has access to, looking for the subscription id.
}

var normalizeScheme = strings.ToLower
var supportedLinkSchemes = map[string]struct{}{
	normalizeScheme("http"):  {},
	normalizeScheme("https"): {},
}

// isSupportedLink interrogates a string to decide if it is a RequestURI that is supported by the Azure template engine
// as defined here:
// https://docs.microsoft.com/en-us/azure/azure-resource-manager/resource-group-linked-templates#external-template-and-external-parameters
func isSupportedLink(subject string) bool {
	parsed, err := url.ParseRequestURI(subject)
	if err != nil {
		return false
	}

	_, ok := supportedLinkSchemes[normalizeScheme(parsed.Scheme)]

	return ok
}

func init() {
	azureCmd.AddCommand(provisionCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// provisionCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// provisionCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	provisionConfig.BindEnv(SubscriptionName, "AZURE_SUBSCRIPTION_ID", "AZ_SUBSCRIPTION_ID")
	provisionConfig.BindEnv(ClientIDName, "AZURE_CLIENT_ID", "AZ_CLIENT_ID")
	provisionConfig.BindEnv(ClientSecretName, "AZURE_CLIENT_SECRET", "AZ_CLIENT_SECRET")
	provisionConfig.BindEnv(TenantIDName, "AZURE_TENANT_ID", "AZ_TENANT_ID")
	provisionConfig.BindEnv(EnvironmentName, "AZURE_ENVIRONMENT", "AZ_ENVIRONMENT")

	var sanitizedClientSecret string
	if rawSecret := provisionConfig.GetString(ClientSecretName); rawSecret != "" {
		const safeCharCount = 10
		if len(rawSecret) > safeCharCount {
			sanitizedClientSecret = fmt.Sprintf("...%s", rawSecret[len(rawSecret)-safeCharCount:])
		} else {
			sanitizedClientSecret = "[key hidden]"
		}
	}

	provisionConfig.SetDefault(EnvironmentName, EnvironmentDefault)
	provisionConfig.SetDefault(DatabaseName, getDatabaseFlavor("."))
	provisionConfig.SetDefault(ResoureGroupName, "buffalo-app") // TODO: generate a random suffix
	provisionConfig.SetDefault(LocationName, LocationDefault)

	provisionCmd.Flags().StringP(ImageName, ImageShorthand, ImageDefault, imageUsage)
	provisionCmd.Flags().StringP(TemplateName, TemplateShorthand, TemplateDefault, templateUsage)
	provisionCmd.Flags().StringP(SubscriptionName, SubscriptionShorthand, provisionConfig.GetString(SubscriptionName), subscriptionUsage)
	provisionCmd.Flags().String(ClientIDName, provisionConfig.GetString(ClientIDName), clientIDUsage)
	provisionCmd.Flags().String(ClientSecretName, sanitizedClientSecret, clientSecretUsage)
	provisionCmd.Flags().Bool(DeviceAuthName, false, deviceAuthUsage)
	provisionCmd.Flags().BoolP(VerboseName, VerboseShortname, false, verboseUsage)
	provisionCmd.Flags().String(TenantIDName, provisionConfig.GetString(TenantIDName), tenantUsage)
	provisionCmd.Flags().StringP(EnvironmentName, EnvironmentShorthand, provisionConfig.GetString(EnvironmentName), environmentUsage)
	provisionCmd.Flags().StringP(DatabaseName, DatabaseShorthand, provisionConfig.GetString(DatabaseName), databaseUsage)
	provisionCmd.Flags().StringP(ResoureGroupName, ResourceGroupShorthand, provisionConfig.GetString(ResoureGroupName), resourceGroupUsage)
	provisionCmd.Flags().StringP(LocationName, LocationShorthand, provisionConfig.GetString(LocationName), locationUsage)

	provisionConfig.BindPFlags(provisionCmd.Flags())
}

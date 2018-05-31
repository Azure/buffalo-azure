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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/resources/mgmt/subscriptions"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2017-05-10/resources"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
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

	// TemplateDefault is the name of the Template to use if no value was provided.
	TemplateDefault = "./azuredeploy.json"

	// TemplateDefaultLink defines the link that will be used if no local rm-template is found, and a link wasn't
	// provided.
	TemplateDefaultLink = "https://aka.ms/buffalo-template"
	templateUsage       = "The Azure Resource Management template which specifies the resources to provision."
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

// These constants define a parameter which forces this program to ignore any ambient Azure settings available as
// environment variables, and instead forces us to use Device Auth instead.
const (
	DeviceAuthName  = "use-device-auth"
	deviceAuthUsage = "Ignore --client-id and --client-secret, interactively authenticate instead."
)

// These constants define a parameter which toggles whether or not status information will be printed as this program
// executes.
const (
	VerboseName      = "verbose"
	VerboseShortname = "v"
	verboseUsage     = "Print out status information as this program executes."
)

var status *log.Logger
var errLog = newFormattedLog(os.Stderr, "error")
var debugLog *log.Logger

var debug string

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

		debugLog.Print("debugging enabled")

		// Authenticate and setup clients
		subscriptionID := provisionConfig.GetString(SubscriptionName)
		clientID := provisionConfig.GetString(ClientIDName)
		clientSecret := provisionConfig.GetString(ClientSecretName)
		templateLocation := provisionConfig.GetString(TemplateName)
		image := provisionConfig.GetString(ImageName)
		databaseType := provisionConfig.GetString(DatabaseName)

		ctx, cancel := context.WithTimeout(context.Background(), 45*time.Minute)
		defer cancel()

		auth, err := getAuthorizer(ctx, subscriptionID, clientID, clientSecret, provisionConfig.GetString(TenantIDName))
		if err != nil {
			errLog.Print("unable to authenticate: ", err)
			return
		}
		status.Print("tenant selected: ", provisionConfig.GetString(TenantIDName))
		status.Print("subscription selected: ", subscriptionID)
		status.Println("template selected: ", templateLocation)
		status.Println("database selected: ", databaseType)
		status.Println("image selected: ", image)

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
			errLog.Printf("unable to fetch or create resource group %s: %v\n", rgName, err)
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

		params := NewDeploymentParameters()
		params.Parameters["database"] = DeploymentParameter{databaseType}
		params.Parameters["imageName"] = DeploymentParameter{image}
		params.Parameters["databaseAdministratorLogin"] = DeploymentParameter{"buffaloAdmin"}
		params.Parameters["databaseAdministratorLoginPassword"] = DeploymentParameter{"M$FT<3sBuffalo"}

		template, err := getDeploymentTemplate(ctx, templateLocation)
		if err != nil {
			errLog.Print("unable to fetch template: ", err)
			return
		}

		template.Parameters = params.Parameters
		template.Mode = resources.Incremental

		fut, err := deployments.CreateOrUpdate(ctx, rgName, "buffalo-app", resources.Deployment{
			Properties: template,
		})

		if err != nil {
			errLog.Print("unable to start deployment: ", err)
			return
		}

		err = fut.WaitForCompletion(ctx, deployments.Client)
		if err != nil {
			errLog.Print("unable to poll for completion progress, your assets may or may not have finished provisioning")
			return
		}

		status.Print("finished deployment")
		exitStatus = 0
	},
	Args: func(cmd *cobra.Command, args []string) error {
		if provisionConfig.GetString(SubscriptionName) == "" {
			return fmt.Errorf("no value found for %q", SubscriptionName)
		}

		hasClientID := provisionConfig.GetString(ClientIDName) != ""
		hasClientSecret := provisionConfig.GetString(ClientSecretName) != ""

		if (hasClientID || hasClientSecret) && !(hasClientID && hasClientSecret) {
			return errors.New("--client-id and --client-secret must be specified together or not at all")
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
		status = newFormattedLog(statusWriter, "information")

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

func getAuthorizer(ctx context.Context, subscriptionID, clientID, clientSecret, tenantID string) (autorest.Authorizer, error) {
	const commonTenant = "common"

	if tenantID == "" {
		debugLog.Print("tenant unset, using common tenant")
		tenantID = commonTenant
	}

	config, err := adal.NewOAuthConfig(environment.ActiveDirectoryEndpoint, tenantID)
	if err != nil {
		return nil, err
	}

	if provisionConfig.GetBool(DeviceAuthName) {
		var intermediate *adal.Token

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

		if tenantID == commonTenant {
			var final autorest.Authorizer
			tenantID, final, err = getTenant(ctx, intermediate, subscriptionID)
			if err != nil {
				return nil, err
			}

			return final, nil
		}
		return autorest.NewBearerAuthorizer(intermediate), nil
	}

	if tenantID == commonTenant {
		return nil, errors.New("tenant inference unsupported with Service Principal authentication")
	}

	auth, err := adal.NewServicePrincipalToken(
		*config,
		clientID,
		clientSecret,
		environment.ResourceManagerEndpoint)
	if err != nil {
		return nil, err
	}
	status.Println("service principal token created for client: ", clientID)
	return autorest.NewBearerAuthorizer(auth), nil
}

func getDatabaseFlavor(buffaloRoot string) string {
	return "postgres" // TODO (#29): parse buffalo app for the database they're using.
}

func getDeploymentTemplate(ctx context.Context, raw string) (*resources.DeploymentProperties, error) {
	if isSupportedLink(raw) {
		debugLog.Print("identified external link")
		buf := bytes.NewBuffer([]byte{})

		err := downloadTemplate(ctx, buf, raw)
		if err != nil {
			return nil, err
		}

		debugLog.Printf("template %d bytes long", buf.Len())

		return &resources.DeploymentProperties{
			Template: json.RawMessage(buf.Bytes()),
		}, nil
	}

	debugLog.Print("identified local file")
	handle, err := os.Open(raw)
	if err != nil {
		return nil, err
	}

	contents, err := ioutil.ReadAll(handle)
	if err != nil {
		return nil, err
	}

	return &resources.DeploymentProperties{
		Template: json.RawMessage(contents),
	}, nil
}

var redirectCodes = map[int]struct{}{
	http.StatusMovedPermanently:  {},
	http.StatusPermanentRedirect: {},
	http.StatusTemporaryRedirect: {},
	http.StatusSeeOther:          {},
	http.StatusFound:             {},
}

var temporaryFailureCodes = map[int]struct{}{
	http.StatusTooManyRequests: {},
	http.StatusGatewayTimeout:  {},
	http.StatusRequestTimeout:  {},
}

var acceptedCodes = map[int]struct{}{
	http.StatusOK: {},
}

func downloadTemplate(ctx context.Context, dest io.Writer, src string) error {
	const maxRedirects = 5
	const maxRetries = 3
	var download func(context.Context, io.Writer, string, uint) error

	status.Print("downloading template from: ", src)

	download = func(ctx context.Context, dest io.Writer, src string, depth uint) (err error) {
		if depth > maxRedirects {
			return errors.New("too many redirects")
		}

		for attempt := 0; attempt < maxRetries; attempt++ {
			var req *http.Request
			var resp *http.Response

			req, err = http.NewRequest(http.MethodGet, src, nil)
			if err != nil {
				return
			}
			req = req.WithContext(ctx)

			resp, err = http.DefaultClient.Do(req)
			if err != nil {
				return
			}

			if _, ok := acceptedCodes[resp.StatusCode]; ok {
				_, err = io.Copy(dest, resp.Body)
				return
			}

			if _, ok := redirectCodes[resp.StatusCode]; ok {
				loc := resp.Header.Get("Location")
				debugLog.Printf("HTTP Status Code %d encountered, following redirection to %s", resp.StatusCode, loc)
				return download(ctx, dest, loc, depth+1)
			}

			if _, ok := temporaryFailureCodes[resp.StatusCode]; ok {
				debugLog.Printf("HTTP Status Code %d encountered, retrying", resp.StatusCode)
				continue
			}

			err = fmt.Errorf("unexpected status code: %d", resp.StatusCode)
			return
		}

		err = errors.New("too many attempts")
		return
	}

	return download(ctx, dest, src, 1)
}

func getTenant(ctx context.Context, common *adal.Token, subscription string) (string, autorest.Authorizer, error) {
	tenants := subscriptions.NewTenantsClient()
	tenants.Authorizer = autorest.NewBearerAuthorizer(common)

	var err error
	var tenantList subscriptions.TenantListResultIterator

	subscriptionClient := subscriptions.NewClient()

	status.Println("using authorization to infer tenant")

	for tenantList, err = tenants.ListComplete(ctx); err == nil && tenantList.NotDone(); err = tenantList.Next() {
		var subscriptionList subscriptions.ListResultIterator
		currentTenant := *tenantList.Value().TenantID
		currentConfig, err := adal.NewOAuthConfig(environment.ActiveDirectoryEndpoint, currentTenant)
		if err != nil {
			return "", nil, err
		}
		currentAuth, err := adal.NewServicePrincipalTokenFromManualToken(*currentConfig, deviceClientID, environment.ResourceManagerEndpoint, adal.Token{
			RefreshToken: common.RefreshToken,
		})
		if err != nil {
			return "", nil, err
		}
		subscriptionClient.Authorizer = autorest.NewBearerAuthorizer(currentAuth)

		for subscriptionList, err = subscriptionClient.ListComplete(ctx); err == nil && subscriptionList.NotDone(); err = subscriptionList.Next() {
			if currentSub := subscriptionList.Value(); currentSub.SubscriptionID != nil && strings.EqualFold(*currentSub.SubscriptionID, subscription) {
				provisionConfig.Set(TenantIDName, *tenantList.Value().TenantID)
				return *tenantList.Value().TenantID, subscriptionClient.Authorizer, nil
			}
		}
	}
	if err != nil {
		return "", nil, err
	}

	return "", nil, fmt.Errorf("unable to find subscription: %s", subscription)
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

func newFormattedLog(output io.Writer, identifier string) *log.Logger {
	const identLen = 4
	for len(identifier) < identLen {
		identifier = identifier + " "
	}
	return log.New(output, fmt.Sprintf("[%s] ", strings.ToUpper(identifier)[:identLen]), log.Ldate|log.Ltime)
}

func init() {
	var debugWriter io.Writer
	if debug == "" {
		debugWriter = ioutil.Discard
	} else {
		debugWriter = os.Stderr
	}
	debugLog = newFormattedLog(debugWriter, "debug")

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

	if _, err := os.Stat(TemplateDefault); err == nil {
		provisionConfig.SetDefault(TemplateName, TemplateDefault)
	} else {
		provisionConfig.SetDefault(TemplateName, TemplateDefaultLink)
	}

	provisionConfig.SetDefault(EnvironmentName, EnvironmentDefault)
	provisionConfig.SetDefault(DatabaseName, getDatabaseFlavor("."))
	provisionConfig.SetDefault(ResoureGroupName, "buffalo-app") // TODO (#30): generate a random suffix
	provisionConfig.SetDefault(LocationName, LocationDefault)

	provisionCmd.Flags().StringP(ImageName, ImageShorthand, ImageDefault, imageUsage)
	provisionCmd.Flags().StringP(TemplateName, TemplateShorthand, provisionConfig.GetString(TemplateName), templateUsage)
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

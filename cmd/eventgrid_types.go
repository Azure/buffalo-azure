package cmd

var wellKnownEvents = map[string]string{
	"Microsoft.ContainerRegistry.ImagePushed":                              "github.com/Azure/azure-sdk-for-go/services/eventgrid/2018-01-01/eventgrid.ContainerRegistryImagePushedEventData",
	"Microsoft.ContainerRegistry.ImageDeleted":                             "github.com/Azure/azure-sdk-for-go/services/eventgrid/2018-01-01/eventgrid.ContainerRegistryImageDeletedEventData",
	"Microsoft.EventGrid.SubscriptionValidationEvent":                      "github.com/Azure/azure-sdk-for-go/services/eventgrid/2018-01-01/eventgrid.SubscriptionValidationEventData",
	"Microsoft.EventGrid.SubscriptionDeletedEvent":                         "github.com/Azure/azure-sdk-for-go/services/eventgrid/2018-01-01/eventgrid.SubscriptionDeletedEventData",
	"Microsoft.EventHub.CaptureFileCreated":                                "github.com/Azure/azure-sdk-for-go/services/eventgrid/2018-01-01/eventgrid.EventHubCaptureFileCreatedEventData",
	"Microsoft.Devices.DeviceCreated":                                      "github.com/Azure/azure-sdk-for-go/services/eventgrid/2018-01-01/eventgrid.IotHubDeviceCreatedEventData",
	"Microsoft.Devices.DeviceDeleted":                                      "github.com/Azure/azure-sdk-for-go/services/eventgrid/2018-01-01/eventgrid.IotHubDeviceDeletedEventData",
	"Microsoft.Media.JobStateChanged":                                      "github.com/Azure/azure-sdk-for-go/services/eventgrid/2018-01-01/eventgrid.MediaJobStateChangeEventData",
	"Microsoft.Resources.ResourceDeleteCancel":                             "github.com/Azure/azure-sdk-for-go/services/eventgrid/2018-01-01/eventgrid.ResourceDeleteCancelData",
	"Microsoft.Resources.ResourceDeleteFailure":                            "github.com/Azure/azure-sdk-for-go/services/eventgrid/2018-01-01/eventgrid.ResourceDeleteFailureData",
	"Microsoft.Resources.ResourceDeleteSuccess":                            "github.com/Azure/azure-sdk-for-go/services/eventgrid/2018-01-01/eventgrid.ResourceDeleteSuccessData",
	"Microsoft.Resources.ResourceWriteCancel":                              "github.com/Azure/azure-sdk-for-go/services/eventgrid/2018-01-01/eventgrid.ResourceWriteCancelData",
	"Microsoft.Resources.ResourceWriteFailure":                             "github.com/Azure/azure-sdk-for-go/services/eventgrid/2018-01-01/eventgrid.ResourceWriteFailureData",
	"Microsoft.Resources.ResourceWriteSuccess":                             "github.com/Azure/azure-sdk-for-go/services/eventgrid/2018-01-01/eventgrid.ResourceWriteSuccessData",
	"Microsoft.ServiceBus.ActiveMessagesAvailableWithNoListeners":          "github.com/Azure/azure-sdk-for-go/services/eventgrid/2018-01-01/eventgrid.ServiceBusActiveMessagesAvailableWithNoListenersEventData",
	"Microsoft.ServiceBus.DeadletterMessagesAvailableWithNoListenersEvent": "github.com/Azure/azure-sdk-for-go/services/eventgrid/2018-01-01/eventgrid.ServiceBusDeadletterMessagesAvailableWithNoListenersEventData",
	"Microsoft.Storage.BlobCreated":                                        "github.com/Azure/azure-sdk-for-go/services/eventgrid/2018-01-01/eventgrid.StorageBlobCreatedEventData",
	"Microsoft.Storage.BlobDeleted":                                        "github.com/Azure/azure-sdk-for-go/services/eventgrid/2018-01-01/eventgrid.StorageBlobDeletedEventData",
}

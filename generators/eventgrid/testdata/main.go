package main

import (
	"fmt"

	egdp "github.com/Azure/azure-sdk-for-go/services/eventgrid/2018-01-01/eventgrid"
	"github.com/Azure/buffalo-azure/sdk/eventgrid"
)

func main() {
	fmt.Printf("%#v\n", eventgrid.TypeDispatchSubscriber{})
	fmt.Printf("%#v\n", egdp.StorageBlobCreatedEventData{})
}

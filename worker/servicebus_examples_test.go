package worker_test

import (
	"context"
	"fmt"
	"os"
	"time"

	servicebus "github.com/Azure/azure-service-bus-go"
	azwork "github.com/Azure/buffalo-azure/worker"
	"github.com/gobuffalo/buffalo/worker"
)

func ExampleServiceBusPublisher_initializeAndSend() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create a client that can communicate with a particular Service Bus Namespace.
	sbConnection := "<your service bus connection string>"
	nsOpts := []servicebus.NamespaceOption{
		servicebus.NamespaceWithConnectionString(sbConnection),
	}

	ns, err := servicebus.NewNamespace(nsOpts...)
	if err != nil {
		fmt.Fprintln(os.Stderr, "unable to create namespace client: ", err)
		return
	}

	// Create a client that can communicate with a particular Service Bus Queue.
	queueName := "<your queue name>"
	queue, err := ns.NewQueue(ctx, queueName)
	if err != nil {
		fmt.Fprintln(os.Stderr, "unable to create queue client", err)
		return
	}

	// Create the client which knows how to publish jobs that will be carried out
	// by a ServiceBusReceiver which knows how to actually do the job.
	myPublisher := azwork.NewServiceBusPublisher(queue)

	// Start sending jobs to be processed somewhere else!
	myPublisher.Perform(worker.Job{
		Queue: queueName,
		Args: worker.Args{
			"source": "https://notarealdomain.com/gomodules/buffalo-azure/worker.zip",
		},
		Handler: "downloader",
	})
}

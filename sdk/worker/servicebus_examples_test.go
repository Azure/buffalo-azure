package worker_test

import (
	"fmt"
	"os"

	azwork "github.com/Azure/buffalo-azure/worker"
	"github.com/gobuffalo/buffalo/worker"
)

func ExampleServiceBusPublisher_initializeAndSend() {
	sbConnection := "<your service bus connection string>"
	queueName := "<your queue name>"

	// Create a client that can communicate with a particular Service Bus Queue.

	// Create the client which knows how to publish jobs that will be carried out
	// by a ServiceBusReceiver which knows how to actually do the job.
	myPublisher, err := azwork.NewServiceBusPublisher(sbConnection, 1)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	// Start sending jobs to be processed somewhere else!
	myPublisher.Perform(worker.Job{
		Queue: queueName,
		Args: worker.Args{
			"source": "https://notarealdomain.com/gomodules/buffalo-azure/worker.zip",
		},
		Handler: "downloader",
	})
}

// +build !offline

package worker

import (
	"context"
	"fmt"
	"testing"
	"time"

	servicebus "github.com/Azure/azure-service-bus-go"
	"github.com/gobuffalo/buffalo/worker"
	"github.com/joho/godotenv"
	"github.com/satori/uuid"
	"github.com/spf13/viper"
)

var config = viper.New()

// All viper names of inputs into this test harness to light-up different tests.
const (
	ServiceBusConnection                    = "sb-conn"
	ServiceBusConnectionEnvironmentVariable = "BUFFALO_AZURE_TEST_SERVICE_BUS_CONNECTION"
)

func init() {
	godotenv.Overload("./.env", "../.env")

	config.BindEnv(ServiceBusConnection, ServiceBusConnectionEnvironmentVariable)
}

func TestServiceBus_StartStop_noQueues(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	subject, err := NewServiceBus()
	if err != nil {
		t.Error(err)
		return
	}

	err = subject.Start(ctx)
	if err != nil {
		t.Error(err)
		return
	}

	err = subject.Stop()
	if err != nil {
		t.Error(err)
		return
	}
}

func TestServiceBus_SendReceive(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	const handlerIdent = "test-receiver"

	if skipIfMissingConfig(t, ServiceBusConnection) {
		return
	}

	ns, err := servicebus.NewNamespace(servicebus.NamespaceWithConnectionString(config.GetString(ServiceBusConnection)))
	if err != nil {
		t.Error(err)
		return
	}

	queueName := uuid.NewV1().String()
	mgmtClient := ns.NewQueueManager()

	if _, err := mgmtClient.Put(ctx, queueName); err != nil {
		t.Error(err)
		return
	}
	defer mgmtClient.Delete(context.Background(), queueName)

	queue, err := ns.NewQueue(ctx, queueName)
	if err != nil {
		t.Error(err)
		return
	}

	subject, err := NewServiceBus(queue)
	if err != nil {
		t.Error(err)
		return
	}

	conduit := make(chan string)

	var handler worker.Handler = func(args worker.Args) error {
		left := args["left"].(string)
		right := args["right"].(string)

		conduit <- fmt.Sprintf("%s %s", left, right)
		return nil
	}

	subject.Register(handlerIdent, handler)

	err = subject.Start(ctx)
	if err != nil {
		t.Error(err)
		return
	}
	defer subject.Stop()

	// Using prime numbers so there is no other combination that could work.
	const left, right = "Hello", "World!"
	const want = left + " " + right

	expected := worker.Job{
		Queue:   queueName,
		Handler: handlerIdent,
		Args: worker.Args{
			"left":  left,
			"right": right,
		},
	}

	err = subject.Perform(expected)
	if err != nil {
		t.Error(err)
		return
	}

	select {
	case <-ctx.Done():
		t.Log("timed out")
		t.Fail()
	case got := <-conduit:
		if got != want {
			t.Logf("got: %s want: %s", got, want)
			t.Fail()
		}
	}
}

func skipIfMissingConfig(t *testing.T, params ...string) bool {
	missingRequired := false
	for i := range params {
		if !config.IsSet(params[i]) {
			t.Skipf("Unable to find required parameter %q", params[i])
			missingRequired = true
		}
	}
	return missingRequired
}

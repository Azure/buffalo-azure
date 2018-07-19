package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	servicebus "github.com/Azure/azure-service-bus-go"
	bufwork "github.com/gobuffalo/buffalo/worker"
)

type queuePool map[string]*servicebus.Queue

type (
	// ServiceBusReceiver is able to listen to a ServiceBus Queue and do Buffalo Jobs.
	ServiceBusReceiver struct {
		pool     queuePool
		handlers map[string]bufwork.Handler
		handles  map[string]*servicebus.ListenerHandle
		ctx      context.Context
		mut      sync.RWMutex
	}

	// ServiceBusPublisher is able to write Buffalo Jobs to ServiceBus Queue.
	ServiceBusPublisher struct {
		pool queuePool
	}

	// ServiceBus is a compound type, which fulfills the entire interface defined by the type
	// `github.com/gobuffalo/buffalo/worker.Worker`.
	//
	// The same connection pool is used for both the ServiceBusPublisher and ServiceBusRececiver.
	ServiceBus struct {
		ServiceBusPublisher
		ServiceBusReceiver
	}
)

// NewServiceBusReceiver is a constructor for the type `ServiceBusReceiver`.
func NewServiceBusReceiver(queues ...*servicebus.Queue) *ServiceBusReceiver {
	return &ServiceBusReceiver{
		pool:     newQueuePool(queues),
		handlers: make(map[string]bufwork.Handler),
		handles:  make(map[string]*servicebus.ListenerHandle, len(queues)),
	}
}

// Register binds a function that should be called when a Job calling for a particular value of Handler is
// received.
func (sbr *ServiceBusReceiver) Register(name string, processor bufwork.Handler) error {
	sbr.mut.Lock()
	defer sbr.mut.Unlock()

	sbr.handlers[name] = processor
	return nil
}

// Start begins listening to a ServiceBus Queue in order to handle github.com/gobuffalo/buffalo/worker.Jobs
// that come accross the wire as Service Bus Messages.
//
// Start returns immediately without blocking. To have this worker Stop, see the Stop() message.
func (sbr *ServiceBusReceiver) Start(ctx context.Context) (err error) {
	sbr.mut.RLock()
	defer sbr.mut.RUnlock()

	sbr.ctx = ctx
	for name, client := range sbr.pool {
		var handle *servicebus.ListenerHandle
		handle, err = client.Receive(ctx, sbr.dispatch)
		if err != nil {
			break
		}
		sbr.handles[name] = handle
	}

	// Having `Start` partially complete could cause connection leaks, or have it be hard to reason about
	// the existing state. For that reason, if even one Queue listener fails to start listening, the clause
	// below will stop any of the queues that had successfully started listening.
	if err != nil {
		sbr.stop()
	}

	return
}

// Stop makes this instance of a ServiceBus listener cease listening for Jobs.
func (sbr *ServiceBusReceiver) Stop() error {
	sbr.mut.Lock()
	defer sbr.mut.Unlock()

	return sbr.stop()
}

// stop closes all open queue Receive operations, and resets the context from when it was started. It
// takes no lock, and therefor is only suitable for internal consumption.
func (sbr *ServiceBusReceiver) stop() error {
	for _, handle := range sbr.handles {
		handle.Close(sbr.ctx)
	}
	sbr.ctx = nil
	return nil
}

func (sbr *ServiceBusReceiver) dispatch(ctx context.Context, message *servicebus.Message) servicebus.DispositionAction {
	var j bufwork.Job
	err := json.Unmarshal(message.Data, &j)
	if err != nil {
		// Poorly formatted message, throw it away.
		return message.DeadLetter(err)
	}

	// Once we've started doing the job, we should complete it and communicate with ServiceBus. For that reason, we can
	// take out a very focused lock, instead of blocking changes to the ServiceBusReceiver.
	sbr.mut.RLock()
	handler, ok := sbr.handlers[j.Handler]
	sbr.mut.RUnlock()

	if ok {
		handler(j.Args)
	} else {
		// A message that can't be handled by this ServiceBusReceiver, try it again somewhere else.
		return message.Abandon()
	}

	return message.Complete()
}

// NewServiceBusPublisher is a constructor for the type `ServiceBusPublisher`.
func NewServiceBusPublisher(queues ...*servicebus.Queue) *ServiceBusPublisher {
	return &ServiceBusPublisher{
		pool: newQueuePool(queues),
	}
}

// Perform schedules a `Job` to be run as soon as possible.
func (sbp *ServiceBusPublisher) Perform(job bufwork.Job) (err error) {
	return sbp.publish(job, nil)
}

// PerformAt schedules a `Job` to be performed at a particular time.
func (sbp *ServiceBusPublisher) PerformAt(job bufwork.Job, at time.Time) error {
	return sbp.publish(job, &servicebus.SystemProperties{
		ScheduledEnqueueTime: &at,
	})
}

// PerformIn schedules a `Job` to be performed after waiting for a set amount of time.
func (sbp *ServiceBusPublisher) PerformIn(job bufwork.Job, in time.Duration) error {
	return sbp.PerformAt(job, time.Now().Add(in))
}

func (sbp *ServiceBusPublisher) publish(job bufwork.Job, messageProperties *servicebus.SystemProperties) error {
	client, ok := sbp.pool[job.Queue]
	if !ok {
		return fmt.Errorf("unknown queue %q", job.Queue)
	}

	marshaled, err := json.Marshal(job)
	if err != nil {
		return err
	}

	message := servicebus.NewMessage(marshaled)
	message.SystemProperties = messageProperties

	return client.Send(context.Background(), message)
}

// NewServiceBus is a constructor for the compound type ServiceBus. It instantiates a ServiceBusPublisher and ServiceBusReceiver
// that share connection resources.
func NewServiceBus(queues ...*servicebus.Queue) (retval *ServiceBus, _ error) {
	pool := newQueuePool(queues)

	retval = new(ServiceBus)

	retval.ServiceBusPublisher = ServiceBusPublisher{
		pool: pool,
	}

	retval.ServiceBusReceiver = ServiceBusReceiver{
		pool:     pool,
		handlers: make(map[string]bufwork.Handler),
		handles:  make(map[string]*servicebus.ListenerHandle, len(queues)),
	}
	return
}

func newQueuePool(queues []*servicebus.Queue) (retval queuePool) {
	retval = make(queuePool, len(queues))

	for i := range queues {
		retval[queues[i].Name] = queues[i]
	}
	return
}

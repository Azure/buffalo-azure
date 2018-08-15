package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/Azure/azure-amqp-common-go/uuid"
	servicebus "github.com/Azure/azure-service-bus-go"
	bufwork "github.com/gobuffalo/buffalo/worker"
)

type queuePool struct {
	sync.RWMutex
	ns      *servicebus.Namespace
	clients map[string]*servicebus.Queue
	handles map[string]*servicebus.ListenerHandle
	cancel  context.CancelFunc
	ctx     context.Context
	session uuid.UUID
}

// ServiceBusQueueAutoDeleteOnIdleTime is the amount of time that a Service Bus Queue
// will stick around when there are no longer any messages in it.
const ServiceBusQueueAutoDeleteOnIdleTime = 20 * time.Minute

// DefaultQueue is the name of the queue that will be published to/listened from when no Queue is provided to the
// ServiceBus and ServiceBusReceiver constructors.
const DefaultQueue = "buffalo-worker"

type (
	// ServiceBusReceiver is able to listen to a ServiceBus Queue and do Buffalo Jobs.
	ServiceBusReceiver struct {
		pool     *queuePool
		handlers map[string]bufwork.Handler
		mut      sync.RWMutex
	}

	// ServiceBusPublisher is able to write Buffalo Jobs to ServiceBus Queue.
	ServiceBusPublisher struct {
		pool           *queuePool
		PublishTimeout time.Duration
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
func NewServiceBusReceiver(connstr string, queues ...string) (*ServiceBusReceiver, error) {
	ns, err := servicebus.NewNamespace(servicebus.NamespaceWithConnectionString(connstr))
	if err != nil {
		return nil, err
	}

	pool := newQueuePool(ns, queues...)

	retval := new(ServiceBusReceiver)
	initializeServiceBusReceiver(retval, ns, pool)
	return retval, nil
}

func initializeServiceBusReceiver(subject *ServiceBusReceiver, ns *servicebus.Namespace, pool *queuePool) {
	subject.pool = pool
	subject.handlers = make(map[string]bufwork.Handler)
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

	return sbr.pool.Receive(ctx, sbr.dispatch)
}

// Stop makes this instance of a ServiceBus listener cease listening for Jobs.
func (sbr *ServiceBusReceiver) Stop() error {
	sbr.mut.Lock()
	defer sbr.mut.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	return sbr.pool.stop(ctx)
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
		if err := handler(j.Args); err == nil {
			return message.Complete()
		}
	}
	return message.Abandon()
}

// NewServiceBusPublisher is a constructor for the type `ServiceBusPublisher`.
func NewServiceBusPublisher(connStr string, queues ...string) (*ServiceBusPublisher, error) {
	ns, err := servicebus.NewNamespace(servicebus.NamespaceWithConnectionString(connStr))
	if err != nil {
		return nil, err
	}

	retval := new(ServiceBusPublisher)
	initializeServiceBusPublisher(retval, ns, newQueuePool(ns, queues...))
	return retval, nil
}

func initializeServiceBusPublisher(subject *ServiceBusPublisher, ns *servicebus.Namespace, pool *queuePool) {
	subject.pool = pool
	subject.PublishTimeout = 5 * time.Minute
}

// Perform schedules a `Job` to be run as soon as possible.
func (sbp *ServiceBusPublisher) Perform(job bufwork.Job) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), sbp.PublishTimeout)
	defer cancel()

	return sbp.publish(ctx, job, nil)
}

// PerformAt schedules a `Job` to be performed at a particular time.
func (sbp *ServiceBusPublisher) PerformAt(job bufwork.Job, at time.Time) error {
	ctx, cancel := context.WithTimeout(context.Background(), sbp.PublishTimeout)
	defer cancel()

	return sbp.publish(ctx, job, &servicebus.SystemProperties{
		ScheduledEnqueueTime: &at,
	})
}

// PerformIn schedules a `Job` to be performed after waiting for a set amount of time.
func (sbp *ServiceBusPublisher) PerformIn(job bufwork.Job, in time.Duration) error {
	return sbp.PerformAt(job, time.Now().Add(in))
}

func (sbp *ServiceBusPublisher) publish(ctx context.Context, job bufwork.Job, messageProperties *servicebus.SystemProperties) error {
	sbp.pool.Lock()
	client, ok := sbp.pool.clients[job.Queue]
	if !ok {
		var err error
		client, err = assertQueue(ctx, sbp.pool.ns, job.Queue)
		if err != nil {
			sbp.pool.Unlock()
			return err
		}
		sbp.pool.clients[job.Queue] = client

	}
	sbp.pool.Unlock()

	sbp.pool.RLock()
	defer sbp.pool.RUnlock()
	marshaled, err := json.Marshal(job)
	if err != nil {
		return err
	}

	message := servicebus.NewMessage(marshaled)
	message.SystemProperties = messageProperties

	return client.Send(ctx, message)
}

// NewServiceBus is a constructor for the compound type ServiceBus. It instantiates a ServiceBusPublisher and ServiceBusReceiver
// that share connection resources.
func NewServiceBus(connStr string, queues ...string) (retval *ServiceBus, _ error) {
	ns, err := servicebus.NewNamespace(servicebus.NamespaceWithConnectionString(connStr))
	if err != nil {
		return nil, err
	}

	pool := newQueuePool(ns, queues...)

	retval = new(ServiceBus)

	initializeServiceBusPublisher(&retval.ServiceBusPublisher, ns, pool)
	initializeServiceBusReceiver(&retval.ServiceBusReceiver, ns, pool)
	return
}

func newQueuePool(ns *servicebus.Namespace, queues ...string) (retval *queuePool) {
	retval = new(queuePool)
	if len(queues) == 0 {
		queues = []string{DefaultQueue}
	}
	retval.clients = make(map[string]*servicebus.Queue, len(queues))
	retval.ns = ns
	retval.UpsertQueue(context.TODO(), queues...)
	return
}

func (qp *queuePool) UpsertQueue(ctx context.Context, names ...string) error {
	qp.Lock()
	defer qp.Unlock()

	for i := range names {
		if _, ok := qp.clients[names[i]]; ok {
			continue
		}

		client, err := assertQueue(ctx, qp.ns, names[i])
		if err != nil {
			return err
		}
		qp.clients[names[i]] = client
	}
	return nil
}

// Receive starts all registered clients receiving. This method blocks only until each client has begun listening.
func (qp *queuePool) Receive(ctx context.Context, handler servicebus.Handler) error {
	qp.RLock()
	defer qp.RUnlock()

	return qp.receiveAll(ctx, handler)
}

func (qp *queuePool) receiveAll(ctx context.Context, handler servicebus.Handler) (err error) {
	// Already Receiving? Move along
	if qp.handles != nil {
		return nil
	}

	// Create a new unique identifier for a Receive session.
	qp.session, err = uuid.NewV4()
	if err != nil {
		return
	}

	qp.handles = make(map[string]*servicebus.ListenerHandle, len(qp.clients))
	qp.ctx, qp.cancel = context.WithCancel(context.Background())

	for k := range qp.clients {
		if err = qp.receive(ctx, k, handler); err != nil {
			qp.stop(ctx)
			return
		}
	}
	return
}

func (qp *queuePool) receive(ctx context.Context, name string, handler servicebus.Handler) error {
	client, ok := qp.clients[name]
	if !ok {
		return fmt.Errorf("queue %q not in pool", name)
	}

	handle, err := client.Receive(ctx, handler)
	if err != nil {
		return err
	}

	go func(handle *servicebus.ListenerHandle, sessionID uuid.UUID) {
		for {
			select {
			case <-qp.ctx.Done():
				return
			case <-handle.Done():
				qp.Lock()
				defer qp.Unlock()

				// Has stop() been called since this was started? If so, we shouldn't keep
				// trying to connect.
				if qp.session != sessionID {
					return
				}

				client, err := qp.ns.NewQueue(qp.ctx, name)
				if err != nil {
					continue
				}
				handle, err = client.Receive(qp.ctx, handler)
			}
		}
	}(handle, qp.session)

	return nil
}

func (qp *queuePool) stop(ctx context.Context) error {
	qp.Lock()
	defer qp.Unlock()

	encountered := uint16(0)
	for queue := range qp.handles {
		if err := qp.handles[queue].Close(ctx); err != nil {
			encountered++
		}
	}
	qp.cancel()

	qp.handles = nil
	qp.ctx = context.Background()
	qp.cancel = func() {}
	qp.session = uuid.UUID{}

	if encountered > 0 {
		return fmt.Errorf("encountered errors while closing %d queue clients, memory may have been leaked", encountered)
	}
	return nil
}

func assertQueue(ctx context.Context, ns *servicebus.Namespace, name string) (*servicebus.Queue, error) {
	qm := ns.NewQueueManager()
	qe, err := qm.Get(ctx, name)
	if err != nil {
		return nil, err
	}
	if qe == nil {
		_, err := qm.Put(ctx, name)
		if err != nil {
			return nil, err
		}
	}
	return ns.NewQueue(ctx, name)
}

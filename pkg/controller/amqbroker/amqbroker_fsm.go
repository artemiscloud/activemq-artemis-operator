package amqbroker

import (
	brokerv1alpha1 "github.com/rh-messaging/amq-broker-operator/pkg/apis/broker/v1alpha1"
	"github.com/rh-messaging/amq-broker-operator/pkg/utils/fsm"
	"k8s.io/apimachinery/pkg/types"
)

// Names of states
const (
	CreatingK8sResources   = "creating_k8s_resources"
	ConfiguringEnvironment = "configuring_broker_environment"
	CreatingContainer      = "creating_container"
	ContainerRunning       = "running"
)

// Completion of CreatingK8sResources state
const (
	None = 0
	CreatedHeadlessService = 1 << 0
	CreatedPersistentVolumeClaim = 1 << 1
	CreatedStatefulSet = 1 << 2
	CreatedConsoleJolokiaService = 1 << 3
	CreatedMuxProtocolService = 1 << 4

	//Complete = CreatedHeadlessService | CreatedConsoleJolokiaService | CreatedMuxProtocolService
	Complete = 	CreatedHeadlessService |
				CreatedPersistentVolumeClaim |
				CreatedConsoleJolokiaService |
				CreatedMuxProtocolService |
				CreatedStatefulSet
)

type AMQBrokerFSM struct {
	m 				fsm.IMachine
	namespacedName 	types.NamespacedName
	customResource 	*brokerv1alpha1.AMQBroker
	r 				*ReconcileAMQBroker
}

// Need to deep-copy the instance?
func MakeAMQBrokerFSM(instance *brokerv1alpha1.AMQBroker, _namespacedName types.NamespacedName, r *ReconcileAMQBroker) AMQBrokerFSM {

	var someIState		fsm.IState

	amqbfsm := AMQBrokerFSM{
		m: fsm.NewMachine(),
	}

	amqbfsm.namespacedName = _namespacedName
	amqbfsm.customResource = instance
	amqbfsm.r = r

	// TODO: Fix disconnect here between passing the parent and being added later as adding implies parenthood
	someRS := MakeCreatingK8sResourcesState(&amqbfsm, _namespacedName)
	someIState = &someRS
	amqbfsm.Add(&someIState)

	return amqbfsm
}

func NewAMQBrokerFSM(instance *brokerv1alpha1.AMQBroker, _namespacedName types.NamespacedName, r *ReconcileAMQBroker) *AMQBrokerFSM {
	amqbfsm := MakeAMQBrokerFSM(instance, _namespacedName, r)
	return &amqbfsm
}
func (amqbfsm *AMQBrokerFSM) Add(s *fsm.IState) {

	amqbfsm.m.Add(s)
}

func (amqbfsm *AMQBrokerFSM) Remove(s *fsm.IState) {

	amqbfsm.m.Remove(s)
}

func (amqbfsm *AMQBrokerFSM) Enter(stateFrom *fsm.IState) {

	// For the moment sequentially set stuff up
	// k8s resource creation and broker environment configuration can probably be done concurrently later

	// Enter == Setup
	amqbfsm.m.Enter(stateFrom)
}

func (amqbfsm *AMQBrokerFSM) Update() {

	// Update == Reconcile
}

func (amqbfsm *AMQBrokerFSM) Exit(stateFrom *fsm.IState) {

	// Exit == Teardown
	amqbfsm.m.Exit(stateFrom)
}


//func (amqbfsm *AMQBrokerFSM) Update(request reconcile.Request, r *ReconcileAMQBroker) (reconcile.Result, error) {
//
//	// Log where we are and what we're doing
//	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
//	reqLogger.Info("Executing AMQBrokerFSM Update")
//
//	// Set it up
//	var err error = nil
//	var reconcileResult reconcile.Result
//	instance := &brokerv1alpha1.AMQBroker{}
//	found := &corev1.Pod{}
//
//	// Do what's needed
//	for {
//		// Fetch the AMQBroker instance
//		if err = r.client.Get(context.TODO(), request.NamespacedName, instance); err != nil {
//			// Add error detail for use later
//			break
//		}
//
//		// Check to see if the k8s resources exist, if not that's our state
//		// Lets assume we are installing for the first time for the moment...
//
//
//		// Check if this Pod already exists
//		if err = r.client.Get(context.TODO(), amqbfsm.namespacedName, found); err == nil {
//			// Don't do anything as the pod exists
//			break
//		}
//
//		break
//	}
//
//	// Handle error, if any
//	if err != nil {
//		if errors.IsNotFound(err) {
//			reconcileResult = reconcile.Result{}
//			reqLogger.Error(err, "AMQBroker Controller Reconcile encountered a IsNotFound, preventing request requeue", "Pod.Namespace", request.Namespace, "Pod.Name", request.Name)
//			// Setting err to nil to prevent requeue
//			err = nil
//		} else {
//			//log.Error(err, "AMQBroker Controller Reconcile errored")
//			reqLogger.Error(err, "AMQBroker Controller Reconcile errored, requeuing request", "Pod.Namespace", request.Namespace, "Pod.Name", request.Name)
//		}
//	}
//
//	// Single exit, return the result and error condition
//	return reconcileResult, err
//
//}


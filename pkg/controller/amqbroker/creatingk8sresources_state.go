package amqbroker

import (
	"context"
	brokerv1alpha1 "github.com/rh-messaging/amq-broker-operator/pkg/apis/broker/v1alpha1"

	"github.com/rh-messaging/amq-broker-operator/pkg/utils/fsm"
	"github.com/rh-messaging/amq-broker-operator/pkg/utils/selectors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)


type CreatingK8sResourcesState struct {

	s				fsm.State
	namespacedName  types.NamespacedName
	parentFSM		*AMQBrokerFSM
	stepsComplete 	uint8
}

func MakeCreatingK8sResourcesState(_parentFSM *AMQBrokerFSM, _namespacedName types.NamespacedName) CreatingK8sResourcesState{

	rs := CreatingK8sResourcesState{
		s: fsm.MakeState(CreatingK8sResources),
		namespacedName: _namespacedName,
		parentFSM: _parentFSM,
		stepsComplete: None,
	}

	return rs
}

func NewCreatingK8sResourcesState(_parentFSM *AMQBrokerFSM, _namespacedName types.NamespacedName) *CreatingK8sResourcesState{

	rs := MakeCreatingK8sResourcesState(_parentFSM, _namespacedName)

	return &rs
}

func (rs *CreatingK8sResourcesState) Enter(stateFrom *fsm.IState) {

	// Log where we are and what we're doing
	reqLogger := log.WithValues("AMQBroker Name", rs.parentFSM.customResource.Name)
	reqLogger.Info("Entering CreateK8sResourceState")

	// Check to see if the headless service already exists
	if _, err := rs.RetrieveHeadlessService(rs.parentFSM.customResource, rs.parentFSM.namespacedName, rs.parentFSM.r); err != nil {
		// err means not found, so create
		if _, createError := rs.CreateHeadlessService(rs.parentFSM.customResource, getDefaultPorts()); createError == nil {
			rs.stepsComplete |= CreatedHeadlessService
		}
	}

}

func (rs *CreatingK8sResourcesState) Update() {

}

func (rs *CreatingK8sResourcesState) Exit(stateFrom *fsm.IState) {

	// Log where we are and what we're doing
	reqLogger := log.WithValues("AMQBroker Name", rs.parentFSM.customResource.Name)
	reqLogger.Info("Exiting CreatingK8sResourceState")

	// Check to see if the headless service already exists
	if _, err := rs.RetrieveHeadlessService(rs.parentFSM.customResource, rs.parentFSM.namespacedName, rs.parentFSM.r); err != nil {
		// err means not found, so mark deleted
		//if _, createError := rs.CreateHeadlessService(rs.parentFSM.customResource, getDefaultPorts()); createError == nil {

		//}
		rs.stepsComplete &^= CreatedHeadlessService
	}
}

func (rs *CreatingK8sResourcesState) CreateHeadlessService(cr *brokerv1alpha1.AMQBroker, servicePorts *[]corev1.ServicePort) (*corev1.Service, error) {

	// Log where we are and what we're doing
	reqLogger := log.WithValues("AMQBroker Name", cr.Name)
	reqLogger.Info("Creating new " + "headless" + " service")

	labels := selectors.LabelsForAMQBroker(cr.Name)

	//port := corev1.ServicePort{
	//	Name:		cr.Name + "-" + "headless" + "-port",
	//	Protocol:	"TCP",
	//	Port:		port_number,
	//	TargetPort:	intstr.FromInt(int(port_number)),
	//}
	//ports := []corev1.ServicePort{}
	//ports = append(ports, port)

	headlessSvc := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:		"Service",
		},
		ObjectMeta: metav1.ObjectMeta {
			Annotations: 	nil,
			Labels: 		labels,
			Name:			"headless" + "-service",
			Namespace:		cr.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Type: 	"ClusterIP",
			Ports: 	*servicePorts,
			Selector: labels,
			ClusterIP: "None",
			PublishNotReadyAddresses: true,
		},
	}

	// Define the headless Service for the StatefulSet
	//headlessSvc := newHeadlessServiceForCR(instance, 5672)
	//headlessSvc := newHeadlessServiceForCR(rs.parentFSM.customResource, getDefaultPorts())
	// Set AMQBroker instance as the owner and controller
	var err error = nil
	if err = controllerutil.SetControllerReference(rs.parentFSM.customResource, headlessSvc, rs.parentFSM.r.scheme); err != nil {
		// Add error detail for use later
		reqLogger.Info("Failed to set controller reference for new " + "headless" + " service")
	}
	reqLogger.Info("Set controller reference for new " + "headless" + " service")

	// Call k8s create for service
	if err = rs.parentFSM.r.client.Create(context.TODO(), headlessSvc); err != nil {
		// Add error detail for use later
		reqLogger.Info("Failed to creating new " + "headless" + " service")

	}
	reqLogger.Info("Created new " + "headless" + " service")

	return headlessSvc, err
}

func (rs* CreatingK8sResourcesState) DeleteHeadlessService(instance *brokerv1alpha1.AMQBroker) {
	// kubectl delete cleans up kubernetes resources, just need to clean up local resources if any
}

//r *ReconcileAMQBroker
func (rs* CreatingK8sResourcesState) RetrieveHeadlessService(instance *brokerv1alpha1.AMQBroker, namespacedName types.NamespacedName, r *ReconcileAMQBroker) (*corev1.Service, error) {

	// Log where we are and what we're doing
	reqLogger := log.WithValues("AMQBroker Name", instance.Name)
	reqLogger.Info("Retrieving " + "headless" + " service")

	var err error = nil
	headlessService := &corev1.Service{}

	// Check if the headless service already exists
	if err = r.client.Get(context.TODO(), namespacedName, headlessService); err != nil {
		if errors.IsNotFound(err) {
			//reconcileResult = reconcile.Result{}
			reqLogger.Error(err, "Headless service IsNotFound", "Namespace", instance.Namespace, "Name", instance.Name)
			// Setting err to nil to prevent requeue
			//err = nil
			//rs.headlessService = nil
		} else {
			//log.Error(err, "AMQBroker Controller Reconcile errored")
			reqLogger.Error(err, "Headless service found", "Namespace", instance.Namespace, "Name", instance.Name)
			//rs.headlessService = found
		}
	}
	//// Handle error, if any
	//if err != nil {
	//	if errors.IsNotFound(err) {
	//		reconcileResult = reconcile.Result{}
	//		reqLogger.Error(err, "AMQBroker Controller Reconcile encountered a IsNotFound, preventing request requeue", "Pod.Namespace", request.Namespace, "Pod.Name", request.Name)
	//		// Setting err to nil to prevent requeue
	//		err = nil
	//	} else {
	//		//log.Error(err, "AMQBroker Controller Reconcile errored")
	//		reqLogger.Error(err, "AMQBroker Controller Reconcile errored, requeuing request", "Pod.Namespace", request.Namespace, "Pod.Name", request.Name)
	//	}
	//}


	//// Define the headless Service for the StatefulSet
	////headlessSvc := newHeadlessServiceForCR(instance, 5672)
	//headlessSvc := newHeadlessServiceForCR(instance, getDefaultPorts())
	//// Set AMQBroker instance as the owner and controller
	//if err = controllerutil.SetControllerReference(instance, headlessSvc, r.scheme); err != nil {
	//	// Add error detail for use later
	//	break
	//}
	//// Call k8s create for service
	//if err = r.client.Create(context.TODO(), headlessSvc); err != nil {
	//	// Add error detail for use later
	//	break
	//}

	return headlessService, err
}

//func MakeCreatingK8sResourcesState() CreatingK8sResourcesState {
//
//	rs := CreatingK8sResourcesState {
//		s: fsm.MakeState(CreatingK8sResources),
//		stepsComplete: 0,
//	}
//
//	return rs
//}

//func MakeAMQBrokerFSM() fsm.Machine {
//
//	m := fsm.MakeMachine()
//
//	m.Add(fsm.NewState(CreatingK8sResources))
//	m.Add(fsm.NewState(ConfiguringEnvironment))
//	m.Add(fsm.NewState(CreatingContainer))
//	m.Add(fsm.NewState(ContainerRunning))
//
//	return m
//}



//func MakeCreatingK8sResourcesState() fsm.State {
//
//	s := fsm.MakeState(CreatingK8sResources)
//
//	return s
//}

//func MakeConfiguringEnvironmentState() fsm.State {
//
//	s := fsm.MakeState(ConfiguringEnvironment)
//
//	return s
//}
//
//func MakeCreatingContainerState() fsm.State {
//
//	s := fsm.MakeState(CreatingContainer)
//
//	return s
//}

//func (m *fsm.Machine) Update(request reconcile.Request) {
//
//}
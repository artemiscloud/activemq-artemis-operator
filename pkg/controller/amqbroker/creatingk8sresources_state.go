package amqbroker

import (
	svc "github.com/rh-messaging/activemq-artemis-operator/pkg/resources/services"
	ss "github.com/rh-messaging/activemq-artemis-operator/pkg/resources/statefulsets"

	"github.com/rh-messaging/activemq-artemis-operator/pkg/utils/fsm"
	"k8s.io/apimachinery/pkg/types"
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

	var err error = nil
	var retrieveError error = nil

	// Check to see if the statefulset already exists
	if _, err := ss.RetrieveStatefulSet(rs.parentFSM.customResource, rs.parentFSM.namespacedName, rs.parentFSM.r.client); err != nil {
		// err means not found, so create
		if _, retrieveError := ss.CreateStatefulSet(rs.parentFSM.customResource, rs.parentFSM.r.client, rs.parentFSM.r.scheme); retrieveError == nil {
			rs.stepsComplete |= CreatedStatefulSet
		}
	}

	// Check to see if the headless service already exists
	if _, err = svc.RetrieveHeadlessService(rs.parentFSM.customResource, rs.parentFSM.namespacedName, rs.parentFSM.r.client); err != nil {
		// err means not found, so create
		if _, retrieveError = svc.CreateHeadlessService(rs.parentFSM.customResource, rs.parentFSM.r.client, rs.parentFSM.r.scheme); retrieveError == nil {
			rs.stepsComplete |= CreatedHeadlessService
		}
	}

	// Check to see if the ping service already exists
	if _, err = svc.RetrievePingService(rs.parentFSM.customResource, rs.parentFSM.namespacedName, rs.parentFSM.r.client); err != nil {
		// err means not found, so create
		if _, retrieveError = svc.CreatePingService(rs.parentFSM.customResource, rs.parentFSM.r.client, rs.parentFSM.r.scheme); retrieveError == nil {
			rs.stepsComplete |= CreatedPingService
		}
	}
	//// These next two should be considered "hard coded" and temporary
	//// Define the console-jolokia Service for this Pod
	//consoleJolokiaSvc := svc.NewServiceForCR(rs.parentFSM.customResource, "console-jolokia", 8161)
	//// Set AMQBroker instance as the owner and controller
	//if err = controllerutil.SetControllerReference(rs.parentFSM.customResource, consoleJolokiaSvc, rs.parentFSM.r.scheme); err != nil {
	//	// Add error detail for use later
	//
	//}
	//// Call k8s create for service
	//if err = rs.parentFSM.r.client.Create(context.TODO(), consoleJolokiaSvc); err != nil {
	//	// Add error detail for use later
	//	rs.stepsComplete |= CreatedConsoleJolokiaService
	//}
	//
	//// Define the console-jolokia Service for this Pod
	//muxProtocolSvc := svc.NewServiceForCR(rs.parentFSM.customResource, "mux-protocol", 61616)
	//// Set AMQBroker instance as the owner and controller
	//if err = controllerutil.SetControllerReference(rs.parentFSM.customResource, muxProtocolSvc, rs.parentFSM.r.scheme); err != nil {
	//	// Add error detail for use later
	//}
	//// Call k8s create for service
	//if err = rs.parentFSM.r.client.Create(context.TODO(), muxProtocolSvc); err != nil {
	//	// Add error detail for use later
	//	rs.stepsComplete |= CreatedMuxProtocolService
	//}

	// Check to see if the persistent volume claim already exists
	//if _, err := rs.RetrievePersistentVolumeClaim(rs.parentFSM.customResource, rs.parentFSM.namespacedName, rs.parentFSM.r); err != nil {
	//	// err means not found, so create
	//	if _, retrieveError := rs.CreatePersistentVolumeClaim(rs.parentFSM.customResource); retrieveError == nil {
	//		rs.stepsComplete |= CreatedPersistentVolumeClaim
	//	}
	//}
}

func (rs *CreatingK8sResourcesState) Update() {

}

func (rs *CreatingK8sResourcesState) Exit(stateFrom *fsm.IState) {

	// Log where we are and what we're doing
	reqLogger := log.WithValues("AMQBroker Name", rs.parentFSM.customResource.Name)
	reqLogger.Info("Exiting CreatingK8sResourceState")

	var err error = nil

	// Check to see if the headless service already exists
	if _, err = svc.RetrieveHeadlessService(rs.parentFSM.customResource, rs.parentFSM.namespacedName, rs.parentFSM.r.client); err != nil {
		// err means not found, so mark deleted
		//if _, createError := rs.CreateHeadlessService(rs.parentFSM.customResource, getDefaultPorts()); createError == nil {

		//}
		rs.stepsComplete &^= CreatedHeadlessService
	}

	// Check to see if the persistent volume claim already exists
	//if _, err = rs.RetrievePersistentVolumeClaim(rs.parentFSM.customResource, rs.parentFSM.namespacedName, rs.parentFSM.r); err != nil {
	//	// err means not found, so mark deleted
	//	rs.stepsComplete &^= CreatedPersistentVolumeClaim
	//}

	// Check to see if the persistent volume claim already exists
	if _, err = ss.RetrieveStatefulSet(rs.parentFSM.customResource, rs.parentFSM.namespacedName, rs.parentFSM.r.client); err != nil {
		// err means not found, so mark deleted
		rs.stepsComplete &^= CreatedStatefulSet
	}
}


package activemqartemis

import (
	"github.com/rh-messaging/activemq-artemis-operator/pkg/resources/ingresses"
	"github.com/rh-messaging/activemq-artemis-operator/pkg/resources/routes"
	svc "github.com/rh-messaging/activemq-artemis-operator/pkg/resources/services"
	ss "github.com/rh-messaging/activemq-artemis-operator/pkg/resources/statefulsets"
	"github.com/rh-messaging/activemq-artemis-operator/pkg/utils/env"
	"github.com/rh-messaging/activemq-artemis-operator/pkg/utils/fsm"
	"k8s.io/apimachinery/pkg/types"
)

type CreatingK8sResourcesState struct {
	s              fsm.State
	namespacedName types.NamespacedName
	parentFSM      *ActiveMQArtemisFSM
	stepsComplete  uint8
}

func MakeCreatingK8sResourcesState(_parentFSM *ActiveMQArtemisFSM, _namespacedName types.NamespacedName) CreatingK8sResourcesState {

	rs := CreatingK8sResourcesState{
		s:              fsm.MakeState(CreatingK8sResources),
		namespacedName: _namespacedName,
		parentFSM:      _parentFSM,
		stepsComplete:  None,
	}

	return rs
}

func NewCreatingK8sResourcesState(_parentFSM *ActiveMQArtemisFSM, _namespacedName types.NamespacedName) *CreatingK8sResourcesState {

	rs := MakeCreatingK8sResourcesState(_parentFSM, _namespacedName)

	return &rs
}

func (rs *CreatingK8sResourcesState) Enter(stateFrom *fsm.IState) error {

	// Log where we are and what we're doing
	reqLogger := log.WithValues("ActiveMQArtemis Name", rs.parentFSM.customResource.Name)
	reqLogger.Info("Entering CreateK8sResourceState")

	var err error = nil
	var retrieveError error = nil

	// Check to see if the statefulset already exists
	if _, err := ss.RetrieveStatefulSet(rs.parentFSM.customResource.Name+"-ss", rs.parentFSM.namespacedName, rs.parentFSM.r.client); err != nil {
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

	// Check to see if the console-jolokia service already exists
	if _, err = svc.RetrieveConsoleJolokiaService(rs.parentFSM.customResource, rs.parentFSM.namespacedName, rs.parentFSM.r.client); err != nil {
		// err means not found, so create
		if _, retrieveError = svc.CreateConsoleJolokiaService(rs.parentFSM.customResource, rs.parentFSM.r.client, rs.parentFSM.r.scheme); retrieveError == nil {
			rs.stepsComplete |= CreatedConsoleJolokiaService
		}
	}

	// Check to see if the mux-protocol service already exists
	if _, err = svc.RetrieveMuxProtocolService(rs.parentFSM.customResource, rs.parentFSM.namespacedName, rs.parentFSM.r.client); err != nil {
		// err means not found, so create
		if _, retrieveError = svc.CreateMuxProtocolService(rs.parentFSM.customResource, rs.parentFSM.r.client, rs.parentFSM.r.scheme); retrieveError == nil {
			rs.stepsComplete |= CreatedMuxProtocolService
		}
	}

	isOpenshift, err1 := env.DetectOpenshift()
	if err1 != nil {
		log.Error(err1, "Failed to get env")
		return
	}

	if isOpenshift {
		log.Info("Evnironment is OpenShift, creating route")
		if _, err = routes.RetrieveRoute(rs.parentFSM.customResource, rs.parentFSM.namespacedName, rs.parentFSM.r.client); err != nil {
			// err means not found, so create routes
			if _, retrieveError = routes.CreateNewRoute(rs.parentFSM.customResource, rs.parentFSM.r.client, rs.parentFSM.r.scheme); retrieveError == nil {
				rs.stepsComplete |= CreatedRouteOrIngress
			}
		}
	} else {
		log.Info("Environment is not OpenShift, creating ingress")

		if _, err = ingresses.RetrieveIngress(rs.parentFSM.customResource, rs.parentFSM.namespacedName, rs.parentFSM.r.client); err != nil {
			// err means not found, so create routes
			if _, retrieveError = ingresses.CreateNewIngress(rs.parentFSM.customResource, rs.parentFSM.r.client, rs.parentFSM.r.scheme); retrieveError == nil {
				rs.stepsComplete |= CreatedRouteOrIngress
			}

		}

		// Check to see if the routes already exists

	}

	// Check to see if the persistent volume claim already exists
	//if _, err := rs.RetrievePersistentVolumeClaim(rs.parentFSM.customResource, rs.parentFSM.namespacedName, rs.parentFSM.r); err != nil {
	//	// err means not found, so create
	//	if _, retrieveError := rs.CreatePersistentVolumeClaim(rs.parentFSM.customResource); retrieveError == nil {
	//		rs.stepsComplete |= CreatedPersistentVolumeClaim
	//	}
	//}

	return nil
}

func (rs *CreatingK8sResourcesState) Update() error {

	return nil
}

func (rs *CreatingK8sResourcesState) Exit(stateTo *fsm.IState) error {

	// Log where we are and what we're doing
	reqLogger := log.WithValues("ActiveMQArtemis Name", rs.parentFSM.customResource.Name)
	reqLogger.Info("Exiting CreatingK8sResourceState")

	//var err error = nil
	//
	//// Check to see if the headless service already exists
	//if _, err = svc.RetrieveHeadlessService(rs.parentFSM.customResource, rs.parentFSM.namespacedName, rs.parentFSM.r.client); err != nil {
	//	// err means not found, so mark deleted
	//	rs.stepsComplete &^= CreatedHeadlessService
	//}
	//
	//// Check to see if the persistent volume claim already exists
	//if _, err = ss.RetrieveStatefulSet("", rs.parentFSM.namespacedName, rs.parentFSM.r.client); err != nil {
	//	// err means not found, so mark deleted
	//	rs.stepsComplete &^= CreatedStatefulSet
	//}
	//
	//// Check to see if the ping service already exists
	//if _, err = svc.RetrievePingService(rs.parentFSM.customResource, rs.parentFSM.namespacedName, rs.parentFSM.r.client); err != nil {
	//	// err means not found, so mark deleted
	//	rs.stepsComplete &^= CreatedPingService
	//}
	//
	//// Check to see if the console-jolokia service already exists
	//if _, err = svc.RetrieveConsoleJolokiaService(rs.parentFSM.customResource, rs.parentFSM.namespacedName, rs.parentFSM.r.client); err != nil {
	//	// err means not found, so mark deleted
	//	rs.stepsComplete &^= CreatedConsoleJolokiaService
	//}
	//
	//// Check to see if the mux-protocol service already exists
	//if _, err = svc.RetrieveMuxProtocolService(rs.parentFSM.customResource, rs.parentFSM.namespacedName, rs.parentFSM.r.client); err != nil {
	//	// err means not found, so mark deleted
	//	rs.stepsComplete &^= CreatedMuxProtocolService
	//}
	//
	//isOpenshift, err1 := env.DetectOpenshift()
	//if err1 != nil {
	//	log.Error(err1, "Failed to get env")
	//	return
	//}
	//
	//if isOpenshift {
	//	log.Info("Evnironment is OpenShift, checking for created route")
	//	if _, err = routes.RetrieveRoute(rs.parentFSM.customResource, rs.parentFSM.namespacedName, rs.parentFSM.r.client); err != nil {
	//		// err means not found, so mark deleted
	//		rs.stepsComplete &^= CreatedRouteOrIngress
	//	}
	//} else {
	//	log.Info("Environment is not OpenShift, checking for created ingress")
	//
	//	if _, err = ingresses.RetrieveIngress(rs.parentFSM.customResource, rs.parentFSM.namespacedName, rs.parentFSM.r.client); err != nil {
	//		// err means not found, so mark deleted
	//		rs.stepsComplete &^= CreatedRouteOrIngress
	//	}
	//
	//	// Check to see if the routes already exists
	//
	//}
	//
	//// Check to see if the persistent volume claim already exists
	////if _, err = rs.RetrievePersistentVolumeClaim(rs.parentFSM.customResource, rs.parentFSM.namespacedName, rs.parentFSM.r); err != nil {
	////	// err means not found, so mark deleted
	////	rs.stepsComplete &^= CreatedPersistentVolumeClaim
	////}

	return nil
}

package activemqartemis

import (
	"context"
	"github.com/rh-messaging/activemq-artemis-operator/pkg/resources/ingresses"
	"github.com/rh-messaging/activemq-artemis-operator/pkg/resources/routes"
	svc "github.com/rh-messaging/activemq-artemis-operator/pkg/resources/services"
	ss "github.com/rh-messaging/activemq-artemis-operator/pkg/resources/statefulsets"
	"github.com/rh-messaging/activemq-artemis-operator/pkg/utils/env"
	"github.com/rh-messaging/activemq-artemis-operator/pkg/utils/fsm"
	appsv1 "k8s.io/api/apps/v1"
	"github.com/rh-messaging/activemq-artemis-operator/pkg/utils/selectors"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"strconv"
)


// This is the state we should be in whenever something happens that
// requires a change to the kubernetes resources
type CreatingK8sResourcesState struct {
	s              fsm.State
	namespacedName types.NamespacedName
	parentFSM      *ActiveMQArtemisFSM
	stepsComplete  uint8
}

func MakeCreatingK8sResourcesState(_parentFSM *ActiveMQArtemisFSM, _namespacedName types.NamespacedName) CreatingK8sResourcesState {

	rs := CreatingK8sResourcesState{
		s:              fsm.MakeState(CreatingK8sResources, CreatingK8sResourcesID),
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

func (rs *CreatingK8sResourcesState) ID() int {
	return CreatingK8sResourcesID
}

// First time entering state
func (rs *CreatingK8sResourcesState) enterFromInvalidState() error {

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

	if rs.parentFSM.customResource.Spec.Size > 0 {
		err = rs.configureServices()
		err = rs.configureExternalNetworkAccess()
	}

	return err
}

func (rs *CreatingK8sResourcesState) enterFromContainerRunningState() error {

	var err error = nil

	// Did we...
	err = rs.configureServices()
	err = rs.configureExternalNetworkAccess()

	return err
}

func (rs *CreatingK8sResourcesState) configureServices() error {

	// Log where we are and what we're doing
	reqLogger := log.WithValues("ActiveMQArtemis Name", rs.parentFSM.customResource.Name)
	reqLogger.Info("CreatingK8sResourcesState configureServices")

	//// Check to see if the console-jolokia service already exists
	//if _, err = svc.RetrieveConsoleJolokiaService(rs.parentFSM.customResource, rs.parentFSM.namespacedName, rs.parentFSM.r.client); err != nil {
	//	// err means not found, so create
	//	if err = svc.CreateServices(rs.parentFSM.customResource, rs.parentFSM.r.client, rs.parentFSM.r.scheme, "console-jolokia", 8161); err == nil {
	//		rs.stepsComplete |= CreatedConsoleJolokiaService
	//	}
	//}
	var err error = nil
	var i int32 = 0
	ordinalString := ""

	portNumber := int32(8161)
	cr := rs.parentFSM.customResource
	client := rs.parentFSM.r.client
	scheme := rs.parentFSM.r.scheme
	labels := selectors.LabelsForActiveMQArtemis(cr.Name)

	for ; i < cr.Spec.Size; i++ {
		ordinalString = strconv.Itoa(int(i))
		labels["statefulset.kubernetes.io/pod-name"] = cr.Name + "-ss" + "-" + ordinalString

		baseServiceName := "console-jolokia"
		serviceDefinition := svc.NewServiceDefinitionForCR(cr, baseServiceName + "-" + ordinalString, portNumber, labels)
		if err = svc.CreateService(cr, client, scheme, serviceDefinition); err != nil {
			reqLogger.Info("Failure to create " + baseServiceName + " service " + ordinalString)
			continue
		}

		baseServiceName = "all-protocol"
		serviceDefinition = svc.NewServiceDefinitionForCR(cr, baseServiceName + "-" + ordinalString, portNumber, labels)
		if err = svc.CreateService(cr, client, scheme, serviceDefinition); err != nil {
			reqLogger.Info("Failure to create " + baseServiceName + " service " + ordinalString)
			continue
		}
	}

	//// Check to see if the mux-protocol service already exists
	//if _, err = svc.RetrieveAllProtocolService(rs.parentFSM.customResource, rs.parentFSM.namespacedName, rs.parentFSM.r.client); err != nil {
	//	// err means not found, so create
	//	if err = svc.CreateServices(rs.parentFSM.customResource, rs.parentFSM.r.client, rs.parentFSM.r.scheme, "all-protocol", 61616); err == nil {
	//		rs.stepsComplete |= CreatedMuxProtocolService
	//	}
	//}

	return err
}

func (rs *CreatingK8sResourcesState) Enter(previousStateID int) error {

	// Log where we are and what we're doing
	reqLogger := log.WithValues("ActiveMQArtemis Name", rs.parentFSM.customResource.Name)
	reqLogger.Info("Entering CreateK8sResourceState from " + strconv.Itoa(previousStateID))

	switch (previousStateID) {
	case InvalidState:
		rs.enterFromInvalidState()
		break
	case ContainerRunningID:
		rs.enterFromContainerRunningState()
		break
	}

	return nil
}

func (rs *CreatingK8sResourcesState) configureExternalNetworkAccess() error {

	var err error = nil
	var isOpenshift bool = false

	if isOpenshift, err = env.DetectOpenshift(); err != nil {
		log.Error(err, "Failed to get env, will try kubernetes")
	}
	if isOpenshift {
		log.Info("Evnironment is OpenShift")
		err = rs.configureRoutes()
	} else {
		log.Info("Environment is not OpenShift, creating ingress")
		err = rs.configureIngress()
	}

	return err
}

func (rs *CreatingK8sResourcesState) configureIngress() error {

	var err error = nil

	if _, err = ingresses.RetrieveIngress(rs.parentFSM.customResource, rs.parentFSM.namespacedName, rs.parentFSM.r.client); err != nil {
		// err means not found, so create routes
		if _, err = ingresses.CreateNewIngress(rs.parentFSM.customResource, rs.parentFSM.r.client, rs.parentFSM.r.scheme); err == nil {
			rs.stepsComplete |= CreatedRouteOrIngress
		}
	}

	return err
}

func (rs *CreatingK8sResourcesState) configureRoutes() error {

	var err error = nil
	var passthroughTLS bool

	for i := 0; i < int(rs.parentFSM.customResource.Spec.Size); i++ {
		passthroughTLS = false
		targetPortName := "console-jolokia" + "-" + strconv.Itoa(i)
		targetServiceName := rs.parentFSM.customResource.Name + "-service-" + targetPortName
		log.Info("Checking route for " + targetPortName)

		//if len(rs.parentFSM.customResource.Spec.SSLConfig.SecretName) != 0 &&
		//	((len(rs.parentFSM.customResource.Spec.SSLConfig.KeyStorePassword) != 0 && len(rs.parentFSM.customResource.Spec.SSLConfig.KeystoreFilename) != 0) ||
		//		(len(rs.parentFSM.customResource.Spec.SSLConfig.TrustStorePassword) != 0 && len(rs.parentFSM.customResource.Spec.SSLConfig.TrustStoreFilename) != 0)) {
		//	passthroughTLS = true
		//}

		consoleJolokiaRoute := routes.NewRouteDefinitionForCR(rs.parentFSM.customResource, selectors.LabelsForActiveMQArtemis(rs.parentFSM.customResource.Name), targetServiceName, targetPortName, passthroughTLS)
		//if err = routes.RetrieveRoute(rs.parentFSM.customResource, rs.parentFSM.namespacedName, rs.parentFSM.r.client, consoleJolokiaRoute); err != nil {
		//	log.Info("Creating route")
		//	// err means not found, so create routes
		//	if err = routes.CreateNewRoute(rs.parentFSM.customResource, rs.parentFSM.r.client, rs.parentFSM.r.scheme, consoleJolokiaRoute); err == nil {
		//		rs.stepsComplete |= CreatedRouteOrIngress
		//	}
		//} else {
		//	err = routes.UpdateRoute(rs.parentFSM.customResource, rs.parentFSM.r.client, consoleJolokiaRoute)
		//}
		if err = routes.RetrieveRoute(rs.parentFSM.customResource, rs.parentFSM.namespacedName, rs.parentFSM.r.client, consoleJolokiaRoute); err != nil {
			routes.DeleteRoute(rs.parentFSM.customResource, rs.parentFSM.r.client, consoleJolokiaRoute)
		}
		if err = routes.CreateNewRoute(rs.parentFSM.customResource, rs.parentFSM.r.client, rs.parentFSM.r.scheme, consoleJolokiaRoute); err == nil {
			rs.stepsComplete |= CreatedRouteOrIngress
		}


		targetPortName = "all-protocol" + "-" + strconv.Itoa(i)
		targetServiceName = rs.parentFSM.customResource.Name + "-service-" + targetPortName
		log.Info("Checking route for " + targetPortName)
		passthroughTLS = true
		allProtocolRoute := routes.NewRouteDefinitionForCR(rs.parentFSM.customResource, selectors.LabelsForActiveMQArtemis(rs.parentFSM.customResource.Name), targetServiceName, targetPortName, passthroughTLS)

		//if err = routes.RetrieveRoute(rs.parentFSM.customResource, rs.parentFSM.namespacedName, rs.parentFSM.r.client, allProtocolRoute); err != nil {
		//	log.Info("Creating route")
		//	// err means not found, so create routes
		//	if err = routes.CreateNewRoute(rs.parentFSM.customResource, rs.parentFSM.r.client, rs.parentFSM.r.scheme, allProtocolRoute); err == nil {
		//		rs.stepsComplete |= CreatedRouteOrIngress
		//	}
		//} else {
		//	err = routes.UpdateRoute(rs.parentFSM.customResource, rs.parentFSM.r.client, allProtocolRoute)
		//}
		if err = routes.RetrieveRoute(rs.parentFSM.customResource, rs.parentFSM.namespacedName, rs.parentFSM.r.client, allProtocolRoute); err != nil {
			routes.DeleteRoute(rs.parentFSM.customResource, rs.parentFSM.r.client, allProtocolRoute)
		}
		if err = routes.CreateNewRoute(rs.parentFSM.customResource, rs.parentFSM.r.client, rs.parentFSM.r.scheme, allProtocolRoute); err == nil {
			rs.stepsComplete |= CreatedRouteOrIngress
		}
	}

	return err
}

func (rs *CreatingK8sResourcesState) Update() (error, int) {

	// Log where we are and what we're doing
	reqLogger := log.WithValues("ActiveMQArtemis Name", rs.parentFSM.customResource.Name)
	reqLogger.Info("Updating CreatingK8sResourcesState")

	currentStatefulSet := &appsv1.StatefulSet{}
	err := rs.parentFSM.r.client.Get(context.TODO(), types.NamespacedName{Name: rs.parentFSM.customResource.Name + "-ss", Namespace: rs.parentFSM.customResource.Namespace}, currentStatefulSet)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Error(err, "Failed to get StatefulSet.", "Deployment.Namespace", currentStatefulSet.Namespace, "Deployment.Name", currentStatefulSet.Name)
		return nil, CreatingK8sResourcesID
	}

	// Ensure the StatefulSet size is the same as the spec
	size := rs.parentFSM.customResource.Spec.Size
	if *currentStatefulSet.Spec.Replicas != size {
		currentStatefulSet.Spec.Replicas = &size
		err = rs.parentFSM.r.client.Update(context.TODO(), currentStatefulSet)
		if err != nil {
			reqLogger.Error(err, "Failed to update StatefulSet.", "Deployment.Namespace", currentStatefulSet.Namespace, "Deployment.Name", currentStatefulSet.Name)
		}
	} else {
		// go to next state
		return nil, ContainerRunningID
	}

	return nil, CreatingK8sResourcesID
}

func (rs *CreatingK8sResourcesState) Exit() error {

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
	//if _, err = svc.RetrieveAllProtocolService(rs.parentFSM.customResource, rs.parentFSM.namespacedName, rs.parentFSM.r.client); err != nil {
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

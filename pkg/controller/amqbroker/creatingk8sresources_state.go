package amqbroker

import (
	"context"
	brokerv1alpha1 "github.com/rh-messaging/amq-broker-operator/pkg/apis/broker/v1alpha1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/rh-messaging/amq-broker-operator/pkg/utils/fsm"
	"github.com/rh-messaging/amq-broker-operator/pkg/utils/selectors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	appsv1 "k8s.io/api/apps/v1"
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
	if _, err := rs.RetrieveStatefulSet(rs.parentFSM.customResource, rs.parentFSM.namespacedName, rs.parentFSM.r); err != nil {
		// err means not found, so create
		if _, retrieveError := rs.CreateStatefulSet(rs.parentFSM.customResource); retrieveError == nil {
			rs.stepsComplete |= CreatedStatefulSet
		}
	}

	// Check to see if the headless service already exists
	if _, err = rs.RetrieveHeadlessService(rs.parentFSM.customResource, rs.parentFSM.namespacedName, rs.parentFSM.r); err != nil {
		// err means not found, so create
		if _, retrieveError = rs.CreateHeadlessService(rs.parentFSM.customResource, getDefaultPorts()); retrieveError == nil {
			rs.stepsComplete |= CreatedHeadlessService
		}
	}

	// These next two should be considered "hard coded" and temporary
	// Define the console-jolokia Service for this Pod
	consoleJolokiaSvc := newServiceForCR(rs.parentFSM.customResource, "console-jolokia", 8161)
	// Set AMQBroker instance as the owner and controller
	if err = controllerutil.SetControllerReference(rs.parentFSM.customResource, consoleJolokiaSvc, rs.parentFSM.r.scheme); err != nil {
		// Add error detail for use later

	}
	// Call k8s create for service
	if err = rs.parentFSM.r.client.Create(context.TODO(), consoleJolokiaSvc); err != nil {
		// Add error detail for use later
		rs.stepsComplete |= CreatedConsoleJolokiaService
	}

	// Define the console-jolokia Service for this Pod
	muxProtocolSvc := newServiceForCR(rs.parentFSM.customResource, "mux-protocol", 61616)
	// Set AMQBroker instance as the owner and controller
	if err = controllerutil.SetControllerReference(rs.parentFSM.customResource, muxProtocolSvc, rs.parentFSM.r.scheme); err != nil {
		// Add error detail for use later
	}
	// Call k8s create for service
	if err = rs.parentFSM.r.client.Create(context.TODO(), muxProtocolSvc); err != nil {
		// Add error detail for use later
		rs.stepsComplete |= CreatedMuxProtocolService
	}

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
	if _, err = rs.RetrieveHeadlessService(rs.parentFSM.customResource, rs.parentFSM.namespacedName, rs.parentFSM.r); err != nil {
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
	if _, err = rs.RetrieveStatefulSet(rs.parentFSM.customResource, rs.parentFSM.namespacedName, rs.parentFSM.r); err != nil {
		// err means not found, so mark deleted
		rs.stepsComplete &^= CreatedStatefulSet
	}
}

// newServiceForPod returns an amqbroker service for the pod just created
func newHeadlessServiceForCR(cr *brokerv1alpha1.AMQBroker, servicePorts *[]corev1.ServicePort) *corev1.Service {

	labels := selectors.LabelsForAMQBroker(cr.Name)

	svc := &corev1.Service{
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

	return svc
}
func (rs *CreatingK8sResourcesState) CreateHeadlessService(cr *brokerv1alpha1.AMQBroker, servicePorts *[]corev1.ServicePort) (*corev1.Service, error) {

	// Log where we are and what we're doing
	reqLogger := log.WithValues("AMQBroker Name", cr.Name)
	reqLogger.Info("Creating new " + "headless" + " service")

	headlessSvc := newHeadlessServiceForCR(cr, getDefaultPorts())

	// Define the headless Service for the StatefulSet
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

func (rs *CreatingK8sResourcesState) DeleteHeadlessService(instance *brokerv1alpha1.AMQBroker) {
	// kubectl delete cleans up kubernetes resources, just need to clean up local resources if any
}

//r *ReconcileAMQBroker
func (rs *CreatingK8sResourcesState) RetrieveHeadlessService(instance *brokerv1alpha1.AMQBroker, namespacedName types.NamespacedName, r *ReconcileAMQBroker) (*corev1.Service, error) {

	// Log where we are and what we're doing
	reqLogger := log.WithValues("AMQBroker Name", instance.Name)
	reqLogger.Info("Retrieving " + "headless" + " service")

	var err error = nil
	headlessService := newHeadlessServiceForCR(instance, getDefaultPorts())//&corev1.Service{}

	// Check if the headless service already exists
	if err = r.client.Get(context.TODO(), namespacedName, headlessService); err != nil {
		if errors.IsNotFound(err) {
			reqLogger.Error(err, "Headless service IsNotFound", "Namespace", instance.Namespace, "Name", instance.Name)
		} else {
			reqLogger.Error(err, "Headless service found", "Namespace", instance.Namespace, "Name", instance.Name)
		}
	}

	return headlessService, err
}

func newPersistentVolumeClaimForCR(cr *brokerv1alpha1.AMQBroker) *corev1.PersistentVolumeClaim {

	labels := selectors.LabelsForAMQBroker(cr.Name)

	pvc := &corev1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:		"PersistentVolumeClaim",
		},
		ObjectMeta: metav1.ObjectMeta{
			Annotations: 	nil,
			Labels:			labels,
			Name:			cr.Name,
			Namespace:		cr.Namespace,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: 	[]corev1.PersistentVolumeAccessMode{"ReadWriteOnce"},
			Resources:		corev1.ResourceRequirements{
				Requests:	corev1.ResourceList{
					corev1.ResourceName(corev1.ResourceStorage): resource.MustParse("2Gi"),
				},
			},
		},
	}

	return pvc
}
func newPersistentVolumeClaimArrayForCR(cr *brokerv1alpha1.AMQBroker, arrayLength int) *[]corev1.PersistentVolumeClaim {

	pvcArray := make([]corev1.PersistentVolumeClaim, 0, arrayLength)

	var pvc *corev1.PersistentVolumeClaim = nil
	for i := 0; i < arrayLength; i++ {
		pvc = newPersistentVolumeClaimForCR(cr)
		pvcArray = append(pvcArray, *pvc)
	}

	return &pvcArray
}
func (rs *CreatingK8sResourcesState) CreatePersistentVolumeClaim(cr *brokerv1alpha1.AMQBroker) (*corev1.PersistentVolumeClaim, error) {

	// Log where we are and what we're doing
	reqLogger := log.WithValues("AMQBroker Name", cr.Name)
	reqLogger.Info("Creating new persistent volume claim")

	var err error = nil

	// Define the PersistentVolumeClaim for this Pod
	brokerPvc := newPersistentVolumeClaimForCR(cr)
	// Set AMQBroker instance as the owner and controller
	if err = controllerutil.SetControllerReference(cr, brokerPvc, rs.parentFSM.r.scheme); err != nil {
		// Add error detail for use later
		reqLogger.Info("Failed to set controller reference for new " + "persistent volume claim")
	}
	reqLogger.Info("Set controller reference for new " + "persistent volume claim")

	// Call k8s create for service
	if err = rs.parentFSM.r.client.Create(context.TODO(), brokerPvc); err != nil {
		// Add error detail for use later
		reqLogger.Info("Failed to creating new " + "persistent volume claim")
	}
	reqLogger.Info("Created new " + "persistent volume claim")

	return brokerPvc, err
}

func (rs *CreatingK8sResourcesState) RetrievePersistentVolumeClaim(instance *brokerv1alpha1.AMQBroker, namespacedName types.NamespacedName, r *ReconcileAMQBroker) (*corev1.PersistentVolumeClaim, error) {

	// Log where we are and what we're doing
	reqLogger := log.WithValues("AMQBroker Name", instance.Name)
	reqLogger.Info("Retrieving " + "persistent volume claim")

	var err error = nil
	pvc := &corev1.PersistentVolumeClaim{}

	// Check if the headless service already exists
	if err = r.client.Get(context.TODO(), namespacedName, pvc); err != nil {
		if errors.IsNotFound(err) {
			reqLogger.Error(err, "Persistent volume claim IsNotFound", "Namespace", instance.Namespace, "Name", instance.Name)
		} else {
			reqLogger.Error(err, "Persistent volume claim found", "Namespace", instance.Namespace, "Name", instance.Name)
		}
	}

	return pvc, err
}

// newServiceForPod returns an amqbroker service for the pod just created
func newServiceForCR(cr *brokerv1alpha1.AMQBroker, name_suffix string, port_number int32) *corev1.Service {

	// Log where we are and what we're doing
	reqLogger := log.WithValues("AMQBroker Name", cr.Name)
	reqLogger.Info("Creating new " + name_suffix + " service")

	labels := selectors.LabelsForAMQBroker(cr.Name)

	port := corev1.ServicePort{
		Name:		cr.Name + "-" + name_suffix + "-port",
		Protocol:	"TCP",
		Port:		port_number,
		TargetPort:	intstr.FromInt(int(port_number)),
	}
	ports := []corev1.ServicePort{}
	ports = append(ports, port)

	svc := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:		"Service",
		},
		ObjectMeta: metav1.ObjectMeta {
			Annotations: 	nil,
			Labels: 		labels,
			Name:			cr.Name + "-" + name_suffix + "-service",
			Namespace:		cr.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Type: 	"LoadBalancer",
			Ports: 	ports,
			Selector: labels,
		},
	}

	return svc
}

func getDefaultPorts() *[]corev1.ServicePort {

	ports := []corev1.ServicePort{
		corev1.ServicePort{
			Name:		"mqtt",
			Protocol:	"TCP",
			Port:		1883,
			TargetPort:	intstr.FromInt(int(1883)),
		},
		corev1.ServicePort{
			Name:		"amqp",
			Protocol:	"TCP",
			Port:		5672,
			TargetPort:	intstr.FromInt(int(5672)),
		},
		corev1.ServicePort{
			Name:		"console-jolokia",
			Protocol:	"TCP",
			Port:		8161,
			TargetPort:	intstr.FromInt(int(8161)),
		},
		corev1.ServicePort{
			Name:		"stomp",
			Protocol:	"TCP",
			Port:		61613,
			TargetPort:	intstr.FromInt(int(61613)),
		},
		corev1.ServicePort{
			Name:		"all",
			Protocol:	"TCP",
			Port:		61616,
			TargetPort:	intstr.FromInt(int(61616)),
		},
	}

	return &ports
}

// newPodForCR returns an amqbroker pod with the same name/namespace as the cr
//func newPodForCR(cr *brokerv1alpha1.AMQBroker) *corev1.Pod {
//
//	// Log where we are and what we're doing
//	reqLogger:= log.WithName(cr.Name)
//	reqLogger.Info("Creating new pod for custom resource")
//
//	dataPath := "/opt/" + cr.Name + "/data"
//	userEnvVar := corev1.EnvVar{"AMQ_USER", "admin", nil}
//	passwordEnvVar := corev1.EnvVar{"AMQ_PASSWORD", "admin", nil}
//	dataPathEnvVar := corev1.EnvVar{ "AMQ_DATA_DIR", dataPath, nil}
//
//	pod := &corev1.Pod{
//		ObjectMeta: metav1.ObjectMeta{
//			Name:      cr.Name,
//			Namespace: cr.Namespace,
//			Labels:    cr.Labels,
//		},
//		Spec: corev1.PodSpec{
//			Containers: []corev1.Container{
//				{
//					Name:    cr.Name + "-container",
//					Image:	 cr.Spec.Image,
//					Command: []string{"/opt/amq/bin/launch.sh", "start"},
//					VolumeMounts:	[]corev1.VolumeMount{
//						corev1.VolumeMount{
//							Name:		cr.Name + "-data-volume",
//							MountPath:	dataPath,
//							ReadOnly:	false,
//						},
//					},
//					Env:     []corev1.EnvVar{userEnvVar, passwordEnvVar, dataPathEnvVar},
//				},
//			},
//			Volumes: []corev1.Volume{
//				corev1.Volume{
//					Name: cr.Name + "-data-volume",
//					VolumeSource: corev1.VolumeSource{
//						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
//							ClaimName: cr.Name + "-pvc",
//							ReadOnly:  false,
//						},
//					},
//				},
//			},
//		},
//	}
//
//	return pod
//}

func newPodTemplateSpecForCR(cr *brokerv1alpha1.AMQBroker) corev1.PodTemplateSpec {

	// Log where we are and what we're doing
	reqLogger:= log.WithName(cr.Name)
	reqLogger.Info("Creating new pod template spec for custom resource")

	//var pts corev1.PodTemplateSpec

	dataPath := "/opt/" + cr.Name + "/data"
	userEnvVar := corev1.EnvVar{"AMQ_USER", "admin", nil}
	passwordEnvVar := corev1.EnvVar{"AMQ_PASSWORD", "admin", nil}
	dataPathEnvVar := corev1.EnvVar{ "AMQ_DATA_DIR", dataPath, nil}

	pts := corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
			Labels:    cr.Labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    cr.Name + "-container",
					Image:	 cr.Spec.Image,
					Command: []string{"/opt/amq/bin/launch.sh", "start"},
					VolumeMounts:	[]corev1.VolumeMount{
						{
							Name:		cr.Name,
							MountPath:	dataPath,
							ReadOnly:	false,
						},
					},
					Env:     []corev1.EnvVar{userEnvVar, passwordEnvVar, dataPathEnvVar},
				},
			},
			Volumes: []corev1.Volume{
				{
					Name: cr.Name,
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: cr.Name,
							ReadOnly:  false,
						},
					},
				},
			},
		},
	}

	return pts
}

func newStatefulSetForCR(cr *brokerv1alpha1.AMQBroker) *appsv1.StatefulSet {

	// Log where we are and what we're doing
	reqLogger:= log.WithName(cr.Name)
	reqLogger.Info("Creating new statefulset for custom resource")

	var replicas int32 = 1

	labels := selectors.LabelsForAMQBroker(cr.Name)

	ss := &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:		"StatefulSet",
			APIVersion:	"v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:			cr.Name + "-statefulset",
			Namespace:		cr.Namespace,
			Labels:			labels,
			Annotations:	nil,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: 		&replicas,
			ServiceName:	cr.Name + "-headless" + "-service",
			Selector:		&metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: newPodTemplateSpecForCR(cr),
			VolumeClaimTemplates: *newPersistentVolumeClaimArrayForCR(cr, 1),
		},
	}

	return ss
}
func (rs *CreatingK8sResourcesState) CreateStatefulSet(cr *brokerv1alpha1.AMQBroker) (*appsv1.StatefulSet, error) {

	// Log where we are and what we're doing
	reqLogger := log.WithValues("AMQBroker Name", cr.Name)
	reqLogger.Info("Creating new statefulset")

	var err error = nil

	// Define the StatefulSet
	ss := newStatefulSetForCR(cr)
	// Set AMQBroker instance as the owner and controller
	if err = controllerutil.SetControllerReference(cr, ss, rs.parentFSM.r.scheme); err != nil {
		// Add error detail for use later
		reqLogger.Info("Failed to set controller reference for new " + "statefulset")
	}
	reqLogger.Info("Set controller reference for new " + "statefulset")

	// Call k8s create for statefulset
	if err = rs.parentFSM.r.client.Create(context.TODO(), ss); err != nil {
		// Add error detail for use later
		reqLogger.Info("Failed to creating new " + "statefulset")
	}
	reqLogger.Info("Created new " + "statefulset")

	return ss, err
}
func (rs *CreatingK8sResourcesState) RetrieveStatefulSet(instance *brokerv1alpha1.AMQBroker, namespacedName types.NamespacedName, r *ReconcileAMQBroker) (*appsv1.StatefulSet, error) {

	// Log where we are and what we're doing
	reqLogger := log.WithValues("AMQBroker Name", instance.Name)
	reqLogger.Info("Retrieving " + "statefulset")

	var err error = nil
	ss := &appsv1.StatefulSet{}

	// Check if the headless service already exists
	if err = r.client.Get(context.TODO(), namespacedName, ss); err != nil {
		if errors.IsNotFound(err) {
			reqLogger.Error(err, "Statefulset claim IsNotFound", "Namespace", instance.Namespace, "Name", instance.Name)
		} else {
			reqLogger.Error(err, "Statefulset claim found", "Namespace", instance.Namespace, "Name", instance.Name)
		}
	}

	return ss, err
}

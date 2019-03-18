package services

import (
	"context"
	brokerv1alpha1 "github.com/rh-messaging/amq-broker-operator/pkg/apis/broker/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/rh-messaging/amq-broker-operator/pkg/utils/selectors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("package services")


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

// newServiceForPod returns an amqbroker service for the pod just created
func newServiceForCR(cr *brokerv1alpha1.AMQBroker, name_suffix string, port_number int32) *corev1.Service {

	// Log where we are and what we're doing
	//reqLogger := log.WithValues("AMQBroker Name", cr.Name)
	//reqLogger.Info("Creating new " + name_suffix + " service")

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

func GetDefaultPorts() *[]corev1.ServicePort {

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

func CreateHeadlessService(cr *brokerv1alpha1.AMQBroker, client client.Client, scheme *runtime.Scheme) (*corev1.Service, error) {

	// Log where we are and what we're doing
	reqLogger := log.WithValues("AMQBroker Name", cr.Name)
	reqLogger.Info("Creating new " + "headless" + " service")

	headlessSvc := newHeadlessServiceForCR(cr, GetDefaultPorts())

	// Define the headless Service for the StatefulSet
	// Set AMQBroker instance as the owner and controller
	var err error = nil
	if err = controllerutil.SetControllerReference(cr, headlessSvc, scheme); err != nil {
		// Add error detail for use later
		reqLogger.Info("Failed to set controller reference for new " + "headless" + " service")
	}
	reqLogger.Info("Set controller reference for new " + "headless" + " service")

	// Call k8s create for service
	if err = client.Create(context.TODO(), headlessSvc); err != nil {
		// Add error detail for use later
		reqLogger.Info("Failed to creating new " + "headless" + " service")

	}
	reqLogger.Info("Created new " + "headless" + " service")

	return headlessSvc, err
}

func DeleteHeadlessService(instance *brokerv1alpha1.AMQBroker) {
	// kubectl delete cleans up kubernetes resources, just need to clean up local resources if any
}

//r *ReconcileAMQBroker
func RetrieveHeadlessService(instance *brokerv1alpha1.AMQBroker, namespacedName types.NamespacedName, client client.Client) (*corev1.Service, error) {

	// Log where we are and what we're doing
	reqLogger := log.WithValues("AMQBroker Name", instance.Name)
	reqLogger.Info("Retrieving " + "headless" + " service")

	var err error = nil
	headlessService := newHeadlessServiceForCR(instance, GetDefaultPorts())//&corev1.Service{}

	// Check if the headless service already exists
	if err = client.Get(context.TODO(), namespacedName, headlessService); err != nil {
		if errors.IsNotFound(err) {
			reqLogger.Error(err, "Headless service IsNotFound", "Namespace", instance.Namespace, "Name", instance.Name)
		} else {
			reqLogger.Error(err, "Headless service found", "Namespace", instance.Namespace, "Name", instance.Name)
		}
	}

	return headlessService, err
}


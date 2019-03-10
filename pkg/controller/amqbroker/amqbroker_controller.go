package amqbroker

import (
	"context"
	"github.com/rh-messaging/amq-broker-operator/pkg/utils/selectors"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	brokerv1alpha1 "github.com/rh-messaging/amq-broker-operator/pkg/apis/broker/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_amqbroker")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new AMQBroker Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileAMQBroker{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("amqbroker-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource AMQBroker
	err = c.Watch(&source.Kind{Type: &brokerv1alpha1.AMQBroker{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner AMQBroker
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &brokerv1alpha1.AMQBroker{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileAMQBroker{}

// ReconcileAMQBroker reconciles a AMQBroker object
type ReconcileAMQBroker struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a AMQBroker object and makes changes based on the state read
// and what is in the AMQBroker.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileAMQBroker) Reconcile(request reconcile.Request) (reconcile.Result, error) {

	// Log where we are and what we're doing
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling AMQBroker")

	// Set it up
	var err error = nil
	var reconcileResult reconcile.Result
	instance := &brokerv1alpha1.AMQBroker{}
	found := &corev1.Pod{}

	// Do what's needed
	for {
		// Fetch the AMQBroker instance
		err = r.client.Get(context.TODO(), request.NamespacedName, instance)
		if err != nil {
			// Add error detail for use later
			break
		}

		// Check if this Pod already exists
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: request.Name, Namespace: request.Namespace}, found)
		if err == nil {
			// Don't do anything as the pod exists
			break
		}

		// Pod didn't exist, create
		// Define a new Pod object
		pod := newPodForCR(instance)
		// Set AMQBroker instance as the owner and controller
		err = controllerutil.SetControllerReference(instance, pod, r.scheme)
		if err != nil {
			// Add error detail for use later
			break
		}

		// Was able to set the owner and controller, call k8s create for pod
		err = r.client.Create(context.TODO(), pod)
		if err != nil {
			// Add error detail for use later
			break
		}

		// Pod created successfully - don't requeue
		// Probably redundant
		err = nil

		// Define the console-jolokia Service for this Pod
		consoleJolokiaSvc := newServiceForCR(instance, "console-jolokia", 8161)
		// Call k8s create for service
		err = r.client.Create(context.TODO(), consoleJolokiaSvc)
		if err != nil {
			// Add error detail for use later
			break
		}

		// Define the console-jolokia Service for this Pod
		muxProtocolSvc := newServiceForCR(instance, "mux-protocol", 61616)
		// Call k8s create for service
		err = r.client.Create(context.TODO(), muxProtocolSvc)
		if err != nil {
			// Add error detail for use later
			break
		}

		break
	}

	// Handle error, if any
	if err != nil {
		if errors.IsNotFound(err) {
			reconcileResult = reconcile.Result{}
			reqLogger.Error(err, "AMQBroker Controller Reconcile encountered a IsNotFound, preventing request requeue", "Pod.Namespace", request.Namespace, "Pod.Name", request.Name)
			// Setting err to nil to prevent requeue
			err = nil
		} else {
			//log.Error(err, "AMQBroker Controller Reconcile errored")
			reqLogger.Error(err, "AMQBroker Controller Reconcile errored, requeuing request", "Pod.Namespace", request.Namespace, "Pod.Name", request.Name)
		}
	}

	// Single exit, return the result and error condition
	return reconcileResult, err
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


// newPodForCR returns an amqbroker pod with the same name/namespace as the cr
func newPodForCR(cr *brokerv1alpha1.AMQBroker) *corev1.Pod {

	// Log where we are and what we're doing
	reqLogger:= log.WithName(cr.Name)
	reqLogger.Info("Creating new pod for custom resource")

	userEnvVar := corev1.EnvVar{"AMQ_USER", "admin", nil}
    passwordEnvVar := corev1.EnvVar{"AMQ_PASSWORD", "admin", nil}

    return &corev1.Pod{
        ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
            Namespace: cr.Namespace,
            Labels:    cr.Labels,
        },
        Spec: corev1.PodSpec{
            Containers: []corev1.Container{
                {
                    Name:    "amq",
                    Image:	 cr.Spec.Image,
                    Command: []string{"/opt/amq/bin/launch.sh", "start"},
                    Env:     []corev1.EnvVar{userEnvVar, passwordEnvVar},
                },
            },
        },
    }
}

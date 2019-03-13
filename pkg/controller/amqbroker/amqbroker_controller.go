package amqbroker

import (
	"context"
	"github.com/rh-messaging/amq-broker-operator/pkg/utils/selectors"
	"k8s.io/apimachinery/pkg/api/resource"
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

	appsv1 "k8s.io/api/apps/v1"
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
		if err = r.client.Get(context.TODO(), request.NamespacedName, instance); err != nil {
			// Add error detail for use later
			break
		}

		// Check if this Pod already exists
		if err = r.client.Get(context.TODO(), types.NamespacedName{Name: request.Name, Namespace: request.Namespace}, found); err == nil {
			// Don't do anything as the pod exists
			break
		}

		// Define the PersistentVolumeClaim for this Pod
		brokerPvc := newPersistentVolumeClaimForCR(instance)
		// Set AMQBroker instance as the owner and controller
		if err = controllerutil.SetControllerReference(instance, brokerPvc, r.scheme); err != nil {
			// Add error detail for use later
			break
		}

		// Call k8s create for service
		if err = r.client.Create(context.TODO(), brokerPvc); err != nil {
			// Add error detail for use later
			break
		}

		// Define the console-jolokia Service for this Pod
		consoleJolokiaSvc := newServiceForCR(instance, "console-jolokia", 8161)
		// Set AMQBroker instance as the owner and controller
		if err = controllerutil.SetControllerReference(instance, consoleJolokiaSvc, r.scheme); err != nil {
			// Add error detail for use later
			break
		}
		// Call k8s create for service
		if err = r.client.Create(context.TODO(), consoleJolokiaSvc); err != nil {
			// Add error detail for use later
			break
		}

		// Define the console-jolokia Service for this Pod
		muxProtocolSvc := newServiceForCR(instance, "mux-protocol", 61616)
		// Set AMQBroker instance as the owner and controller
		if err = controllerutil.SetControllerReference(instance, muxProtocolSvc, r.scheme); err != nil {
			// Add error detail for use later
			break
		}
		// Call k8s create for service
		if err = r.client.Create(context.TODO(), muxProtocolSvc); err != nil {
			// Add error detail for use later
			break
		}

		// Define the headless Service for the StatefulSet
		//headlessSvc := newHeadlessServiceForCR(instance, 5672)
		headlessSvc := newHeadlessServiceForCR(instance, getDefaultPorts())
		// Set AMQBroker instance as the owner and controller
		if err = controllerutil.SetControllerReference(instance, headlessSvc, r.scheme); err != nil {
			// Add error detail for use later
			break
		}
		// Call k8s create for service
		if err = r.client.Create(context.TODO(), headlessSvc); err != nil {
			// Add error detail for use later
			break
		}


		//// Pod didn't exist, create
		//// Define a new Pod object
		//pod := newPodForCR(instance)
		//// Set AMQBroker instance as the owner and controller
		//if err = controllerutil.SetControllerReference(instance, pod, r.scheme); err != nil {
		//	// Add error detail for use later
		//	break
		//}
		//// Was able to set the owner and controller, call k8s create for pod
		//if err = r.client.Create(context.TODO(), pod); err != nil {
		//	// Add error detail for use later
		//	break
		//}

		// Statefulset didn't exist, create
		// Define a new Pod object
		ss := newStatefulSetForCR(instance)
		// Set AMQBroker instance as the owner and controller
		if err = controllerutil.SetControllerReference(instance, ss, r.scheme); err != nil {
			// Add error detail for use later
			break
		}
		// Was able to set the owner and controller, call k8s create for pod
		if err = r.client.Create(context.TODO(), ss); err != nil {
			// Add error detail for use later
			break
		}


		// Pod created successfully - don't requeue
		// Probably redundant
		err = nil


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
			//Template: corev1.PodTemplateSpec{},
			Template: newPodTemplateSpecForCR(cr),
		},
	}

	return ss
}

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
						corev1.VolumeMount{
							Name:		cr.Name + "-data-volume",
							MountPath:	dataPath,
							ReadOnly:	false,
						},
					},
					Env:     []corev1.EnvVar{userEnvVar, passwordEnvVar, dataPathEnvVar},
				},
			},
			Volumes: []corev1.Volume{
				corev1.Volume{
					Name: cr.Name + "-data-volume",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: cr.Name + "-pvc",
							ReadOnly:  false,
						},
					},
				},
			},
		},
	}

	return pts
}

func newPersistentVolumeClaimForCR(cr *brokerv1alpha1.AMQBroker) *corev1.PersistentVolumeClaim {

	// Log where we are and what we're doing
	reqLogger := log.WithValues("AMQBroker Name", cr.Name)
	reqLogger.Info("Creating new persistent volume claim")

	labels := selectors.LabelsForAMQBroker(cr.Name)

	pvc := &corev1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:		"PersistentVolumeClaim",
		},
		ObjectMeta: metav1.ObjectMeta{
			Annotations: 	nil,
			Labels:			labels,
			Name:			cr.Name + "-pvc",
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

// newServiceForPod returns an amqbroker service for the pod just created
func newHeadlessServiceForCR(cr *brokerv1alpha1.AMQBroker, servicePorts *[]corev1.ServicePort) *corev1.Service {

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

// newPodForCR returns an amqbroker pod with the same name/namespace as the cr
func newPodForCR(cr *brokerv1alpha1.AMQBroker) *corev1.Pod {

	// Log where we are and what we're doing
	reqLogger:= log.WithName(cr.Name)
	reqLogger.Info("Creating new pod for custom resource")

	dataPath := "/opt/" + cr.Name + "/data"
	userEnvVar := corev1.EnvVar{"AMQ_USER", "admin", nil}
    passwordEnvVar := corev1.EnvVar{"AMQ_PASSWORD", "admin", nil}
    dataPathEnvVar := corev1.EnvVar{ "AMQ_DATA_DIR", dataPath, nil}

    pod := &corev1.Pod{
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
                    	corev1.VolumeMount{
							Name:		cr.Name + "-data-volume",
                    		MountPath:	dataPath,
                    		ReadOnly:	false,
						},
					},
                    Env:     []corev1.EnvVar{userEnvVar, passwordEnvVar, dataPathEnvVar},
                },
            },
            Volumes: []corev1.Volume{
            	corev1.Volume{
            		Name: cr.Name + "-data-volume",
            		VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: cr.Name + "-pvc",
							ReadOnly:  false,
						},
					},
				},
			},
        },
    }

    return pod
}

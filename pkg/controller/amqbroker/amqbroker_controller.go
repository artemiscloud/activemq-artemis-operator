package amqbroker

import (
	"context"
	"k8s.io/apimachinery/pkg/api/errors"

	//"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	brokerv1alpha1 "github.com/rh-messaging/activemq-artemis-operator/pkg/apis/broker/v1alpha1"

	corev1 "k8s.io/api/core/v1"

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
//var namespacedNameToFSM map[types.NamespacedName]*fsm.Machine
var namespacedNameToFSM = make(map[types.NamespacedName]*ActiveMQArtemisFSM)

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new ActiveMQArtemis Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileActiveMQArtemis{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("amqbroker-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource ActiveMQArtemis
	err = c.Watch(&source.Kind{Type: &brokerv1alpha1.ActiveMQArtemis{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner ActiveMQArtemis
	err = c.Watch(&source.Kind{Type: &appsv1.StatefulSet{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &brokerv1alpha1.ActiveMQArtemis{},
	})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource Pods and requeue the owner ActiveMQArtemis
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &brokerv1alpha1.ActiveMQArtemis{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileActiveMQArtemis{}

// ReconcileActiveMQArtemis reconciles a ActiveMQArtemis object
type ReconcileActiveMQArtemis struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a ActiveMQArtemis object and makes changes based on the state read
// and what is in the ActiveMQArtemis.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileActiveMQArtemis) Reconcile(request reconcile.Request) (reconcile.Result, error) {

	// Log where we are and what we're doing
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling ActiveMQArtemis")

	var err error = nil
	var reconcileResult reconcile.Result
	var namespacedNameFSM *ActiveMQArtemisFSM = nil
	var amqbfsm *ActiveMQArtemisFSM = nil

	instance := &brokerv1alpha1.ActiveMQArtemis{}
	namespacedName := types.NamespacedName{
		Name: request.Name,//"example-amqbroker",
		Namespace: request.Namespace,//"abo-1",
	}

	// Fetch the ActiveMQArtemis instance
	// When first creating this will have err == nil
	// When deleting after creation this will have err NotFound
	// When deleting before creation reconcile won't be called
	if err = r.client.Get(context.TODO(), request.NamespacedName, instance); err != nil {

		if errors.IsNotFound(err) {
			reqLogger.Error(err, "ActiveMQArtemis Controller Reconcile encountered a IsNotFound, checking to see if we should delete namespacedName tracking", "request.Namespace", request.Namespace, "request.Name", request.Name)

			reconcileResult = reconcile.Result{}

			// See if we have been tracking this NamespacedName
			if namespacedNameFSM = namespacedNameToFSM[namespacedName]; namespacedNameFSM != nil {
				reqLogger.Error(err, "Removing namespacedName tracking", "request.Namespace", request.Namespace, "request.Name", request.Name)
				// If so we should no longer track it
				//namespacedNameToFSM[namespacedName] = nil
				amqbfsm = namespacedNameToFSM[namespacedName]
				amqbfsm.Exit(nil)
				delete(namespacedNameToFSM, namespacedName)
				amqbfsm = nil
			}

			// Setting err to nil to prevent requeue
			err = nil
		} else {
			reqLogger.Error(err, "ActiveMQArtemis Controller Reconcile errored thats not IsNotFound, requeuing request", "Request Namespace", request.Namespace, "Request Name", request.Name)
			// Leaving err as !nil causes requeue
		}

		// Add error detail for use later
		return reconcileResult, err
	}

	// Do lookup to see if we have a fsm for the incoming name in the incoming namespace
	// if not, create it
	// for the given fsm, do an update
	// - update first level sets? what if the operator has gone away and come back? stateless?
	if namespacedNameFSM = namespacedNameToFSM[namespacedName]; namespacedNameFSM == nil {
		amqbfsm = NewActiveMQArtemisFSM(instance, namespacedName, r)
		namespacedNameFSM = amqbfsm
		namespacedNameToFSM[namespacedName] = namespacedNameFSM

		amqbfsm.Enter(nil)
	}

	//namespacedNameFSM.Update(request, r)

	// Set it up
	//var err error = nil
	//var reconcileResult reconcile.Result
	//instance := &brokerv1alpha1.ActiveMQArtemis{}
	//found := &corev1.Pod{}
	//
	//// Do what's needed
	//for {
	//	// Fetch the ActiveMQArtemis instance
	//	if err = r.client.Get(context.TODO(), request.NamespacedName, instance); err != nil {
	//		// Add error detail for use later
	//		break
	//	}
	//
	//	// Check if this Pod already exists
	//	if err = r.client.Get(context.TODO(), namespacedName, found); err == nil {
	//		// Don't do anything as the pod exists
	//		break
	//	}
	//
	//	// Define the PersistentVolumeClaim for this Pod
	//	brokerPvc := newPersistentVolumeClaimForCR(instance)
	//	// Set ActiveMQArtemis instance as the owner and controller
	//	if err = controllerutil.SetControllerReference(instance, brokerPvc, r.scheme); err != nil {
	//		// Add error detail for use later
	//		break
	//	}
	//
	//	// Call k8s create for service
	//	if err = r.client.Create(context.TODO(), brokerPvc); err != nil {
	//		// Add error detail for use later
	//		break
	//	}
	//
	//	// Define the console-jolokia Service for this Pod
	//	consoleJolokiaSvc := newServiceForCR(instance, "console-jolokia", 8161)
	//	// Set ActiveMQArtemis instance as the owner and controller
	//	if err = controllerutil.SetControllerReference(instance, consoleJolokiaSvc, r.scheme); err != nil {
	//		// Add error detail for use later
	//		break
	//	}
	//	// Call k8s create for service
	//	if err = r.client.Create(context.TODO(), consoleJolokiaSvc); err != nil {
	//		// Add error detail for use later
	//		break
	//	}
	//
	//	// Define the console-jolokia Service for this Pod
	//	muxProtocolSvc := newServiceForCR(instance, "mux-protocol", 61616)
	//	// Set ActiveMQArtemis instance as the owner and controller
	//	if err = controllerutil.SetControllerReference(instance, muxProtocolSvc, r.scheme); err != nil {
	//		// Add error detail for use later
	//		break
	//	}
	//	// Call k8s create for service
	//	if err = r.client.Create(context.TODO(), muxProtocolSvc); err != nil {
	//		// Add error detail for use later
	//		break
	//	}
	//
	//	// Define the headless Service for the StatefulSet
	//	//headlessSvc := newHeadlessServiceForCR(instance, 5672)
	//	headlessSvc := newHeadlessServiceForCR(instance, getDefaultPorts())
	//	// Set ActiveMQArtemis instance as the owner and controller
	//	if err = controllerutil.SetControllerReference(instance, headlessSvc, r.scheme); err != nil {
	//		// Add error detail for use later
	//		break
	//	}
	//	// Call k8s create for service
	//	if err = r.client.Create(context.TODO(), headlessSvc); err != nil {
	//		// Add error detail for use later
	//		break
	//	}
	//
	//
	//	//// Pod didn't exist, create
	//	//// Define a new Pod object
	//	//pod := newPodForCR(instance)
	//	//// Set ActiveMQArtemis instance as the owner and controller
	//	//if err = controllerutil.SetControllerReference(instance, pod, r.scheme); err != nil {
	//	//	// Add error detail for use later
	//	//	break
	//	//}
	//	//// Was able to set the owner and controller, call k8s create for pod
	//	//if err = r.client.Create(context.TODO(), pod); err != nil {
	//	//	// Add error detail for use later
	//	//	break
	//	//}
	//
	//	// Statefulset didn't exist, create
	//	// Define a new Pod object
	//	ss := newStatefulSetForCR(instance)
	//	// Set ActiveMQArtemis instance as the owner and controller
	//	if err = controllerutil.SetControllerReference(instance, ss, r.scheme); err != nil {
	//		// Add error detail for use later
	//		break
	//	}
	//	// Was able to set the owner and controller, call k8s create for pod
	//	if err = r.client.Create(context.TODO(), ss); err != nil {
	//		// Add error detail for use later
	//		break
	//	}
	//
	//
	//	// Pod created successfully - don't requeue
	//	// Probably redundant
	//	err = nil
	//
	//
	//	break
	//}
	//
	//// Handle error, if any
	//if err != nil {
	//	if errors.IsNotFound(err) {
	//		reconcileResult = reconcile.Result{}
	//		reqLogger.Error(err, "ActiveMQArtemis Controller Reconcile encountered a IsNotFound, preventing request requeue", "Pod.Namespace", request.Namespace, "Pod.Name", request.Name)
	//		// Setting err to nil to prevent requeue
	//		err = nil
	//	} else {
	//		//log.Error(err, "ActiveMQArtemis Controller Reconcile errored")
	//		reqLogger.Error(err, "ActiveMQArtemis Controller Reconcile errored, requeuing request", "Pod.Namespace", request.Namespace, "Pod.Name", request.Name)
	//	}
	//}

	// Single exit, return the result and error condition
	return reconcileResult, err
}

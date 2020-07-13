package resources

import (
	"context"
	brokerv2alpha1 "github.com/rh-messaging/activemq-artemis-operator/pkg/apis/broker/v2alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("package k8s_actions")

//func Create(cr *brokerv2alpha1.ActiveMQArtemis, client client.Client, scheme *runtime.Scheme, objectDefinition runtime.Object) error {
func Create(owner v1.Object, namespacedName types.NamespacedName, client client.Client, scheme *runtime.Scheme, objectDefinition runtime.Object) error {

	// Log where we are and what we're doing
	reqLogger := log.WithValues("ActiveMQArtemis Name", namespacedName.Name)
	objectTypeString := reflect.TypeOf(objectDefinition).String()
	reqLogger.Info("Creating new " + objectTypeString)

	// Define the headless Service for the StatefulSet
	// Set ActiveMQArtemis instance as the owner and controller
	var err error = nil
	if err = controllerutil.SetControllerReference(owner, objectDefinition.(v1.Object), scheme); err != nil {
		// Add error detail for use later
		reqLogger.Info("Failed to set controller reference for new " + objectTypeString)
	}
	reqLogger.Info("Set controller reference for new " + objectTypeString)

	// Call k8s create for service
	if err = client.Create(context.TODO(), objectDefinition); err != nil {
		// Add error detail for use later
		reqLogger.Info("Failed to creating new " + objectTypeString)
	}
	reqLogger.Info("Created new " + objectTypeString)

	return err
}

// TODO: Evaluate performance impact of using reflect
//func Retrieve(cr *brokerv2alpha1.ActiveMQArtemis, namespacedName types.NamespacedName, client client.Client, objectDefinition runtime.Object) error {
func Retrieve(namespacedName types.NamespacedName, client client.Client, objectDefinition runtime.Object) error {

	// Log where we are and what we're doing
	reqLogger := log.WithValues("ActiveMQArtemis Name", namespacedName.Name)
	objectTypeString := reflect.TypeOf(objectDefinition).String()
	reqLogger.Info("Retrieving " + objectTypeString)

	var err error = nil
	if err = client.Get(context.TODO(), namespacedName, objectDefinition); err != nil {
		if errors.IsNotFound(err) || runtime.IsNotRegisteredError(err) {
			//reqLogger.Info(objectTypeString+" IsNotFound", "Namespace", cr.Namespace, "Name", cr.Name)
			reqLogger.Info(objectTypeString+" IsNotFound", "Namespace", namespacedName.Namespace, "Name", namespacedName.Name)
		} else {
			reqLogger.Info(objectTypeString+" found", "Namespace", namespacedName.Namespace, "Name", namespacedName.Name)
		}
	}

	return err
}


func Update(namespacedName types.NamespacedName, client client.Client, objectDefinition runtime.Object) error {

	reqLogger := log.WithValues("ActiveMQArtemis Name", namespacedName.Name)
	objectTypeString := reflect.TypeOf(objectDefinition).String()
	reqLogger.Info("Updating " + objectTypeString)

	var err error = nil
	if err = client.Update(context.TODO(), objectDefinition); err != nil {
		reqLogger.Error(err, "Failed to update "+objectTypeString)
	}

	return err
}

func Delete(namespacedName types.NamespacedName, client client.Client, objectDefinition runtime.Object) error {

	reqLogger := log.WithValues("ActiveMQArtemis Name", namespacedName.Name)
	objectTypeString := reflect.TypeOf(objectDefinition).String()
	reqLogger.Info("Deleting " + objectTypeString)

	var err error = nil
	if err = client.Delete(context.TODO(), objectDefinition); err != nil {
		reqLogger.Error(err, "Failed to delete "+objectTypeString)
	}

	return err
}

func Enable(customResource *brokerv2alpha1.ActiveMQArtemis, client client.Client, scheme *runtime.Scheme, namespacedName types.NamespacedName, objectDefinition runtime.Object) (bool, error) {
	causedUpdate, err := configureExposure(customResource, client, scheme, namespacedName, objectDefinition, true)
	return causedUpdate, err
}

func Disable(customResource *brokerv2alpha1.ActiveMQArtemis, client client.Client, scheme *runtime.Scheme, namespacedName types.NamespacedName, objectDefinition runtime.Object) (bool, error) {
	causedUpdate, err := configureExposure(customResource, client, scheme, namespacedName, objectDefinition, false)
	return causedUpdate, err
}

func configureExposure(owner v1.Object, client client.Client, scheme *runtime.Scheme, namespacedName types.NamespacedName, objectDefinition runtime.Object, enable bool) (bool, error) {

	// Log where we are and what we're doing
	reqLogger := log.WithValues("ActiveMQArtemis Name ", namespacedName.Name)
	objectTypeString := reflect.TypeOf(objectDefinition).String()
	reqLogger.Info("Configuring " + objectTypeString)

	var err error = nil
	serviceIsNotFound := false
	causedUpdate := true

	//if err = Retrieve(customResource, namespacedName, client, objectDefinition); err != nil {
	if err = Retrieve(namespacedName, client, objectDefinition); err != nil {
		if errors.IsNotFound(err) {
			reqLogger.Info(namespacedName.Name + " " + "not found")
			serviceIsNotFound = true
		}
	}

	// We want a service to be exposed and currently it is not found
	if enable && serviceIsNotFound {
		reqLogger.Info("Creating " + namespacedName.Name)
		if err = Create(owner, namespacedName, client, scheme, objectDefinition); err != nil {
			//reqLogger.Info("Failure to create " + baseServiceName + " service " + ordinalString)
			causedUpdate = true
		}
	}

	// We do NOT want a service to be exposed and the service IS found
	if !enable && !serviceIsNotFound {
		reqLogger.Info("Deleting " + namespacedName.Name)
		if err = Delete(namespacedName, client, objectDefinition); err != nil {
			//reqLogger.Info("Failure to delete " + baseServiceName + " service " + ordinalString)
			causedUpdate = true
		}
	}

	return causedUpdate, err
}

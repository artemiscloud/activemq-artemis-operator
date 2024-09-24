/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"time"

	brokerv1beta1 "github.com/artemiscloud/activemq-artemis-operator/api/v1beta1"
	"github.com/artemiscloud/activemq-artemis-operator/pkg/draincontroller"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var StopCh chan struct{}

var controllers map[string]*draincontroller.Controller = make(map[string]*draincontroller.Controller)

type ActiveMQArtemisMessageMigrationControl struct {
	client.Client
	Scheme    *runtime.Scheme
	Config    *rest.Config
	log       logr.Logger
	localOnly bool
}

func NewActiveMQArtemisMessageMigrationControl(client client.Client, scheme *runtime.Scheme, config *rest.Config, logger logr.Logger, localOnly bool) *ActiveMQArtemisMessageMigrationControl {
	return &ActiveMQArtemisMessageMigrationControl{
		Client:    client,
		Scheme:    scheme,
		Config:    config,
		log:       logger,
		localOnly: localOnly,
	}
}

func (mm *ActiveMQArtemisMessageMigrationControl) Setup(brokerCr *brokerv1beta1.ActiveMQArtemis, params map[string]string) error {

	reqLogger := mm.log.WithValues("Setup", brokerCr.Name, "Namespace", brokerCr.Namespace)

	kubeClient, err := kubernetes.NewForConfig(mm.Config)
	if err != nil {
		reqLogger.Error(err, "Error building kubernetes clientset")
		return err
	}

	kubeInformerFactory, drainControllerInstance, isNewController := mm.getDrainController(kubeClient, brokerCr, params)

	if isNewController {
		reqLogger.V(2).Info("Starting async factory...")
		go kubeInformerFactory.Start(*drainControllerInstance.GetStopCh())

		reqLogger.V(2).Info("Running drain controller async so multiple controllers can run...")
		go mm.runDrainController(drainControllerInstance)
	}

	reqLogger.V(2).Info("Message Migration is set up")
	return nil
}

func (mm *ActiveMQArtemisMessageMigrationControl) Cancel(brokerNs *types.NamespacedName) {
	reqLogger := mm.log.WithValues("Cancel", brokerNs.Name, "Namespace", brokerNs.Namespace)
	if drainer := mm.findDrainerController(brokerNs); drainer != nil {
		reqLogger.V(1).Info("Cancel message migration")
		drainer.RemoveInstance(brokerNs)
	}
}

func (mm *ActiveMQArtemisMessageMigrationControl) findDrainerController(brokerNs *types.NamespacedName) *draincontroller.Controller {
	controllerKey := mm.getControllerKey(brokerNs.Namespace)
	if inst, ok := controllers[controllerKey]; ok {
		return inst
	}
	return nil
}

func (mm *ActiveMQArtemisMessageMigrationControl) getControllerKey(ns string) string {
	controllerKey := "*"
	if mm.localOnly {
		controllerKey = ns
	}
	return controllerKey
}

func (mm *ActiveMQArtemisMessageMigrationControl) getDrainController(kubeClient *kubernetes.Clientset, instance *brokerv1beta1.ActiveMQArtemis, params map[string]string) (kubeinformers.SharedInformerFactory, *draincontroller.Controller, bool) {
	var kubeInformerFactory kubeinformers.SharedInformerFactory
	var controllerInstance *draincontroller.Controller

	controllerKey := mm.getControllerKey(instance.Namespace)

	if inst, ok := controllers[controllerKey]; ok {
		mm.log.V(2).Info("Drain controller already exists", "namespace", controllerKey)
		inst.AddInstance(instance, params)
		return nil, nil, false
	}

	if mm.localOnly {
		// localOnly means there is only one target namespace and it is the same as operator's
		mm.log.V(2).Info("getting localOnly informer factory", "namespace", controllerKey)
		kubeInformerFactory = kubeinformers.NewSharedInformerFactoryWithOptions(kubeClient, time.Second*30, kubeinformers.WithNamespace(instance.Namespace))
	} else {
		mm.log.V(2).Info("Creating informer factory to operate on StatefulSets across all namespaces")
		kubeInformerFactory = kubeinformers.NewSharedInformerFactory(kubeClient, time.Second*30)
	}

	mm.log.V(2).Info("new drain controller...", "labels", instance.Labels)
	controllerInstance = draincontroller.NewController(controllerKey, instance, kubeClient, kubeInformerFactory, instance.Namespace, mm.Client, mm.log, mm.localOnly)
	controllers[controllerKey] = controllerInstance

	mm.log.V(2).Info("Adding scaledown instance to controller", "controller", controllerInstance, "scaledown", instance)
	controllerInstance.AddInstance(instance, params)

	return kubeInformerFactory, controllerInstance, true
}

func (mm *ActiveMQArtemisMessageMigrationControl) runDrainController(controller *draincontroller.Controller) {
	if err := controller.Run(1); err != nil {
		mm.log.Error(err, "Error running controller")
	}
}

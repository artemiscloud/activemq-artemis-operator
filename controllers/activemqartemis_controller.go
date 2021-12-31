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
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	brokerv1beta1 "github.com/artemiscloud/activemq-artemis-operator/api/v1beta1"
)

var glog = ctrl.Log.WithName("controller_v1beta1activemqartemis")

var namespacedNameToFSM = make(map[types.NamespacedName]*ActiveMQArtemisFSM)

type ActiveMQArtemisConfigHandler interface {
	IsApplicableFor(brokerNamespacedName types.NamespacedName) bool
	Config(initContainers []corev1.Container, outputDirRoot string, yacfgProfileVersion string, yacfgProfileName string) (value []string)
}

var namespaceToConfigHandler = make(map[types.NamespacedName]ActiveMQArtemisConfigHandler)

func GetBrokerConfigHandler(brokerNamespacedName types.NamespacedName) (handler ActiveMQArtemisConfigHandler) {
	for _, handler := range namespaceToConfigHandler {
		if handler.IsApplicableFor(brokerNamespacedName) {
			return handler
		}
	}
	return nil
}

func UpdatePodForSecurity(securityHandlerNamespacedName types.NamespacedName, handler ActiveMQArtemisConfigHandler) error {
	success := true
	for nsn, fsm := range namespacedNameToFSM {
		if handler.IsApplicableFor(nsn) {
			fsm.SetPodInvalid(true)
			glog.Info("Need update fsm for security", "fsm", nsn)
			if err, _ := fsm.Update(); err != nil {
				success = false
				glog.Error(err, "error in updating security", "cr", fsm.namespacedName)
			}
		}
	}
	if success {
		return nil
	}
	err := fmt.Errorf("error in update security, please see log for details")
	return err
}

func RemoveBrokerConfigHandler(namespacedName types.NamespacedName) {
	glog.Info("Removing config handler", "name", namespacedName)
	oldHandler, ok := namespaceToConfigHandler[namespacedName]
	if ok {
		delete(namespaceToConfigHandler, namespacedName)
		glog.Info("Handler removed, updating fsm if exists")
		UpdatePodForSecurity(namespacedName, oldHandler)
	}
}

func AddBrokerConfigHandler(namespacedName types.NamespacedName, handler ActiveMQArtemisConfigHandler, toReconcile bool) error {
	if _, ok := namespaceToConfigHandler[namespacedName]; ok {
		glog.V(1).Info("There is an old config handler, it'll be replaced")
	}
	namespaceToConfigHandler[namespacedName] = handler
	glog.V(1).Info("A new config handler has been added", "handler", handler)
	if toReconcile {
		glog.V(1).Info("Updating broker security")
		return UpdatePodForSecurity(namespacedName, handler)
	}
	return nil
}

// ActiveMQArtemisReconciler reconciles a ActiveMQArtemis object
type ActiveMQArtemisReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Result ctrl.Result
}

//+kubebuilder:rbac:groups=broker.amq.io,resources=activemqartemises,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=broker.amq.io,resources=activemqartemises/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=broker.amq.io,resources=activemqartemises/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ActiveMQArtemis object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *ActiveMQArtemisReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// your logic here

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ActiveMQArtemisReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&brokerv1beta1.ActiveMQArtemis{}).
		Owns(&appsv1.StatefulSet{}).
		Owns(&corev1.Pod{}).
		Complete(r)
}

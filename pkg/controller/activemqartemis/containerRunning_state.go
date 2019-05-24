package activemqartemis


import (
	"context"
	"k8s.io/apimachinery/pkg/api/errors"
	"github.com/rh-messaging/activemq-artemis-operator/pkg/utils/fsm"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	appsv1 "k8s.io/api/apps/v1"
)

type ContainerRunningState struct {
	s 				fsm.State
	namespacedName	types.NamespacedName
	parentFSM		*ActiveMQArtemisFSM
}

func MakeContainerRunningState(_parentFSM *ActiveMQArtemisFSM, _namespacedName types.NamespacedName) ContainerRunningState {

	rs := ContainerRunningState{
		s:					fsm.MakeState(ContainerRunning),
		namespacedName: 	_namespacedName,
		parentFSM: 	_parentFSM,
	}

	return rs
}

func NewContainerRunningState(_parentFSM *ActiveMQArtemisFSM, _namespacedName types.NamespacedName) *ContainerRunningState {

	rs := MakeContainerRunningState(_parentFSM, _namespacedName)

	return &rs
}

func (rs *ContainerRunningState) Enter(stateFrom *fsm.IState) {

	// Log where we are and what we're doing
	reqLogger := log.WithValues("ActiveMQArtemis Name", rs.parentFSM.customResource.Name)
	reqLogger.Info("Entering ContainerRunningState")

}

//func (rs *ContainerRunningState) Update() {
func (rs *ContainerRunningState) Update() error { // signature mismatch with error return

	// Log where we are and what we're doing
	reqLogger := log.WithValues("ActiveMQArtemis Name", rs.parentFSM.customResource.Name)
	reqLogger.Info("Updating ContainerRunningState")

	found := &appsv1.StatefulSet{}
	err := rs.parentFSM.r.client.Get(context.TODO(), types.NamespacedName{Name: rs.parentFSM.customResource.Name, Namespace: rs.parentFSM.customResource.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		//return reconcile.Result{Requeue: true}, nil
		return nil
	}

	// Ensure the StatefulSet size is the same as the spec
	size := rs.parentFSM.customResource.Spec.Size
	if *found.Spec.Replicas != size {
		found.Spec.Replicas = &size
		err = rs.parentFSM.r.client.Update(context.TODO(), found)
		if err != nil {
			reqLogger.Error(err, "Failed to update StatefulSet.", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
			//return reconcile.Result{}, err
			return err
		}
		// Spec updated - return and requeue
		//return reconcile.Result{Requeue: true}, nil
		return nil
	}

	return nil
}

func (rs *ContainerRunningState) Exit(stateTo *fsm.IState) {

	// Log where we are and what we're doing
	reqLogger := log.WithValues("ActiveMQArtemis Name", rs.parentFSM.customResource.Name)
	reqLogger.Info("Exiting ContainerRunningState")

}
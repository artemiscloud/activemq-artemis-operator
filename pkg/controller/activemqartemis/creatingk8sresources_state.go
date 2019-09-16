package activemqartemis

import (
	"context"
	"github.com/rh-messaging/activemq-artemis-operator/pkg/resources"
	"github.com/rh-messaging/activemq-artemis-operator/pkg/resources/environments"
	"github.com/rh-messaging/activemq-artemis-operator/pkg/resources/pods"
	"github.com/rh-messaging/activemq-artemis-operator/pkg/resources/secrets"
	"github.com/rh-messaging/activemq-artemis-operator/pkg/resources/serviceports"
	svc "github.com/rh-messaging/activemq-artemis-operator/pkg/resources/services"
	ss "github.com/rh-messaging/activemq-artemis-operator/pkg/resources/statefulsets"
	"github.com/rh-messaging/activemq-artemis-operator/pkg/resources/volumes"
	"github.com/rh-messaging/activemq-artemis-operator/pkg/utils/fsm"
	"github.com/rh-messaging/activemq-artemis-operator/pkg/utils/random"
	"github.com/rh-messaging/activemq-artemis-operator/pkg/utils/selectors"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"strconv"
	"time"
)

var reconciler = ActiveMQArtemisReconciler{
	statefulSetUpdates: 0,
}

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

func (rs *CreatingK8sResourcesState) generateNames() {

	// Initialize the kubernetes names
	ss.NameBuilder.Base(rs.parentFSM.customResource.Name).Suffix("ss").Generate()
	svc.HeadlessNameBuilder.Prefix(rs.parentFSM.customResource.Name).Base("hdls").Suffix("svc").Generate()
	svc.PingNameBuilder.Prefix(rs.parentFSM.customResource.Name).Base("ping").Suffix("svc").Generate()
	pods.NameBuilder.Base(rs.parentFSM.customResource.Name).Suffix("container").Generate()
	secrets.CredentialsNameBuilder.Prefix(rs.parentFSM.customResource.Name).Base("credentials").Suffix("secret").Generate()
	secrets.ConsoleNameBuilder.Prefix(rs.parentFSM.customResource.Name).Base("console").Suffix("secret").Generate()
	secrets.NettyNameBuilder.Prefix(rs.parentFSM.customResource.Name).Base("netty").Suffix("secret").Generate()
}

func (rs *CreatingK8sResourcesState) generateSecrets() {

	credentialsSecretName := secrets.CredentialsNameBuilder.Name()
	namespacedName := types.NamespacedName{
		Name:      credentialsSecretName,
		Namespace: rs.namespacedName.Namespace,
	}

	clusterUser := random.GenerateRandomString(8)
	clusterPassword := random.GenerateRandomString(8)
	adminUser := ""
	adminPassword := ""

	if "" == rs.parentFSM.customResource.Spec.AdminUser {
		adminUser = environments.Defaults.AMQ_USER
	} else {
		adminUser = rs.parentFSM.customResource.Spec.AdminUser
	}
	if "" == rs.parentFSM.customResource.Spec.AdminPassword {
		adminPassword = environments.Defaults.AMQ_PASSWORD
	} else {
		adminPassword = rs.parentFSM.customResource.Spec.AdminPassword
	}

	stringDataMap := map[string]string{
		"clusterUser":     clusterUser,
		"clusterPassword": clusterPassword,
		"AMQ_USER":        adminUser,
		"AMQ_PASSWORD":    adminPassword,
	}
	// TODO: Remove this hack
	environments.GLOBAL_AMQ_CLUSTER_USER = clusterUser
	environments.GLOBAL_AMQ_CLUSTER_PASSWORD = clusterPassword

	secrets.Create(rs.parentFSM.customResource, namespacedName, stringDataMap, rs.parentFSM.r.client, rs.parentFSM.r.scheme)
}

// First time entering state
func (rs *CreatingK8sResourcesState) enterFromInvalidState() error {

	var err error = nil
	var retrieveError error = nil

	rs.generateNames()
	selectors.LabelBuilder.Base(rs.parentFSM.customResource.Name).Suffix("app").Generate()
	rs.generateSecrets()

	// TODO: Ensure consistent path usage; remove this hack
	volumes.GLOBAL_DATA_PATH = environments.GetPropertyForCR("AMQ_DATA_DIR", rs.parentFSM.customResource, "/opt/"+rs.parentFSM.customResource.Name+"/data")

	statefulsetDefinition := ss.NewStatefulSetForCR(rs.parentFSM.customResource)
	_ = reconciler.Process(rs.parentFSM.customResource, rs.parentFSM.r.client, rs.parentFSM.r.scheme, statefulsetDefinition)
	// Check to see if the statefulset already exists
	if err := resources.Retrieve(rs.parentFSM.customResource, rs.parentFSM.namespacedName, rs.parentFSM.r.client, statefulsetDefinition); err != nil {
		// err means not found, so create
		if retrieveError = resources.Create(rs.parentFSM.customResource, rs.parentFSM.r.client, rs.parentFSM.r.scheme, statefulsetDefinition); retrieveError == nil {
			rs.stepsComplete |= CreatedStatefulSet

			//TODO: Remove this blatant hack
			ss.GLOBAL_CRNAME = rs.parentFSM.customResource.Name
		}
	}

	headlessServiceDefinition := svc.NewHeadlessServiceForCR(rs.parentFSM.customResource, serviceports.GetDefaultPorts(rs.parentFSM.customResource))
	// Check to see if the headless service already exists
	if err = resources.Retrieve(rs.parentFSM.customResource, rs.parentFSM.namespacedName, rs.parentFSM.r.client, headlessServiceDefinition); err != nil {
		// err means not found, so create
		if retrieveError = resources.Create(rs.parentFSM.customResource, rs.parentFSM.r.client, rs.parentFSM.r.scheme, headlessServiceDefinition); retrieveError == nil {
			rs.stepsComplete |= CreatedHeadlessService
		}
	}

	// Check to see if the ping service already exists
	labels := selectors.LabelBuilder.Labels()
	pingServiceDefinition := svc.NewPingServiceDefinitionForCR(rs.parentFSM.customResource, labels, labels)
	if err = resources.Retrieve(rs.parentFSM.customResource, rs.parentFSM.namespacedName, rs.parentFSM.r.client, pingServiceDefinition); err != nil {
		// err means not found, so create
		if retrieveError = resources.Create(rs.parentFSM.customResource, rs.parentFSM.r.client, rs.parentFSM.r.scheme, pingServiceDefinition); retrieveError == nil {
			rs.stepsComplete |= CreatedPingService
		}
	}

	return err
}

func (rs *CreatingK8sResourcesState) Enter(previousStateID int) error {

	// Log where we are and what we're doing
	reqLogger := log.WithValues("ActiveMQArtemis Name", rs.parentFSM.customResource.Name)
	reqLogger.Info("Entering CreateK8sResourceState from " + strconv.Itoa(previousStateID))

	switch previousStateID {
	case NotCreatedID:
		rs.enterFromInvalidState()
		break
		//case ScalingID:
		// No brokers running; safe to touch journals etc...
	}

	return nil
}

func (rs *CreatingK8sResourcesState) Update() (error, int) {

	// Log where we are and what we're doing
	reqLogger := log.WithValues("ActiveMQArtemis Name", rs.parentFSM.customResource.Name)
	reqLogger.Info("Updating CreatingK8sResourcesState")

	var err error = nil
	var nextStateID int = CreatingK8sResourcesID
	var statefulSetUpdates uint32 = 0

	currentStatefulSet := &appsv1.StatefulSet{}
	ssNamespacedName := types.NamespacedName{Name: ss.NameBuilder.Name(), Namespace: rs.parentFSM.customResource.Namespace}
	err = rs.parentFSM.r.client.Get(context.TODO(), ssNamespacedName, currentStatefulSet)
	for {
		if err != nil && errors.IsNotFound(err) {
			reqLogger.Error(err, "Failed to get StatefulSet.", "Deployment.Namespace", currentStatefulSet.Namespace, "Deployment.Name", currentStatefulSet.Name)
			err = nil
			break
		}

		// Do we need to check for and bounce an observed generation change here?
		if (rs.stepsComplete&CreatedStatefulSet > 0) &&
			(rs.stepsComplete&CreatedHeadlessService) > 0 &&
			(rs.stepsComplete&CreatedPingService > 0) {

			statefulSetUpdates = reconciler.Process(rs.parentFSM.customResource, rs.parentFSM.r.client, rs.parentFSM.r.scheme, currentStatefulSet)
			if statefulSetUpdates > 0 {
				if err := resources.Update(rs.parentFSM.customResource, rs.parentFSM.r.client, currentStatefulSet); err != nil {
					reqLogger.Error(err, "Failed to update StatefulSet.", "Deployment.Namespace", currentStatefulSet.Namespace, "Deployment.Name", currentStatefulSet.Name)
					break
				}
			}
			if rs.parentFSM.customResource.Spec.DeploymentPlan.Size != currentStatefulSet.Status.ReadyReplicas {
				if rs.parentFSM.customResource.Spec.DeploymentPlan.Size > 0 {
					nextStateID = ScalingID
				}
				break
			}
		} else {
			// Not ready... requeue to wait? What other action is required - try to recreate?
			rs.parentFSM.r.result = reconcile.Result{Requeue: true, RequeueAfter: time.Second * 5}
			rs.enterFromInvalidState()
			reqLogger.Info("CreatingK8sResourcesState requesting reconcile requeue for 5 seconds due to k8s resources not created")
			break
		}

		break
	}
	pods.UpdatePodStatus(rs.parentFSM.customResource, rs.parentFSM.r.client, ssNamespacedName)

	return err, nextStateID
}

func (rs *CreatingK8sResourcesState) Exit() error {

	// Log where we are and what we're doing
	reqLogger := log.WithValues("ActiveMQArtemis Name", rs.parentFSM.customResource.Name)
	reqLogger.Info("Exiting CreatingK8sResourceState")

	return nil
}

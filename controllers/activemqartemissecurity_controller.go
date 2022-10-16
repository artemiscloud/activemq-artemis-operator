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
	"reflect"

	brokerv1beta1 "github.com/artemiscloud/activemq-artemis-operator/api/v1beta1"
	"github.com/artemiscloud/activemq-artemis-operator/pkg/resources"
	"github.com/artemiscloud/activemq-artemis-operator/pkg/resources/environments"
	"github.com/artemiscloud/activemq-artemis-operator/pkg/resources/secrets"
	"github.com/artemiscloud/activemq-artemis-operator/pkg/utils/common"
	"github.com/artemiscloud/activemq-artemis-operator/pkg/utils/random"
	"github.com/artemiscloud/activemq-artemis-operator/pkg/utils/selectors"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var elog = ctrl.Log.WithName("controller_v1alpha1activemqartemissecurity")

// ActiveMQArtemisSecurityReconciler reconciles a ActiveMQArtemisSecurity object
type ActiveMQArtemisSecurityReconciler struct {
	client.Client
	Scheme           *runtime.Scheme
	BrokerReconciler *ActiveMQArtemisReconciler
}

//+kubebuilder:rbac:groups=broker.amq.io,namespace=activemq-artemis-operator,resources=activemqartemissecurities,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=broker.amq.io,namespace=activemq-artemis-operator,resources=activemqartemissecurities/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=broker.amq.io,namespace=activemq-artemis-operator,resources=activemqartemissecurities/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ActiveMQArtemisSecurity object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *ActiveMQArtemisSecurityReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {

	reqLogger := ctrl.Log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name, "Reconciling", "ActiveMQArtemisSecurity")
	reqLogger.Info("===Reconciling security===")

	instance := &brokerv1beta1.ActiveMQArtemisSecurity{}

	if err := r.Client.Get(context.TODO(), request.NamespacedName, instance); err != nil {
		if errors.IsNotFound(err) {
			//unregister the CR
			r.BrokerReconciler.RemoveBrokerConfigHandler(request.NamespacedName)
			// Setting err to nil to prevent requeue
			err = nil
		} else {
			reqLogger.Error(err, "Reconcile errored thats not IsNotFound, requeuing request", "Request Namespace", request.Namespace, "Request Name", request.Name)
		}

		if err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{RequeueAfter: common.GetReconcileResyncPeriod()}, nil
	}

	newHandler := &ActiveMQArtemisSecurityConfigHandler{
		SecurityCR:     instance.DeepCopy(),
		NamespacedName: request.NamespacedName,
		owner:          r,
	}

	brokerNamespacedName := types.NamespacedName{
		Namespace: request.Namespace,
		Name:      "",
	}
	if securityHandler := GetSecurityConfigHandler(brokerNamespacedName); securityHandler != nil {
		if reflect.DeepEqual(securityHandler, newHandler) {
			reqLogger.Info("Will not reconcile the same security config")
			return ctrl.Result{RequeueAfter: common.GetReconcileResyncPeriod()}, nil
		}
	}
	reqLogger.Info("Operator doesn't have the security handler, add it")
	if err := r.BrokerReconciler.AddBrokerConfigHandler(request.NamespacedName, newHandler); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: common.GetReconcileResyncPeriod()}, nil
}

type ActiveMQArtemisSecurityConfigHandler struct {
	SecurityCR     *brokerv1beta1.ActiveMQArtemisSecurity
	NamespacedName types.NamespacedName
	owner          *ActiveMQArtemisSecurityReconciler
}

func (r *ActiveMQArtemisSecurityConfigHandler) Clone() common.ActiveMQArtemisConfigHandler {
	clone := &ActiveMQArtemisSecurityConfigHandler{
		SecurityCR:     r.SecurityCR.DeepCopy(),
		NamespacedName: r.NamespacedName,
		owner:          r.owner,
	}
	return clone
}

func (r *ActiveMQArtemisSecurityConfigHandler) Merge(original common.ActiveMQArtemisConfigHandler) {
	secHandler := original.(*ActiveMQArtemisSecurityConfigHandler)

	incomingLoginModules := secHandler.SecurityCR.Spec.LoginModules
	for _, lm := range incomingLoginModules.PropertiesLoginModules {
		found := -1
		for i, lmExisting := range r.SecurityCR.Spec.LoginModules.PropertiesLoginModules {
			if lmExisting.Name == lm.Name {
				found = i
				break
			}
		}
		if found >= 0 {
			mergePropertiesLoginModules(&r.SecurityCR.Spec.LoginModules.PropertiesLoginModules[found], &lm)
		} else {
			r.SecurityCR.Spec.LoginModules.PropertiesLoginModules = append(r.SecurityCR.Spec.LoginModules.PropertiesLoginModules, lm)
		}
	}
	for _, lm := range incomingLoginModules.GuestLoginModules {
		found := -1
		for i, lmExisting := range r.SecurityCR.Spec.LoginModules.GuestLoginModules {
			if lmExisting.Name == lm.Name {
				found = i
				break
			}
		}
		if found >= 0 {
			r.SecurityCR.Spec.LoginModules.GuestLoginModules[found] = lm
		} else {
			r.SecurityCR.Spec.LoginModules.GuestLoginModules = append(r.SecurityCR.Spec.LoginModules.GuestLoginModules, lm)
		}
	}
	for _, lm := range incomingLoginModules.KeycloakLoginModules {
		found := -1
		for i, lmExisting := range r.SecurityCR.Spec.LoginModules.KeycloakLoginModules {
			if lmExisting.Name == lm.Name {
				found = i
				break
			}
		}
		if found >= 0 {
			r.SecurityCR.Spec.LoginModules.KeycloakLoginModules[found] = lm
		} else {
			r.SecurityCR.Spec.LoginModules.KeycloakLoginModules = append(r.SecurityCR.Spec.LoginModules.KeycloakLoginModules, lm)
		}
	}

	incomingSecurityDomains := secHandler.SecurityCR.Spec.SecurityDomains
	if incomingSecurityDomains.BrokerDomain.Name != nil && *incomingSecurityDomains.BrokerDomain.Name != "" {
		r.SecurityCR.Spec.SecurityDomains.BrokerDomain = incomingSecurityDomains.BrokerDomain
	}
	if incomingSecurityDomains.ConsoleDomain.Name != nil && *incomingSecurityDomains.ConsoleDomain.Name != "" {
		r.SecurityCR.Spec.SecurityDomains.ConsoleDomain = incomingSecurityDomains.ConsoleDomain
	}

	incomingSecuritySettings := secHandler.SecurityCR.Spec.SecuritySettings
	existingBrokerSettings := r.SecurityCR.Spec.SecuritySettings.Broker
	for _, entry := range incomingSecuritySettings.Broker {
		found := -1
		for i, existingEntry := range existingBrokerSettings {
			if entry.Match == existingEntry.Match {
				found = i
				break
			}
		}
		if found >= 0 {
			for _, incomingPerm := range entry.Permissions {
				permFound := -1
				for k, existingPerm := range r.SecurityCR.Spec.SecuritySettings.Broker[found].Permissions {
					if incomingPerm.OperationType == existingPerm.OperationType {
						permFound = k
						break
					}
				}
				if permFound >= 0 {
					//merge perm roles
					for _, incomingRole := range incomingPerm.Roles {
						roleFound := false
						for _, existingRole := range r.SecurityCR.Spec.SecuritySettings.Broker[found].Permissions[permFound].Roles {
							if existingRole == incomingRole {
								roleFound = true
								break
							}
							if !roleFound {
								r.SecurityCR.Spec.SecuritySettings.Broker[found].Permissions[permFound].Roles = append(r.SecurityCR.Spec.SecuritySettings.Broker[found].Permissions[permFound].Roles, incomingRole)
							}
						}
					}
				} else {
					r.SecurityCR.Spec.SecuritySettings.Broker[found].Permissions = append(r.SecurityCR.Spec.SecuritySettings.Broker[found].Permissions, incomingPerm)
				}
			}
		} else {
			r.SecurityCR.Spec.SecuritySettings.Broker = append(r.SecurityCR.Spec.SecuritySettings.Broker, entry)
		}
	}
	existingHawtioRoles := r.SecurityCR.Spec.SecuritySettings.Management.HawtioRoles
	for _, role := range incomingSecuritySettings.Management.HawtioRoles {
		notFound := true
		for _, existingRole := range existingHawtioRoles {
			if role == existingRole {
				notFound = false
				break
			}
		}
		if notFound {
			r.SecurityCR.Spec.SecuritySettings.Management.HawtioRoles = append(r.SecurityCR.Spec.SecuritySettings.Management.HawtioRoles, role)
		}
	}
	incomingConnector := incomingSecuritySettings.Management.Connector
	if incomingConnector.Host != nil {
		r.SecurityCR.Spec.SecuritySettings.Management.Connector.Host = incomingConnector.Host
	}
	if incomingConnector.Port != nil {
		r.SecurityCR.Spec.SecuritySettings.Management.Connector.Port = incomingConnector.Port
	}
	if incomingConnector.RmiRegistryPort != nil {
		r.SecurityCR.Spec.SecuritySettings.Management.Connector.RmiRegistryPort = incomingConnector.RmiRegistryPort
	}
	if incomingConnector.JmxRealm != nil {
		r.SecurityCR.Spec.SecuritySettings.Management.Connector.JmxRealm = incomingConnector.JmxRealm
	}
	if incomingConnector.ObjectName != nil {
		r.SecurityCR.Spec.SecuritySettings.Management.Connector.ObjectName = incomingConnector.ObjectName
	}
	if incomingConnector.AuthenticatorType != nil {
		r.SecurityCR.Spec.SecuritySettings.Management.Connector.AuthenticatorType = incomingConnector.AuthenticatorType
	}
	if incomingConnector.Secured != nil {
		r.SecurityCR.Spec.SecuritySettings.Management.Connector.Secured = incomingConnector.Secured
	}
	if incomingConnector.KeyStoreProvider != nil {
		r.SecurityCR.Spec.SecuritySettings.Management.Connector.KeyStoreProvider = incomingConnector.KeyStoreProvider
	}
	if incomingConnector.KeyStorePath != nil {
		r.SecurityCR.Spec.SecuritySettings.Management.Connector.KeyStorePath = incomingConnector.KeyStorePath
	}
	if incomingConnector.KeyStorePassword != nil {
		r.SecurityCR.Spec.SecuritySettings.Management.Connector.KeyStorePassword = incomingConnector.KeyStorePassword
	}
	if incomingConnector.TrustStoreProvider != nil {
		r.SecurityCR.Spec.SecuritySettings.Management.Connector.TrustStoreProvider = incomingConnector.TrustStoreProvider
	}
	if incomingConnector.TrustStorePath != nil {
		r.SecurityCR.Spec.SecuritySettings.Management.Connector.TrustStorePath = incomingConnector.TrustStorePath
	}
	if incomingConnector.TrustStorePassword != nil {
		r.SecurityCR.Spec.SecuritySettings.Management.Connector.TrustStorePassword = incomingConnector.TrustStorePassword
	}
	if incomingConnector.PasswordCodec != nil {
		r.SecurityCR.Spec.SecuritySettings.Management.Connector.PasswordCodec = incomingConnector.PasswordCodec
	}

	incomingDefaultAccess := incomingSecuritySettings.Management.Authorisation.DefaultAccess
	existingDefaultAccess := r.SecurityCR.Spec.SecuritySettings.Management.Authorisation.DefaultAccess
	for _, da := range incomingDefaultAccess {
		found := -1
		for i, existingDa := range existingDefaultAccess {
			if *existingDa.Method == *da.Method {
				found = i
				break
			}
		}
		if found >= 0 {
			existingRoles := r.SecurityCR.Spec.SecuritySettings.Management.Authorisation.DefaultAccess[found].Roles
			for _, incomingRole := range da.Roles {
				rfound := false
				for _, existingRole := range existingRoles {
					if existingRole == incomingRole {
						rfound = true
						break
					}
					if !rfound {
						r.SecurityCR.Spec.SecuritySettings.Management.Authorisation.DefaultAccess[found].Roles = append(r.SecurityCR.Spec.SecuritySettings.Management.Authorisation.DefaultAccess[found].Roles, incomingRole)
					}
				}
			}
		} else {
			r.SecurityCR.Spec.SecuritySettings.Management.Authorisation.DefaultAccess = append(r.SecurityCR.Spec.SecuritySettings.Management.Authorisation.DefaultAccess, da)
		}
	}

	incomingAllowedList := incomingSecuritySettings.Management.Authorisation.AllowedList
	existingAllowedList := r.SecurityCR.Spec.SecuritySettings.Management.Authorisation.AllowedList
	for _, al := range incomingAllowedList {
		found := false
		for _, existingAl := range existingAllowedList {
			if al.Domain != nil && existingAl.Domain != nil && *al.Domain == *existingAl.Domain {
				if al.Key != nil && existingAl.Key != nil && *al.Key == *existingAl.Key {
					found = true
					break
				}
			}
		}
		if !found {
			r.SecurityCR.Spec.SecuritySettings.Management.Authorisation.AllowedList = append(r.SecurityCR.Spec.SecuritySettings.Management.Authorisation.AllowedList, al)
		}
	}

	incomingRoleAccess := incomingSecuritySettings.Management.Authorisation.RoleAccess
	existingRoleAccess := r.SecurityCR.Spec.SecuritySettings.Management.Authorisation.RoleAccess
	clog.Info("==== Meging RoleAccess...", "incoming", incomingRoleAccess, "existing", existingRoleAccess)
	for x, ra := range incomingRoleAccess {
		clog.Info("checking incoming ra", "indx", x, "ra", ra)
		found := -1
		for i, existingRa := range existingRoleAccess {
			if ra.Domain != nil && existingRa.Domain != nil && *ra.Domain == *existingRa.Domain {
				clog.Info("same domain at existing ", "index", i)
				if ra.Key == nil && existingRa.Key == nil {
					clog.Info("ok they match as key all nils")
					found = i
					break
				}
				if ra.Key != nil && existingRa.Key != nil && *ra.Key == *existingRa.Key {
					clog.Info("ok1 the keys are same")
					found = i
					break
				}
			}
		}
		if found >= 0 {
			existingList := r.SecurityCR.Spec.SecuritySettings.Management.Authorisation.RoleAccess[found].AccessList
			incomingList := ra.AccessList
			clog.Info("we found and to merge", "inx", found, "existing", existingList, "incoming", incomingList)
			for _, incomingAccess := range incomingList {
				clog.Info("checking out incoming access", "value", incomingAccess)
				found1 := -1
				for m, existingAccess := range existingList {
					if *existingAccess.Method == *incomingAccess.Method {
						found1 = m
						clog.Info("There is existing one", "method", existingAccess.Method)
						break
					}
				}
				if found1 >= 0 {
					existingRoles := r.SecurityCR.Spec.SecuritySettings.Management.Authorisation.RoleAccess[found].AccessList[found1].Roles
					clog.Info("we found one to merge roles", "existing roles", existingRoles, "incoming roles", incomingAccess.Roles)
					for _, incomingRole := range incomingAccess.Roles {
						found2 := false
						for _, existingRole := range existingRoles {
							if existingRole == incomingRole {
								found2 = true
								break
							}
						}
						if !found2 {
							clog.Info("not found, adding", "Incoming role", incomingRole, "not in method", r.SecurityCR.Spec.SecuritySettings.Management.Authorisation.RoleAccess[found].AccessList[found1].Method)
							r.SecurityCR.Spec.SecuritySettings.Management.Authorisation.RoleAccess[found].AccessList[found1].Roles = append(r.SecurityCR.Spec.SecuritySettings.Management.Authorisation.RoleAccess[found].AccessList[found1].Roles, incomingRole)
						}
					}

				} else {
					r.SecurityCR.Spec.SecuritySettings.Management.Authorisation.RoleAccess[found].AccessList = append(r.SecurityCR.Spec.SecuritySettings.Management.Authorisation.RoleAccess[found].AccessList, incomingAccess)
				}
			}
		} else {
			r.SecurityCR.Spec.SecuritySettings.Management.Authorisation.RoleAccess = append(r.SecurityCR.Spec.SecuritySettings.Management.Authorisation.RoleAccess, ra)
		}
	}
}

func mergePropertiesLoginModules(target *brokerv1beta1.PropertiesLoginModuleType, incoming *brokerv1beta1.PropertiesLoginModuleType) {
	reqLogger := ctrl.Log.WithName("merge properties login module")
	for _, inUser := range incoming.Users {
		reqLogger.V(1).Info("incoming user", "name", inUser.Name)
		found := -1
		for i, existingUser := range target.Users {
			if existingUser.Name == inUser.Name {
				reqLogger.V(1).Info("found existing user", "i", i)
				found = i
				break
			}
		}
		if found >= 0 {
			target.Users[found].Password = inUser.Password
			target.Users[found].Roles = combineRoles(target.Users[found].Roles, inUser.Roles)
		} else {
			target.Users = append(target.Users, inUser)
		}
	}
}

func combineRoles(roles1 []string, roles2 []string) []string {
	rolesMap := make(map[string]string)
	for _, role1 := range roles1 {
		rolesMap[role1] = role1
	}
	for _, role2 := range roles2 {
		rolesMap[role2] = role2
	}
	result := make([]string, 0)
	for k := range rolesMap {
		result = append(result, k)
	}
	return result
}

func (r *ActiveMQArtemisSecurityConfigHandler) IsApplicableFor(brokerNamespacedName types.NamespacedName) bool {
	reqLogger := ctrl.Log.WithValues("IsApplicableFor", brokerNamespacedName)

	applyTo := r.SecurityCR.Spec.ApplyToCrNames
	reqLogger.V(1).Info("applyTo", "len", len(applyTo), "sec", r.SecurityCR.Spec)

	//currently security doesnt apply to other namespaces than its own
	if r.NamespacedName.Namespace != brokerNamespacedName.Namespace {
		reqLogger.V(1).Info("this security cr is not applicable for broker because it's not in my namespace")
		return false
	}
	if brokerNamespacedName.Name == "" {
		reqLogger.V(1).Info("For fetching the security CR")
		return true
	}
	if len(applyTo) == 0 {
		reqLogger.V(1).Info("this security cr is applicable for broker because no applyTo is configured")
		return true
	}
	for _, crName := range applyTo {
		reqLogger.V(1).Info("Going through applyTo", "crName", crName)
		if crName == "*" || crName == "" || crName == brokerNamespacedName.Name {
			reqLogger.V(1).Info("this security cr is applicable for broker as it's either match-all or match name")
			return true
		}
	}
	reqLogger.V(1).Info("all applyToCrNames checked, no match. Not applicable")
	return false
}

func (r *ActiveMQArtemisSecurityConfigHandler) processCrPasswords() *brokerv1beta1.ActiveMQArtemisSecurity {
	result := r.SecurityCR.DeepCopy()

	if len(result.Spec.LoginModules.PropertiesLoginModules) > 0 {
		for i, pm := range result.Spec.LoginModules.PropertiesLoginModules {
			if len(pm.Users) > 0 {
				for j, user := range pm.Users {
					if user.Password == nil {
						result.Spec.LoginModules.PropertiesLoginModules[i].Users[j].Password = r.getPassword("security-properties-"+pm.Name, user.Name)
					}
				}
			}
		}
	}

	if len(result.Spec.LoginModules.KeycloakLoginModules) > 0 {
		for _, pm := range result.Spec.LoginModules.KeycloakLoginModules {
			keycloakSecretName := "security-keycloak-" + pm.Name
			if pm.Configuration.ClientKeyStore != nil {
				if pm.Configuration.ClientKeyPassword == nil {
					pm.Configuration.ClientKeyPassword = r.getPassword(keycloakSecretName, "client-key-password")
				}
				if pm.Configuration.ClientKeyStorePassword == nil {
					pm.Configuration.ClientKeyStorePassword = r.getPassword(keycloakSecretName, "client-key-store-password")
				}
			}
			if pm.Configuration.TrustStore != nil {
				if pm.Configuration.TrustStorePassword == nil {
					pm.Configuration.TrustStorePassword = r.getPassword(keycloakSecretName, "trust-store-password")
				}
			}
			//need to process pm.Configuration.Credentials too. later.
			if len(pm.Configuration.Credentials) > 0 {
				for i, kv := range pm.Configuration.Credentials {
					if kv.Value == nil {
						pm.Configuration.Credentials[i].Value = r.getPassword(keycloakSecretName, "credentials-"+kv.Key)
					}
				}
			}
		}
	}
	return result
}

func (r *ActiveMQArtemisSecurityConfigHandler) GetDefaultLabels() map[string]string {
	defaultLabelData := selectors.LabelerData{}
	defaultLabelData.Base(r.SecurityCR.Name).Suffix("app").Generate()
	return defaultLabelData.Labels()

}

//retrive value from secret, generate value if not exist.
func (r *ActiveMQArtemisSecurityConfigHandler) getPassword(secretName string, key string) *string {
	//check if the secret exists.
	namespacedName := types.NamespacedName{
		Name:      secretName,
		Namespace: r.NamespacedName.Namespace,
	}
	// Attempt to retrieve the secret
	stringDataMap := make(map[string]string)

	secretDefinition := secrets.NewSecret(namespacedName, secretName, stringDataMap, r.GetDefaultLabels())

	if err := resources.Retrieve(namespacedName, r.owner.Client, secretDefinition); err != nil {
		if errors.IsNotFound(err) {
			//create the secret
			resources.Create(r.SecurityCR, r.owner.Client, r.owner.Scheme, secretDefinition)
		}
	} else {
		slog.Info("Found secret " + secretName)

		if elem, ok := secretDefinition.Data[key]; ok {
			//the value exists
			value := string(elem)
			return &value
		}
	}
	//now need generate value
	value := random.GenerateRandomString(8)
	//update the secret
	if secretDefinition.Data == nil {
		secretDefinition.Data = make(map[string][]byte)
	}
	secretDefinition.Data[key] = []byte(value)
	slog.Info("Updating secret", "secret", namespacedName.Name)
	if err := resources.Update(r.owner.Client, secretDefinition); err != nil {
		slog.Error(err, "failed to update secret", "secret", secretName)
	}
	return &value
}

func (r *ActiveMQArtemisSecurityConfigHandler) Config(initContainers []corev1.Container, outputDirRoot string, yacfgProfileVersion string, yacfgProfileName string) (value []string) {
	ctrl.Log.Info("Reconciling ActiveMQArtemisSecurity", "cr", r.SecurityCR)
	result := r.processCrPasswords()
	outputDir := outputDirRoot + "/security"
	var configCmds = []string{"echo \"making dir " + outputDir + "\"", "mkdir -p " + outputDir}
	filePath := outputDir + "/security-config.yaml"
	cmdPersistCRAsYaml, err := r.persistCR(filePath, result)
	if err != nil {
		slog.Error(err, "Error marshalling security CR", "cr", r.SecurityCR)
		return nil
	}
	slog.Info("get the command", "value", cmdPersistCRAsYaml)
	configCmds = append(configCmds, cmdPersistCRAsYaml)
	configCmds = append(configCmds, "/opt/amq-broker/script/cfg/config-security.sh")
	envVarName := "SECURITY_CFG_YAML"
	envVar := corev1.EnvVar{
		Name:      envVarName,
		Value:     filePath,
		ValueFrom: nil,
	}
	environments.Create(initContainers, &envVar)

	envVarName = "YACFG_PROFILE_VERSION"
	envVar = corev1.EnvVar{
		envVarName,
		yacfgProfileVersion,
		nil,
	}
	environments.Create(initContainers, &envVar)

	envVarName = "YACFG_PROFILE_NAME"
	envVar = corev1.EnvVar{
		envVarName,
		yacfgProfileName,
		nil,
	}
	environments.Create(initContainers, &envVar)

	ctrl.Log.Info("returning config cmds", "value", configCmds)
	return configCmds
}

func (r *ActiveMQArtemisSecurityConfigHandler) persistCR(filePath string, cr *brokerv1beta1.ActiveMQArtemisSecurity) (value string, err error) {

	// remove superfluous data that can trip up the shell
	stripped := cr.DeepCopy()
	stripped.ObjectMeta = metav1.ObjectMeta{}

	data, err := yaml.Marshal(stripped)
	if err != nil {
		return "", err
	}
	return "echo \"" + string(data) + "\" > " + filePath, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ActiveMQArtemisSecurityReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&brokerv1beta1.ActiveMQArtemisSecurity{}).
		Complete(r)
}

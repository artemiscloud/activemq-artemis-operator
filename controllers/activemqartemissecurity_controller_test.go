/*
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
// +kubebuilder:docs-gen:collapse=Apache License

/*
As usual, we start with the necessary imports. We also define some utility variables.
*/
package controllers

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"

	brokerv1beta1 "github.com/artemiscloud/activemq-artemis-operator/api/v1beta1"
	"github.com/artemiscloud/activemq-artemis-operator/pkg/utils/common"
	"github.com/artemiscloud/activemq-artemis-operator/pkg/utils/namer"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

var boolFalse = false
var boolTrue = true

var _ = Describe("security controller", func() {

	BeforeEach(func() {
		if verbose {
			fmt.Println("Time with MicroSeconds: ", time.Now().Format("2006-01-02 15:04:05.000000"), " test:", CurrentSpecReport())
		}
	})

	AfterEach(func() {
	})

	Context("broker with security custom resources", Label("broker-security-res"), func() {

		It("security after recreating broker cr", func() {

			By("deploy a security cr")
			securityCr, createdSecurityCr := DeploySecurity(nameFromTest(), defaultNamespace, func(candidate *brokerv1beta1.ActiveMQArtemisSecurity) {
			})

			By("deploy a broker cr")
			brokerCr, createdBrokerCr := DeployCustomBroker(defaultNamespace, nil)

			By("checking the security gets applied")
			requestedSs := &appsv1.StatefulSet{}
			Eventually(func() bool {
				key := types.NamespacedName{Name: namer.CrToSS(createdBrokerCr.Name), Namespace: defaultNamespace}
				err := k8sClient.Get(ctx, key, requestedSs)
				if err != nil {
					return false
				}

				initContainer := requestedSs.Spec.Template.Spec.InitContainers[0]
				secApplied := false
				for _, arg := range initContainer.Args {
					if strings.Contains(arg, "mkdir -p /init_cfg_root/security/security") {
						secApplied = true
						break
					}
				}
				return secApplied
			}, timeout, interval).Should(BeTrue())

			By("delete the broker cr")
			Expect(k8sClient.Delete(ctx, createdBrokerCr)).Should(Succeed())
			Eventually(func() bool {
				return checkCrdDeleted(brokerCr.Name, defaultNamespace, createdBrokerCr)
			}, timeout, interval).Should(BeTrue())

			By("re-deploy the broker cr")
			brokerCr, createdBrokerCr = DeployCustomBroker(defaultNamespace, func(candidate *brokerv1beta1.ActiveMQArtemis) {
				candidate.Name = brokerCr.Name
			})

			By("verify the security is re-applied")
			Eventually(func() bool {
				key := types.NamespacedName{Name: namer.CrToSS(createdBrokerCr.Name), Namespace: defaultNamespace}
				err := k8sClient.Get(ctx, key, requestedSs)
				if err != nil {
					return false
				}

				initContainer := requestedSs.Spec.Template.Spec.InitContainers[0]
				secApplied := false
				for _, arg := range initContainer.Args {
					if strings.Contains(arg, "mkdir -p /init_cfg_root/security/security") {
						secApplied = true
						break
					}
				}
				return secApplied
			}, timeout, interval).Should(BeTrue())

			//cleanup
			Expect(k8sClient.Delete(ctx, createdBrokerCr)).Should(Succeed())
			Eventually(func() bool {
				return checkCrdDeleted(brokerCr.Name, defaultNamespace, createdBrokerCr)
			}, existingClusterTimeout, existingClusterInterval).Should(BeTrue())

			Expect(k8sClient.Delete(ctx, createdSecurityCr)).Should(Succeed())
			Eventually(func() bool {
				return checkCrdDeleted(securityCr.Name, defaultNamespace, createdSecurityCr)
			}, existingClusterTimeout, existingClusterInterval).Should(BeTrue())
		})
	})

	Context("Multiple Security CRs Merge Test", Label("multi-sec-test-context"), func() {

		password1 := "password1"
		password2 := "password2"
		password3 := "password3"
		password4 := "password4"

		propModuleName := "login-prop-module"
		propModuleName1 := "login-prop-module1"
		propModuleName2 := "login-prop-module2"
		guestModuleName := "guest-module"
		guestUser := "guestUser"
		guestRole := "guestRole"
		guestRole1 := "guestRole1"

		domainName := "activemq"
		domainName1 := "activemq1"
		consoleDomainName := "console"
		moduleFlagSufficient := "sufficient"
		moduleFlagRequired := "required"

		viewerRole := "viewer"
		hawtioDomain := "hawtio"
		artemisDomain := "org.apache.activemq.artemis"
		jbossDomain := "org.jboss"
		allListMethod := "list*"
		allGetMethod := "get*"
		allSetMethod := "set*"
		allBrowseMethod := "browse*"

		hawtioGuestRole := "guest"
		hawtioGuest1Role := "guest1"
		hawtioAdminRole := "admin"
		connectorHost := "localhost"
		connectorHost1 := "localhost1"
		var connectorPort int32 = 2099
		var connectorPort1 int32 = 3099
		jmxRealm := "activemq"
		allowedDomain := "org.apache"
		allowedKey := "sub=queue"
		allowedKey1 := "sub=topic"
		allowedKey2 := "sub=queue1"

		keycloakBrokerName := "keycloak-broker"
		keycloakConsoleName := "keycloak-console"
		directAccessType := "directAccess"
		bearerTokenType := "bearerToken"

		nsName := types.NamespacedName{
			Name:      "ex-aao",
			Namespace: defaultNamespace,
		}

		ns2Name := types.NamespacedName{
			Name:      "ex-aao2",
			Namespace: defaultNamespace,
		}

		klCfgRealm := "artemis-keycloak-demo"
		klCfgRealmx := "artemis-keycloak-xxxx"
		klCfgResource := "artemis-broker"
		klCfgConsoleResource := "artemis-console"
		klCfgAuthServerUrl := "http://10.109.131.209:8888/auth"
		klCfgAuthServerUrl1 := "http://10.109.131.209:8080/auth"
		klCfgAuthServerUrlConsole := "http://keycloak:5555/auth"
		klCfgAuthServerUrlConsole1 := "http://keycloak.3387.com/auth"
		klCfgPrincipalAttribute := "preferred_username"
		klCfgSslRequired := "external"
		klCfgCredentialKey := "secret"
		klCfgCredentialValue := "9699685c-8a30-45cf-bf19-xxxxxxxxxxxx"
		klCfgCredentialValue1 := "9699685c-8a30-45cf-bf19-0d38bbac5fdc"
		var klCfgConfiPort int32 = 0

		It("testing multiple security CRs adding and deleting", Label("adding-and-deleting"), func() {
			By("deploy 3 security CRs")
			secCrd1, createdSecCrd1 := DeploySecurity("", defaultNamespace, func(candidate *brokerv1beta1.ActiveMQArtemisSecurity) {
				candidate.Spec.ApplyToCrNames = []string{nsName.Name}
			})
			secCrd2, createdSecCrd2 := DeploySecurity("", defaultNamespace, func(candidate *brokerv1beta1.ActiveMQArtemisSecurity) {
				candidate.Spec.ApplyToCrNames = []string{nsName.Name}
			})
			secCrd3, createdSecCrd3 := DeploySecurity("", defaultNamespace, func(candidate *brokerv1beta1.ActiveMQArtemisSecurity) {
				candidate.Spec.ApplyToCrNames = []string{nsName.Name}
			})

			key1 := createdSecCrd1.Name + "." + defaultNamespace
			key2 := createdSecCrd2.Name + "." + defaultNamespace
			key3 := createdSecCrd3.Name + "." + defaultNamespace

			value1, value2, value3 := "", "", ""
			value1Ok, value2Ok, value3Ok := false, false, false
			Eventually(func(g Gomega) {
				_, appliedData := GetBrokerSecurityConfigHandler(nsName)
				g.Expect(len(appliedData)).To(Equal(3))

				value1, value1Ok = appliedData[key1]
				Expect(value1Ok).To(BeTrue())
				Expect(value1).To(Equal(createdSecCrd1.ResourceVersion))

				value2, value2Ok = appliedData[key2]
				Expect(value2Ok).To(BeTrue())
				Expect(value2).To(Equal(createdSecCrd2.ResourceVersion))

				value3, value3Ok = appliedData[key3]
				Expect(value3Ok).To(BeTrue())
				Expect(value3).To(Equal(createdSecCrd3.ResourceVersion))
			}, timeout, interval).Should(Succeed())

			By("deploy the broker")
			brokerCrd, createdBrokerCrd := DeployCustomBroker(defaultNamespace, func(candidate *brokerv1beta1.ActiveMQArtemis) {
				candidate.Name = "ex-aao"
				candidate.Spec.DeploymentPlan.Size = 1
			})

			By("checking secrets get created")
			secretKey := types.NamespacedName{
				Name:      brokerCrd.Name + "-applied-security",
				Namespace: defaultNamespace,
			}
			secret := corev1.Secret{}
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, secretKey, &secret)).Should(Succeed())
				g.Expect(len(secret.Data)).To(Equal(3))
				g.Expect(string(string(secret.Data[key1]))).To(Equal(value1))
				g.Expect(string(secret.Data[key2])).To(Equal(value2))
				g.Expect(string(secret.Data[key3])).To(Equal(value3))
				//make sure the secrets is owned by the broker CR
				g.Expect(len(secret.OwnerReferences)).To(Equal(1))
				ownerRef := secret.OwnerReferences[0]
				g.Expect(ownerRef.Kind).To(Equal("ActiveMQArtemis"))
				g.Expect(ownerRef.Name).To(Equal("ex-aao"))
				g.Expect(ownerRef.UID).To(Equal(createdBrokerCrd.UID))
			}, timeout, interval).Should(Succeed())

			By("delete security 3")
			Expect(k8sClient.Delete(ctx, createdSecCrd3)).Should(Succeed())
			Eventually(func() bool {
				return checkCrdDeleted(secCrd3.Name, defaultNamespace, createdSecCrd3)
			}, timeout, interval).Should(BeTrue())
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, secretKey, &secret)).Should(Succeed())
				g.Expect(len(secret.Data)).To(Equal(2))
				g.Expect(string(secret.Data[key1])).To(Equal(value1))
				g.Expect(string(secret.Data[key2])).To(Equal(value2))
			}, timeout, interval).Should(Succeed())

			By("delete security 2 " + createdSecCrd2.Name)
			Expect(k8sClient.Delete(ctx, createdSecCrd2)).Should(Succeed())
			Eventually(func() bool {
				return checkCrdDeleted(secCrd2.Name, defaultNamespace, createdSecCrd2)
			}, timeout, interval).Should(BeTrue())
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, secretKey, &secret)).Should(Succeed())
				g.Expect(len(secret.Data)).To(Equal(1))
				g.Expect(string(secret.Data[key1])).To(Equal(value1))
			}, timeout, interval).Should(Succeed())

			By("delete security 1")
			Expect(k8sClient.Delete(ctx, createdSecCrd1)).Should(Succeed())
			Eventually(func() bool {
				return checkCrdDeleted(secCrd1.Name, defaultNamespace, createdSecCrd1)
			}, timeout, interval).Should(BeTrue())
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, secretKey, &secret)).ShouldNot(Succeed())
			}, timeout, interval).Should(Succeed())

			By("clean up broker")
			Expect(k8sClient.Delete(ctx, createdBrokerCrd))
			Eventually(func() bool {
				return checkCrdDeleted(brokerCrd.Name, defaultNamespace, createdBrokerCrd)
			}, timeout, interval).Should(BeTrue())
		})

		It("testing keycloak merge", Label("sec-keycloak-module-merge"), func() {

			By("deploy 2 security CRs to merge with keycloak security configurations")
			secCrd1, createdSecCrd1 := DeploySecurity("", defaultNamespace, func(secCrdToDeploy *brokerv1beta1.ActiveMQArtemisSecurity) {
				secCrdToDeploy.Spec.ApplyToCrNames = []string{nsName.Name}

				secCrdToDeploy.Spec.LoginModules.PropertiesLoginModules = nil

				keycloakLoginModules := []brokerv1beta1.KeycloakLoginModuleType{
					{
						Name:       keycloakBrokerName,
						ModuleType: &directAccessType,
						Configuration: brokerv1beta1.KeycloakModuleConfigurationType{
							Realm:                   &klCfgRealm,
							Resource:                &klCfgResource,
							AuthServerUrl:           &klCfgAuthServerUrl,
							UseResourceRoleMappings: &boolTrue,
							PrincipalAttribute:      &klCfgPrincipalAttribute,
							SslRequired:             &klCfgSslRequired,
							Credentials: []brokerv1beta1.KeyValueType{
								{
									Key:   klCfgCredentialKey,
									Value: &klCfgCredentialValue,
								},
							},
						},
					},
					{
						Name:       keycloakConsoleName,
						ModuleType: &bearerTokenType,
						Configuration: brokerv1beta1.KeycloakModuleConfigurationType{
							Realm:                   &klCfgRealmx,
							Resource:                &klCfgConsoleResource,
							AuthServerUrl:           &klCfgAuthServerUrlConsole,
							PrincipalAttribute:      &klCfgPrincipalAttribute,
							UseResourceRoleMappings: &boolFalse,
							SslRequired:             &klCfgSslRequired,
							ConfidentialPort:        &klCfgConfiPort,
						},
					},
				}

				secCrdToDeploy.Spec.LoginModules.KeycloakLoginModules = keycloakLoginModules

				secCrdToDeploy.Spec.SecurityDomains.BrokerDomain = brokerv1beta1.BrokerDomainType{
					Name: &domainName,
					LoginModules: []brokerv1beta1.LoginModuleReferenceType{
						{
							Name:   &keycloakBrokerName,
							Flag:   &moduleFlagRequired,
							Debug:  &boolFalse,
							Reload: &boolTrue,
						},
					},
				}
				secCrdToDeploy.Spec.SecurityDomains.ConsoleDomain = brokerv1beta1.BrokerDomainType{
					Name: &consoleDomainName,
					LoginModules: []brokerv1beta1.LoginModuleReferenceType{
						{
							Name: &keycloakConsoleName,
							Flag: &moduleFlagSufficient,
						},
					},
				}
			})

			secCrd2, createdSecCrd2 := DeploySecurity("", defaultNamespace, func(secCrdToDeploy *brokerv1beta1.ActiveMQArtemisSecurity) {
				applyToCrs := []string{nsName.Name, ns2Name.Name}
				secCrdToDeploy.Spec.ApplyToCrNames = applyToCrs

				secCrdToDeploy.Spec.LoginModules.PropertiesLoginModules = nil

				keycloakLoginModules := []brokerv1beta1.KeycloakLoginModuleType{
					{
						Name:       keycloakBrokerName,
						ModuleType: &directAccessType,
						Configuration: brokerv1beta1.KeycloakModuleConfigurationType{
							Realm:                   &klCfgRealm,
							Resource:                &klCfgResource,
							AuthServerUrl:           &klCfgAuthServerUrl1,
							UseResourceRoleMappings: &boolTrue,
							PrincipalAttribute:      &klCfgPrincipalAttribute,
							SslRequired:             &klCfgSslRequired,
							Credentials: []brokerv1beta1.KeyValueType{
								{
									Key:   klCfgCredentialKey,
									Value: &klCfgCredentialValue1,
								},
							},
						},
					},
					{
						Name:       keycloakConsoleName,
						ModuleType: &bearerTokenType,
						Configuration: brokerv1beta1.KeycloakModuleConfigurationType{
							Realm:                   &klCfgRealm,
							Resource:                &klCfgConsoleResource,
							AuthServerUrl:           &klCfgAuthServerUrlConsole1,
							PrincipalAttribute:      &klCfgPrincipalAttribute,
							UseResourceRoleMappings: &boolTrue,
							SslRequired:             &klCfgSslRequired,
							ConfidentialPort:        &klCfgConfiPort,
						},
					},
				}

				secCrdToDeploy.Spec.LoginModules.KeycloakLoginModules = keycloakLoginModules

				secCrdToDeploy.Spec.SecurityDomains.BrokerDomain = brokerv1beta1.BrokerDomainType{
					Name: &domainName,
					LoginModules: []brokerv1beta1.LoginModuleReferenceType{
						{
							Name:   &keycloakBrokerName,
							Flag:   &moduleFlagRequired,
							Debug:  &boolTrue,
							Reload: &boolTrue,
						},
					},
				}
				secCrdToDeploy.Spec.SecurityDomains.ConsoleDomain = brokerv1beta1.BrokerDomainType{
					Name: &consoleDomainName,
					LoginModules: []brokerv1beta1.LoginModuleReferenceType{
						{
							Name: &keycloakConsoleName,
							Flag: &moduleFlagRequired,
						},
					},
				}
			})

			By("checking the keycloak merge result")
			Eventually(func(g Gomega) {
				_, appliedData := GetBrokerSecurityConfigHandler(nsName)
				g.Expect(len(appliedData)).To(Equal(2))
				key1 := createdSecCrd1.Name + "." + defaultNamespace
				key2 := createdSecCrd2.Name + "." + defaultNamespace

				value1, value1Ok := appliedData[key1]
				g.Expect(value1Ok).To(BeTrue())
				g.Expect(value1).To(Equal(createdSecCrd1.ResourceVersion))
				value2, value2Ok := appliedData[key2]
				g.Expect(value2Ok).To(BeTrue())
				g.Expect(value2).To(Equal(createdSecCrd2.ResourceVersion))

				_, appliedData = GetBrokerSecurityConfigHandler(ns2Name)
				g.Expect(len(appliedData)).To(Equal(1))
				key2 = createdSecCrd2.Name + "." + defaultNamespace
				value2, value2Ok = appliedData[key2]
				g.Expect(value2Ok).To(BeTrue())
				g.Expect(value2).To(Equal(createdSecCrd2.ResourceVersion))
			}, timeout, interval).Should(Succeed())

			for _, brokerNs := range []types.NamespacedName{nsName, ns2Name} {
				Eventually(func(g Gomega) {
					configHandler, _ := GetBrokerSecurityConfigHandler(brokerNs)
					securityHandler := configHandler.(*ActiveMQArtemisSecurityConfigHandler)
					g.Expect(securityHandler).NotTo(BeNil())

					secSpec := securityHandler.SecurityCR.Spec
					g.Expect(len(secSpec.ApplyToCrNames)).To(BeEquivalentTo(1))
					g.Expect(secSpec.ApplyToCrNames[0]).To(BeEquivalentTo(brokerNs.Name))

					keycloakModules := secSpec.LoginModules.KeycloakLoginModules
					g.Expect(len(keycloakModules)).To(BeEquivalentTo(2))

					brokerKeycloakModule := keycloakModules[0]
					consoleKeycloakModule := keycloakModules[1]

					g.Expect(brokerKeycloakModule.Name).To(BeEquivalentTo(keycloakBrokerName))
					g.Expect(*brokerKeycloakModule.ModuleType).To(BeEquivalentTo(directAccessType))
					g.Expect(*brokerKeycloakModule.Configuration.Realm).To(BeEquivalentTo(klCfgRealm))
					g.Expect(*brokerKeycloakModule.Configuration.Resource).To(BeEquivalentTo(klCfgResource))
					g.Expect(*brokerKeycloakModule.Configuration.AuthServerUrl).To(BeEquivalentTo(klCfgAuthServerUrl1))
					g.Expect(*brokerKeycloakModule.Configuration.UseResourceRoleMappings).To(BeTrue())
					g.Expect(*brokerKeycloakModule.Configuration.PrincipalAttribute).To(BeEquivalentTo(klCfgPrincipalAttribute))
					g.Expect(*brokerKeycloakModule.Configuration.SslRequired).To(BeEquivalentTo(klCfgSslRequired))
					g.Expect(len(brokerKeycloakModule.Configuration.Credentials)).To(BeEquivalentTo(1))
					g.Expect(brokerKeycloakModule.Configuration.Credentials[0].Key).To(BeEquivalentTo(klCfgCredentialKey))
					g.Expect(*brokerKeycloakModule.Configuration.Credentials[0].Value).To(BeEquivalentTo(klCfgCredentialValue1))

					g.Expect(consoleKeycloakModule.Name).To(BeEquivalentTo(keycloakConsoleName))
					g.Expect(*consoleKeycloakModule.ModuleType).To(BeEquivalentTo(bearerTokenType))
					g.Expect(*consoleKeycloakModule.Configuration.Realm).To(BeEquivalentTo(klCfgRealm))
					g.Expect(*consoleKeycloakModule.Configuration.Resource).To(BeEquivalentTo(klCfgConsoleResource))
					g.Expect(*consoleKeycloakModule.Configuration.AuthServerUrl).To(BeEquivalentTo(klCfgAuthServerUrlConsole1))
					g.Expect(*consoleKeycloakModule.Configuration.PrincipalAttribute).To(BeEquivalentTo(klCfgPrincipalAttribute))
					g.Expect(*consoleKeycloakModule.Configuration.UseResourceRoleMappings).To(BeTrue())
					g.Expect(*consoleKeycloakModule.Configuration.SslRequired).To(BeEquivalentTo(klCfgSslRequired))
					g.Expect(*consoleKeycloakModule.Configuration.ConfidentialPort).To(BeEquivalentTo(0))

					brokerDomain := secSpec.SecurityDomains.BrokerDomain
					g.Expect(*brokerDomain.Name).To(BeEquivalentTo(domainName))
					g.Expect(len(brokerDomain.LoginModules)).To(BeEquivalentTo(1))
					g.Expect(*brokerDomain.LoginModules[0].Name).To(BeEquivalentTo(keycloakBrokerName))
					g.Expect(*brokerDomain.LoginModules[0].Flag).To(BeEquivalentTo(moduleFlagRequired))
					g.Expect(*brokerDomain.LoginModules[0].Debug).To(BeTrue())
					g.Expect(*brokerDomain.LoginModules[0].Reload).To(BeTrue())

					consoleDomain := secSpec.SecurityDomains.ConsoleDomain
					g.Expect(*consoleDomain.Name).To(BeEquivalentTo(consoleDomainName))
					g.Expect(len(consoleDomain.LoginModules)).To(BeEquivalentTo(1))
					g.Expect(*consoleDomain.LoginModules[0].Name).To(BeEquivalentTo(keycloakConsoleName))
					g.Expect(*consoleDomain.LoginModules[0].Flag).To(BeEquivalentTo(moduleFlagRequired))
					g.Expect(consoleDomain.LoginModules[0].Debug).To(BeNil())
					g.Expect(consoleDomain.LoginModules[0].Reload).To(BeNil())

				}, timeout, interval).Should(Succeed())
			}

			Eventually(func(g Gomega) {
				anotherNs := types.NamespacedName{
					Name:      "ex-aao",
					Namespace: "another-namespace",
				}
				configHandler, _ := GetBrokerSecurityConfigHandler(anotherNs)
				securityHandler := configHandler.(*ActiveMQArtemisSecurityConfigHandler)
				g.Expect(securityHandler).To(BeNil())
			})

			By("delete 2nd security CR")
			Expect(k8sClient.Delete(ctx, createdSecCrd2)).Should(Succeed())
			Eventually(func() bool {
				return checkCrdDeleted(secCrd2.Name, defaultNamespace, createdSecCrd2)
			}, timeout, interval).Should(BeTrue())

			// security for ex-aao2 is gone
			Eventually(func(g Gomega) {
				configHandler, _ := GetBrokerSecurityConfigHandler(ns2Name)
				securityHandler := configHandler.(*ActiveMQArtemisSecurityConfigHandler)
				g.Expect(securityHandler).To(BeNil())
			}, timeout, interval).Should(Succeed())

			// security for ex-aao restored to first cr
			Eventually(func(g Gomega) {
				configHandler, _ := GetBrokerSecurityConfigHandler(nsName)
				securityHandler := configHandler.(*ActiveMQArtemisSecurityConfigHandler)
				g.Expect(securityHandler).NotTo(BeNil())

				secSpec := securityHandler.SecurityCR.Spec
				g.Expect(reflect.DeepEqual(secSpec, secCrd1.Spec)).To(BeTrue())
			}, timeout, interval).Should(Succeed())

			// delete 1st security CR
			Expect(k8sClient.Delete(ctx, createdSecCrd1)).Should(Succeed())
			Eventually(func() bool {
				return checkCrdDeleted(secCrd1.Name, defaultNamespace, createdSecCrd1)
			}, timeout, interval).Should(BeTrue())
		})

		It("testing management merge", Label("sec-management-merge"), func() {

			By("deploy 2 security CRs to merge with management security configurations")
			secCrd1, createdSecCrd1 := DeploySecurity("", defaultNamespace, func(secCrdToDeploy *brokerv1beta1.ActiveMQArtemisSecurity) {
				applyToCrs := []string{"ex-aao"}
				secCrdToDeploy.Spec.ApplyToCrNames = applyToCrs

				secCrdToDeploy.Spec.LoginModules.PropertiesLoginModules = nil

				secCrdToDeploy.Spec.SecurityDomains.BrokerDomain = brokerv1beta1.BrokerDomainType{
					Name:         nil,
					LoginModules: nil,
				}

				managementSettings := brokerv1beta1.ManagementSecuritySettingsType{
					HawtioRoles: []string{
						hawtioGuestRole,
					},
					Connector: brokerv1beta1.ConnectorConfigType{
						Host:    &connectorHost,
						Port:    &connectorPort,
						Secured: &boolTrue,
					},
					Authorisation: brokerv1beta1.AuthorisationConfigType{
						DefaultAccess: []brokerv1beta1.DefaultAccessType{
							{
								Method: &allBrowseMethod,
								Roles: []string{
									"guest",
								},
							},
						},
						AllowedList: []brokerv1beta1.AllowedListEntryType{
							{
								Domain: &allowedDomain,
								Key:    &allowedKey,
							},
							{
								Domain: &allowedDomain,
								Key:    &allowedKey1,
							},
						},
						RoleAccess: []brokerv1beta1.RoleAccessType{
							{
								Domain: &artemisDomain,
								AccessList: []brokerv1beta1.DefaultAccessType{
									{
										Method: &allListMethod,
										Roles: []string{
											"amq",
										},
									},
									{
										Method: &allGetMethod,
										Roles: []string{
											"amq",
										},
									},
								},
							},
						},
					},
				}

				secCrdToDeploy.Spec.SecuritySettings.Management = managementSettings
			})

			secCrd2, createdSecCrd2 := DeploySecurity("", defaultNamespace, func(secCrdToDeploy *brokerv1beta1.ActiveMQArtemisSecurity) {
				applyToCrs := []string{"ex-aao", "ex-aao2"}
				secCrdToDeploy.Spec.ApplyToCrNames = applyToCrs

				secCrdToDeploy.Spec.LoginModules.PropertiesLoginModules = nil

				secCrdToDeploy.Spec.SecurityDomains.BrokerDomain = brokerv1beta1.BrokerDomainType{
					Name:         nil,
					LoginModules: nil,
				}

				managementSettings := brokerv1beta1.ManagementSecuritySettingsType{
					HawtioRoles: []string{
						hawtioGuest1Role,
						hawtioAdminRole,
					},
					Connector: brokerv1beta1.ConnectorConfigType{
						Host:     &connectorHost1,
						Port:     &connectorPort1,
						JmxRealm: &jmxRealm,
					},
					Authorisation: brokerv1beta1.AuthorisationConfigType{
						DefaultAccess: []brokerv1beta1.DefaultAccessType{
							{
								Method: &allBrowseMethod,
								Roles: []string{
									"guest1",
								},
							},
							{
								Method: &allGetMethod,
								Roles: []string{
									"guest1",
								},
							},
						},
						AllowedList: []brokerv1beta1.AllowedListEntryType{
							{
								Domain: &allowedDomain,
								Key:    &allowedKey,
							},
							{
								Domain: &allowedDomain,
								Key:    &allowedKey2,
							},
						},
						RoleAccess: []brokerv1beta1.RoleAccessType{
							{
								Domain: &artemisDomain,
								AccessList: []brokerv1beta1.DefaultAccessType{
									{
										Method: &allListMethod,
										Roles: []string{
											"amq",
											"root",
										},
									},
									{
										Method: &allGetMethod,
										Roles: []string{
											"guest",
											"admin",
										},
									},
									{
										Method: &allSetMethod,
										Roles: []string{
											"amq",
											"guest",
										},
									},
								},
							},
							{
								Domain: &jbossDomain,
								AccessList: []brokerv1beta1.DefaultAccessType{
									{
										Method: &allGetMethod,
										Roles: []string{
											"admin",
										},
									},
								},
							},
						},
					},
				}

				secCrdToDeploy.Spec.SecuritySettings.Management = managementSettings
			})

			By("checking the management merge result")
			Eventually(func(g Gomega) {
				configHandler, _ := GetBrokerSecurityConfigHandler(nsName)
				securityHandler := configHandler.(*ActiveMQArtemisSecurityConfigHandler)
				g.Expect(securityHandler).NotTo(BeNil())

				secSpec := securityHandler.SecurityCR.Spec
				g.Expect(len(secSpec.ApplyToCrNames)).To(BeEquivalentTo(1))
				g.Expect(secSpec.ApplyToCrNames[0]).To(BeEquivalentTo(nsName.Name))

				mgmtSettings := secSpec.SecuritySettings.Management

				g.Expect(len(mgmtSettings.HawtioRoles)).To(BeEquivalentTo(3))
				g.Expect(mgmtSettings.HawtioRoles[0]).To(BeEquivalentTo(hawtioGuestRole))
				g.Expect(mgmtSettings.HawtioRoles[1]).To(BeEquivalentTo(hawtioGuest1Role))
				g.Expect(mgmtSettings.HawtioRoles[2]).To(BeEquivalentTo(hawtioAdminRole))

				g.Expect(*mgmtSettings.Connector.Host).To(BeEquivalentTo(connectorHost1))
				g.Expect(*mgmtSettings.Connector.Port).To(BeEquivalentTo(connectorPort1))
				g.Expect(*mgmtSettings.Connector.Secured).To(BeTrue())
				g.Expect(*mgmtSettings.Connector.JmxRealm).To(Equal(jmxRealm))

				g.Expect(len(mgmtSettings.Authorisation.DefaultAccess)).To(BeEquivalentTo(2))
				g.Expect(*mgmtSettings.Authorisation.DefaultAccess[0].Method).To(BeEquivalentTo(allBrowseMethod))
				g.Expect(len(mgmtSettings.Authorisation.DefaultAccess[0].Roles)).To(BeEquivalentTo(2))
				g.Expect(mgmtSettings.Authorisation.DefaultAccess[0].Roles[0]).To(BeEquivalentTo(hawtioGuestRole))
				g.Expect(mgmtSettings.Authorisation.DefaultAccess[0].Roles[1]).To(BeEquivalentTo(hawtioGuest1Role))

				g.Expect(*mgmtSettings.Authorisation.DefaultAccess[1].Method).To(BeEquivalentTo(allGetMethod))
				g.Expect(len(mgmtSettings.Authorisation.DefaultAccess[1].Roles)).To(BeEquivalentTo(1))
				g.Expect(mgmtSettings.Authorisation.DefaultAccess[1].Roles[0]).To(BeEquivalentTo(hawtioGuest1Role))

				g.Expect(len(mgmtSettings.Authorisation.AllowedList)).To(BeEquivalentTo(3))
				g.Expect(*mgmtSettings.Authorisation.AllowedList[0].Domain).To(BeEquivalentTo("org.apache"))
				g.Expect(*mgmtSettings.Authorisation.AllowedList[0].Key).To(BeEquivalentTo("sub=queue"))
				g.Expect(*mgmtSettings.Authorisation.AllowedList[1].Domain).To(BeEquivalentTo("org.apache"))
				g.Expect(*mgmtSettings.Authorisation.AllowedList[1].Key).To(BeEquivalentTo("sub=topic"))
				g.Expect(*mgmtSettings.Authorisation.AllowedList[2].Domain).To(BeEquivalentTo("org.apache"))
				g.Expect(*mgmtSettings.Authorisation.AllowedList[2].Key).To(BeEquivalentTo("sub=queue1"))

				g.Expect(len(mgmtSettings.Authorisation.RoleAccess)).To(BeEquivalentTo(2))
				g.Expect(*mgmtSettings.Authorisation.RoleAccess[0].Domain).To(BeEquivalentTo(artemisDomain))
				g.Expect(len(mgmtSettings.Authorisation.RoleAccess[0].AccessList)).To(BeEquivalentTo(3))

				g.Expect(*mgmtSettings.Authorisation.RoleAccess[0].AccessList[0].Method).To(BeEquivalentTo(allListMethod))
				g.Expect(len(mgmtSettings.Authorisation.RoleAccess[0].AccessList[0].Roles)).To(BeEquivalentTo(2))
				g.Expect(mgmtSettings.Authorisation.RoleAccess[0].AccessList[0].Roles[0]).To(BeEquivalentTo("amq"))
				g.Expect(mgmtSettings.Authorisation.RoleAccess[0].AccessList[0].Roles[1]).To(BeEquivalentTo("root"))

				g.Expect(*mgmtSettings.Authorisation.RoleAccess[0].AccessList[1].Method).To(BeEquivalentTo(allGetMethod))
				g.Expect(len(mgmtSettings.Authorisation.RoleAccess[0].AccessList[1].Roles)).To(BeEquivalentTo(3))
				g.Expect(mgmtSettings.Authorisation.RoleAccess[0].AccessList[1].Roles[0]).To(BeEquivalentTo("amq"))
				g.Expect(mgmtSettings.Authorisation.RoleAccess[0].AccessList[1].Roles[1]).To(BeEquivalentTo("guest"))
				g.Expect(mgmtSettings.Authorisation.RoleAccess[0].AccessList[1].Roles[2]).To(BeEquivalentTo("admin"))

				g.Expect(*mgmtSettings.Authorisation.RoleAccess[0].AccessList[2].Method).To(BeEquivalentTo(allSetMethod))
				g.Expect(len(mgmtSettings.Authorisation.RoleAccess[0].AccessList[2].Roles)).To(BeEquivalentTo(2))
				g.Expect(mgmtSettings.Authorisation.RoleAccess[0].AccessList[2].Roles[0]).To(BeEquivalentTo("amq"))
				g.Expect(mgmtSettings.Authorisation.RoleAccess[0].AccessList[2].Roles[1]).To(BeEquivalentTo("guest"))

				g.Expect(*mgmtSettings.Authorisation.RoleAccess[1].Domain).To(BeEquivalentTo("org.jboss"))
				g.Expect(len(mgmtSettings.Authorisation.RoleAccess[1].AccessList)).To(BeEquivalentTo(1))
				g.Expect(*mgmtSettings.Authorisation.RoleAccess[1].AccessList[0].Method).To(BeEquivalentTo(allGetMethod))
				g.Expect(len(mgmtSettings.Authorisation.RoleAccess[1].AccessList[0].Roles)).To(BeEquivalentTo(1))
				g.Expect(mgmtSettings.Authorisation.RoleAccess[1].AccessList[0].Roles[0]).To(BeEquivalentTo("admin"))
			}, timeout, interval).Should(Succeed())

			By("checking effective security CR for ex-aao2")

			Eventually(func(g Gomega) {
				configHandler, _ := GetBrokerSecurityConfigHandler(ns2Name)
				securityHandler := configHandler.(*ActiveMQArtemisSecurityConfigHandler)
				g.Expect(securityHandler).NotTo(BeNil())

				secSpec := securityHandler.SecurityCR.Spec
				g.Expect(len(secSpec.ApplyToCrNames)).To(BeEquivalentTo(1))
				g.Expect(secSpec.ApplyToCrNames[0]).To(BeEquivalentTo(ns2Name.Name))

				//the effective CR should be the same as second one except ApplyToCrNames
				secSpec.ApplyToCrNames = []string{
					"ex-aao", "ex-aao2",
				}
				g.Expect(reflect.DeepEqual(secSpec, secCrd2.Spec))

			}, timeout, interval).Should(Succeed())

			By("delete 2nd security CR")
			Expect(k8sClient.Delete(ctx, createdSecCrd2)).Should(Succeed())
			Eventually(func() bool {
				return checkCrdDeleted(secCrd2.Name, defaultNamespace, createdSecCrd2)
			}, timeout, interval).Should(BeTrue())

			// security for ex-aao2 is gone
			Eventually(func(g Gomega) {
				configHandler, _ := GetBrokerSecurityConfigHandler(ns2Name)
				securityHandler := configHandler.(*ActiveMQArtemisSecurityConfigHandler)
				g.Expect(securityHandler).To(BeNil())
			}, timeout, interval).Should(Succeed())

			// security for ex-aao restored to first cr
			Eventually(func(g Gomega) {
				configHandler, _ := GetBrokerSecurityConfigHandler(nsName)
				securityHandler := configHandler.(*ActiveMQArtemisSecurityConfigHandler)
				g.Expect(securityHandler).NotTo(BeNil())

				secSpec := securityHandler.SecurityCR.Spec
				g.Expect(reflect.DeepEqual(secSpec, secCrd1.Spec)).To(BeTrue())
			}, timeout, interval).Should(Succeed())

			// delete 1st security CR
			Expect(k8sClient.Delete(ctx, createdSecCrd1)).Should(Succeed())
			Eventually(func() bool {
				return checkCrdDeleted(secCrd1.Name, defaultNamespace, createdSecCrd1)
			}, timeout, interval).Should(BeTrue())
		})

		It("testing security settings merge", Label("sec-settings-merge"), func() {

			By("deploy 2 security CRs to merge with security settings")
			secCrd1, createdSecCrd1 := DeploySecurity("", defaultNamespace, func(secCrdToDeploy *brokerv1beta1.ActiveMQArtemisSecurity) {
				applyToCrs := []string{"ex-aao"}
				secCrdToDeploy.Spec.ApplyToCrNames = applyToCrs

				secCrdToDeploy.Spec.LoginModules.PropertiesLoginModules = nil

				secCrdToDeploy.Spec.SecurityDomains.BrokerDomain = brokerv1beta1.BrokerDomainType{
					Name:         nil,
					LoginModules: nil,
				}

				securitySettings := []brokerv1beta1.BrokerSecuritySettingType{
					{
						Match: "Info",
						Permissions: []brokerv1beta1.PermissionType{
							{
								OperationType: "createDurableQueue",
								Roles: []string{
									"amq",
								},
							},
							{
								OperationType: "deleteDurableQueue",
								Roles: []string{
									"amq",
								},
							},
						},
					},
					{
						Match: "#",
						Permissions: []brokerv1beta1.PermissionType{
							{
								OperationType: "createDurableQueue",
								Roles: []string{
									"root",
								},
							},
							{
								OperationType: "deleteDurableQueue",
								Roles: []string{
									"root",
								},
							},
							{
								OperationType: "createNonDurableQueue",
								Roles: []string{
									"root",
								},
							},
							{
								OperationType: "deleteNonDurableQueue",
								Roles: []string{
									"root",
								},
							},
						},
					},
				}
				secCrdToDeploy.Spec.SecuritySettings.Broker = securitySettings
			})

			secCrd2, createdSecCrd2 := DeploySecurity("", defaultNamespace, func(secCrdToDeploy *brokerv1beta1.ActiveMQArtemisSecurity) {
				applyToCrs := []string{"ex-aao", "ex-aao2"}
				secCrdToDeploy.Spec.ApplyToCrNames = applyToCrs

				secCrdToDeploy.Spec.LoginModules.PropertiesLoginModules = nil

				secCrdToDeploy.Spec.SecurityDomains.BrokerDomain = brokerv1beta1.BrokerDomainType{
					Name:         nil,
					LoginModules: nil,
				}

				securitySettings := []brokerv1beta1.BrokerSecuritySettingType{
					{
						Match: "Info2",
						Permissions: []brokerv1beta1.PermissionType{
							{
								OperationType: "send",
								Roles: []string{
									"guest",
								},
							},
							{
								OperationType: "consume",
								Roles: []string{
									"amq",
								},
							},
						},
					},
					{
						Match: "#",
						Permissions: []brokerv1beta1.PermissionType{
							{
								OperationType: "deleteNonDurableQueue",
								Roles: []string{
									"root",
									"admin",
								},
							},
							{
								OperationType: "createTempQueue",
								Roles: []string{
									"root",
								},
							},
						},
					},
				}
				secCrdToDeploy.Spec.SecuritySettings.Broker = securitySettings
			})

			By("checking the merge result")

			Eventually(func(g Gomega) {
				configHandler, _ := GetBrokerSecurityConfigHandler(nsName)
				securityHandler := configHandler.(*ActiveMQArtemisSecurityConfigHandler)
				g.Expect(securityHandler).NotTo(BeNil())

				secSpec := securityHandler.SecurityCR.Spec
				g.Expect(len(secSpec.ApplyToCrNames)).To(BeEquivalentTo(1))
				g.Expect(secSpec.ApplyToCrNames[0]).To(BeEquivalentTo(nsName.Name))

				secSettings := secSpec.SecuritySettings.Broker
				g.Expect(len(secSettings)).To(BeEquivalentTo(3))
				g.Expect(secSettings[0].Match).To(BeEquivalentTo("Info"))
				g.Expect(len(secSettings[0].Permissions)).To(BeEquivalentTo(2))
				g.Expect(secSettings[0].Permissions[0].OperationType).To(BeEquivalentTo("createDurableQueue"))
				g.Expect(len(secSettings[0].Permissions[0].Roles)).To(BeEquivalentTo(1))
				g.Expect(secSettings[0].Permissions[0].Roles[0]).To(BeEquivalentTo("amq"))
				g.Expect(secSettings[0].Permissions[1].OperationType).To(BeEquivalentTo("deleteDurableQueue"))
				g.Expect(len(secSettings[0].Permissions[1].Roles)).To(BeEquivalentTo(1))
				g.Expect(secSettings[0].Permissions[1].Roles[0]).To(BeEquivalentTo("amq"))

				g.Expect(secSettings[1].Match).To(BeEquivalentTo("#"))
				g.Expect(len(secSettings[1].Permissions)).To(BeEquivalentTo(5))
				g.Expect(secSettings[1].Permissions[0].OperationType).To(BeEquivalentTo("createDurableQueue"))
				g.Expect(len(secSettings[1].Permissions[0].Roles)).To(BeEquivalentTo(1))
				g.Expect(secSettings[1].Permissions[0].Roles[0]).To(BeEquivalentTo("root"))
				g.Expect(secSettings[1].Permissions[1].OperationType).To(BeEquivalentTo("deleteDurableQueue"))
				g.Expect(len(secSettings[1].Permissions[1].Roles)).To(BeEquivalentTo(1))
				g.Expect(secSettings[1].Permissions[1].Roles[0]).To(BeEquivalentTo("root"))
				g.Expect(secSettings[1].Permissions[2].OperationType).To(BeEquivalentTo("createNonDurableQueue"))
				g.Expect(len(secSettings[1].Permissions[2].Roles)).To(BeEquivalentTo(1))
				g.Expect(secSettings[1].Permissions[2].Roles[0]).To(BeEquivalentTo("root"))
				g.Expect(secSettings[1].Permissions[3].OperationType).To(BeEquivalentTo("deleteNonDurableQueue"))
				g.Expect(len(secSettings[1].Permissions[3].Roles)).To(BeEquivalentTo(2))
				g.Expect(secSettings[1].Permissions[3].Roles[0]).To(BeEquivalentTo("root"))
				g.Expect(secSettings[1].Permissions[3].Roles[1]).To(BeEquivalentTo("admin"))
				g.Expect(secSettings[1].Permissions[4].OperationType).To(BeEquivalentTo("createTempQueue"))
				g.Expect(len(secSettings[1].Permissions[4].Roles)).To(BeEquivalentTo(1))
				g.Expect(secSettings[1].Permissions[4].Roles[0]).To(BeEquivalentTo("root"))

				g.Expect(secSettings[2].Match).To(BeEquivalentTo("Info2"))
				g.Expect(len(secSettings[2].Permissions)).To(BeEquivalentTo(2))
				g.Expect(secSettings[2].Permissions[0].OperationType).To(BeEquivalentTo("send"))
				g.Expect(len(secSettings[2].Permissions[0].Roles)).To(BeEquivalentTo(1))
				g.Expect(secSettings[2].Permissions[0].Roles[0]).To(BeEquivalentTo("guest"))
				g.Expect(secSettings[2].Permissions[1].OperationType).To(BeEquivalentTo("consume"))
				g.Expect(len(secSettings[2].Permissions[1].Roles)).To(BeEquivalentTo(1))
				g.Expect(secSettings[2].Permissions[1].Roles[0]).To(BeEquivalentTo("amq"))
			}, timeout, interval).Should(Succeed())

			By("checking effective security CR for ex-aao2")

			Eventually(func(g Gomega) {
				configHandler, _ := GetBrokerSecurityConfigHandler(ns2Name)
				securityHandler := configHandler.(*ActiveMQArtemisSecurityConfigHandler)
				g.Expect(securityHandler).NotTo(BeNil())

				secSpec := securityHandler.SecurityCR.Spec
				g.Expect(len(secSpec.ApplyToCrNames)).To(BeEquivalentTo(1))
				g.Expect(secSpec.ApplyToCrNames[0]).To(BeEquivalentTo(ns2Name.Name))

				secSettings := secSpec.SecuritySettings.Broker

				g.Expect(len(secSettings)).To(BeEquivalentTo(2))

				g.Expect(secSettings[0].Match).To(BeEquivalentTo("Info2"))
				g.Expect(len(secSettings[0].Permissions)).To(BeEquivalentTo(2))
				g.Expect(secSettings[0].Permissions[0].OperationType).To(BeEquivalentTo("send"))
				g.Expect(len(secSettings[0].Permissions[0].Roles)).To(BeEquivalentTo(1))
				g.Expect(secSettings[0].Permissions[0].Roles[0]).To(BeEquivalentTo("guest"))
				g.Expect(secSettings[0].Permissions[1].OperationType).To(BeEquivalentTo("consume"))
				g.Expect(len(secSettings[0].Permissions[1].Roles)).To(BeEquivalentTo(1))
				g.Expect(secSettings[0].Permissions[1].Roles[0]).To(BeEquivalentTo("amq"))

				g.Expect(secSettings[1].Match).To(BeEquivalentTo("#"))
				g.Expect(len(secSettings[1].Permissions)).To(BeEquivalentTo(2))
				g.Expect(secSettings[1].Permissions[0].OperationType).To(BeEquivalentTo("deleteNonDurableQueue"))
				g.Expect(len(secSettings[1].Permissions[0].Roles)).To(BeEquivalentTo(2))
				g.Expect(secSettings[1].Permissions[0].Roles[0]).To(BeEquivalentTo("root"))
				g.Expect(secSettings[1].Permissions[0].Roles[1]).To(BeEquivalentTo("admin"))
				g.Expect(secSettings[1].Permissions[1].OperationType).To(BeEquivalentTo("createTempQueue"))
				g.Expect(len(secSettings[1].Permissions[1].Roles)).To(BeEquivalentTo(1))
				g.Expect(secSettings[1].Permissions[1].Roles[0]).To(BeEquivalentTo("root"))
			}, timeout, interval).Should(Succeed())

			By("delete 2nd security CR")
			Expect(k8sClient.Delete(ctx, createdSecCrd2)).Should(Succeed())
			Eventually(func() bool {
				return checkCrdDeleted(secCrd2.Name, defaultNamespace, createdSecCrd2)
			}, timeout, interval).Should(BeTrue())

			// security for ex-aao2 is gone
			Eventually(func(g Gomega) {
				configHandler, _ := GetBrokerSecurityConfigHandler(ns2Name)
				securityHandler := configHandler.(*ActiveMQArtemisSecurityConfigHandler)
				g.Expect(securityHandler).To(BeNil())
			}, timeout, interval).Should(Succeed())

			// security for ex-aao restored to first cr
			Eventually(func(g Gomega) {
				configHandler, _ := GetBrokerSecurityConfigHandler(nsName)
				securityHandler := configHandler.(*ActiveMQArtemisSecurityConfigHandler)
				g.Expect(securityHandler).NotTo(BeNil())

				secSpec := securityHandler.SecurityCR.Spec
				g.Expect(reflect.DeepEqual(secSpec, secCrd1.Spec)).To(BeTrue())
			}, timeout, interval).Should(Succeed())

			// delete 1st security CR
			Expect(k8sClient.Delete(ctx, createdSecCrd1)).Should(Succeed())
			Eventually(func() bool {
				return checkCrdDeleted(secCrd1.Name, defaultNamespace, createdSecCrd1)
			}, timeout, interval).Should(BeTrue())
		})

		It("testing security domain merge", Label("sec-domain-merge"), func() {
			nsName := types.NamespacedName{
				Name:      "ex-aao",
				Namespace: defaultNamespace,
			}

			ns2Name := types.NamespacedName{
				Name:      "ex-aao2",
				Namespace: defaultNamespace,
			}

			By("deploy 2 security CRs to merge with security domains")
			secCrd1, createdSecCrd1 := DeploySecurity("", defaultNamespace, func(secCrdToDeploy *brokerv1beta1.ActiveMQArtemisSecurity) {
				applyToCrs := []string{"ex-aao"}
				secCrdToDeploy.Spec.ApplyToCrNames = applyToCrs

				propLoginModules := []brokerv1beta1.PropertiesLoginModuleType{
					{
						Name: propModuleName1,
						Users: []brokerv1beta1.UserType{
							{
								Name:     "user1",
								Password: &password1,
								Roles: []string{
									"root1",
								},
							},
						},
					},
				}
				secCrdToDeploy.Spec.LoginModules.PropertiesLoginModules = propLoginModules

				brokerDomain := brokerv1beta1.BrokerDomainType{
					Name: &domainName,
					LoginModules: []brokerv1beta1.LoginModuleReferenceType{
						{
							Name:   &propModuleName1,
							Flag:   &moduleFlagRequired,
							Debug:  &boolTrue,
							Reload: &boolTrue,
						},
					},
				}

				secCrdToDeploy.Spec.SecurityDomains.BrokerDomain = brokerDomain
			})

			secCrd2, createdSecCrd2 := DeploySecurity("", defaultNamespace, func(secCrdToDeploy *brokerv1beta1.ActiveMQArtemisSecurity) {
				applyToCrs := []string{"ex-aao", "ex-aao2"}
				secCrdToDeploy.Spec.ApplyToCrNames = applyToCrs

				propLoginModules := []brokerv1beta1.PropertiesLoginModuleType{
					{
						Name: propModuleName2,
						Users: []brokerv1beta1.UserType{
							{
								Name:     "user2",
								Password: &password2,
								Roles: []string{
									"root2",
								},
							},
						},
					},
				}
				secCrdToDeploy.Spec.LoginModules.PropertiesLoginModules = propLoginModules

				brokerDomain := brokerv1beta1.BrokerDomainType{
					Name: &domainName1,
					LoginModules: []brokerv1beta1.LoginModuleReferenceType{
						{
							Name:  &propModuleName2,
							Flag:  &moduleFlagSufficient,
							Debug: &boolTrue,
						},
					},
				}

				secCrdToDeploy.Spec.SecurityDomains.BrokerDomain = brokerDomain

				consoleDomain := brokerv1beta1.BrokerDomainType{
					Name: &consoleDomainName,
					LoginModules: []brokerv1beta1.LoginModuleReferenceType{
						{
							Name:  &propModuleName1,
							Flag:  &moduleFlagRequired,
							Debug: &boolTrue,
						},
					},
				}

				secCrdToDeploy.Spec.SecurityDomains.ConsoleDomain = consoleDomain
			})

			By("checking the merge result")

			Eventually(func(g Gomega) {
				configHandler, _ := GetBrokerSecurityConfigHandler(nsName)
				securityHandler := configHandler.(*ActiveMQArtemisSecurityConfigHandler)
				g.Expect(securityHandler).NotTo(BeNil())

				secSpec := securityHandler.SecurityCR.Spec
				g.Expect(len(secSpec.ApplyToCrNames)).To(BeEquivalentTo(1))
				g.Expect(secSpec.ApplyToCrNames[0]).To(BeEquivalentTo(nsName.Name))

				g.Expect(len(secSpec.LoginModules.PropertiesLoginModules)).To(BeEquivalentTo(2))
				g.Expect(secSpec.LoginModules.PropertiesLoginModules[0].Name).To(BeEquivalentTo(propModuleName1))

				users := secSpec.LoginModules.PropertiesLoginModules[0].Users
				g.Expect(len(users)).To(BeEquivalentTo(1))

				g.Expect(users[0].Name).To(BeEquivalentTo("user1"))
				g.Expect(*users[0].Password).To(BeEquivalentTo(password1))
				g.Expect(len(users[0].Roles)).To(BeEquivalentTo(1))
				g.Expect(CheckRolesMatch(users[0].Roles, "root1"))

				g.Expect(secSpec.LoginModules.PropertiesLoginModules[1].Name).To(BeEquivalentTo(propModuleName2))

				users = secSpec.LoginModules.PropertiesLoginModules[1].Users
				g.Expect(len(users)).To(BeEquivalentTo(1))

				g.Expect(users[0].Name).To(BeEquivalentTo("user2"))
				g.Expect(*users[0].Password).To(BeEquivalentTo(password2))
				g.Expect(len(users[0].Roles)).To(BeEquivalentTo(1))
				g.Expect(CheckRolesMatch(users[0].Roles, "root2"))

				g.Expect(*secSpec.SecurityDomains.BrokerDomain.Name).To(BeEquivalentTo(domainName1))
				g.Expect(len(secSpec.SecurityDomains.BrokerDomain.LoginModules)).To(BeEquivalentTo(1))
				g.Expect(*secSpec.SecurityDomains.BrokerDomain.LoginModules[0].Name).To(BeEquivalentTo(propModuleName2))
				g.Expect(*secSpec.SecurityDomains.BrokerDomain.LoginModules[0].Flag).To(BeEquivalentTo(moduleFlagSufficient))
				g.Expect(*secSpec.SecurityDomains.BrokerDomain.LoginModules[0].Debug).To(BeEquivalentTo(boolTrue))
				g.Expect(secSpec.SecurityDomains.BrokerDomain.LoginModules[0].Reload).To(BeNil())

				g.Expect(*secSpec.SecurityDomains.ConsoleDomain.Name).To(BeEquivalentTo(consoleDomainName))
				g.Expect(len(secSpec.SecurityDomains.ConsoleDomain.LoginModules)).To(BeEquivalentTo(1))
				g.Expect(*secSpec.SecurityDomains.ConsoleDomain.LoginModules[0].Name).To(BeEquivalentTo(propModuleName1))
				g.Expect(*secSpec.SecurityDomains.ConsoleDomain.LoginModules[0].Flag).To(BeEquivalentTo(moduleFlagRequired))
				g.Expect(*secSpec.SecurityDomains.ConsoleDomain.LoginModules[0].Debug).To(BeEquivalentTo(boolTrue))
				g.Expect(secSpec.SecurityDomains.ConsoleDomain.LoginModules[0].Reload).To(BeNil())

			}, timeout, interval).Should(Succeed())

			By("checking effective security CR for ex-aao2")

			Eventually(func(g Gomega) {
				configHandler, _ := GetBrokerSecurityConfigHandler(ns2Name)
				securityHandler := configHandler.(*ActiveMQArtemisSecurityConfigHandler)
				g.Expect(securityHandler).NotTo(BeNil())

				secSpec := securityHandler.SecurityCR.Spec
				g.Expect(len(secSpec.ApplyToCrNames)).To(BeEquivalentTo(1))
				g.Expect(secSpec.ApplyToCrNames[0]).To(BeEquivalentTo(ns2Name.Name))

				g.Expect(len(secSpec.LoginModules.PropertiesLoginModules)).To(BeEquivalentTo(1))
				g.Expect(secSpec.LoginModules.PropertiesLoginModules[0].Name).To(BeEquivalentTo(propModuleName2))

				users := secSpec.LoginModules.PropertiesLoginModules[0].Users
				g.Expect(len(users)).To(BeEquivalentTo(1))

				g.Expect(*users[0].Password).To(BeEquivalentTo(password2))
				g.Expect(len(users[0].Roles)).To(BeEquivalentTo(1))
				g.Expect(CheckRolesMatch(users[0].Roles, "root2"))

				g.Expect(*secSpec.SecurityDomains.BrokerDomain.Name).To(BeEquivalentTo(domainName1))
				g.Expect(len(secSpec.SecurityDomains.BrokerDomain.LoginModules)).To(BeEquivalentTo(1))
				g.Expect(*secSpec.SecurityDomains.BrokerDomain.LoginModules[0].Name).To(BeEquivalentTo(propModuleName2))
				g.Expect(*secSpec.SecurityDomains.BrokerDomain.LoginModules[0].Flag).To(BeEquivalentTo(moduleFlagSufficient))
				g.Expect(*secSpec.SecurityDomains.BrokerDomain.LoginModules[0].Debug).To(BeEquivalentTo(boolTrue))
				g.Expect(secSpec.SecurityDomains.BrokerDomain.LoginModules[0].Reload).To(BeNil())

				g.Expect(*secSpec.SecurityDomains.ConsoleDomain.Name).To(BeEquivalentTo(consoleDomainName))
				g.Expect(len(secSpec.SecurityDomains.ConsoleDomain.LoginModules)).To(BeEquivalentTo(1))
				g.Expect(*secSpec.SecurityDomains.ConsoleDomain.LoginModules[0].Name).To(BeEquivalentTo(propModuleName1))
				g.Expect(*secSpec.SecurityDomains.ConsoleDomain.LoginModules[0].Flag).To(BeEquivalentTo(moduleFlagRequired))
				g.Expect(*secSpec.SecurityDomains.ConsoleDomain.LoginModules[0].Debug).To(BeEquivalentTo(boolTrue))
				g.Expect(secSpec.SecurityDomains.ConsoleDomain.LoginModules[0].Reload).To(BeNil())

			}, timeout, interval).Should(Succeed())

			By("delete 2nd security CR")
			Expect(k8sClient.Delete(ctx, createdSecCrd2)).Should(Succeed())
			Eventually(func() bool {
				return checkCrdDeleted(secCrd2.Name, defaultNamespace, createdSecCrd2)
			}, timeout, interval).Should(BeTrue())

			// security for ex-aao2 is gone
			Eventually(func(g Gomega) {
				configHandler, _ := GetBrokerSecurityConfigHandler(ns2Name)
				securityHandler := configHandler.(*ActiveMQArtemisSecurityConfigHandler)
				g.Expect(securityHandler).To(BeNil())
			}, timeout, interval).Should(Succeed())

			// security for ex-aao restored to first cr
			Eventually(func(g Gomega) {
				configHandler, _ := GetBrokerSecurityConfigHandler(nsName)
				securityHandler := configHandler.(*ActiveMQArtemisSecurityConfigHandler)
				g.Expect(securityHandler).NotTo(BeNil())

				secSpec := securityHandler.SecurityCR.Spec
				g.Expect(reflect.DeepEqual(secSpec, secCrd1.Spec)).To(BeTrue())
			}, timeout, interval).Should(Succeed())

			// delete 1st security CR
			Expect(k8sClient.Delete(ctx, createdSecCrd1)).Should(Succeed())
			Eventually(func() bool {
				return checkCrdDeleted(secCrd1.Name, defaultNamespace, createdSecCrd1)
			}, timeout, interval).Should(BeTrue())
		})

		It("testing properties login module merge", Label("prop-module-merge"), func() {
			By("deploying 2 security CRs with properties login modules")

			secCrd1, createdSecCrd1 := DeploySecurity("", defaultNamespace, func(secCrdToDeploy *brokerv1beta1.ActiveMQArtemisSecurity) {
				applyToCrs := []string{"ex-aao"}
				secCrdToDeploy.Spec.ApplyToCrNames = applyToCrs

				propLoginModules := []brokerv1beta1.PropertiesLoginModuleType{
					{
						Name: propModuleName,
						Users: []brokerv1beta1.UserType{
							{
								Name:     "user1",
								Password: &password1,
								Roles: []string{
									"role1",
									"role2",
								},
							},
							{
								Name:     "user2",
								Password: &password2,
								Roles: []string{
									"role3",
									"role4",
								},
							},
							{
								Name:     "user3",
								Password: &password3,
								Roles: []string{
									"role1",
									"role3",
									"role5",
								},
							},
						},
					},
				}
				secCrdToDeploy.Spec.LoginModules.PropertiesLoginModules = propLoginModules

				brokerDomain := brokerv1beta1.BrokerDomainType{
					Name: &domainName,
					LoginModules: []brokerv1beta1.LoginModuleReferenceType{
						{
							Name:  &propModuleName,
							Flag:  &moduleFlagSufficient,
							Debug: &boolFalse,
						},
					},
				}

				secCrdToDeploy.Spec.SecurityDomains.BrokerDomain = brokerDomain
			})

			secCrd2, createdSecCrd2 := DeploySecurity("", defaultNamespace, func(secCrdToDeploy *brokerv1beta1.ActiveMQArtemisSecurity) {
				applyToCrs := []string{"ex-aao", "ex-aao2"}
				secCrdToDeploy.Spec.ApplyToCrNames = applyToCrs

				propLoginModules := []brokerv1beta1.PropertiesLoginModuleType{
					{
						Name: propModuleName,
						Users: []brokerv1beta1.UserType{
							{
								Name:     "user1",
								Password: &password4,
								Roles: []string{
									"role3",
									"role4",
									"role5",
								},
							},
							{
								Name:     "user2",
								Password: &password3,
								Roles: []string{
									"role1",
									"role2",
									"role3",
									"role5",
								},
							},
							{
								Name:     "user3",
								Password: &password2,
								Roles: []string{
									"role2",
									"role3",
									"role4",
								},
							},
							{
								Name:     "user4",
								Password: &password1,
								Roles: []string{
									"role1",
									"role2",
									"role3",
									"role4",
									"role5",
								},
							},
						},
					},
				}
				secCrdToDeploy.Spec.LoginModules.PropertiesLoginModules = propLoginModules

				brokerDomain := brokerv1beta1.BrokerDomainType{
					Name: &domainName1,
					LoginModules: []brokerv1beta1.LoginModuleReferenceType{
						{
							Name:  &propModuleName,
							Flag:  &moduleFlagRequired,
							Debug: &boolTrue,
						},
					},
				}

				secCrdToDeploy.Spec.SecurityDomains.BrokerDomain = brokerDomain
			})

			By("checking the merger util works")
			nsName := types.NamespacedName{
				Name:      "ex-aao",
				Namespace: defaultNamespace,
			}

			By("checking effective security CR")

			Eventually(func(g Gomega) {
				configHandler, _ := GetBrokerSecurityConfigHandler(nsName)
				g.Expect(configHandler).NotTo(BeNil())
				securityHandler := configHandler.(*ActiveMQArtemisSecurityConfigHandler)

				secSpec := securityHandler.SecurityCR.Spec
				g.Expect(len(secSpec.ApplyToCrNames)).To(BeEquivalentTo(1))
				g.Expect(secSpec.ApplyToCrNames[0]).To(BeEquivalentTo(nsName.Name))

				g.Expect(len(secSpec.LoginModules.PropertiesLoginModules)).To(BeEquivalentTo(1))
				g.Expect(secSpec.LoginModules.PropertiesLoginModules[0].Name).To(BeEquivalentTo(propModuleName))

				users := secSpec.LoginModules.PropertiesLoginModules[0].Users
				g.Expect(len(users)).To(BeEquivalentTo(4))

				userMap := make(map[string]brokerv1beta1.UserType)
				for _, user := range users {
					userMap[user.Name] = user
				}

				user1, ok1 := userMap["user1"]
				g.Expect(ok1).To(BeTrue())
				g.Expect(*user1.Password).To(BeEquivalentTo(password4))
				g.Expect(len(user1.Roles)).To(BeEquivalentTo(5))
				g.Expect(CheckRolesMatch(user1.Roles, "role1", "role2", "role3", "role4", "role5"))

				user2, ok2 := userMap["user2"]
				g.Expect(ok2).To(BeTrue())
				g.Expect(*user2.Password).To(BeEquivalentTo(password3))
				g.Expect(len(user2.Roles)).To(BeEquivalentTo(5))
				g.Expect(CheckRolesMatch(user2.Roles, "role1", "role2", "role3", "role4", "role5"))

				user3, ok3 := userMap["user3"]
				g.Expect(ok3).To(BeTrue())
				g.Expect(*user3.Password).To(BeEquivalentTo(password2))
				g.Expect(len(user3.Roles)).To(BeEquivalentTo(5))
				g.Expect(CheckRolesMatch(user3.Roles, "role1", "role2", "role3", "role4", "role5"))

				user4, ok4 := userMap["user4"]
				g.Expect(ok4).To(BeTrue())
				g.Expect(*user4.Password).To(BeEquivalentTo(password1))
				g.Expect(len(user4.Roles)).To(BeEquivalentTo(5))
				g.Expect(CheckRolesMatch(user4.Roles, "role1", "role2", "role3", "role4", "role5"))

				g.Expect(*secSpec.SecurityDomains.BrokerDomain.Name).To(BeEquivalentTo(domainName1))
				g.Expect(len(secSpec.SecurityDomains.BrokerDomain.LoginModules)).To(BeEquivalentTo(1))
				g.Expect(*secSpec.SecurityDomains.BrokerDomain.LoginModules[0].Name).To(BeEquivalentTo(propModuleName))
				g.Expect(*secSpec.SecurityDomains.BrokerDomain.LoginModules[0].Flag).To(BeEquivalentTo(moduleFlagRequired))
				g.Expect(*secSpec.SecurityDomains.BrokerDomain.LoginModules[0].Debug).To(BeEquivalentTo(boolTrue))

			}, timeout, interval).Should(Succeed())

			ns2Name := types.NamespacedName{
				Name:      "ex-aao2",
				Namespace: defaultNamespace,
			}

			By("checking effective security CR for ex-aao2")

			Eventually(func(g Gomega) {
				configHandler, _ := GetBrokerSecurityConfigHandler(ns2Name)
				g.Expect(configHandler).NotTo(BeNil())
				securityHandler := configHandler.(*ActiveMQArtemisSecurityConfigHandler)

				secSpec := securityHandler.SecurityCR.Spec
				g.Expect(len(secSpec.ApplyToCrNames)).To(BeEquivalentTo(1))
				g.Expect(secSpec.ApplyToCrNames[0]).To(BeEquivalentTo(ns2Name.Name))

				g.Expect(len(secSpec.LoginModules.PropertiesLoginModules)).To(BeEquivalentTo(1))
				g.Expect(secSpec.LoginModules.PropertiesLoginModules[0].Name).To(BeEquivalentTo(propModuleName))

				users := secSpec.LoginModules.PropertiesLoginModules[0].Users
				g.Expect(len(users)).To(BeEquivalentTo(4))

				userMap := make(map[string]brokerv1beta1.UserType)
				for _, user := range users {
					userMap[user.Name] = user
				}

				user1, ok1 := userMap["user1"]
				g.Expect(ok1).To(BeTrue())
				g.Expect(*user1.Password).To(BeEquivalentTo(password4))
				g.Expect(len(user1.Roles)).To(BeEquivalentTo(3))
				g.Expect(CheckRolesMatch(user1.Roles, "role3", "role4", "role5"))

				user2, ok2 := userMap["user2"]
				g.Expect(ok2).To(BeTrue())
				g.Expect(*user2.Password).To(BeEquivalentTo(password3))
				g.Expect(len(user2.Roles)).To(BeEquivalentTo(4))
				g.Expect(CheckRolesMatch(user2.Roles, "role1", "role2", "role3", "role5"))

				user3, ok3 := userMap["user3"]
				g.Expect(ok3).To(BeTrue())
				g.Expect(*user3.Password).To(BeEquivalentTo(password2))
				g.Expect(len(user3.Roles)).To(BeEquivalentTo(3))
				g.Expect(CheckRolesMatch(user3.Roles, "role2", "role3", "role4"))

				user4, ok4 := userMap["user4"]
				g.Expect(ok4).To(BeTrue())
				g.Expect(*user4.Password).To(BeEquivalentTo(password1))
				g.Expect(len(user4.Roles)).To(BeEquivalentTo(5))
				g.Expect(CheckRolesMatch(user4.Roles, "role1", "role2", "role3", "role4", "role5"))

				g.Expect(*secSpec.SecurityDomains.BrokerDomain.Name).To(BeEquivalentTo(domainName1))
				g.Expect(len(secSpec.SecurityDomains.BrokerDomain.LoginModules)).To(BeEquivalentTo(1))
				g.Expect(*secSpec.SecurityDomains.BrokerDomain.LoginModules[0].Name).To(BeEquivalentTo(propModuleName))
				g.Expect(*secSpec.SecurityDomains.BrokerDomain.LoginModules[0].Flag).To(BeEquivalentTo(moduleFlagRequired))
				g.Expect(*secSpec.SecurityDomains.BrokerDomain.LoginModules[0].Debug).To(BeEquivalentTo(boolTrue))

			}, timeout, interval).Should(Succeed())

			if os.Getenv("USE_EXISTING_CLUSTER") == "true" {

				By("deploy broker")
				brokerCr, createdCr := DeployCustomBroker(defaultNamespace, func(candidate *brokerv1beta1.ActiveMQArtemis) {
					candidate.Spec.DeploymentPlan.ReadinessProbe = &corev1.Probe{
						InitialDelaySeconds: 5,
						TimeoutSeconds:      5,
						PeriodSeconds:       5,
					}
					candidate.Name = "ex-aao"

				})

				By("fetch the statefulset")
				key := types.NamespacedName{
					Name:      namer.CrToSS(brokerCr.Name),
					Namespace: defaultNamespace,
				}
				currentSS := &appsv1.StatefulSet{}
				Eventually(func(g Gomega) {
					g.Expect(k8sClient.Get(ctx, key, currentSS)).Should(Succeed())
					g.Expect(currentSS.Status.ReadyReplicas).Should(BeEquivalentTo(1))

				}, existingClusterTimeout, existingClusterInterval).Should(Succeed())

				podWithOrdinal := namer.CrToSS(brokerCr.Name) + "-0"

				By("checking all users in the artemis-users.properties")
				command := []string{"cat", "amq-broker/etc/artemis-users.properties"}
				Eventually(func(g Gomega) {
					stdOutContent := ExecOnPod(podWithOrdinal, brokerCr.Name, defaultNamespace, command, g)
					g.Expect(stdOutContent).Should(ContainSubstring("user1 ="), stdOutContent)
					g.Expect(stdOutContent).Should(ContainSubstring("user2 ="), stdOutContent)
					g.Expect(stdOutContent).Should(ContainSubstring("user3 ="), stdOutContent)
					g.Expect(stdOutContent).Should(ContainSubstring("user4 ="), stdOutContent)
				}, existingClusterTimeout, existingClusterInterval).Should(Succeed())

				By("checking all roles in the artemis-roles.properties")
				command = []string{"cat", "amq-broker/etc/artemis-roles.properties"}
				Eventually(func(g Gomega) {
					stdOutContent := ExecOnPod(podWithOrdinal, brokerCr.Name, defaultNamespace, command, g)
					g.Expect(stdOutContent).Should(ContainSubstring("role1 = user1,user2,user3,user4"), stdOutContent)
					g.Expect(stdOutContent).Should(ContainSubstring("role2 = user1,user2,user3,user4"), stdOutContent)
					g.Expect(stdOutContent).Should(ContainSubstring("role3 = user1,user2,user3,user4"), stdOutContent)
					g.Expect(stdOutContent).Should(ContainSubstring("role4 = user1,user2,user3,user4"), stdOutContent)
				}, existingClusterTimeout, existingClusterInterval).Should(Succeed())

				Expect(k8sClient.Delete(ctx, createdCr)).Should(Succeed())
				Eventually(func() bool {
					return checkCrdDeleted(brokerCr.Name, defaultNamespace, createdCr)
				}, existingClusterTimeout, existingClusterInterval).Should(BeTrue())

				By("deploy another broker and check security")
				brokerCr2, createdCr2 := DeployCustomBroker(defaultNamespace, func(candidate *brokerv1beta1.ActiveMQArtemis) {
					candidate.Spec.DeploymentPlan.ReadinessProbe = &corev1.Probe{
						InitialDelaySeconds: 5,
						TimeoutSeconds:      5,
						PeriodSeconds:       5,
					}
					candidate.Name = "ex-aao2"

				})

				By("fetch the statefulset")
				key2 := types.NamespacedName{
					Name:      namer.CrToSS(brokerCr2.Name),
					Namespace: defaultNamespace,
				}
				currentSS2 := &appsv1.StatefulSet{}

				Eventually(func(g Gomega) {
					g.Expect(k8sClient.Get(ctx, key2, currentSS2)).Should(Succeed())
					g.Expect(currentSS.Status.ReadyReplicas).Should(BeEquivalentTo(1))

				}, existingClusterTimeout, existingClusterInterval).Should(Succeed())

				pod2WithOrdinal := namer.CrToSS(brokerCr2.Name) + "-0"

				By("checking all users in the artemis-users.properties")
				command = []string{"cat", "amq-broker/etc/artemis-users.properties"}
				Eventually(func(g Gomega) {
					stdOutContent := ExecOnPod(pod2WithOrdinal, brokerCr2.Name, defaultNamespace, command, g)
					g.Expect(stdOutContent).Should(ContainSubstring("user1 ="), stdOutContent)
					g.Expect(stdOutContent).Should(ContainSubstring("user2 ="), stdOutContent)
					g.Expect(stdOutContent).Should(ContainSubstring("user3 ="), stdOutContent)
					g.Expect(stdOutContent).Should(ContainSubstring("user4 ="), stdOutContent)
				}, existingClusterTimeout, existingClusterInterval).Should(Succeed())

				By("checking all roles in the artemis-roles.properties")
				command = []string{"cat", "amq-broker/etc/artemis-roles.properties"}
				Eventually(func(g Gomega) {
					stdOutContent := ExecOnPod(pod2WithOrdinal, brokerCr2.Name, defaultNamespace, command, g)
					g.Expect(stdOutContent).Should(ContainSubstring("role1 = user2,user4"), stdOutContent)
					g.Expect(stdOutContent).Should(ContainSubstring("role2 = user2,user3,user4"), stdOutContent)
					g.Expect(stdOutContent).Should(ContainSubstring("role3 = user1,user2,user3,user4"), stdOutContent)
					g.Expect(stdOutContent).Should(ContainSubstring("role4 = user1,user3,user4"), stdOutContent)
				}, existingClusterTimeout, existingClusterInterval).Should(Succeed())

				Expect(k8sClient.Delete(ctx, createdCr2)).Should(Succeed())
				Eventually(func() bool {
					return checkCrdDeleted(brokerCr2.Name, defaultNamespace, createdCr2)
				}, existingClusterTimeout, existingClusterInterval).Should(BeTrue())
			}

			By("checking resources get removed")
			Expect(k8sClient.Delete(ctx, createdSecCrd1)).Should(Succeed())
			Eventually(func() bool {
				return checkCrdDeleted(secCrd1.Name, defaultNamespace, createdSecCrd1)
			}, timeout, interval).Should(BeTrue())

			Expect(k8sClient.Delete(ctx, createdSecCrd2)).Should(Succeed())
			Eventually(func() bool {
				return checkCrdDeleted(secCrd2.Name, defaultNamespace, createdSecCrd2)
			}, timeout, interval).Should(BeTrue())
		})

		It("testing general login modules merge", Label("multi-module-merge"), func() {
			By("deploying 1 security CR with one properties login module")

			secCrd1, createdSecCrd1 := DeploySecurity("", defaultNamespace, func(secCrdToDeploy *brokerv1beta1.ActiveMQArtemisSecurity) {
				applyToCrs := []string{"ex-aao"}
				secCrdToDeploy.Spec.ApplyToCrNames = applyToCrs

				propLoginModules := []brokerv1beta1.PropertiesLoginModuleType{
					{
						Name: propModuleName,
						Users: []brokerv1beta1.UserType{
							{
								Name:     "user1",
								Password: &password1,
								Roles: []string{
									"root",
								},
							},
						},
					},
				}
				secCrdToDeploy.Spec.LoginModules.PropertiesLoginModules = propLoginModules

				secCrdToDeploy.Spec.SecurityDomains.BrokerDomain = brokerv1beta1.BrokerDomainType{
					Name:         nil,
					LoginModules: nil,
				}
			})

			By("checking the security config")
			nsName := types.NamespacedName{
				Name:      "ex-aao",
				Namespace: defaultNamespace,
			}

			By("checking first security CR")
			Eventually(func(g Gomega) {
				configHandler, _ := GetBrokerSecurityConfigHandler(nsName)
				g.Expect(configHandler).NotTo(BeNil())
				securityHandler := configHandler.(*ActiveMQArtemisSecurityConfigHandler)

				secSpec := securityHandler.SecurityCR.Spec
				g.Expect(len(secSpec.ApplyToCrNames)).To(BeEquivalentTo(1))
				g.Expect(secSpec.ApplyToCrNames[0]).To(BeEquivalentTo(nsName.Name))

				g.Expect(len(secSpec.LoginModules.PropertiesLoginModules)).To(BeEquivalentTo(1))
				g.Expect(secSpec.LoginModules.PropertiesLoginModules[0].Name).To(BeEquivalentTo(propModuleName))

				users := secSpec.LoginModules.PropertiesLoginModules[0].Users
				g.Expect(len(users)).To(BeEquivalentTo(1))

				userMap := make(map[string]brokerv1beta1.UserType)
				for _, user := range users {
					userMap[user.Name] = user
				}

				user1, ok1 := userMap["user1"]
				g.Expect(ok1).To(BeTrue())
				g.Expect(*user1.Password).To(BeEquivalentTo(password1))
				g.Expect(len(user1.Roles)).To(BeEquivalentTo(1))
				g.Expect(CheckRolesMatch(user1.Roles, "root"))

				g.Expect(secSpec.SecurityDomains.BrokerDomain.Name).To(BeNil())
			}, timeout, interval).Should(Succeed())

			secCrd2, createdSecCrd2 := DeploySecurity("", defaultNamespace, func(secCrdToDeploy *brokerv1beta1.ActiveMQArtemisSecurity) {
				applyToCrs := []string{"ex-aao", "ex-aao2"}
				secCrdToDeploy.Spec.ApplyToCrNames = applyToCrs

				propLoginModules := []brokerv1beta1.PropertiesLoginModuleType{
					{
						Name: propModuleName2,
						Users: []brokerv1beta1.UserType{
							{
								Name:     "user2",
								Password: &password2,
								Roles: []string{
									"root2",
								},
							},
						},
					},
				}
				guestLoginModules := []brokerv1beta1.GuestLoginModuleType{
					{
						Name:      guestModuleName,
						GuestUser: &guestUser,
						GuestRole: &guestRole,
					},
				}
				secCrdToDeploy.Spec.LoginModules.PropertiesLoginModules = propLoginModules
				secCrdToDeploy.Spec.LoginModules.GuestLoginModules = guestLoginModules

				brokerDomain := brokerv1beta1.BrokerDomainType{
					Name: &domainName,
					LoginModules: []brokerv1beta1.LoginModuleReferenceType{
						{
							Name:  &propModuleName2,
							Flag:  &moduleFlagSufficient,
							Debug: &boolTrue,
						},
					},
				}

				secCrdToDeploy.Spec.SecurityDomains.BrokerDomain = brokerDomain
			})

			nsName = types.NamespacedName{
				Name:      "ex-aao",
				Namespace: defaultNamespace,
			}

			By("checking effective security CR")
			Eventually(func(g Gomega) {
				configHandler, _ := GetBrokerSecurityConfigHandler(nsName)
				g.Expect(configHandler).NotTo(BeNil())
				securityHandler := configHandler.(*ActiveMQArtemisSecurityConfigHandler)

				secSpec := securityHandler.SecurityCR.Spec
				g.Expect(len(secSpec.ApplyToCrNames)).To(BeEquivalentTo(1))
				g.Expect(secSpec.ApplyToCrNames[0]).To(BeEquivalentTo(nsName.Name))

				g.Expect(len(secSpec.LoginModules.PropertiesLoginModules)).To(BeEquivalentTo(2))

				g.Expect(secSpec.LoginModules.PropertiesLoginModules[0].Name).To(BeEquivalentTo(propModuleName))

				users := secSpec.LoginModules.PropertiesLoginModules[0].Users
				g.Expect(len(users)).To(BeEquivalentTo(1))

				userMap := make(map[string]brokerv1beta1.UserType)
				for _, user := range users {
					userMap[user.Name] = user
				}

				user1, ok1 := userMap["user1"]
				g.Expect(ok1).To(BeTrue())
				g.Expect(*user1.Password).To(BeEquivalentTo(password1))
				g.Expect(len(user1.Roles)).To(BeEquivalentTo(1))
				g.Expect(CheckRolesMatch(user1.Roles, "role1"))

				g.Expect(secSpec.LoginModules.PropertiesLoginModules[1].Name).To(BeEquivalentTo(propModuleName2))

				users = secSpec.LoginModules.PropertiesLoginModules[1].Users
				g.Expect(len(users)).To(BeEquivalentTo(1))

				userMap = make(map[string]brokerv1beta1.UserType)
				for _, user := range users {
					userMap[user.Name] = user
				}

				user2, ok2 := userMap["user2"]
				g.Expect(ok2).To(BeTrue())
				g.Expect(*user2.Password).To(BeEquivalentTo(password2))
				g.Expect(len(user1.Roles)).To(BeEquivalentTo(1))
				g.Expect(CheckRolesMatch(user1.Roles, "role2"))

				g.Expect(len(secSpec.LoginModules.GuestLoginModules)).To(BeEquivalentTo(1))
				g.Expect(*secSpec.LoginModules.GuestLoginModules[0].GuestUser).To(BeEquivalentTo(guestUser))
				g.Expect(*secSpec.LoginModules.GuestLoginModules[0].GuestRole).To(BeEquivalentTo(guestRole))

				g.Expect(*secSpec.SecurityDomains.BrokerDomain.Name).To(BeEquivalentTo(domainName))
				g.Expect(len(secSpec.SecurityDomains.BrokerDomain.LoginModules)).To(BeEquivalentTo(1))
				g.Expect(*secSpec.SecurityDomains.BrokerDomain.LoginModules[0].Name).To(BeEquivalentTo(propModuleName2))
				g.Expect(*secSpec.SecurityDomains.BrokerDomain.LoginModules[0].Flag).To(BeEquivalentTo(moduleFlagSufficient))
				g.Expect(*secSpec.SecurityDomains.BrokerDomain.LoginModules[0].Debug).To(BeEquivalentTo(boolTrue))

			}, timeout, interval).Should(Succeed())

			//3rd security cr
			secCrd3, createdSecCrd3 := DeploySecurity("", defaultNamespace, func(secCrdToDeploy *brokerv1beta1.ActiveMQArtemisSecurity) {
				applyToCrs := []string{"ex-aao"}
				secCrdToDeploy.Spec.ApplyToCrNames = applyToCrs

				guestLoginModules := []brokerv1beta1.GuestLoginModuleType{
					{
						Name:      guestModuleName,
						GuestUser: &guestUser,
						GuestRole: &guestRole1,
					},
				}
				secCrdToDeploy.Spec.LoginModules.GuestLoginModules = guestLoginModules
				secCrdToDeploy.Spec.LoginModules.PropertiesLoginModules = nil

				consoleDomain := brokerv1beta1.BrokerDomainType{
					Name: &consoleDomainName,
					LoginModules: []brokerv1beta1.LoginModuleReferenceType{
						{
							Name:  &propModuleName,
							Flag:  &moduleFlagSufficient,
							Debug: &boolTrue,
						},
					},
				}

				secCrdToDeploy.Spec.SecurityDomains.ConsoleDomain = consoleDomain

				secCrdToDeploy.Spec.SecurityDomains.BrokerDomain = brokerv1beta1.BrokerDomainType{
					Name:         nil,
					LoginModules: nil,
				}

				brokerSecuritySettings := []brokerv1beta1.BrokerSecuritySettingType{
					{
						Match: "#",
						Permissions: []brokerv1beta1.PermissionType{
							{
								OperationType: "createAddress",
								Roles: []string{
									"root2",
								},
							},
						},
					},
					{
						Match: "activemq.management.#",
						Permissions: []brokerv1beta1.PermissionType{
							{
								OperationType: "createNonDurableQueue",
								Roles: []string{
									"admin",
								},
							},
							{
								OperationType: "createAddress",
								Roles: []string{
									"admin",
								},
							},
							{
								OperationType: "manage",
								Roles: []string{
									"admin",
								},
							},
						},
					},
				}

				secCrdToDeploy.Spec.SecuritySettings.Broker = brokerSecuritySettings

				managementSecuritySettings := brokerv1beta1.ManagementSecuritySettingsType{
					HawtioRoles: []string{
						viewerRole,
					},
					Authorisation: brokerv1beta1.AuthorisationConfigType{
						AllowedList: []brokerv1beta1.AllowedListEntryType{
							{
								Domain: &hawtioDomain,
							},
						},
						RoleAccess: []brokerv1beta1.RoleAccessType{
							{
								Domain: &artemisDomain,
								AccessList: []brokerv1beta1.DefaultAccessType{
									{
										Method: &allListMethod,
										Roles: []string{
											"admin",
											"viewer",
										},
									},
								},
							},
						},
					},
				}
				secCrdToDeploy.Spec.SecuritySettings.Management = managementSecuritySettings
			})

			nsName = types.NamespacedName{
				Name:      "ex-aao",
				Namespace: defaultNamespace,
			}

			By("checking effective security CR after 3 security CRs are deployed")
			Eventually(func(g Gomega) {
				configHandler, appliedData := GetBrokerSecurityConfigHandler(nsName)
				g.Expect(configHandler).NotTo(BeNil())

				g.Expect(len(appliedData)).To(Equal(3))
				key1 := createdSecCrd1.Name + "." + defaultNamespace
				key2 := createdSecCrd2.Name + "." + defaultNamespace
				key3 := createdSecCrd3.Name + "." + defaultNamespace

				value1, value1Ok := appliedData[key1]
				g.Expect(value1Ok).To(BeTrue())
				g.Expect(value1).To(Equal(createdSecCrd1.ResourceVersion))

				value2, value2Ok := appliedData[key2]
				g.Expect(value2Ok).To(BeTrue())
				g.Expect(value2).To(Equal(createdSecCrd2.ResourceVersion))

				value3, value3Ok := appliedData[key3]
				g.Expect(value3Ok).To(BeTrue())
				g.Expect(value3).To(Equal(createdSecCrd3.ResourceVersion))

				securityHandler := configHandler.(*ActiveMQArtemisSecurityConfigHandler)

				secSpec := securityHandler.SecurityCR.Spec
				g.Expect(len(secSpec.ApplyToCrNames)).To(BeEquivalentTo(1))
				g.Expect(secSpec.ApplyToCrNames[0]).To(BeEquivalentTo(nsName.Name))

				g.Expect(len(secSpec.LoginModules.PropertiesLoginModules)).To(BeEquivalentTo(2))

				g.Expect(secSpec.LoginModules.PropertiesLoginModules[0].Name).To(BeEquivalentTo(propModuleName))

				users := secSpec.LoginModules.PropertiesLoginModules[0].Users
				g.Expect(len(users)).To(BeEquivalentTo(1))

				userMap := make(map[string]brokerv1beta1.UserType)
				for _, user := range users {
					userMap[user.Name] = user
				}

				user1, ok1 := userMap["user1"]
				g.Expect(ok1).To(BeTrue())
				g.Expect(*user1.Password).To(BeEquivalentTo(password1))
				g.Expect(len(user1.Roles)).To(BeEquivalentTo(1))
				g.Expect(CheckRolesMatch(user1.Roles, "role1"))

				g.Expect(secSpec.LoginModules.PropertiesLoginModules[1].Name).To(BeEquivalentTo(propModuleName2))

				users = secSpec.LoginModules.PropertiesLoginModules[1].Users
				g.Expect(len(users)).To(BeEquivalentTo(1))

				userMap = make(map[string]brokerv1beta1.UserType)
				for _, user := range users {
					userMap[user.Name] = user
				}

				user2, ok2 := userMap["user2"]
				g.Expect(ok2).To(BeTrue())
				g.Expect(*user2.Password).To(BeEquivalentTo(password2))
				g.Expect(len(user1.Roles)).To(BeEquivalentTo(1))
				g.Expect(CheckRolesMatch(user1.Roles, "role2"))

				g.Expect(len(secSpec.LoginModules.GuestLoginModules)).To(BeEquivalentTo(1))
				g.Expect(*secSpec.LoginModules.GuestLoginModules[0].GuestUser).To(BeEquivalentTo(guestUser))
				g.Expect(*secSpec.LoginModules.GuestLoginModules[0].GuestRole).To(BeEquivalentTo(guestRole1))

				g.Expect(*secSpec.SecurityDomains.BrokerDomain.Name).To(BeEquivalentTo(domainName))
				g.Expect(len(secSpec.SecurityDomains.BrokerDomain.LoginModules)).To(BeEquivalentTo(1))
				g.Expect(*secSpec.SecurityDomains.BrokerDomain.LoginModules[0].Name).To(BeEquivalentTo(propModuleName2))
				g.Expect(*secSpec.SecurityDomains.BrokerDomain.LoginModules[0].Flag).To(BeEquivalentTo(moduleFlagSufficient))
				g.Expect(*secSpec.SecurityDomains.BrokerDomain.LoginModules[0].Debug).To(BeEquivalentTo(boolTrue))

				g.Expect(*secSpec.SecurityDomains.ConsoleDomain.Name).To(BeEquivalentTo(consoleDomainName))
				g.Expect(len(secSpec.SecurityDomains.ConsoleDomain.LoginModules)).To(BeEquivalentTo(1))
				g.Expect(*secSpec.SecurityDomains.ConsoleDomain.LoginModules[0].Name).To(BeEquivalentTo(propModuleName))
				g.Expect(*secSpec.SecurityDomains.ConsoleDomain.LoginModules[0].Flag).To(BeEquivalentTo(moduleFlagSufficient))
				g.Expect(*secSpec.SecurityDomains.ConsoleDomain.LoginModules[0].Debug).To(BeEquivalentTo(boolTrue))

				brokerSecuritySettings := secSpec.SecuritySettings.Broker
				g.Expect(len(brokerSecuritySettings)).To(BeEquivalentTo(2))

				g.Expect(brokerSecuritySettings[0].Match).To(BeEquivalentTo("#"))
				g.Expect(len(brokerSecuritySettings[0].Permissions)).To(BeEquivalentTo(1))
				g.Expect(brokerSecuritySettings[0].Permissions[0].OperationType).To(BeEquivalentTo("createAddress"))
				g.Expect(len(brokerSecuritySettings[0].Permissions[0].Roles)).To(BeEquivalentTo(1))
				g.Expect(brokerSecuritySettings[0].Permissions[0].Roles[0]).To(BeEquivalentTo("root2"))

				g.Expect(brokerSecuritySettings[1].Match).To(BeEquivalentTo("activemq.management.#"))
				g.Expect(len(brokerSecuritySettings[1].Permissions)).To(BeEquivalentTo(3))
				g.Expect(brokerSecuritySettings[1].Permissions[0].OperationType).To(BeEquivalentTo("createNonDurableQueue"))
				g.Expect(len(brokerSecuritySettings[1].Permissions[0].Roles)).To(BeEquivalentTo(1))
				g.Expect(brokerSecuritySettings[1].Permissions[0].Roles[0]).To(BeEquivalentTo("admin"))

				g.Expect(brokerSecuritySettings[1].Permissions[1].OperationType).To(BeEquivalentTo("createAddress"))
				g.Expect(len(brokerSecuritySettings[1].Permissions[1].Roles)).To(BeEquivalentTo(1))
				g.Expect(brokerSecuritySettings[1].Permissions[1].Roles[0]).To(BeEquivalentTo("admin"))

				g.Expect(brokerSecuritySettings[1].Permissions[2].OperationType).To(BeEquivalentTo("manage"))
				g.Expect(len(brokerSecuritySettings[1].Permissions[2].Roles)).To(BeEquivalentTo(1))
				g.Expect(brokerSecuritySettings[1].Permissions[2].Roles[0]).To(BeEquivalentTo("admin"))

				managementSecuritySettings := secSpec.SecuritySettings.Management
				g.Expect(len(managementSecuritySettings.HawtioRoles)).To(BeEquivalentTo(1))
				g.Expect(managementSecuritySettings.HawtioRoles[0]).To(BeEquivalentTo("viewer"))
				g.Expect(len(managementSecuritySettings.Authorisation.AllowedList)).To(BeEquivalentTo(1))
				g.Expect(*managementSecuritySettings.Authorisation.AllowedList[0].Domain).To(BeEquivalentTo(hawtioDomain))
				g.Expect(len(managementSecuritySettings.Authorisation.RoleAccess)).To(BeEquivalentTo(1))
				g.Expect(*managementSecuritySettings.Authorisation.RoleAccess[0].Domain).To(BeEquivalentTo("org.apache.activemq.artemis"))
				g.Expect(len(managementSecuritySettings.Authorisation.RoleAccess[0].AccessList)).To(BeEquivalentTo(1))
				g.Expect(*managementSecuritySettings.Authorisation.RoleAccess[0].AccessList[0].Method).To(BeEquivalentTo("list*"))
				g.Expect(len(managementSecuritySettings.Authorisation.RoleAccess[0].AccessList[0].Roles)).To(BeEquivalentTo(2))
				g.Expect(managementSecuritySettings.Authorisation.RoleAccess[0].AccessList[0].Roles[0]).To(BeEquivalentTo("admin"))
				g.Expect(managementSecuritySettings.Authorisation.RoleAccess[0].AccessList[0].Roles[1]).To(BeEquivalentTo("viewer"))
			}, timeout, interval).Should(Succeed())

			By("checking resources get removed")

			Expect(k8sClient.Delete(ctx, createdSecCrd1)).Should(Succeed())
			Eventually(func() bool {
				return checkCrdDeleted(secCrd1.Name, defaultNamespace, createdSecCrd1)
			}, timeout, interval).Should(BeTrue())

			Expect(k8sClient.Delete(ctx, createdSecCrd2)).Should(Succeed())
			Eventually(func() bool {
				return checkCrdDeleted(secCrd2.Name, defaultNamespace, createdSecCrd2)
			}, timeout, interval).Should(BeTrue())

			Expect(k8sClient.Delete(ctx, createdSecCrd3)).Should(Succeed())
			Eventually(func() bool {
				return checkCrdDeleted(secCrd3.Name, defaultNamespace, createdSecCrd3)
			}, timeout, interval).Should(BeTrue())

		})
	})

	Context("Reconcile Test", func() {
		It("testing security applied after broker", Label("security-apply-restart"), func() {
			By("deploying the broker cr")
			crd, createdCrd := DeployCustomBroker(defaultNamespace, func(brokerCrdToDeploy *brokerv1beta1.ActiveMQArtemis) {

				brokerCrdToDeploy.Spec.Acceptors = []brokerv1beta1.AcceptorType{
					{
						Name:       "all",
						Expose:     true,
						Port:       61616,
						Protocols:  "all",
						SSLEnabled: false,
					},
				}
				brokerCrdToDeploy.Spec.AdminUser = "admin"
				brokerCrdToDeploy.Spec.AdminPassword = "admin"
				brokerCrdToDeploy.Spec.Console.Expose = true
				brokerCrdToDeploy.Spec.DeploymentPlan.Clustered = &boolFalse
				brokerCrdToDeploy.Spec.DeploymentPlan.JolokiaAgentEnabled = true
				brokerCrdToDeploy.Spec.DeploymentPlan.MessageMigration = &boolTrue
				brokerCrdToDeploy.Spec.DeploymentPlan.PersistenceEnabled = true
				brokerCrdToDeploy.Spec.DeploymentPlan.RequireLogin = true
				brokerCrdToDeploy.Spec.DeploymentPlan.Size = 1
			})

			if os.Getenv("USE_EXISTING_CLUSTER") == "true" {
				By("make sure the broker is up and running")
				Eventually(func(g Gomega) {
					key := types.NamespacedName{Name: namer.CrToSS(crd.Name), Namespace: defaultNamespace}
					sfsFound := &appsv1.StatefulSet{}

					g.Expect(k8sClient.Get(ctx, key, sfsFound)).Should(Succeed())
					g.Expect(sfsFound.Status.ReadyReplicas).Should(BeEquivalentTo(1))
				}, existingClusterTimeout, existingClusterInterval).Should(Succeed())

				By("Checking no security gets applied to broker at the moment" + createdCrd.Name)
				podWithOrdinal := namer.CrToSS(crd.Name) + "-0"
				command := []string{"ls", "amq-broker/etc"}

				Eventually(func(g Gomega) {
					stdOutContent := ExecOnPod(podWithOrdinal, crd.Name, defaultNamespace, command, g)
					g.Expect(stdOutContent).ShouldNot(ContainSubstring("keycloak"))
				}, existingClusterTimeout, existingClusterInterval).Should(Succeed())
			}

			By("deploying the security cr")
			secCrd, createdSecCrd := DeploySecurity("ex-keycloak", defaultNamespace, func(secCrdToDeploy *brokerv1beta1.ActiveMQArtemisSecurity) {

				brokerModuleName := "keycloak-broker"
				consoleModuleName := "keycloak-console"
				directAccess := "directAccess"
				realm := "artemis-keycloak-demo"
				brokerResource := "artemis-broker"
				brokerAuthUrl := "http://10.111.246.36:8080/auth"
				principalAttribute := "preferred_username"
				external := "external"
				token := "9699685c-8a30-45cf-bf19-0d38bbac5fdc"
				bearerToken := "bearerToken"
				consoleResource := "artemis-console"
				consoleAuthUrl := "http://keycloak.3387.com/auth"
				var confidentialPort int32 = 0

				brokerDomainName := "activemq"
				consoleDomainName := "console"
				requiredFlag := "required"
				mgmtDomain := "org.apache.activemq.artemis"
				listMethod := "list*"
				sendMethod := "sendMessage*"
				browseMethod := "browse*"

				secCrdToDeploy.Spec.LoginModules = brokerv1beta1.LoginModulesType{
					KeycloakLoginModules: []brokerv1beta1.KeycloakLoginModuleType{
						{
							Name:       brokerModuleName,
							ModuleType: &directAccess,
							Configuration: brokerv1beta1.KeycloakModuleConfigurationType{
								Realm:                   &realm,
								Resource:                &brokerResource,
								AuthServerUrl:           &brokerAuthUrl,
								UseResourceRoleMappings: &boolTrue,
								PrincipalAttribute:      &principalAttribute,
								SslRequired:             &external,
								Credentials: []brokerv1beta1.KeyValueType{
									{
										Key:   "secret",
										Value: &token,
									},
								},
							},
						},
						{
							Name:       consoleModuleName,
							ModuleType: &bearerToken,
							Configuration: brokerv1beta1.KeycloakModuleConfigurationType{
								Realm:                   &realm,
								Resource:                &consoleResource,
								AuthServerUrl:           &consoleAuthUrl,
								PrincipalAttribute:      &principalAttribute,
								UseResourceRoleMappings: &boolTrue,
								SslRequired:             &external,
								ConfidentialPort:        &confidentialPort,
							},
						},
					},
				}
				secCrdToDeploy.Spec.SecurityDomains = brokerv1beta1.SecurityDomainsType{
					BrokerDomain: brokerv1beta1.BrokerDomainType{
						Name: &brokerDomainName,
						LoginModules: []brokerv1beta1.LoginModuleReferenceType{
							{
								Name: &brokerModuleName,
								Flag: &requiredFlag,
							},
						},
					},
					ConsoleDomain: brokerv1beta1.BrokerDomainType{
						Name: &consoleDomainName,
						LoginModules: []brokerv1beta1.LoginModuleReferenceType{
							{
								Name: &consoleModuleName,
								Flag: &requiredFlag,
							},
						},
					},
				}
				secCrdToDeploy.Spec.SecuritySettings = brokerv1beta1.SecuritySettingsType{
					Broker: []brokerv1beta1.BrokerSecuritySettingType{
						{
							Match: "Info",
							Permissions: []brokerv1beta1.PermissionType{
								{
									OperationType: "createDurableQueue",
									Roles: []string{
										"amq",
									},
								},
								{
									OperationType: "deleteDurableQueue",
									Roles: []string{
										"amq",
									},
								},
								{
									OperationType: "createNonDurableQueue",
									Roles: []string{
										"amq",
									},
								},
								{
									OperationType: "deleteNonDurableQueue",
									Roles: []string{
										"amq",
									},
								},
								{
									OperationType: "send",
									Roles: []string{
										"amq",
									},
								},
								{
									OperationType: "consume",
									Roles: []string{
										"amq",
									},
								},
							},
						},
						{
							Match: "activemq.management.#",
							Permissions: []brokerv1beta1.PermissionType{
								{
									OperationType: "createNonDurableQueue",
									Roles: []string{
										"amq",
									},
								},
								{
									OperationType: "createAddress",
									Roles: []string{
										"amq",
									},
								},
								{
									OperationType: "consume",
									Roles: []string{
										"amq",
									},
								},
								{
									OperationType: "manage",
									Roles: []string{
										"amq",
									},
								},
								{
									OperationType: "send",
									Roles: []string{
										"amq",
									},
								},
							},
						},
					},
					Management: brokerv1beta1.ManagementSecuritySettingsType{
						HawtioRoles: []string{
							"guest",
						},
						Authorisation: brokerv1beta1.AuthorisationConfigType{
							RoleAccess: []brokerv1beta1.RoleAccessType{
								{
									Domain: &mgmtDomain,
									AccessList: []brokerv1beta1.DefaultAccessType{
										{
											Method: &listMethod,
											Roles: []string{
												"guest",
											},
										},
										{
											Method: &sendMethod,
											Roles: []string{
												"guest",
											},
										},
										{
											Method: &browseMethod,
											Roles: []string{
												"guest",
											},
										},
									},
								},
							},
						},
					},
				}
			})
			By("checking security is applied")
			requestedSs := &appsv1.StatefulSet{}

			Eventually(func() bool {
				key := types.NamespacedName{Name: namer.CrToSS(createdCrd.Name), Namespace: defaultNamespace}
				err := k8sClient.Get(ctx, key, requestedSs)
				if err != nil {
					return false
				}

				initContainer := requestedSs.Spec.Template.Spec.InitContainers[0]
				secApplied := false
				for _, arg := range initContainer.Args {
					if strings.Contains(arg, "mkdir -p /init_cfg_root/security/security") {
						secApplied = true
						break
					}
				}
				return secApplied
			}, timeout, interval).Should(BeTrue())

			if os.Getenv("USE_EXISTING_CLUSTER") == "true" {
				By("Checking ready on SS")
				Eventually(func(g Gomega) {
					key := types.NamespacedName{Name: namer.CrToSS(crd.Name), Namespace: defaultNamespace}
					sfsFound := &appsv1.StatefulSet{}

					g.Expect(k8sClient.Get(ctx, key, sfsFound)).Should(Succeed())
					g.Expect(sfsFound.Status.ReadyReplicas).Should(BeEquivalentTo(1))
				}, existingClusterTimeout, existingClusterInterval).Should(Succeed())

				By("Checking security gets applied to broker " + createdCrd.Name)
				podWithOrdinal := namer.CrToSS(crd.Name) + "-0"
				command := []string{"ls", "amq-broker/etc"}

				Eventually(func(g Gomega) {
					stdOutContent := ExecOnPod(podWithOrdinal, crd.Name, defaultNamespace, command, g)
					g.Expect(stdOutContent).Should(ContainSubstring("keycloak"))
				}, existingClusterTimeout, existingClusterInterval).Should(Succeed())
			}

			By("checking resources get removed")
			Expect(k8sClient.Delete(ctx, crd)).Should(Succeed())
			Eventually(func() bool {
				return checkCrdDeleted(crd.Name, defaultNamespace, createdCrd)
			}, timeout, interval).Should(BeTrue())

			Expect(k8sClient.Delete(ctx, createdSecCrd)).Should(Succeed())
			Eventually(func() bool {
				return checkCrdDeleted(secCrd.ObjectMeta.Name, defaultNamespace, createdSecCrd)
			}, timeout, interval).Should(BeTrue())

		})

		It("reconcile twice with nothing changed", Label("reconcile_same_sec_twice"), func() {

			By("Creating security cr")
			ctx := context.Background()
			crd := generateSecuritySpec("", defaultNamespace)

			brokerDomainName := "activemq"
			loginModuleName := "module1"
			loginModuleFlag := "sufficient"

			loginModuleList := make([]brokerv1beta1.PropertiesLoginModuleType, 1)
			propLoginModule := brokerv1beta1.PropertiesLoginModuleType{
				Name: loginModuleName,
				Users: []brokerv1beta1.UserType{
					{
						Name:     "user1",
						Password: nil,
						Roles: []string{
							"admin", "amq",
						},
					},
				},
			}
			loginModuleList = append(loginModuleList, propLoginModule)
			crd.Spec.LoginModules.PropertiesLoginModules = loginModuleList

			crd.Spec.SecurityDomains.BrokerDomain = brokerv1beta1.BrokerDomainType{
				Name: &brokerDomainName,
				LoginModules: []brokerv1beta1.LoginModuleReferenceType{
					{
						Name: &loginModuleName,
						Flag: &loginModuleFlag,
					},
				},
			}

			By("Deploying the CRD " + crd.ObjectMeta.Name)
			Expect(k8sClient.Create(ctx, crd)).Should(Succeed())

			createdCrd := &brokerv1beta1.ActiveMQArtemisSecurity{}

			By("Making sure that the CRD gets deployed " + crd.ObjectMeta.Name)
			Eventually(func() bool {
				return getPersistedVersionedCrd(crd.ObjectMeta.Name, defaultNamespace, createdCrd)
			}, timeout, interval).Should(BeTrue())
			Expect(createdCrd.Name).Should(Equal(crd.ObjectMeta.Name))

			var securityHandler common.ActiveMQArtemisConfigHandler
			Eventually(func() bool {
				securityHandler, _ = GetBrokerSecurityConfigHandler(types.NamespacedName{
					Name:      crd.ObjectMeta.Name,
					Namespace: defaultNamespace,
				})
				return securityHandler != nil
			}, timeout, interval).Should(BeTrue())

			realHandler, ok := securityHandler.(*ActiveMQArtemisSecurityConfigHandler)
			Expect(ok).To(BeTrue())

			By("Redeploying the same CR")
			request := ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name:      crd.ObjectMeta.Name,
					Namespace: defaultNamespace,
				},
			}

			result, err := securityReconciler.Reconcile(context.Background(), request)

			Expect(err).NotTo(HaveOccurred())
			Expect(result.RequeueAfter).To(Equal(common.GetReconcileResyncPeriod()))

			newHandler, _ := GetBrokerSecurityConfigHandler(types.NamespacedName{
				Name:      crd.ObjectMeta.Name,
				Namespace: defaultNamespace,
			})
			Expect(newHandler).NotTo(BeNil())

			newRealHandler, ok2 := newHandler.(*ActiveMQArtemisSecurityConfigHandler)
			Expect(ok2).To(BeTrue())

			equal := reflect.DeepEqual(realHandler, newRealHandler)
			Expect(equal).To(BeTrue())

			By("check it has gone")
			Expect(k8sClient.Delete(ctx, createdCrd))
			Eventually(func() bool {
				return checkCrdDeleted(crd.ObjectMeta.Name, defaultNamespace, createdCrd)
			}, timeout, interval).Should(BeTrue())

		})

		It("Testing applyToCrNames working properly", func() {

			By("Creating security cr")
			crd := generateSecuritySpec("", defaultNamespace)

			brokerDomainName := "activemq"
			loginModuleName := "module1"
			loginModuleFlag := "sufficient"

			loginModuleList := make([]brokerv1beta1.PropertiesLoginModuleType, 1)
			propLoginModule := brokerv1beta1.PropertiesLoginModuleType{
				Name: loginModuleName,
				Users: []brokerv1beta1.UserType{
					{
						Name:     "user1",
						Password: nil,
						Roles: []string{
							"admin", "amq",
						},
					},
				},
			}
			loginModuleList = append(loginModuleList, propLoginModule)
			crd.Spec.LoginModules.PropertiesLoginModules = loginModuleList

			crd.Spec.SecurityDomains.BrokerDomain = brokerv1beta1.BrokerDomainType{
				Name: &brokerDomainName,
				LoginModules: []brokerv1beta1.LoginModuleReferenceType{
					{
						Name: &loginModuleName,
						Flag: &loginModuleFlag,
					},
				},
			}
			crd.Name = "security"

			defaultbrokerNamespace := defaultNamespace
			broker3Namespace := "broker3-namespace"
			broker4Namespace := "broker4-namespace"
			broker0Name := "broker0"
			broker1Name := "broker1"
			broker2Name := "broker2"
			broker3Name := "broker0"
			broker4Name := "broker0"

			broker0 := types.NamespacedName{
				Name:      broker0Name,
				Namespace: defaultbrokerNamespace,
			}
			broker1 := types.NamespacedName{
				Name:      broker1Name,
				Namespace: defaultbrokerNamespace,
			}
			broker2 := types.NamespacedName{
				Name:      broker2Name,
				Namespace: defaultbrokerNamespace,
			}
			broker3 := types.NamespacedName{
				Name:      broker3Name,
				Namespace: broker3Namespace,
			}
			broker4 := types.NamespacedName{
				Name:      broker4Name,
				Namespace: broker4Namespace,
			}

			secHandler := ActiveMQArtemisSecurityConfigHandler{
				SecurityCR: crd,
				NamespacedName: types.NamespacedName{
					Name:      crd.Name,
					Namespace: defaultNamespace,
				},
				owner: nil,
			}

			By("Default security applies to all in the same namespace but none in others")
			Expect(crd.Spec.ApplyToCrNames).To(BeEmpty())

			Expect(secHandler.IsApplicableFor(broker0)).To(BeTrue())
			Expect(secHandler.IsApplicableFor(broker1)).To(BeTrue())
			Expect(secHandler.IsApplicableFor(broker2)).To(BeTrue())
			Expect(secHandler.IsApplicableFor(broker3)).To(BeFalse())
			Expect(secHandler.IsApplicableFor(broker4)).To(BeFalse())

			By("ApplyToCrNames being empty applies to all in the same namespace but none in others")
			crd.Spec.ApplyToCrNames = []string{""}
			Expect(crd.Spec.ApplyToCrNames[0]).To(BeEmpty())

			Expect(secHandler.IsApplicableFor(broker0)).To(BeTrue())
			Expect(secHandler.IsApplicableFor(broker1)).To(BeTrue())
			Expect(secHandler.IsApplicableFor(broker2)).To(BeTrue())
			Expect(secHandler.IsApplicableFor(broker3)).To(BeFalse())
			Expect(secHandler.IsApplicableFor(broker4)).To(BeFalse())

			By("ApplyToCrNames being * applies to all in the same namespace but none in others")
			crd.Spec.ApplyToCrNames = []string{"*"}
			Expect(crd.Spec.ApplyToCrNames[0]).To(Equal("*"))

			Expect(secHandler.IsApplicableFor(broker0)).To(BeTrue())
			Expect(secHandler.IsApplicableFor(broker1)).To(BeTrue())
			Expect(secHandler.IsApplicableFor(broker2)).To(BeTrue())
			Expect(secHandler.IsApplicableFor(broker3)).To(BeFalse())
			Expect(secHandler.IsApplicableFor(broker4)).To(BeFalse())

			By("ApplyToCrNames being broker0 only applies to broker0 in the same namespace")
			crd.Spec.ApplyToCrNames = []string{"broker0"}
			Expect(crd.Spec.ApplyToCrNames[0]).To(Equal("broker0"))

			Expect(secHandler.IsApplicableFor(broker0)).To(BeTrue())
			Expect(secHandler.IsApplicableFor(broker1)).To(BeFalse())
			Expect(secHandler.IsApplicableFor(broker2)).To(BeFalse())
			Expect(secHandler.IsApplicableFor(broker3)).To(BeFalse())
			Expect(secHandler.IsApplicableFor(broker4)).To(BeFalse())

			By("ApplyToCrNames being broker0, broker1 only applies to broker0 and broker1 in the same namespace")
			crd.Spec.ApplyToCrNames = []string{"broker0", "broker1"}
			Expect(crd.Spec.ApplyToCrNames[0]).To(Equal("broker0"))
			Expect(crd.Spec.ApplyToCrNames[1]).To(Equal("broker1"))

			Expect(secHandler.IsApplicableFor(broker0)).To(BeTrue())
			Expect(secHandler.IsApplicableFor(broker1)).To(BeTrue())
			Expect(secHandler.IsApplicableFor(broker2)).To(BeFalse())
			Expect(secHandler.IsApplicableFor(broker3)).To(BeFalse())
			Expect(secHandler.IsApplicableFor(broker4)).To(BeFalse())

		})

		It("Reconcile security on multiple broker CRs", Label("security-to-multiple-crs"), func() {

			By("Deploying 3 brokers======")
			broker1Cr, createdBroker1Cr := DeployBroker("ex-aaom", defaultNamespace)
			broker2Cr, createdBroker2Cr := DeployBroker("ex-aaom1", defaultNamespace)
			broker3Cr, createdBroker3Cr := DeployBroker("ex-aaom2", defaultNamespace)

			By("deploying security after brokers are deployed=======")
			secCrd, createdSecCrd := DeploySecurity("", defaultNamespace, func(secCrdToDeploy *brokerv1beta1.ActiveMQArtemisSecurity) {
				applyToCrs := make([]string, 0)
				applyToCrs = append(applyToCrs, "ex-aaom")
				applyToCrs = append(applyToCrs, "ex-aaom2")
				secCrdToDeploy.Spec.ApplyToCrNames = applyToCrs
			})

			requestedSs := &appsv1.StatefulSet{}
			By("Checking security gets applied to broker1 " + broker1Cr.Name)
			key := types.NamespacedName{Name: namer.CrToSS(createdBroker1Cr.Name), Namespace: defaultNamespace}
			Eventually(func(g Gomega) {
				By("Checking SS for security")
				g.Expect(k8sClient.Get(ctx, key, requestedSs)).Should(Succeed())

				By("SS retrieved fresh")
				initContainer := requestedSs.Spec.Template.Spec.InitContainers[0]
				secApplied := false
				for _, arg := range initContainer.Args {
					By("Checking arg " + arg)
					if strings.Contains(arg, "mkdir -p /init_cfg_root/security/security") {
						By("args match! " + arg)
						secApplied = true
						break
					}
				}
				g.Expect(secApplied).To(BeTrue())
			}, timeout, interval).Should(Succeed())

			By("Checking security doesn't gets applied to broker2 " + broker2Cr.Name)
			Eventually(func() bool {
				key := types.NamespacedName{Name: namer.CrToSS(createdBroker2Cr.Name), Namespace: defaultNamespace}
				err := k8sClient.Get(ctx, key, requestedSs)
				if err != nil {
					return false
				}
				initContainer := requestedSs.Spec.Template.Spec.InitContainers[0]
				secApplied := false
				for _, arg := range initContainer.Args {
					if strings.Contains(arg, "mkdir -p /init_cfg_root/security/security") {
						secApplied = true
						break
					}
				}
				return secApplied

			}, timeout, interval).Should(BeFalse())

			By("Checking security gets applied to broker3 " + broker3Cr.Name)
			Eventually(func() bool {
				key := types.NamespacedName{Name: namer.CrToSS(createdBroker3Cr.Name), Namespace: defaultNamespace}
				err := k8sClient.Get(ctx, key, requestedSs)
				if err != nil {
					fmt.Printf("error retrieving broker3 ss %v\n", err)
					return false
				}
				initContainer := requestedSs.Spec.Template.Spec.InitContainers[0]
				secApplied := false
				for _, arg := range initContainer.Args {
					if strings.Contains(arg, "mkdir -p /init_cfg_root/security/security") {
						secApplied = true
						break
					}
				}
				return secApplied

			}, timeout, interval).Should(BeTrue())

			By("check it has gone")
			Expect(k8sClient.Delete(ctx, createdBroker1Cr)).Should(Succeed())
			Eventually(func() bool {
				return checkCrdDeleted(broker1Cr.ObjectMeta.Name, defaultNamespace, createdBroker1Cr)
			}, timeout, interval).Should(BeTrue())

			Expect(k8sClient.Delete(ctx, createdBroker2Cr)).Should(Succeed())
			Eventually(func() bool {
				return checkCrdDeleted(broker2Cr.ObjectMeta.Name, defaultNamespace, createdBroker2Cr)
			}, timeout, interval).Should(BeTrue())

			Expect(k8sClient.Delete(ctx, createdBroker3Cr)).Should(Succeed())
			Eventually(func() bool {
				return checkCrdDeleted(broker3Cr.ObjectMeta.Name, defaultNamespace, createdBroker3Cr)
			}, timeout, interval).Should(BeTrue())

			Expect(k8sClient.Delete(ctx, createdSecCrd)).Should(Succeed())
			Eventually(func() bool {
				return checkCrdDeleted(secCrd.ObjectMeta.Name, defaultNamespace, createdSecCrd)
			}, timeout, interval).Should(BeTrue())

		})

		It("Reconcile security on broker with non shell safe annotations", func() {

			By("Deploying broker")
			brokerCrd := generateArtemisSpec(defaultNamespace)
			brokerCrd.Spec.DeploymentPlan.Size = 1
			// make is speedy for real cluster checks
			brokerCrd.Spec.DeploymentPlan.ReadinessProbe = &v1.Probe{
				InitialDelaySeconds: 1,
				PeriodSeconds:       1,
				TimeoutSeconds:      5,
			}
			Expect(k8sClient.Create(ctx, &brokerCrd)).Should(Succeed())

			createdBrokerCr := &brokerv1beta1.ActiveMQArtemis{}

			Eventually(func() bool {
				return getPersistedVersionedCrd(brokerCrd.ObjectMeta.Name, defaultNamespace, createdBrokerCr)
			}, timeout, interval).Should(BeTrue())
			Expect(brokerCrd.Name).Should(Equal(createdBrokerCr.ObjectMeta.Name))

			_, createdSecCrd := DeploySecurity("", defaultNamespace, func(secCrdToDeploy *brokerv1beta1.ActiveMQArtemisSecurity) {
				applyToCrs := make([]string, 0)
				applyToCrs = append(applyToCrs, createdBrokerCr.ObjectMeta.Name)
				secCrdToDeploy.Spec.ApplyToCrNames = applyToCrs
				secCrdToDeploy.Annotations = map[string]string{
					"testannotation": "pltf-amq (1)",
				}
			})

			requestedSs := &appsv1.StatefulSet{}

			By("Checking security gets applied to broker1 " + createdBrokerCr.Name)
			Eventually(func() bool {
				key := types.NamespacedName{Name: namer.CrToSS(createdBrokerCr.Name), Namespace: defaultNamespace}
				err := k8sClient.Get(ctx, key, requestedSs)
				if err != nil {
					return false
				}

				initContainer := requestedSs.Spec.Template.Spec.InitContainers[0]
				secApplied := false
				emptyMetadata := false
				for _, arg := range initContainer.Args {

					if strings.Contains(arg, "mkdir -p /init_cfg_root/security/security") {
						secApplied = true

						if !(strings.Contains(arg, "testannotation")) {
							emptyMetadata = true
							break
						}
					}
				}
				return secApplied && emptyMetadata
			}, timeout, interval).Should(BeTrue())

			if os.Getenv("USE_EXISTING_CLUSTER") == "true" {
				By("Checking status of CR because we expect it to deploy on a real cluster")
				key := types.NamespacedName{Name: createdBrokerCr.ObjectMeta.Name, Namespace: defaultNamespace}

				Eventually(func(g Gomega) {
					g.Expect(k8sClient.Get(ctx, key, createdBrokerCr)).Should(Succeed())

					g.Expect(len(createdBrokerCr.Status.PodStatus.Ready)).Should(BeEquivalentTo(1))
				}, existingClusterTimeout, existingClusterInterval).Should(Succeed())
			}

			Expect(k8sClient.Delete(ctx, createdBrokerCr)).Should(Succeed())
			Expect(k8sClient.Delete(ctx, createdSecCrd)).Should(Succeed())

		})

		It("Reconcile security with management role access", func() {

			if os.Getenv("USE_EXISTING_CLUSTER") != "true" {
				return
			}

			By("Deploying broker")
			brokerCrd := generateArtemisSpec(defaultNamespace)
			brokerCrd.Spec.DeploymentPlan.Size = 1
			Expect(k8sClient.Create(ctx, &brokerCrd)).Should(Succeed())

			createdCrd := &brokerv1beta1.ActiveMQArtemis{}

			By("Checking the pod is up and running")
			Eventually(func(g Gomega) {
				brokerKey := types.NamespacedName{Name: brokerCrd.Name, Namespace: brokerCrd.Namespace}
				g.Expect(k8sClient.Get(ctx, brokerKey, createdCrd)).Should(Succeed())
				g.Expect(len(createdCrd.Status.PodStatus.Ready)).Should(BeEquivalentTo(1))

			}, existingClusterTimeout, existingClusterInterval).Should(Succeed())

			By("Deploying security with management role access")
			mgmtDomain := "org.apache.activemq.artemis"
			method1 := "list*"
			method2 := "sendMessage*"
			method3 := "browse*"
			accessList := []brokerv1beta1.DefaultAccessType{
				{
					Method: &method1,
					Roles:  []string{"guest"},
				},
				{
					Method: &method2,
					Roles:  []string{"guest"},
				},
				{
					Method: &method3,
					Roles:  []string{"guest"},
				},
			}
			roleAccess := []brokerv1beta1.RoleAccessType{
				{
					Domain:     &mgmtDomain,
					AccessList: accessList,
				},
			}

			allowedDomain := "org.apache.activemq.artemis.allowed"
			allowedList := []brokerv1beta1.AllowedListEntryType{
				{
					Domain: &allowedDomain,
				},
			}

			_, createdSecCrd := DeploySecurity("", defaultNamespace, func(secCrdToDeploy *brokerv1beta1.ActiveMQArtemisSecurity) {

				secCrdToDeploy.Spec.SecuritySettings.Management = brokerv1beta1.ManagementSecuritySettingsType{
					Authorisation: brokerv1beta1.AuthorisationConfigType{
						AllowedList: allowedList,
						RoleAccess:  roleAccess,
					},
				}
			})

			By("Checking the pod get started")
			brokerKey := types.NamespacedName{Name: brokerCrd.Name, Namespace: brokerCrd.Namespace}
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, brokerKey, createdCrd)).Should(Succeed())
				g.Expect(len(createdCrd.Status.PodStatus.Ready)).Should(BeEquivalentTo(1))
			}, existingClusterTimeout, existingClusterInterval).Should(Succeed())

			By("Checking the management.xml has the correct role-access element")
			gvk := schema.GroupVersionKind{
				Group:   "",
				Version: "v1",
				Kind:    "Pod",
			}
			restClient, err := apiutil.RESTClientForGVK(gvk, false, restConfig, serializer.NewCodecFactory(scheme.Scheme))
			Expect(err).To(BeNil())

			podOrdinal := strconv.FormatInt(int64(0), 10)
			podName := namer.CrToSS(brokerCrd.Name) + "-" + podOrdinal

			brokerName := brokerCrd.Name
			Eventually(func(g Gomega) {
				execReq := restClient.
					Post().
					Namespace(defaultNamespace).
					Resource("pods").
					Name(podName).
					SubResource("exec").
					VersionedParams(&corev1.PodExecOptions{
						Container: brokerName + "-container",
						Command:   []string{"cat", "amq-broker/etc/management.xml"},
						Stdin:     true,
						Stdout:    true,
						Stderr:    true,
					}, runtime.NewParameterCodec(scheme.Scheme))

				exec, err := remotecommand.NewSPDYExecutor(restConfig, "POST", execReq.URL())

				if err != nil {
					fmt.Printf("error while creating remote command executor: %v", err)
				}
				Expect(err).To(BeNil())
				var capturedOut bytes.Buffer

				err = exec.Stream(remotecommand.StreamOptions{
					Stdin:  os.Stdin,
					Stdout: &capturedOut,
					Stderr: os.Stderr,
					Tty:    false,
				})
				g.Expect(err).To(BeNil())

				By("Checking for output from pod")
				g.Eventually(func(g Gomega) {
					By("Checking for output from pod")
					g.Expect(capturedOut.Len() > 0)
				}, existingClusterTimeout, existingClusterInterval).Should(Succeed())

				By("Checking pod output")
				content := capturedOut.String()
				g.Expect(content).Should(ContainSubstring("<match domain=\"org.apache.activemq.artemis\""))
				g.Expect(content).Should(ContainSubstring("<access method=\"list*\" roles=\"guest\""))
				g.Expect(content).Should(ContainSubstring("<access method=\"sendMessage*\" roles=\"guest\""))
				g.Expect(content).Should(ContainSubstring("<access method=\"browse*\" roles=\"guest\""))
				g.Expect(content).Should(ContainSubstring("<entry domain=\"org.apache.activemq.artemis.allowed\""))
			}, existingClusterTimeout, existingClusterInterval).Should(Succeed())

			Expect(k8sClient.Delete(ctx, createdCrd)).Should(Succeed())
			Expect(k8sClient.Delete(ctx, createdSecCrd)).Should(Succeed())
		})

	})

	It("reconcile after Broker CR deployed, verify force reconcile", Label("failtest-1"), func() {

		By("Creating Broker CR")
		ctx := context.Background()
		brokerCrd := generateArtemisSpec(defaultNamespace)

		Expect(k8sClient.Create(ctx, &brokerCrd)).Should(Succeed())

		createdBrokerCrd := &brokerv1beta1.ActiveMQArtemis{}

		Eventually(func() (int, error) {
			key := types.NamespacedName{Name: brokerCrd.ObjectMeta.Name, Namespace: defaultNamespace}
			err := k8sClient.Get(ctx, key, createdBrokerCrd)

			if err != nil {
				return -1, err
			}

			return len(createdBrokerCrd.Status.PodStatus.Stopped), nil
		}, timeout, interval).Should(Equal(1))

		// after stable status, determine version
		createdSs := &appsv1.StatefulSet{}
		ssKey := types.NamespacedName{Name: namer.CrToSS(createdBrokerCrd.Name), Namespace: defaultNamespace}

		By("Making sure that the ss gets deployed " + createdBrokerCrd.ObjectMeta.Name)
		Eventually(func(g Gomega) {
			g.Expect(k8sClient.Get(ctx, ssKey, createdSs)).Should(Succeed())
		}, timeout, interval).Should(Succeed())

		versionSsDeployed := createdSs.ObjectMeta.ResourceVersion
		By("tracking cc resource version: " + versionSsDeployed)

		By("Creating security cr")
		crd := generateSecuritySpec("", defaultNamespace)

		brokerDomainName := "activemq"
		loginModuleName := "module1"
		loginModuleFlag := "sufficient"

		loginModuleList := make([]brokerv1beta1.PropertiesLoginModuleType, 1)
		propLoginModule := brokerv1beta1.PropertiesLoginModuleType{
			Name: loginModuleName,
			Users: []brokerv1beta1.UserType{
				{
					Name:     "user1",
					Password: nil,
					Roles: []string{
						"admin", "amq",
					},
				},
			},
		}
		loginModuleList = append(loginModuleList, propLoginModule)
		crd.Spec.LoginModules.PropertiesLoginModules = loginModuleList

		crd.Spec.SecurityDomains.BrokerDomain = brokerv1beta1.BrokerDomainType{
			Name: &brokerDomainName,
			LoginModules: []brokerv1beta1.LoginModuleReferenceType{
				{
					Name: &loginModuleName,
					Flag: &loginModuleFlag,
				},
			},
		}

		By("Deploying the CRD " + crd.ObjectMeta.Name)
		Expect(k8sClient.Create(ctx, crd)).Should(Succeed())

		createdCrd := &brokerv1beta1.ActiveMQArtemisSecurity{}

		By("Making sure that the CRD gets deployed " + crd.ObjectMeta.Name)
		Eventually(func() bool {
			return getPersistedVersionedCrd(crd.ObjectMeta.Name, defaultNamespace, createdCrd)
		}, timeout, interval).Should(BeTrue())
		Expect(createdCrd.Name).Should(Equal(crd.ObjectMeta.Name))

		// make sure broker gets new SS for createdBrokerCrd
		Eventually(func(g Gomega) {

			g.Expect(k8sClient.Get(ctx, ssKey, createdSs)).Should(Succeed())
			By("Verifying != ss set resoruce version, deployed: " + versionSsDeployed + ", current: " + createdSs.GetResourceVersion())
			g.Expect(versionSsDeployed).ShouldNot(Equal(createdSs.GetResourceVersion()))

		}, timeout, interval).Should(Succeed())

		By("check it has gone")
		Expect(k8sClient.Delete(ctx, createdBrokerCrd))
		Expect(k8sClient.Delete(ctx, createdCrd))
		Eventually(func() bool {
			return checkCrdDeleted(crd.ObjectMeta.Name, defaultNamespace, createdCrd)
		}, timeout, interval).Should(BeTrue())

	})

})

func DeploySecurity(secName string, targetNamespace string, customFunc func(candidate *brokerv1beta1.ActiveMQArtemisSecurity)) (*brokerv1beta1.ActiveMQArtemisSecurity, *brokerv1beta1.ActiveMQArtemisSecurity) {
	ctx := context.Background()
	secCrd := generateSecuritySpec(secName, targetNamespace)

	brokerDomainName := "activemq"
	loginModuleName := "module1"
	loginModuleFlag := "sufficient"
	okDefaultPwd := "ok"

	loginModuleList := make([]brokerv1beta1.PropertiesLoginModuleType, 1)
	propLoginModule := brokerv1beta1.PropertiesLoginModuleType{
		Name: loginModuleName,
		Users: []brokerv1beta1.UserType{
			{
				Name:     "user1",
				Password: &okDefaultPwd,
				Roles: []string{
					"admin", "amq",
				},
			},
		},
	}
	loginModuleList = append(loginModuleList, propLoginModule)
	secCrd.Spec.LoginModules.PropertiesLoginModules = loginModuleList

	secCrd.Spec.SecurityDomains.BrokerDomain = brokerv1beta1.BrokerDomainType{
		Name: &brokerDomainName,
		LoginModules: []brokerv1beta1.LoginModuleReferenceType{
			{
				Name: &loginModuleName,
				Flag: &loginModuleFlag,
			},
		},
	}

	customFunc(secCrd)

	Expect(k8sClient.Create(ctx, secCrd)).Should(Succeed())

	createdSecCrd := &brokerv1beta1.ActiveMQArtemisSecurity{}

	Eventually(func() bool {
		return getPersistedVersionedCrd(secCrd.ObjectMeta.Name, targetNamespace, createdSecCrd)
	}, timeout, interval).Should(BeTrue())
	Expect(createdSecCrd.Name).Should(Equal(secCrd.ObjectMeta.Name))

	return secCrd, createdSecCrd
}

func generateSecuritySpec(secName string, targetNamespace string) *brokerv1beta1.ActiveMQArtemisSecurity {

	spec := brokerv1beta1.ActiveMQArtemisSecuritySpec{}

	theName := secName
	if secName == "" {
		theName = randString()
	}

	toCreate := brokerv1beta1.ActiveMQArtemisSecurity{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ActiveMQArtemisSecurity",
			APIVersion: brokerv1beta1.GroupVersion.Identifier(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      theName,
			Namespace: targetNamespace,
		},
		Spec: spec,
	}

	return &toCreate
}

func CheckRolesMatch(roles []string, expected ...string) bool {
	if len(roles) != len(expected) {
		return false
	}
	for _, expectedRole := range expected {
		found := false
		for _, role := range roles {
			if role == expectedRole {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

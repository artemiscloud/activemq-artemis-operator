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

package controllers

import (
	"context"
	"fmt"
	"os"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"

	brokerv1beta1 "github.com/artemiscloud/activemq-artemis-operator/api/v1beta1"
	"github.com/artemiscloud/activemq-artemis-operator/pkg/utils/common"
	"github.com/artemiscloud/activemq-artemis-operator/version"
)

var _ = Describe("jdbc fast failover", func() {

	BeforeEach(func() {
		BeforeEachSpec()
	})

	AfterEach(func() {
		AfterEachSpec()
	})

	Context("artemis", Label("slow"), func() {
		It("cr with db store", func() {
			if os.Getenv("USE_EXISTING_CLUSTER") == "true" {

				By("deploying db and service")
				dbName := "pdb"
				var dbPort int32 = 5432
				dbAppLables := map[string]string{"app": dbName}
				db := appsv1.Deployment{

					TypeMeta:   metav1.TypeMeta{Kind: "Deployment", APIVersion: "apps/v1"},
					ObjectMeta: metav1.ObjectMeta{Name: dbName, Namespace: defaultNamespace, Labels: dbAppLables},
					Spec: appsv1.DeploymentSpec{
						Selector: &metav1.LabelSelector{MatchLabels: dbAppLables},
						Replicas: common.Int32ToPtr(1),
						Template: corev1.PodTemplateSpec{ObjectMeta: metav1.ObjectMeta{Labels: dbAppLables},
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name:  dbName,
										Image: "docker.io/postgres:16.2",
										Env: []corev1.EnvVar{
											{
												Name:  "POSTGRES_PASSWORD",
												Value: "postgres",
											},
										},
										Ports: []corev1.ContainerPort{
											{
												ContainerPort: dbPort,
											},
										},
										ReadinessProbe: &corev1.Probe{
											ProbeHandler: corev1.ProbeHandler{
												TCPSocket: &corev1.TCPSocketAction{
													Port: intstr.IntOrString{
														IntVal: dbPort,
													},
												},
											},
											InitialDelaySeconds: 15,
											PeriodSeconds:       10,
										},
									},
								},
							}},
					},
				}
				Expect(k8sClient.Create(ctx, &db)).Should(Succeed())

				dbService := corev1.Service{
					TypeMeta:   metav1.TypeMeta{Kind: "Service", APIVersion: "core/v1"},
					ObjectMeta: metav1.ObjectMeta{Name: dbName, Namespace: defaultNamespace},
					Spec: corev1.ServiceSpec{
						Selector: dbAppLables,
						Ports: []corev1.ServicePort{
							{
								Name:     dbName,
								Protocol: "TCP",
								Port:     dbPort,
							},
						},
					},
				}

				Expect(k8sClient.Create(ctx, &dbService)).Should(Succeed())

				By("deploying artemis")

				peerLabel := "fast-ha-peer"
				peerPrefix := NextSpecResourceName()
				ctx := context.Background()
				peerA := generateArtemisSpec(defaultNamespace)
				peerA.Name = peerPrefix + "-peer-a"

				peerA.Spec.Acceptors = []brokerv1beta1.AcceptorType{{Name: "tcp", Port: 61616, Expose: true}}

				if !isOpenshift {
					peerA.Spec.IngressDomain = defaultTestIngressDomain
				}

				peerA.Spec.Console.Expose = true

				peerA.Spec.DeploymentPlan.PersistenceEnabled = boolFalse
				peerA.Spec.DeploymentPlan.Clustered = &boolFalse

				peerA.Spec.DeploymentPlan.Labels = map[string]string{peerLabel: peerPrefix}

				peerA.Spec.Env = []corev1.EnvVar{
					{
						Name:  "ARTEMIS_EXTRA_LIBS",
						Value: "/amq/init/config/extra-libs",
					},
				}

				By("configuring the broker")
				peerA.Spec.BrokerProperties = []string{
					"addressConfigurations.JOBS.routingTypes=ANYCAST",
					"addressConfigurations.JOBS.queueConfigs.JOBS.routingType=ANYCAST",

					"HAPolicyConfiguration=SHARED_STORE_PRIMARY",
					"storeConfiguration=DATABASE",
					"storeConfiguration.jdbcDriverClassName=org.postgresql.Driver",
					"storeConfiguration.jdbcConnectionUrl=jdbc:postgresql://" + dbName + ":" + fmt.Sprintf("%d", dbPort) + "/postgres?user=postgres&password=postgres",
				}

				By("patching ss to add init container to pull down jdbc jar")

				kindMatchSs := "StatefulSet"
				peerA.Spec.ResourceTemplates = []brokerv1beta1.ResourceTemplate{
					{
						Selector: &brokerv1beta1.ResourceSelector{
							Kind: &kindMatchSs,
						},
						Patch: &unstructured.Unstructured{Object: map[string]interface{}{
							"kind": "StatefulSet",
							"spec": map[string]interface{}{
								"template": map[string]interface{}{
									"spec": map[string]interface{}{
										"initContainers": []interface{}{
											map[string]interface{}{
												"name":  "download-jdbc-driver",
												"image": version.LatestInitImage,
												"command": []interface{}{
													"sh",
													"-c",
													"mkdir -p /amq/init/config/extra-libs && curl -o /amq/init/config/extra-libs/postgresql-42.7.1.jar https://jdbc.postgresql.org/download/postgresql-42.7.1.jar",
												},
												"volumeMounts": []interface{}{
													map[string]interface{}{
														"name":      "amq-cfg-dir",
														"mountPath": "/amq/init/config",
													},
												},
											},
										},
									},
								},
							},
						},
						},
					},
				}

				peerB := &brokerv1beta1.ActiveMQArtemis{}

				// a clone of peerA
				peerA.DeepCopyInto(peerB)
				peerB.Name = peerPrefix + "-peer-b"

				By("provisioning the broker peer-a")
				Expect(k8sClient.Create(ctx, &peerA)).Should(Succeed())

				By("provisioning the broker peer-b")
				Expect(k8sClient.Create(ctx, peerB)).Should(Succeed())

				By("verifying one broker ready, it is a race to lock the db")
				Eventually(func(g Gomega) {

					createdBrokerCrd := &brokerv1beta1.ActiveMQArtemis{}

					g.Expect(k8sClient.Get(ctx, types.NamespacedName{
						Name:      peerPrefix + "-peer-a",
						Namespace: defaultNamespace}, createdBrokerCrd)).Should(Succeed())
					if verbose {
						fmt.Printf("\npeer-a CR Status:%v", createdBrokerCrd.Status)
					}
					readyPeer := meta.IsStatusConditionTrue(createdBrokerCrd.Status.Conditions, brokerv1beta1.ReadyConditionType)

					if !readyPeer {
						g.Expect(k8sClient.Get(ctx, types.NamespacedName{
							Name:      peerPrefix + "-peer-b",
							Namespace: defaultNamespace}, createdBrokerCrd)).Should(Succeed())
						if verbose {
							fmt.Printf("\npeer-b CR Status:%v", createdBrokerCrd.Status)
						}
						readyPeer = meta.IsStatusConditionTrue(createdBrokerCrd.Status.Conditions, brokerv1beta1.ReadyConditionType)
					}
					g.Expect(readyPeer).Should(BeTrue())

				}, existingClusterTimeout, existingClusterInterval).Should(Succeed())

				brokerService := "broker-ha-jdbc"
				By("provisioning loadbalanced service for this CR, for use within the cluster via dns")
				svc := &corev1.Service{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
						Kind:       "Service",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      brokerService,
						Namespace: defaultNamespace,
					},
					Spec: corev1.ServiceSpec{
						Selector: map[string]string{
							peerLabel: peerPrefix, // shared by both CRs
						},
						Ports: []corev1.ServicePort{
							{
								Port:       62616,
								TargetPort: intstr.IntOrString{IntVal: 61616},
							},
						},
					},
				}

				Expect(k8sClient.Create(ctx, svc)).Should(Succeed())

				By("verifying service is ok - status does not reflect endpoints which is not ideal")
				createdService := &corev1.Service{}
				Eventually(func(g Gomega) {

					g.Expect(k8sClient.Get(ctx, types.NamespacedName{Name: svc.Name,
						Namespace: defaultNamespace}, createdService)).Should(Succeed())

					if verbose {
						fmt.Printf("\nsvc Status:%v", createdService.Status)
					}
					g.Expect(len(createdService.Status.LoadBalancer.Ingress)).Should(BeNumerically("==", 0))
					g.Expect(len(createdService.Status.Conditions)).Should(BeNumerically("==", 0))

				}, existingClusterTimeout, existingClusterInterval).Should(Succeed())

				CleanResource(svc, svc.Name, defaultNamespace)

				CleanResource(&peerA, peerA.Name, defaultNamespace)
				CleanResource(peerB, peerB.Name, defaultNamespace)

				CleanResource(&db, db.Name, defaultNamespace)
				CleanResource(&dbService, dbService.Name, defaultNamespace)
			}
		})
	})
})

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
	"context"
	"math/rand"
	"strings"

	"fmt"
	//"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	brokerv1beta1 "github.com/artemiscloud/activemq-artemis-operator/api/v1beta1"
	"github.com/artemiscloud/activemq-artemis-operator/pkg/utils/namer"
)

// Uncomment this and the "test" import if you want to debug this set of tests
//func TestArtemisController(t *testing.T) {
// 	RegisterFailHandler(Fail)
// 	RunSpecs(t, "Artemis Controller Suite")
// }

var _ = Describe("artemis controller", func() {

	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		namespace = "default"
		timeout   = time.Second * 30
		duration  = time.Second * 10
		interval  = time.Millisecond * 250
	)

	Context("Liveness Probe Tests", func() {
		It("Override Liveness Probe No Exec", func() {
			By("By creating a crd with Liveness Probe")
			ctx := context.Background()
			crd := generateArtemisSpec(namespace)
			crd.Spec.AdminUser = "admin"
			crd.Spec.AdminPassword = "password"
			livenessProbe := corev1.Probe{}
			livenessProbe.PeriodSeconds = 5
			livenessProbe.InitialDelaySeconds = 6
			livenessProbe.TimeoutSeconds = 7
			livenessProbe.SuccessThreshold = 8
			livenessProbe.FailureThreshold = 9
			crd.Spec.DeploymentPlan.LivenessProbe = livenessProbe
			createdCrd := &brokerv1beta1.ActiveMQArtemis{}
			createdSs := &appsv1.StatefulSet{}

			By("Deploying the CRD")
			Expect(k8sClient.Create(ctx, &crd)).Should(Succeed())

			By("Making sure that the CRD gets deployed")
			Eventually(checkCrdCreated(crd.ObjectMeta.Name, namespace, createdCrd), timeout, interval).Should(BeTrue())
			Expect(createdCrd.Name).Should(Equal(crd.ObjectMeta.Name))

			By("Checking that Stateful Set is Created with the Liveness Probe")
			Eventually(func() bool {
				key := types.NamespacedName{Name: namer.CrToSS(createdCrd.Name), Namespace: namespace}

				err := k8sClient.Get(ctx, key, createdSs)

				if err != nil {
					return false
				}
				return createdSs.Spec.Template.Spec.Containers[0].LivenessProbe.TCPSocket != nil
			}, timeout, interval).Should(Equal(true))

			By("Making sure the Liveness probe is correct")
			Expect(len(createdSs.Spec.Template.Spec.Containers) == 1).Should(BeTrue())
			Expect(createdSs.Spec.Template.Spec.Containers[0].LivenessProbe.Handler.TCPSocket.Port.String() == "8161").Should(BeTrue())
			Expect(createdSs.Spec.Template.Spec.Containers[0].LivenessProbe.PeriodSeconds == 5).Should(BeTrue())
			Expect(createdSs.Spec.Template.Spec.Containers[0].LivenessProbe.InitialDelaySeconds == 6).Should(BeTrue())
			Expect(createdSs.Spec.Template.Spec.Containers[0].LivenessProbe.TimeoutSeconds == 7).Should(BeTrue())
			Expect(createdSs.Spec.Template.Spec.Containers[0].LivenessProbe.SuccessThreshold == 8).Should(BeTrue())
			Expect(createdSs.Spec.Template.Spec.Containers[0].LivenessProbe.FailureThreshold == 9).Should(BeTrue())

			By("Checking that Stateful Set is updated with the Liveness Probe")
			Eventually(func() bool {
				key := types.NamespacedName{Name: namer.CrToSS(createdCrd.Name), Namespace: namespace}

				err := k8sClient.Get(ctx, key, createdSs)

				if err != nil {
					return false
				}
				return createdSs.Spec.Template.Spec.Containers[0].LivenessProbe.TCPSocket != nil
			}, timeout, interval).Should(Equal(true))

			By("Updating the CR")
			Eventually(checkCrdCreated(crd.ObjectMeta.Name, namespace, createdCrd), timeout, interval).Should(BeTrue())
			original := generateOriginalArtemisSpec(namespace, createdCrd.Name)

			original.Spec.DeploymentPlan.LivenessProbe.PeriodSeconds = 15
			original.Spec.DeploymentPlan.LivenessProbe.InitialDelaySeconds = 16
			original.Spec.DeploymentPlan.LivenessProbe.TimeoutSeconds = 17
			original.Spec.DeploymentPlan.LivenessProbe.SuccessThreshold = 18
			original.Spec.DeploymentPlan.LivenessProbe.FailureThreshold = 19
			exec := corev1.ExecAction{
				Command: []string{"/broker/bin/artemis check node"},
			}
			original.Spec.DeploymentPlan.LivenessProbe.Exec = &exec
			original.ObjectMeta.ResourceVersion = createdCrd.GetResourceVersion()
			By("Redeploying the CRD")
			Expect(k8sClient.Update(ctx, original)).Should(Succeed())

			Eventually(func() bool {
				key := types.NamespacedName{Name: namer.CrToSS(createdCrd.Name), Namespace: namespace}

				err := k8sClient.Get(ctx, key, createdSs)

				if err != nil {
					return false
				}
				return createdSs.Spec.Template.Spec.Containers[0].LivenessProbe.PeriodSeconds == 15
			}, timeout, interval).Should(Equal(true))

			Expect(createdSs.Spec.Template.Spec.Containers[0].LivenessProbe.Handler.Exec != nil).Should(BeTrue())
			Expect(createdSs.Spec.Template.Spec.Containers[0].LivenessProbe.Handler.Exec.Command[0] == "/broker/bin/artemis check node").Should(BeTrue())
			Expect(createdSs.Spec.Template.Spec.Containers[0].LivenessProbe.PeriodSeconds == 15).Should(BeTrue())
			Expect(createdSs.Spec.Template.Spec.Containers[0].LivenessProbe.InitialDelaySeconds == 16).Should(BeTrue())
			Expect(createdSs.Spec.Template.Spec.Containers[0].LivenessProbe.TimeoutSeconds == 17).Should(BeTrue())
			Expect(createdSs.Spec.Template.Spec.Containers[0].LivenessProbe.SuccessThreshold == 18).Should(BeTrue())
			Expect(createdSs.Spec.Template.Spec.Containers[0].LivenessProbe.FailureThreshold == 19).Should(BeTrue())

			By("check it has gone")
			Expect(k8sClient.Delete(ctx, createdCrd))
			Eventually(checkCrdDeleted(crd.ObjectMeta.Name, namespace, createdCrd), timeout, interval).Should(BeTrue())
		})

		It("Override Liveness Probe Exec", func() {
			By("By creating a crd with Liveness Probe")
			ctx := context.Background()
			crd := generateArtemisSpec(namespace)
			crd.Spec.AdminUser = "admin"
			crd.Spec.AdminPassword = "password"
			exec := corev1.ExecAction{
				Command: []string{"/broker/bin/artemis check node"},
			}
			livenessProbe := corev1.Probe{}
			livenessProbe.Exec = &exec
			crd.Spec.DeploymentPlan.LivenessProbe = livenessProbe
			createdCrd := &brokerv1beta1.ActiveMQArtemis{}
			createdSs := &appsv1.StatefulSet{}
			By("Deploying the CRD")
			Expect(k8sClient.Create(ctx, &crd)).Should(Succeed())

			By("Making sure that the CRD gets deployed")
			Eventually(checkCrdCreated(crd.ObjectMeta.Name, namespace, createdCrd), timeout, interval).Should(BeTrue())
			Expect(createdCrd.Name).Should(Equal(crd.ObjectMeta.Name))

			By("Checking that Stateful Set is Created with the Liveness Probe")
			Eventually(func() bool {
				key := types.NamespacedName{Name: namer.CrToSS(createdCrd.Name), Namespace: namespace}

				err := k8sClient.Get(ctx, key, createdSs)

				if err != nil {
					return false
				}
				return createdSs.Spec.Template.Spec.Containers[0].LivenessProbe.Exec != nil
			}, timeout, interval).Should(Equal(true))

			By("Making sure the Liveness probe is correct")
			Expect(len(createdSs.Spec.Template.Spec.Containers) == 1).Should(BeTrue())
			Expect(createdSs.Spec.Template.Spec.Containers[0].LivenessProbe.Handler.Exec.Command[0] == "/broker/bin/artemis check node").Should(BeTrue())
			Expect(k8sClient.Delete(ctx, createdCrd))

			By("check it has gone")
			Eventually(checkCrdDeleted(crd.ObjectMeta.Name, namespace, createdCrd), timeout, interval).Should(BeTrue())
		})

		It("Default Liveness Probe", func() {
			By("By creating a crd without Liveness Probe")
			ctx := context.Background()
			crd := generateArtemisSpec(namespace)
			crd.Spec.AdminUser = "admin"
			crd.Spec.AdminPassword = "password"
			createdCrd := &brokerv1beta1.ActiveMQArtemis{}
			createdSs := &appsv1.StatefulSet{}
			Expect(k8sClient.Create(ctx, &crd)).Should(Succeed())
			Eventually(checkCrdCreated(crd.ObjectMeta.Name, namespace, createdCrd), timeout, interval).Should(BeTrue())
			Expect(createdCrd.Name).Should(Equal(crd.ObjectMeta.Name))

			By("Checking that the Liveness Probe is created")
			Eventually(func() bool {
				key := types.NamespacedName{Name: namer.CrToSS(createdCrd.Name), Namespace: namespace}

				err := k8sClient.Get(ctx, key, createdSs)

				if err != nil {
					return false
				}
				return createdSs.Spec.Template.Spec.Containers[0].LivenessProbe.TCPSocket != nil
			}, timeout, interval).Should(Equal(true))

			Expect(len(createdSs.Spec.Template.Spec.Containers) == 1).Should(BeTrue())
			Expect(createdSs.Spec.Template.Spec.Containers[0].LivenessProbe.Handler.TCPSocket.Port.String() == "8161").Should(BeTrue())
			Expect(k8sClient.Delete(ctx, createdCrd))

			By("check it has gone")
			Eventually(checkCrdDeleted(crd.ObjectMeta.Name, namespace, createdCrd), timeout, interval).Should(BeTrue())
		})
	})

	Context("Readiness Probe Tests", func() {
		It("Override Readiness Probe No Exec", func() {
			By("By creating a crd with Readiness Probe")
			ctx := context.Background()
			crd := generateArtemisSpec(namespace)
			crd.Spec.AdminUser = "admin"
			crd.Spec.AdminPassword = "password"
			readinessProbe := corev1.Probe{}
			readinessProbe.PeriodSeconds = 5
			readinessProbe.InitialDelaySeconds = 6
			readinessProbe.TimeoutSeconds = 7
			readinessProbe.SuccessThreshold = 8
			readinessProbe.FailureThreshold = 9
			crd.Spec.DeploymentPlan.ReadinessProbe = readinessProbe
			createdCrd := &brokerv1beta1.ActiveMQArtemis{}
			createdSs := &appsv1.StatefulSet{}
			By("Deploying the CRD")
			Expect(k8sClient.Create(ctx, &crd)).Should(Succeed())

			By("Making sure that the CRD gets deployed")
			Eventually(checkCrdCreated(crd.ObjectMeta.Name, namespace, createdCrd), timeout, interval).Should(BeTrue())
			Expect(createdCrd.Name).Should(Equal(crd.ObjectMeta.Name))

			By("Checking that Stateful Set is Created with the Readiness Probe")
			Eventually(func() bool {
				key := types.NamespacedName{Name: namer.CrToSS(createdCrd.Name), Namespace: namespace}

				err := k8sClient.Get(ctx, key, createdSs)

				if err != nil {
					return false
				}
				return createdSs.Spec.Template.Spec.Containers[0].ReadinessProbe.Exec != nil
			}, timeout, interval).Should(Equal(true))

			By("Making sure the Readiness probe is correct")
			Expect(len(createdSs.Spec.Template.Spec.Containers) == 1).Should(BeTrue())
			Expect(createdSs.Spec.Template.Spec.Containers[0].ReadinessProbe.Handler.Exec.Command[0] == "/bin/bash").Should(BeTrue())
			Expect(createdSs.Spec.Template.Spec.Containers[0].ReadinessProbe.Handler.Exec.Command[1] == "-c").Should(BeTrue())
			Expect(createdSs.Spec.Template.Spec.Containers[0].ReadinessProbe.Handler.Exec.Command[2] == "/opt/amq/bin/readinessProbe.sh").Should(BeTrue())
			Expect(k8sClient.Delete(ctx, createdCrd))

			By("check it has gone")
			Eventually(checkCrdDeleted(crd.ObjectMeta.Name, namespace, createdCrd), timeout, interval).Should(BeTrue())
		})

		It("Override Readiness Probe Exec", func() {
			By("By creating a crd with Readiness Probe")
			ctx := context.Background()
			crd := generateArtemisSpec(namespace)
			crd.Spec.AdminUser = "admin"
			crd.Spec.AdminPassword = "password"
			exec := corev1.ExecAction{
				Command: []string{"/broker/bin/artemis check node"},
			}
			readinessProbe := corev1.Probe{}
			readinessProbe.Exec = &exec
			crd.Spec.DeploymentPlan.ReadinessProbe = readinessProbe
			createdCrd := &brokerv1beta1.ActiveMQArtemis{}
			createdSs := &appsv1.StatefulSet{}
			By("Deploying the CRD")
			Expect(k8sClient.Create(ctx, &crd)).Should(Succeed())

			By("Making sure that the CRD gets deployed")
			Eventually(checkCrdCreated(crd.ObjectMeta.Name, namespace, createdCrd), timeout, interval).Should(BeTrue())
			Expect(createdCrd.Name).Should(Equal(crd.ObjectMeta.Name))

			By("Checking that Stateful Set is Created with the Readiness Probe")
			Eventually(func() bool {
				key := types.NamespacedName{Name: namer.CrToSS(createdCrd.Name), Namespace: namespace}

				err := k8sClient.Get(ctx, key, createdSs)

				if err != nil {
					return false
				}
				return createdSs.Spec.Template.Spec.Containers[0].ReadinessProbe.Exec != nil
			}, timeout, interval).Should(Equal(true))

			By("Making sure the Readiness probe is correct")
			Expect(len(createdSs.Spec.Template.Spec.Containers) == 1).Should(BeTrue())
			Expect(createdSs.Spec.Template.Spec.Containers[0].ReadinessProbe.Handler.Exec.Command[0] == "/broker/bin/artemis check node").Should(BeTrue())
			Expect(k8sClient.Delete(ctx, createdCrd))

			By("check it has gone")
			Eventually(checkCrdDeleted(crd.ObjectMeta.Name, namespace, createdCrd), timeout, interval).Should(BeTrue())
		})

		It("Default Readiness Probe", func() {
			By("By creating a crd without Readiness Probe")
			ctx := context.Background()
			crd := generateArtemisSpec(namespace)
			crd.Spec.AdminUser = "admin"
			crd.Spec.AdminPassword = "password"
			createdCrd := &brokerv1beta1.ActiveMQArtemis{}
			createdSs := &appsv1.StatefulSet{}
			Expect(k8sClient.Create(ctx, &crd)).Should(Succeed())

			Eventually(checkCrdCreated(crd.ObjectMeta.Name, namespace, createdCrd), timeout, interval).Should(BeTrue())
			Expect(createdCrd.Name).Should(Equal(crd.ObjectMeta.Name))

			By("Checking that the Readiness Probe is created")
			Eventually(func() bool {
				key := types.NamespacedName{Name: namer.CrToSS(createdCrd.Name), Namespace: namespace}

				err := k8sClient.Get(ctx, key, createdSs)

				if err != nil {
					return false
				}
				return createdSs.Spec.Template.Spec.Containers[0].ReadinessProbe.Exec != nil
			}, timeout, interval).Should(Equal(true))

			By("Making sure the Readiness probe is correct")
			Expect(len(createdSs.Spec.Template.Spec.Containers) == 1).Should(BeTrue())
			Expect(createdSs.Spec.Template.Spec.Containers[0].ReadinessProbe.Handler.Exec.Command[0] == "/bin/bash").Should(BeTrue())
			Expect(createdSs.Spec.Template.Spec.Containers[0].ReadinessProbe.Handler.Exec.Command[1] == "-c").Should(BeTrue())
			Expect(createdSs.Spec.Template.Spec.Containers[0].ReadinessProbe.Handler.Exec.Command[2] == "/opt/amq/bin/readinessProbe.sh").Should(BeTrue())
			Expect(k8sClient.Delete(ctx, createdCrd))

			By("check it has gone")
			Eventually(checkCrdDeleted(crd.ObjectMeta.Name, namespace, createdCrd), timeout, interval).Should(BeTrue())
		})
	})

	Context("With deployed controller", func() {
		It("Expect pod desc", func() {
			By("By creating a new crd")
			ctx := context.Background()
			crd := generateArtemisSpec(namespace)
			Expect(k8sClient.Create(ctx, &crd)).Should(Succeed())

			createdCrd := &brokerv1beta1.ActiveMQArtemis{}

			Eventually(checkCrdCreated(crd.ObjectMeta.Name, namespace, createdCrd), timeout, interval).Should(BeTrue())
			Expect(createdCrd.Name).Should(Equal(crd.ObjectMeta.Name))

			// would like more status updates on createdCrd

			By("By checking the status of stateful set")
			Eventually(func() (int, error) {
				key := types.NamespacedName{Name: namer.CrToSS(createdCrd.Name), Namespace: namespace}
				createdSs := &appsv1.StatefulSet{}

				err := k8sClient.Get(ctx, key, createdSs)
				if err != nil {
					fmt.Printf("Error getting ss: %v\n", err)
					return -1, err
				}
				fmt.Printf("CR: %v\n", createdSs)

				// presence is good enough... check on this status just for kicks
				return int(createdSs.Status.Replicas), err
			}, duration, interval).Should(Equal(0))

			By("Checking stopped status of CR because we expect it to fail to deploy")
			Eventually(func() (int, error) {
				key := types.NamespacedName{Name: crd.ObjectMeta.Name, Namespace: namespace}
				err := k8sClient.Get(ctx, key, createdCrd)

				if err != nil {
					fmt.Printf("Error getting CR: %v\n", err)
					return -1, err
				}
				fmt.Printf("CR: %v\n", createdCrd)

				if len(createdCrd.Status.PodStatus.Stopped) > 0 {
					fmt.Printf("Stopped: %v\n", createdCrd.Status.PodStatus.Stopped[0])
				}
				return len(createdCrd.Status.PodStatus.Stopped), nil
			}, timeout, interval).Should(Equal(1))
			Expect(k8sClient.Delete(ctx, &crd)).Should(Succeed())

			By("check it has gone")
			Eventually(checkCrdDeleted(crd.ObjectMeta.Name, namespace, createdCrd), timeout, interval).Should(BeTrue())
		})
	})

})

func generateArtemisSpec(namespace string) brokerv1beta1.ActiveMQArtemis {

	spec := brokerv1beta1.ActiveMQArtemisSpec{}

	toCreate := brokerv1beta1.ActiveMQArtemis{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ActiveMQArtemis",
			APIVersion: brokerv1beta1.GroupVersion.Identifier(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      randString(),
			Namespace: namespace,
		},
		Spec: spec,
	}

	return toCreate
}

func generateOriginalArtemisSpec(namespace string, name string) *brokerv1beta1.ActiveMQArtemis {

	spec := brokerv1beta1.ActiveMQArtemisSpec{}

	toCreate := brokerv1beta1.ActiveMQArtemis{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ActiveMQArtemis",
			APIVersion: brokerv1beta1.GroupVersion.Identifier(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: spec,
	}

	return &toCreate
}

func randString() string {
	rand.Seed(time.Now().UnixNano())
	chars := []rune("abcdefghijklmnopqrstuvwxyz")
	length := 6
	var b strings.Builder
	b.WriteString("broker-")
	for i := 0; i < length; i++ {
		b.WriteRune(chars[rand.Intn(len(chars))])
	}
	return b.String()
}

func checkCrdCreated(name string, nameSpace string, crd *brokerv1beta1.ActiveMQArtemis) bool {
	key := types.NamespacedName{Name: name, Namespace: nameSpace}
	err := k8sClient.Get(ctx, key, crd)
	return err == nil
}

func checkCrdDeleted(name string, namespace string, crd *brokerv1beta1.ActiveMQArtemis) bool {
	//fetched := &pscv1alpha1.PreScaledCronJob{}
	err := k8sClient.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, crd)
	return errors.IsNotFound(err)
}

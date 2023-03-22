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
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/types"

	"encoding/json"

	brokerv1beta1 "github.com/artemiscloud/activemq-artemis-operator/api/v1beta1"
	drainctrl "github.com/artemiscloud/activemq-artemis-operator/pkg/draincontroller"
	"github.com/artemiscloud/activemq-artemis-operator/pkg/resources/environments"
	"github.com/artemiscloud/activemq-artemis-operator/pkg/utils/common"
	"github.com/artemiscloud/activemq-artemis-operator/pkg/utils/namer"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/remotecommand"
)

var _ = Describe("Scale down controller", func() {

	BeforeEach(func() {
		BeforeEachSpec()
	})

	Context("Scale down test", func() {
		It("deploy plan 2 clustered", Label("basic-scaledown-check"), func() {

			// note: we force a non local scaledown cr to exercise creds generation
			// hense only valid with DEPLOY_OPERATOR = false
			// 	see suite_test.go: os.Setenv("OPERATOR_WATCH_NAMESPACE", "SomeValueToCauesEqualitytoFailInIsLocalSoDrainControllerSortsCreds")
			if os.Getenv("USE_EXISTING_CLUSTER") == "true" {

				brokerName := NextSpecResourceName()
				ctx := context.Background()

				brokerCrd := generateOriginalArtemisSpec(defaultNamespace, brokerName)

				booleanTrue := true
				brokerCrd.Spec.DeploymentPlan.Clustered = &booleanTrue
				brokerCrd.Spec.DeploymentPlan.Size = common.Int32ToPtr(2)
				brokerCrd.Spec.DeploymentPlan.PersistenceEnabled = true
				brokerCrd.Spec.DeploymentPlan.ReadinessProbe = &corev1.Probe{
					InitialDelaySeconds: 1,
					PeriodSeconds:       5,
				}
				Expect(k8sClient.Create(ctx, brokerCrd)).Should(Succeed())

				createdBrokerCrd := &brokerv1beta1.ActiveMQArtemis{}

				By("verifying two ready")
				Eventually(func(g Gomega) {

					getPersistedVersionedCrd(brokerCrd.ObjectMeta.Name, defaultNamespace, createdBrokerCrd)
					By("Check ready status")
					g.Expect(len(createdBrokerCrd.Status.PodStatus.Ready)).Should(BeEquivalentTo(2))

				}, existingClusterTimeout, interval).Should(Succeed())

				By("Sending a message to 1")
				podWithOrdinal := namer.CrToSS(brokerCrd.Name) + "-1"

				sendCmd := []string{"amq-broker/bin/artemis", "producer", "--user", "Jay", "--password", "activemq", "--url", "tcp://" + podWithOrdinal + ":61616", "--message-count", "1", "--destination", "DLQ", "--verbose"}

				stdout, _, err := RunCommandInPod(podWithOrdinal, brokerName+"-container", sendCmd)
				Expect(err).To(BeNil())
				Expect(*stdout).Should(ContainSubstring("Produced: 1 messages"))

				By("Scaling down to ss-0")
				podWithOrdinal = namer.CrToSS(brokerCrd.Name) + "-0"

				Eventually(func(g Gomega) {

					getPersistedVersionedCrd(brokerCrd.ObjectMeta.Name, defaultNamespace, createdBrokerCrd)
					createdBrokerCrd.Spec.DeploymentPlan.Size = common.Int32ToPtr(1)
					// not checking return from update as it will error on repeat as there is no change
					// which is expected
					g.Expect(k8sClient.Update(ctx, createdBrokerCrd)).Should(Succeed())

					By("checking statefulset scaled down to 1 pod")
					g.Expect(len(createdBrokerCrd.Status.PodStatus.Ready)).Should(BeEquivalentTo(1))

				}, existingClusterTimeout, interval*2).Should(Succeed())

				By("Checking messsage count on broker 0")
				Eventually(func(g Gomega) {
					// This moment a drainer pod will come up and do the message migration
					// checking message count on broker 0 to make sure scale down finally happens.
					checkMessageCountOnPod(brokerName, podWithOrdinal, "DLQ", 1, g)
				}, existingClusterTimeout, interval*2).Should(Succeed())

				By("Receiving a message from 0")

				rcvCmd := []string{"amq-broker/bin/artemis", "consumer", "--user", "Jay", "--password", "activemq", "--url", "tcp://" + podWithOrdinal + ":61616", "--message-count", "1", "--destination", "DLQ", "--receive-timeout", "10000", "--break-on-null", "--verbose"}
				stdout, _, err = RunCommandInPod(podWithOrdinal, brokerName+"-container", rcvCmd)

				Expect(err).To(BeNil())

				Expect(*stdout).Should(ContainSubstring("JMS Message ID:"))

				Eventually(func(g Gomega) {
					getPersistedVersionedCrd(brokerCrd.ObjectMeta.Name, defaultNamespace, createdBrokerCrd)
					By("flipping message migration state (from default true) on brokerCr")
					booleanFalse := false
					createdBrokerCrd.Spec.DeploymentPlan.MessageMigration = &booleanFalse
					g.Expect(k8sClient.Update(ctx, createdBrokerCrd)).Should(Succeed())
					By("Unset message migration in broker cr")

				}, existingClusterTimeout, existingClusterInterval).Should(Succeed())

				By("verify drain job gone")
				drainJobKey := types.NamespacedName{Name: "job-" + namer.CrToSS(brokerCrd.Name) + "-1", Namespace: defaultNamespace}
				drainJob := &batchv1.Job{}
				Eventually(func(g Gomega) {
					g.Expect(k8sClient.Get(ctx, drainJobKey, drainJob)).ShouldNot(Succeed())
				}, existingClusterTimeout, existingClusterInterval).Should(Succeed())

				By("verify drain pod is deleted")
				jobName := "job-" + namer.CrToSS(brokerCrd.Name) + "-1"
				drainPodSelector := labels.NewSelector()
				jobLabel, err := labels.NewRequirement("job-name", selection.Equals, []string{jobName})
				Expect(err).To(BeNil())
				drainPodSelector = drainPodSelector.Add(*jobLabel)
				Eventually(func(g Gomega) {
					pods := &corev1.PodList{}
					k8sClient.List(context.TODO(), pods, &client.ListOptions{LabelSelector: drainPodSelector, Namespace: defaultNamespace})
					g.Expect(len(pods.Items)).To(Equal(0))
				}, existingClusterTimeout, existingClusterInterval).Should(Succeed())

				By("clean up broker")
				Expect(k8sClient.Delete(ctx, createdBrokerCrd)).Should(Succeed())
			}

		})

		It("scaledown failure test", Label("basic-scaledown-failure"), func() {

			if os.Getenv("USE_EXISTING_CLUSTER") == "true" {

				brokerName := NextSpecResourceName()
				ctx := context.Background()

				brokerCrd, createdBrokerCrd := DeployCustomBroker(defaultNamespace, func(candidate *brokerv1beta1.ActiveMQArtemis) {
					candidate.Name = brokerName
					candidate.Spec.DeploymentPlan.Clustered = &boolTrue
					candidate.Spec.DeploymentPlan.Size = common.Int32ToPtr(2)
					candidate.Spec.DeploymentPlan.PersistenceEnabled = true
					candidate.Spec.DeploymentPlan.MessageMigration = &boolTrue
					candidate.Spec.DeploymentPlan.ReadinessProbe = &corev1.Probe{
						InitialDelaySeconds: 1,
						PeriodSeconds:       5,
					}
				})

				By("verifying two ready")
				Eventually(func(g Gomega) {
					getPersistedVersionedCrd(brokerCrd.ObjectMeta.Name, defaultNamespace, createdBrokerCrd)
					By("Check ready status")
					g.Expect(len(createdBrokerCrd.Status.PodStatus.Ready)).Should(BeEquivalentTo(2))

				}, existingClusterTimeout, interval).Should(Succeed())

				By("Sending a message to 1")
				podWithOrdinal := namer.CrToSS(brokerCrd.Name) + "-1"

				sendCmd := []string{"amq-broker/bin/artemis", "producer", "--user", "Jay", "--password", "activemq", "--url", "tcp://" + podWithOrdinal + ":61616", "--message-count", "1", "--destination", "DLQ", "--verbose"}

				content, _, err := RunCommandInPod(podWithOrdinal, brokerName+"-container", sendCmd)
				Expect(err).To(BeNil())
				Expect(*content).Should(ContainSubstring("Produced: 1 messages"))

				By("Make scaledown to fail")
				drainController := scaleDownRconciler.GetDrainController("*")
				Expect(drainController).NotTo(BeNil())

				failDrainCommand := []string{
					"/bin/sh",
					"-c",
					"echo \"To fail the drainer\" ; exit 1",
				}

				drainController.SetDrainCommand(failDrainCommand)

				By("Scaling down to fail")
				Eventually(func(g Gomega) {
					getPersistedVersionedCrd(brokerCrd.ObjectMeta.Name, defaultNamespace, createdBrokerCrd)
					createdBrokerCrd.Spec.DeploymentPlan.Size = common.Int32ToPtr(1)
					// not checking return from update as it will error on repeat as there is no change
					// which is expected
					g.Expect(k8sClient.Update(ctx, createdBrokerCrd)).Should(Succeed())
				}, existingClusterTimeout, interval*2).Should(Succeed())

				By("Checking job status")
				jobKey := types.NamespacedName{Name: "job-" + namer.CrToSS(brokerCrd.Name) + "-1", Namespace: defaultNamespace}
				job := &batchv1.Job{}
				Eventually(func(g Gomega) {
					g.Expect(k8sClient.Get(ctx, jobKey, job)).Should(Succeed())
					g.Expect(job.Status.Failed).To(Equal(int32(1)))
				}, existingClusterTimeout, existingClusterInterval).Should(Succeed())

				By("checking pod 0 didn't get the message")
				podWithOrdinal = namer.CrToSS(brokerCrd.Name) + "-0"
				Eventually(func(g Gomega) {
					checkMessageCountOnPod(brokerName, podWithOrdinal, "DLQ", 0, g)
				}, existingClusterTimeout, existingClusterInterval).Should(Succeed())

				By("Create new job spec to make it pass")
				ssKey := types.NamespacedName{
					Name:      namer.CrToSS(brokerCrd.Name),
					Namespace: defaultNamespace,
				}
				currentSS := &appsv1.StatefulSet{}
				Eventually(func(g Gomega) {
					g.Expect(k8sClient.Get(ctx, ssKey, currentSS)).Should(Succeed())
				}, existingClusterTimeout, existingClusterInterval).Should(Succeed())

				drainController.SetDrainCommand(drainctrl.DefaultDrainCommand)

				newJobSpec, err := drainController.NewDrainJob(currentSS, 1, job.Name+"-new")
				Expect(err).To(BeNil())
				jobs := k8sClientSet.BatchV1().Jobs(defaultNamespace)

				_, err = jobs.Create(context.TODO(), newJobSpec, metav1.CreateOptions{})
				Expect(err).To(BeNil())

				By("checking new job eventually complete")
				newJobKey := types.NamespacedName{Name: job.Name + "-new", Namespace: defaultNamespace}
				Eventually(func(g Gomega) {
					g.Expect(k8sClient.Get(ctx, newJobKey, newJobSpec)).To(Succeed())
					g.Expect(newJobSpec.Status.Succeeded).To(Equal(int32(1)))
					g.Expect(newJobSpec.Status.Failed).To(Equal(int32(0)))
				}, existingClusterTimeout, existingClusterInterval).Should(Succeed())

				By("checking pod 0 got the message")
				Eventually(func(g Gomega) {
					checkMessageCountOnPod(brokerName, podWithOrdinal, "DLQ", 1, g)
				}, existingClusterTimeout, existingClusterInterval).Should(Succeed())

				By("clean up")
				deletePolicy := metav1.DeletePropagationBackground
				Expect(k8sClient.Delete(ctx, job, &client.DeleteOptions{PropagationPolicy: &deletePolicy}))
				Expect(k8sClient.Delete(ctx, newJobSpec, &client.DeleteOptions{PropagationPolicy: &deletePolicy})).Should(Succeed())
				Expect(k8sClient.Delete(ctx, createdBrokerCrd)).Should(Succeed())
			}
		})
	})

	It("Toleration ok, verify scaledown", Label("scaledown-toleration"), func() {

		// some required services on crc get evicted which invalidates this test of taints
		isOpenshift, err := environments.DetectOpenshift()
		Expect(err).Should(BeNil())

		if !isOpenshift && os.Getenv("USE_EXISTING_CLUSTER") == "true" {

			By("Tainting the node with no schedule")
			Eventually(func(g Gomega) {

				// find our node, take the first one...
				nodes := &corev1.NodeList{}
				g.Expect(k8sClient.List(ctx, nodes, &client.ListOptions{})).Should(Succeed())
				g.Expect(len(nodes.Items) > 0).Should(BeTrue())

				node := nodes.Items[0]
				g.Expect(len(node.Spec.Taints)).Should(BeEquivalentTo(0))
				node.Spec.Taints = []corev1.Taint{{Key: "artemis", Value: "please", Effect: corev1.TaintEffectNoSchedule}}
				g.Expect(k8sClient.Update(ctx, &node)).Should(Succeed())
			}, timeout*2, interval).Should(Succeed())

			By("Creating a crd plan 2,clustered, with matching tolerations")
			ctx := context.Background()
			crd := generateArtemisSpec(defaultNamespace)
			clustered := true
			crd.Spec.DeploymentPlan.Clustered = &clustered
			crd.Spec.DeploymentPlan.Size = common.Int32ToPtr(2)
			crd.Spec.DeploymentPlan.PersistenceEnabled = true
			crd.Spec.DeploymentPlan.ReadinessProbe = &corev1.Probe{
				InitialDelaySeconds: 1,
				PeriodSeconds:       5,
			}

			crd.Spec.DeploymentPlan.Tolerations = []corev1.Toleration{
				{
					Key:      "artemis",
					Value:    "please",
					Operator: corev1.TolerationOpEqual,
					Effect:   corev1.TaintEffectNoSchedule,
				},
			}

			By("Deploying the CRD " + crd.ObjectMeta.Name)
			Expect(k8sClient.Create(ctx, &crd)).Should(Succeed())

			brokerKey := types.NamespacedName{Name: crd.Name, Namespace: crd.Namespace}
			createdCrd := &brokerv1beta1.ActiveMQArtemis{}

			By("veryify pods started as matching taints/tolerations in play")
			Eventually(func(g Gomega) {

				g.Expect(k8sClient.Get(ctx, brokerKey, createdCrd)).Should(Succeed())
				g.Expect(len(createdCrd.Status.PodStatus.Ready)).Should(BeEquivalentTo(2))

			}, existingClusterTimeout, existingClusterInterval).Should(Succeed())

			// need to produce and consume, even if scaledown controller is blocked, it can run once taints are removed
			podWithOrdinal := namer.CrToSS(brokerKey.Name) + "-1"
			By("Sending a message to Host: " + podWithOrdinal)

			sendCmd := []string{"amq-broker/bin/artemis", "producer", "--user", "Jay", "--password", "activemq", "--url", "tcp://" + podWithOrdinal + ":61616", "--message-count", "1"}
			content, _, err := RunCommandInPod(podWithOrdinal, brokerKey.Name+"-container", sendCmd)

			Expect(err).To(BeNil())

			Expect(*content).Should(ContainSubstring("Produced: 1 messages"))

			By("Scaling down to 1")
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, brokerKey, createdCrd)).Should(Succeed())
				createdCrd.Spec.DeploymentPlan.Size = common.Int32ToPtr(1)
				g.Expect(k8sClient.Update(ctx, createdCrd)).Should(Succeed())
				By("Scale down to ss-0 update complete")
			}, existingClusterTimeout, existingClusterInterval).Should(Succeed())

			By("Checking ready 1")
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, brokerKey, createdCrd)).Should(Succeed())
				g.Expect(len(createdCrd.Status.PodStatus.Ready)).Should(BeEquivalentTo(1))
			}, existingClusterTimeout, existingClusterInterval).Should(Succeed())

			podWithOrdinal = namer.CrToSS(createdCrd.Name) + "-0"
			By("Receiving a message from Host: " + podWithOrdinal)

			recvCmd := []string{"amq-broker/bin/artemis", "consumer", "--user", "Jay", "--password", "activemq", "--url", "tcp://" + podWithOrdinal + ":61616", "--message-count", "1", "--receive-timeout", "10000", "--break-on-null", "--verbose"}

			content, _, err = RunCommandInPod(podWithOrdinal, brokerKey.Name+"-container", recvCmd)

			Expect(err).To(BeNil())

			Expect(*content).Should(ContainSubstring("JMS Message ID:"))

			By("reverting taints on node")
			Eventually(func(g Gomega) {

				// find our node, take the first one...
				nodes := &corev1.NodeList{}
				g.Expect(k8sClient.List(ctx, nodes, &client.ListOptions{})).Should(Succeed())
				g.Expect(len(nodes.Items) > 0).Should(BeTrue())

				node := nodes.Items[0]
				g.Expect(len(node.Spec.Taints)).Should(BeEquivalentTo(1))
				node.Spec.Taints = []corev1.Taint{}
				g.Expect(k8sClient.Update(ctx, &node)).Should(Succeed())
			}, timeout*2, interval).Should(Succeed())

			//here the drain job/pod should have gone.
			//should check gone not exitst
			By("checking drain job gone")
			jobKey := types.NamespacedName{Name: "job-" + namer.CrToSS(crd.Name) + "-1", Namespace: defaultNamespace}
			job := &batchv1.Job{}
			Eventually(func(g Gomega) {
				err := k8sClient.Get(ctx, jobKey, job)
				g.Expect(errors.IsNotFound(err)).To(BeTrue(), "error", err)
			}, existingClusterTimeout, existingClusterInterval).Should(Succeed())

			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, brokerKey, createdCrd)).Should(Succeed())
				By("flipping message migration state (from default true) on brokerCr")
				booleanFalse := false
				createdCrd.Spec.DeploymentPlan.MessageMigration = &booleanFalse
				g.Expect(k8sClient.Update(ctx, createdCrd)).Should(Succeed())
				By("Unset message migration in broker cr")

			}, existingClusterTimeout, existingClusterInterval).Should(Succeed())

			Expect(k8sClient.Delete(ctx, createdCrd)).Should(Succeed())
		}
	})

})

func RunCommandInPod(podName string, containerName string, command []string) (*string, *string, error) {
	gvk := schema.GroupVersionKind{
		Group:   "",
		Version: "v1",
		Kind:    "Pod",
	}
	restClient, err := apiutil.RESTClientForGVK(gvk, false, testEnv.Config, serializer.NewCodecFactory(testEnv.Scheme))
	Expect(err).To(BeNil())
	execReq := restClient.
		Post().
		Namespace(defaultNamespace).
		Resource("pods").
		Name(podName).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: containerName,
			Command:   command,
			Stdin:     true,
			Stdout:    true,
			Stderr:    true,
		}, runtime.NewParameterCodec(testEnv.Scheme))

	exec, err := remotecommand.NewSPDYExecutor(testEnv.Config, "POST", execReq.URL())

	if err != nil {
		return nil, nil, err
	}

	var capturedOut bytes.Buffer
	var capturedErr bytes.Buffer

	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  os.Stdin,
		Stdout: &capturedOut,
		Stderr: &capturedErr,
		Tty:    false,
	})

	stdout := capturedOut.String()
	stderr := capturedErr.String()

	return &stdout, &stderr, err
}

func checkMessageCountOnPod(crName string, podName string, queueName string, count int, g Gomega) {
	httpOrigin := "Origin: http://" + podName + ":8161"
	reqUrl := "http://" + podName + ":8161/console/jolokia/read/org.apache.activemq.artemis:broker=\"amq-broker\",address=\"" + queueName + "\",component=addresses,queue=\"" + queueName + "\",routing-type=\"anycast\",subcomponent=queues/MessageCount"
	queryCmd := []string{"curl", "-H", httpOrigin, "-u", "any:any", reqUrl}
	stdout, stderr, err := RunCommandInPod(podName, crName+"-container", queryCmd)
	g.Expect(err).To(BeNil(), "stdout", stdout, "stderr", stderr)
	var response map[string]interface{}
	Expect(json.Unmarshal([]byte(*stdout), &response)).To(Succeed())
	g.Expect(response["status"]).To(BeEquivalentTo(200))
	g.Expect(response["value"]).To(BeEquivalentTo(count))
}

package twointerior

import (
	"github.com/fgiorgetti/qpid-dispatch-go-tests/pkg/framework"
	"github.com/interconnectedcloud/qdr-operator/pkg/apis/interconnectedcloud/v1alpha1"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
)

var (
	// FrameworkQdrOne the instance of the framework that holds deployment of QdrOne
	FrameworkQdrOne *framework.Framework
	// FrameworkQdrTwo the instance of the framework that holds deployment of QdrTwo
	FrameworkQdrTwo *framework.Framework
	// QdrOneSpec specification for Interconnect deployed against framework one
	QdrOneSpec      *v1alpha1.InterconnectSpec
	// QdrTwoSpec specification for Interconnect deployed against framework two
	QdrTwoSpec      *v1alpha1.InterconnectSpec
)

// Constants that defines common information for Interconnects deployed against the different namespaces
const (
	QdrOneName = "qdrone"
	QdrTwoName = "qdrtwo"
)

// This topology is meant to be deployed on distinct contexts on the
// same cluster.
var _ = ginkgo.BeforeEach(func() {

	// Creating two distinct frameworks for same cluster/context
	FrameworkQdrOne = framework.NewFramework("two-interior", framework.TestContext.GetContexts()[0])
	FrameworkQdrTwo = framework.NewFramework("two-interior", framework.TestContext.GetContexts()[0])

})

// Deploys QdrOneSpec, retrieves the generated service URL and create QdrTwoSpec
// setting up a connector to the QdrOneSpec inter-router listener.
var _ = ginkgo.JustBeforeEach(func() {

	// Contexts
	ctxOne := FrameworkQdrOne.GetFirstContext()
	ctxTwo := FrameworkQdrTwo.GetFirstContext()

	// Initialize the Interconnect resources
	QdrOneSpec = &v1alpha1.InterconnectSpec{
		DeploymentPlan: v1alpha1.DeploymentPlanType{
			Size:      1,
			Image:     framework.TestContext.QdrImage,
			Role:      "interior",
			Placement: "Any",
		},
	}

	QdrTwoSpec = &v1alpha1.InterconnectSpec{
		DeploymentPlan: v1alpha1.DeploymentPlanType{
			Size:      1,
			Image:     framework.TestContext.QdrImage,
			Role:      "interior",
			Placement: "Any",
		},
		InterRouterConnectors: []v1alpha1.Connector{
			{
				Host:           QdrOneName + "." + ctxOne.Namespace + ".svc.cluster.local",
				Port:           55672,
				VerifyHostname: false,
			},
		},
	}

	// Creating QdrOneSpec
	ic, err := ctxOne.CreateInterconnectFromSpec(1, QdrOneName, *QdrOneSpec)
	gomega.Expect(err).To(gomega.BeNil())
	gomega.Expect(ic).NotTo(gomega.BeNil())

	// Wait for QdrOne deployment
	err = framework.WaitForDeployment(ctxOne.Clients.KubeClient, ctxOne.Namespace, QdrOneName, 1, framework.RetryInterval, framework.Timeout)
	gomega.Expect(err).To(gomega.BeNil())

	// Creating QdrTwoSpec
	ic, err = ctxTwo.CreateInterconnectFromSpec(1, QdrTwoName, *QdrTwoSpec)
	gomega.Expect(err).To(gomega.BeNil())
	gomega.Expect(ic).NotTo(gomega.BeNil())

	// Wait for QdrTwo deployment
	err = framework.WaitForDeployment(ctxTwo.Clients.KubeClient, ctxTwo.Namespace, QdrTwoName, 1, framework.RetryInterval, framework.Timeout)
	gomega.Expect(err).To(gomega.BeNil())

})

// After each test completes, run cleanup actions to save resources
var _ = ginkgo.AfterEach(func() {
	FrameworkQdrOne.AfterEach()
	FrameworkQdrTwo.AfterEach()
})

package routerbroker

import (
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
)

// Just to illustrate the structure for this test suite
var _ = ginkgo.Describe("RouterBrokerDeployment", func() {

	ginkgo.It("Deploys one interior router and one broker", func() {
		gomega.Expect(int32(1)).To(gomega.Equal(int32(1)))
	})

	ginkgo.It("Validate Auto-Links", func() {
		gomega.Expect(int32(1)).To(gomega.Equal(int32(1)))
	})

	ginkgo.It("Validate Link-Routes", func() {
		gomega.Expect(int32(1)).To(gomega.Equal(int32(1)))
	})
})

package v2alpha5

import (
	"github.com/artemiscloud/activemq-artemis-operator/api/v1beta1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("ActiveMQArtemis Conversion", func() {
	Context("v2alpha5 Redelivery conversion to v1alpha1", func() {
		It("converts a default object", func() {
			objectKey := client.ObjectKey{Name: "example-v2alpha5", Namespace: "foo-ns"}
			spoke := ActiveMQArtemis{
				ObjectMeta: v1.ObjectMeta{
					Name:      objectKey.Name,
					Namespace: objectKey.Namespace,
				},
			}

			hub := &v1beta1.ActiveMQArtemis{}
			Expect(spoke.ConvertTo(hub)).To(Succeed())
		})
	})

	Context("v1alpha1 Redelivery conversion to v2alpha5", func() {
		It("converts a default object", func() {
			objectKey := client.ObjectKey{Name: "example-v1beta1", Namespace: "foo-ns"}
			hub := &v1beta1.ActiveMQArtemis{
				ObjectMeta: v1.ObjectMeta{
					Name:      objectKey.Name,
					Namespace: objectKey.Namespace,
				},
			}

			spoke := &ActiveMQArtemis{}
			Expect(spoke.ConvertFrom(hub)).To(Succeed())
		})
	})
})

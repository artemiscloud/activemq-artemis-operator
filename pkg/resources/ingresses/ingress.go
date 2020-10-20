package ingresses

import (
	svc "github.com/artemiscloud/activemq-artemis-operator/pkg/resources/services"
	//"github.com/artemiscloud/activemq-artemis-operator/pkg/utils/selectors"
	extv1b1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"os"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("package ingresses")

// Create newIngressForCR method to create exposed ingress
//func NewIngressForCR(cr *v2alpha1.ActiveMQArtemis, target string) *extv1b1.Ingress {
func NewIngressForCR(namespacedName types.NamespacedName, labels map[string]string, targetServiceName string, targetPortName string) *extv1b1.Ingress {

	ingress := &extv1b1.Ingress{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "extensions/v1beta1",
			Kind:       "Ingress",
		},
		ObjectMeta: metav1.ObjectMeta{
			Labels:    labels,
			Name:      targetServiceName + "-ing",
			Namespace: namespacedName.Namespace,
		},
		Spec: extv1b1.IngressSpec{
			Rules: []extv1b1.IngressRule{
				{
					Host: os.Getenv("KUBERNETES_SERVICE_HOST"),
					IngressRuleValue: extv1b1.IngressRuleValue{
						HTTP: &extv1b1.HTTPIngressRuleValue{
							Paths: []extv1b1.HTTPIngressPath{
								extv1b1.HTTPIngressPath{
									Path: "/",
									Backend: extv1b1.IngressBackend{
										ServiceName: svc.HeadlessNameBuilder.Name(),
										ServicePort: intstr.FromString(targetPortName),
									},
								},
							},
						},
					},
				},
			},
		},
	}
	return ingress
}

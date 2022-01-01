package ingresses

import (
	//"github.com/artemiscloud/activemq-artemis-operator/pkg/utils/selectors"

	extv1b1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	logf "sigs.k8s.io/controller-runtime"
)

var log = logf.Log.WithName("package ingresses")

//leave this for backward compatibility
func NewIngressForCR(namespacedName types.NamespacedName, labels map[string]string, targetServiceName string, targetPortName string) *extv1b1.Ingress {
	return NewIngressForCRWithSSL(namespacedName, labels, targetServiceName, targetPortName, false)
}

// Create newIngressForCR method to create exposed ingress
//func NewIngressForCR(cr *v2alpha1.ActiveMQArtemis, target string) *extv1b1.Ingress {
func NewIngressForCRWithSSL(namespacedName types.NamespacedName, labels map[string]string, targetServiceName string, targetPortName string, sslEnabled bool) *extv1b1.Ingress {

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
					IngressRuleValue: extv1b1.IngressRuleValue{
						HTTP: &extv1b1.HTTPIngressRuleValue{
							Paths: []extv1b1.HTTPIngressPath{
								extv1b1.HTTPIngressPath{
									Path: "/",
									Backend: extv1b1.IngressBackend{
										ServiceName: targetServiceName,
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
	if sslEnabled {
		//ingress assumes TLS termination at the ingress point and uses default https port 443
		//it requires a host name that matches the certificate's CN value
		ingress.Spec.Rules[0].Host = "www.mgmtconsole.com"
		tls := []extv1b1.IngressTLS{
			{
				Hosts: []string{
					"www.mgmtconsole.com",
				},
				SecretName: "console-secret",
			},
		}
		ingress.Spec.TLS = tls
	}
	return ingress
}

package ingresses

import (
	//"github.com/artemiscloud/activemq-artemis-operator/pkg/utils/selectors"

	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	logf "sigs.k8s.io/controller-runtime"
)

var log = logf.Log.WithName("package ingresses")

// Create newIngressForCR method to create exposed ingress
//func NewIngressForCR(cr *v2alpha1.ActiveMQArtemis, target string) *extv1b1.Ingress {
func NewIngressForCRWithSSL(namespacedName types.NamespacedName, labels map[string]string, targetServiceName string, targetPortName string, sslEnabled bool) *netv1.Ingress {

	portName := ""
	portNumber := -1
	portValue := intstr.FromString(targetPortName)
	if portNumber = portValue.IntValue(); portNumber == 0 {
		portName = portValue.String()
	}

	pathType := netv1.PathTypePrefix

	ingress := &netv1.Ingress{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "networking.k8s.io/v1",
			Kind:       "Ingress",
		},
		ObjectMeta: metav1.ObjectMeta{
			Labels:    labels,
			Name:      targetServiceName + "-ing",
			Namespace: namespacedName.Namespace,
		},
		Spec: netv1.IngressSpec{
			Rules: []netv1.IngressRule{
				{
					IngressRuleValue: netv1.IngressRuleValue{
						HTTP: &netv1.HTTPIngressRuleValue{
							Paths: []netv1.HTTPIngressPath{
								{
									Path:     "/",
									PathType: &pathType,
									Backend: netv1.IngressBackend{
										Service: &netv1.IngressServiceBackend{
											Name: targetServiceName,
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
	if portName == "" {
		ingress.Spec.Rules[0].HTTP.Paths[0].Backend.Service.Port = netv1.ServiceBackendPort{
			Number: int32(portNumber),
		}
	} else {
		ingress.Spec.Rules[0].HTTP.Paths[0].Backend.Service.Port = netv1.ServiceBackendPort{
			Name: portName,
		}
	}
	ingress.Spec.Rules[0].Host = "www.mgmtconsole.com"
	if sslEnabled {
		//ingress assumes TLS termination at the ingress point and uses default https port 443
		//it requires a host name that matches the certificate's CN value
		//it also uses a self-signed cert if you don't provide one
		//this is the default configuration in minikube
		tls := []netv1.IngressTLS{
			{
				Hosts: []string{
					"www.mgmtconsole.com",
				},
			},
		}
		ingress.Spec.TLS = tls
	}
	return ingress
}

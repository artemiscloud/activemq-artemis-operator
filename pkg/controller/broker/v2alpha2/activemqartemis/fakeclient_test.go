package v2alpha2activemqartemis

import (
	brokerv2alpha2 "github.com/artemiscloud/activemq-artemis-operator/pkg/apis/broker/v2alpha2"
	routev1 "github.com/openshift/api/route/v1"

	"testing"

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

//buildReconcileWithFakeClientWithMocks return reconcile with fake client, schemes and mock objects
func buildReconcileWithFakeClientWithMocks(objs []runtime.Object, t *testing.T) *ReconcileActiveMQArtemis {

	registerObjs := []runtime.Object{&brokerv2alpha2.ActiveMQArtemis{}, &corev1.Service{}, &appsv1.StatefulSet{}, &appsv1.StatefulSetList{}, &corev1.Pod{}, &routev1.Route{}, &routev1.RouteList{}, &corev1.PersistentVolumeClaimList{}, &corev1.ServiceList{}}
	registerObjs = append(registerObjs)
	brokerv2alpha2.SchemeBuilder.Register(registerObjs...)
	brokerv2alpha2.SchemeBuilder.Register()

	scheme, err := brokerv2alpha2.SchemeBuilder.Build()
	if err != nil {
		assert.Fail(t, "unable to build scheme")
	}
	client := fake.NewFakeClientWithScheme(scheme, objs...)

	return &ReconcileActiveMQArtemis{
		client: client,
		scheme: scheme,
	}

}

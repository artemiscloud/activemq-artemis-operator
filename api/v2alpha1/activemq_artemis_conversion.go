package v2alpha1

import (
	"github.com/artemiscloud/activemq-artemis-operator/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var log = logf.Log.WithName("v2alpha1Conversion")

func (r *ActiveMQArtemis) ConvertTo(dst conversion.Hub) error {
	log.Info("==== trying converting from v2alpha1 to v1beta1")
	target := dst.(*v1beta1.ActiveMQArtemis)
	log.Info("target got", "dst", target.APIVersion)
	return nil
}

//may not need it if the Hub (storage version) is the latest
func (r *ActiveMQArtemis) ConvertFrom(src conversion.Hub) error {
	return nil
}

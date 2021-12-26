package v2alpha4

import (
	"github.com/artemiscloud/activemq-artemis-operator/api/v2alpha5"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var log = logf.Log.WithName("v2alpha4Conversion")

func (r *ActiveMQArtemis) ConvertTo(dst conversion.Hub) error {
	log.Info("==== trying converting from v2alpha4 to v2alpha5")
	target := dst.(*v2alpha5.ActiveMQArtemis)
	log.Info("target got", "dst", target.APIVersion)
	//target.ObjectMeta = r.ObjectMeta
	return nil
}

//may not need it if the Hub (storage version) is the latest
func (r *ActiveMQArtemis) ConvertFrom(src conversion.Hub) error {
	return nil
}

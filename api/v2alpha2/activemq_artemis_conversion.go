package v2alpha2

import (
	"sigs.k8s.io/controller-runtime/pkg/conversion"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var log = logf.Log.WithName("v2alpha2Conversion")

func (r *ActiveMQArtemis) ConvertTo(dst conversion.Hub) error {
	log.V(1).Info("ConvertTo not implemented")
	return nil
}

//may not need it if the Hub (storage version) is the latest
func (r *ActiveMQArtemis) ConvertFrom(src conversion.Hub) error {
	log.V(1).Info("ConvertFrom not implemented")
	return nil
}

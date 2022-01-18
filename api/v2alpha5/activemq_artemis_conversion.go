package v2alpha5

import (
	"runtime/debug"

	"github.com/artemiscloud/activemq-artemis-operator/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var log = logf.Log.WithName("v2alpha5Conversion")

func (r *ActiveMQArtemis) ConvertTo(dst conversion.Hub) error {
	log.Info("====* trying converting from v2alpha5 to v1beta1")
	debug.PrintStack()
	log.Info("ConverTo called ====================see above stack....")
	target := dst.(*v1beta1.ActiveMQArtemis)
	log.Info("target got", "dst", *target)

	target.ObjectMeta = r.ObjectMeta
	target.Name = r.Name
	target.Namespace = r.Namespace
	target.Spec.DeploymentPlan.Size = r.Spec.DeploymentPlan.Size
	target.Spec.DeploymentPlan.Image = r.Spec.DeploymentPlan.Image
	target.Spec.DeploymentPlan.PersistenceEnabled = r.Spec.DeploymentPlan.PersistenceEnabled

	log.Info("**** converted to v1beta1", "target", *target)
	//Todo: covert all artemis cr data to target

	return nil
}

//may not need it if the Hub (storage version) is the latest
func (r *ActiveMQArtemis) ConvertFrom(src conversion.Hub) error {
	log.Info("<<<<< trying convert from v1beta1 to v2alpha5", "src", src)
	debug.PrintStack()
	log.Info("ConvertFrom called ====================see above stack....")
	source := src.(*v1beta1.ActiveMQArtemis)

	log.Info("target v2alpha5 before convert", "r", *r)

	r.ObjectMeta = source.ObjectMeta
	r.Name = source.Name
	r.Namespace = source.Namespace
	r.Spec.DeploymentPlan.Size = source.Spec.DeploymentPlan.Size
	r.Spec.DeploymentPlan.Image = source.Spec.DeploymentPlan.Image
	r.Spec.DeploymentPlan.PersistenceEnabled = source.Spec.DeploymentPlan.PersistenceEnabled

	log.Info("**** converted to v2alpha5", "target", *r)

	return nil
}

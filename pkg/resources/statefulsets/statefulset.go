package statefulsets

import (
	pvc "github.com/rh-messaging/amq-broker-operator/pkg/resources/persistentvolumeclaims"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"context"
	brokerv1alpha1 "github.com/rh-messaging/amq-broker-operator/pkg/apis/broker/v1alpha1"
	"github.com/rh-messaging/amq-broker-operator/pkg/utils/selectors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	appsv1 "k8s.io/api/apps/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("package statefulsets")



// newPodForCR returns an amqbroker pod with the same name/namespace as the cr
//func newPodForCR(cr *brokerv1alpha1.AMQBroker) *corev1.Pod {
//
//	// Log where we are and what we're doing
//	reqLogger:= log.WithName(cr.Name)
//	reqLogger.Info("Creating new pod for custom resource")
//
//	dataPath := "/opt/" + cr.Name + "/data"
//	userEnvVar := corev1.EnvVar{"AMQ_USER", "admin", nil}
//	passwordEnvVar := corev1.EnvVar{"AMQ_PASSWORD", "admin", nil}
//	dataPathEnvVar := corev1.EnvVar{ "AMQ_DATA_DIR", dataPath, nil}
//
//	pod := &corev1.Pod{
//		ObjectMeta: metav1.ObjectMeta{
//			Name:      cr.Name,
//			Namespace: cr.Namespace,
//			Labels:    cr.Labels,
//		},
//		Spec: corev1.PodSpec{
//			Containers: []corev1.Container{
//				{
//					Name:    cr.Name + "-container",
//					Image:	 cr.Spec.Image,
//					Command: []string{"/opt/amq/bin/launch.sh", "start"},
//					VolumeMounts:	[]corev1.VolumeMount{
//						corev1.VolumeMount{
//							Name:		cr.Name + "-data-volume",
//							MountPath:	dataPath,
//							ReadOnly:	false,
//						},
//					},
//					Env:     []corev1.EnvVar{userEnvVar, passwordEnvVar, dataPathEnvVar},
//				},
//			},
//			Volumes: []corev1.Volume{
//				corev1.Volume{
//					Name: cr.Name + "-data-volume",
//					VolumeSource: corev1.VolumeSource{
//						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
//							ClaimName: cr.Name + "-pvc",
//							ReadOnly:  false,
//						},
//					},
//				},
//			},
//		},
//	}
//
//	return pod
//}


func newPodTemplateSpecForCR(cr *brokerv1alpha1.AMQBroker) corev1.PodTemplateSpec {

	// Log where we are and what we're doing
	reqLogger:= log.WithName(cr.Name)
	reqLogger.Info("Creating new pod template spec for custom resource")

	//var pts corev1.PodTemplateSpec

	dataPath := "/opt/" + cr.Name + "/data"
	userEnvVar := corev1.EnvVar{"AMQ_USER", "admin", nil}
	passwordEnvVar := corev1.EnvVar{"AMQ_PASSWORD", "admin", nil}
	dataPathEnvVar := corev1.EnvVar{ "AMQ_DATA_DIR", dataPath, nil}

	pts := corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
			Labels:    cr.Labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    cr.Name + "-container",
					Image:	 cr.Spec.Image,
					Command: []string{"/opt/amq/bin/launch.sh", "start"},
					VolumeMounts:	[]corev1.VolumeMount{
						{
							Name:		cr.Name,
							MountPath:	dataPath,
							ReadOnly:	false,
						},
					},
					Env:     []corev1.EnvVar{userEnvVar, passwordEnvVar, dataPathEnvVar},
				},
			},
			Volumes: []corev1.Volume{
				{
					Name: cr.Name,
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: cr.Name,
							ReadOnly:  false,
						},
					},
				},
			},
		},
	}

	return pts
}

func newStatefulSetForCR(cr *brokerv1alpha1.AMQBroker) *appsv1.StatefulSet {

	// Log where we are and what we're doing
	reqLogger:= log.WithName(cr.Name)
	reqLogger.Info("Creating new statefulset for custom resource")

	var replicas int32 = 1

	labels := selectors.LabelsForAMQBroker(cr.Name)

	ss := &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:		"StatefulSet",
			APIVersion:	"v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:			cr.Name + "-statefulset",
			Namespace:		cr.Namespace,
			Labels:			labels,
			Annotations:	nil,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: 		&replicas,
			ServiceName:	cr.Name + "-headless" + "-service",
			Selector:		&metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: newPodTemplateSpecForCR(cr),
			VolumeClaimTemplates: *pvc.NewPersistentVolumeClaimArrayForCR(cr, 1),
		},
	}

	return ss
}
func CreateStatefulSet(cr *brokerv1alpha1.AMQBroker, client client.Client, scheme *runtime.Scheme) (*appsv1.StatefulSet, error) {

	// Log where we are and what we're doing
	reqLogger := log.WithValues("AMQBroker Name", cr.Name)
	reqLogger.Info("Creating new statefulset")

	var err error = nil

	// Define the StatefulSet
	ss := newStatefulSetForCR(cr)
	// Set AMQBroker instance as the owner and controller
	if err = controllerutil.SetControllerReference(cr, ss, scheme); err != nil {
		// Add error detail for use later
		reqLogger.Info("Failed to set controller reference for new " + "statefulset")
	}
	reqLogger.Info("Set controller reference for new " + "statefulset")

	// Call k8s create for statefulset
	if err = client.Create(context.TODO(), ss); err != nil {
		// Add error detail for use later
		reqLogger.Info("Failed to creating new " + "statefulset")
	}
	reqLogger.Info("Created new " + "statefulset")

	return ss, err
}
func RetrieveStatefulSet(instance *brokerv1alpha1.AMQBroker, namespacedName types.NamespacedName, client client.Client) (*appsv1.StatefulSet, error) {

	// Log where we are and what we're doing
	reqLogger := log.WithValues("AMQBroker Name", instance.Name)
	reqLogger.Info("Retrieving " + "statefulset")

	var err error = nil
	ss := &appsv1.StatefulSet{}

	// Check if the headless service already exists
	if err = client.Get(context.TODO(), namespacedName, ss); err != nil {
		if errors.IsNotFound(err) {
			reqLogger.Error(err, "Statefulset claim IsNotFound", "Namespace", instance.Namespace, "Name", instance.Name)
		} else {
			reqLogger.Error(err, "Statefulset claim found", "Namespace", instance.Namespace, "Name", instance.Name)
		}
	}

	return ss, err
}

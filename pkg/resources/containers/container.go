package containers

import (
	corev1 "k8s.io/api/core/v1"
)

const (
	graceTime       = 30
	TCPLivenessPort = 8161
)

//func MakeContainer(cr *brokerv2alpha1.ActiveMQArtemis) corev1.Container {
func MakeContainer(customResourceName string, imageName string, envVarArray []corev1.EnvVar) corev1.Container {

	container := corev1.Container{
		Name:    customResourceName + "-container",
		Image:   imageName, //cr.Spec.DeploymentPlan.Image,
		Command: []string{"/opt/amq/bin/launch.sh", "start"},
		Env:     envVarArray, //environments.MakeEnvVarArrayForCR(cr),
	}

	return container
}

func MakeInitContainer(containerName string, imageName string, envVarArray []corev1.EnvVar) corev1.Container {

	container := corev1.Container{
		Name:  containerName,
		Image: imageName,
		Env:   envVarArray,
	}

	return container
}

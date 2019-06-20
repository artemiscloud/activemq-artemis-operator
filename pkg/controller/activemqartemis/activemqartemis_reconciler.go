package activemqartemis

import (
	brokerv1alpha1 "github.com/rh-messaging/activemq-artemis-operator/pkg/apis/broker/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

const (
	statefulSetSizeUpdated          = 1 << 0
	statefulSetClusterConfigUpdated = 1 << 1
	statefulSetSSLConfigUpdated     = 1 << 2
)

type ActiveMQArtemisReconciler struct {
	statefulSetUpdates uint32
}

type ActiveMQArtemisIReconciler interface {
	Process(customResource *brokerv1alpha1.ActiveMQArtemis, currentStatefulSet *appsv1.StatefulSet) uint32
}

func (reconciler *ActiveMQArtemisReconciler) Process(customResource *brokerv1alpha1.ActiveMQArtemis, currentStatefulSet *appsv1.StatefulSet) uint32 {

	// Ensure the StatefulSet size is the same as the spec
	if *currentStatefulSet.Spec.Replicas != customResource.Spec.Size {
		currentStatefulSet.Spec.Replicas = &customResource.Spec.Size
		reconciler.statefulSetUpdates |= statefulSetSizeUpdated
	}

	if clusterConfigSyncCausedUpdateOn(customResource, currentStatefulSet) {
		reconciler.statefulSetUpdates |= statefulSetClusterConfigUpdated
	}

	if sslConfigSyncCausedUpdateOn(customResource, currentStatefulSet) {
		reconciler.statefulSetUpdates |= statefulSetSSLConfigUpdated
	}

	return reconciler.statefulSetUpdates
}

// https://stackoverflow.com/questions/37334119/how-to-delete-an-element-from-a-slice-in-golang
func remove(s []corev1.EnvVar, i int) []corev1.EnvVar {
	s[i] = s[len(s)-1]
	// We do not need to put s[i] at the end, as it will be discarded anyway
	return s[:len(s)-1]
}

func clusterConfigSyncCausedUpdateOn(customResource *brokerv1alpha1.ActiveMQArtemis, currentStatefulSet *appsv1.StatefulSet) bool {

	foundClustered := false
	foundClusterUser := false
	foundClusterPassword := false

	clusteredNeedsUpdate := false
	clusterUserNeedsUpdate := false
	clusterPasswordNeedsUpdate := false

	statefulSetUpdated := false

	// TODO: Remove yuck
	// ensure password and username are valid if can't via openapi validation?
	if customResource.Spec.ClusterConfig.ClusterPassword != "" &&
		customResource.Spec.ClusterConfig.ClusterUserName != "" {

		envVarArray := []corev1.EnvVar{}
		// Find the existing values
		for _, v := range currentStatefulSet.Spec.Template.Spec.Containers[0].Env {
			if v.Name == "AMQ_CLUSTERED" {
				foundClustered = true
				if v.Value == "false" {
					clusteredNeedsUpdate = true
				}
			}
			if v.Name == "AMQ_CLUSTER_USER" {
				foundClusterUser = true
				if v.Value != customResource.Spec.ClusterConfig.ClusterUserName {
					clusterUserNeedsUpdate = true
				}
			}
			if v.Name == "AMQ_CLUSTER_PASSWORD" {
				foundClusterPassword = true
				if v.Value != customResource.Spec.ClusterConfig.ClusterPassword {
					clusterPasswordNeedsUpdate = true
				}
			}
		}

		if !foundClustered || clusteredNeedsUpdate {
			newClusteredValue := corev1.EnvVar{
				"AMQ_CLUSTERED",
				"true",
				nil,
			}
			envVarArray = append(envVarArray, newClusteredValue)
			statefulSetUpdated = true
		}

		if !foundClusterUser || clusterUserNeedsUpdate {
			newClusteredValue := corev1.EnvVar{
				"AMQ_CLUSTER_USER",
				customResource.Spec.ClusterConfig.ClusterUserName,
				nil,
			}
			envVarArray = append(envVarArray, newClusteredValue)
			statefulSetUpdated = true
		}

		if !foundClusterPassword || clusterPasswordNeedsUpdate {
			newClusteredValue := corev1.EnvVar{
				"AMQ_CLUSTER_PASSWORD",
				customResource.Spec.ClusterConfig.ClusterPassword,
				nil,
			}
			envVarArray = append(envVarArray, newClusteredValue)
			statefulSetUpdated = true
		}

		if statefulSetUpdated {
			envVarArrayLen := len(envVarArray)
			if envVarArrayLen > 0 {
				containerArrayLen := len(currentStatefulSet.Spec.Template.Spec.Containers)
				for i := 0; i < containerArrayLen; i++ {
					for j := 0; j < envVarArrayLen; j++ {
						currentStatefulSet.Spec.Template.Spec.Containers[i].Env = append(currentStatefulSet.Spec.Template.Spec.Containers[i].Env, envVarArray[j])
					}
				}
			}
		}
	} else {

		for i := 0; i < len(currentStatefulSet.Spec.Template.Spec.Containers); i++ {
			for j := len(currentStatefulSet.Spec.Template.Spec.Containers[i].Env) - 1; j >= 0; j-- {
				if "AMQ_CLUSTERED" == currentStatefulSet.Spec.Template.Spec.Containers[i].Env[j].Name ||
					"AMQ_CLUSTER_USER" == currentStatefulSet.Spec.Template.Spec.Containers[i].Env[j].Name ||
					"AMQ_CLUSTER_PASSWORD" == currentStatefulSet.Spec.Template.Spec.Containers[i].Env[j].Name {
					currentStatefulSet.Spec.Template.Spec.Containers[i].Env = remove(currentStatefulSet.Spec.Template.Spec.Containers[i].Env, j)
					statefulSetUpdated = true
				}
			}
		}
	}

	return statefulSetUpdated
}

func sslConfigSyncCausedUpdateOn(customResource *brokerv1alpha1.ActiveMQArtemis, currentStatefulSet *appsv1.StatefulSet) bool {

	foundKeystore := false
	foundKeystorePassword := false
	foundKeystoreTruststoreDir := false
	foundTruststore := false
	foundTruststorePassword := false

	keystoreNeedsUpdate := false
	keystorePasswordNeedsUpdate := false
	keystoreTruststoreDirNeedsUpdate := false
	truststoreNeedsUpdate := false
	truststorePasswordNeedsUpdate := false

	statefulSetUpdated := false

	// TODO: Remove yuck
	// ensure password and username are valid if can't via openapi validation?
	if customResource.Spec.SSLConfig.KeyStorePassword != "" &&
		customResource.Spec.SSLConfig.KeystoreFilename != "" {

		envVarArray := []corev1.EnvVar{}
		// Find the existing values
		for _, v := range currentStatefulSet.Spec.Template.Spec.Containers[0].Env {
			if v.Name == "AMQ_KEYSTORE" {
				foundKeystore = true
				if v.Value != customResource.Spec.SSLConfig.KeystoreFilename {
					keystoreNeedsUpdate = true
				}
			}
			if v.Name == "AMQ_KEYSTORE_PASSWORD" {
				foundKeystorePassword = true
				if v.Value != customResource.Spec.SSLConfig.KeyStorePassword {
					keystorePasswordNeedsUpdate = true
				}
			}
			if v.Name == "AMQ_KEYSTORE_TRUSTSTORE_DIR" {
				foundKeystoreTruststoreDir = true
				if v.Value != "/etc/amq-secret-volume" {
					keystoreTruststoreDirNeedsUpdate = true
				}
			}
		}

		if !foundKeystore || keystoreNeedsUpdate {
			newClusteredValue := corev1.EnvVar{
				"AMQ_KEYSTORE",
				customResource.Spec.SSLConfig.KeystoreFilename,
				nil,
			}
			envVarArray = append(envVarArray, newClusteredValue)
			statefulSetUpdated = true
		}

		if !foundKeystorePassword || keystorePasswordNeedsUpdate {
			newClusteredValue := corev1.EnvVar{
				"AMQ_KEYSTORE_PASSWORD",
				customResource.Spec.SSLConfig.KeyStorePassword,
				nil,
			}
			envVarArray = append(envVarArray, newClusteredValue)
			statefulSetUpdated = true
		}

		if !foundKeystoreTruststoreDir || keystoreTruststoreDirNeedsUpdate {
			newClusteredValue := corev1.EnvVar{
				"AMQ_KEYSTORE_TRUSTSTORE_DIR",
				"/etc/amq-secret-volume",
				nil,
			}
			envVarArray = append(envVarArray, newClusteredValue)
			statefulSetUpdated = true
		}

		if statefulSetUpdated {
			envVarArrayLen := len(envVarArray)
			if envVarArrayLen > 0 {
				containerArrayLen := len(currentStatefulSet.Spec.Template.Spec.Containers)
				for i := 0; i < containerArrayLen; i++ {
					for j := 0; j < envVarArrayLen; j++ {
						currentStatefulSet.Spec.Template.Spec.Containers[i].Env = append(currentStatefulSet.Spec.Template.Spec.Containers[i].Env, envVarArray[j])
					}
				}
			}
		}
	} else {

		for i := 0; i < len(currentStatefulSet.Spec.Template.Spec.Containers); i++ {
			for j := len(currentStatefulSet.Spec.Template.Spec.Containers[i].Env) - 1; j >= 0; j-- {
				if "AMQ_KEYSTORE" == currentStatefulSet.Spec.Template.Spec.Containers[i].Env[j].Name ||
					"AMQ_KEYSTORE_PASSWORD" == currentStatefulSet.Spec.Template.Spec.Containers[i].Env[j].Name ||
					"AMQ_KEYSTORE_TRUSTSTORE_DIR" == currentStatefulSet.Spec.Template.Spec.Containers[i].Env[j].Name {
					currentStatefulSet.Spec.Template.Spec.Containers[i].Env = remove(currentStatefulSet.Spec.Template.Spec.Containers[i].Env, j)
					statefulSetUpdated = true
				}
			}
		}
	}

	if customResource.Spec.SSLConfig.TrustStorePassword != "" &&
		customResource.Spec.SSLConfig.TrustStoreFilename != "" {

		envVarArray := []corev1.EnvVar{}
		// Find the existing values
		for _, v := range currentStatefulSet.Spec.Template.Spec.Containers[0].Env {
			if v.Name == "AMQ_TRUSTSTORE" {
				foundTruststore = true
				if v.Value != customResource.Spec.SSLConfig.TrustStoreFilename {
					truststoreNeedsUpdate = true
				}
			}
			if v.Name == "AMQ_TRUSTSTORE_PASSWORD" {
				foundTruststorePassword = true
				if v.Value != customResource.Spec.SSLConfig.TrustStorePassword {
					truststorePasswordNeedsUpdate = true
				}
			}
		}

		if !foundTruststore || truststoreNeedsUpdate {
			newClusteredValue := corev1.EnvVar{
				"AMQ_TRUSTSTORE",
				customResource.Spec.SSLConfig.TrustStoreFilename,
				nil,
			}
			envVarArray = append(envVarArray, newClusteredValue)
			statefulSetUpdated = true
		}

		if !foundTruststorePassword || truststorePasswordNeedsUpdate {
			newClusteredValue := corev1.EnvVar{
				"AMQ_TRUSTSTORE_PASSWORD",
				customResource.Spec.SSLConfig.TrustStorePassword,
				nil,
			}
			envVarArray = append(envVarArray, newClusteredValue)
			statefulSetUpdated = true
		}

		if statefulSetUpdated {
			envVarArrayLen := len(envVarArray)
			if envVarArrayLen > 0 {
				containerArrayLen := len(currentStatefulSet.Spec.Template.Spec.Containers)
				for i := 0; i < containerArrayLen; i++ {
					for j := 0; j < envVarArrayLen; j++ {
						currentStatefulSet.Spec.Template.Spec.Containers[i].Env = append(currentStatefulSet.Spec.Template.Spec.Containers[i].Env, envVarArray[j])
					}
				}
			}
		}
	} else {

		for i := 0; i < len(currentStatefulSet.Spec.Template.Spec.Containers); i++ {
			for j := len(currentStatefulSet.Spec.Template.Spec.Containers[i].Env) - 1; j >= 0; j-- {
				if "AMQ_TRUSTSTORE" == currentStatefulSet.Spec.Template.Spec.Containers[i].Env[j].Name ||
					"AMQ_TRUSTSTORE_PASSWORD" == currentStatefulSet.Spec.Template.Spec.Containers[i].Env[j].Name {
					currentStatefulSet.Spec.Template.Spec.Containers[i].Env = remove(currentStatefulSet.Spec.Template.Spec.Containers[i].Env, j)
					statefulSetUpdated = true
				}
			}
		}
	}

	if statefulSetUpdated {
		sslConfigSyncEnsureSecretVolumeMountExists(customResource, currentStatefulSet)
	}

	return statefulSetUpdated
}

func sslConfigSyncEnsureSecretVolumeMountExists(customResource *brokerv1alpha1.ActiveMQArtemis, currentStatefulSet *appsv1.StatefulSet) {

	secretVolumeExists := false
	secretVolumeMountExists := false

	for i := 0; i < len(currentStatefulSet.Spec.Template.Spec.Volumes); i++ {
		if currentStatefulSet.Spec.Template.Spec.Volumes[i].Name == customResource.Spec.SSLConfig.SecretName {
			secretVolumeExists = true
			break
		}
	}
	if !secretVolumeExists {
		volume := corev1.Volume{
			Name: "broker-secret-volume",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: customResource.Spec.SSLConfig.SecretName,
				},
			},
		}

		currentStatefulSet.Spec.Template.Spec.Volumes = append(currentStatefulSet.Spec.Template.Spec.Volumes, volume)
	}

	for i := 0; i < len(currentStatefulSet.Spec.Template.Spec.Containers); i++ {
		for j := 0; j < len(currentStatefulSet.Spec.Template.Spec.Containers[i].VolumeMounts); j++ {
			if currentStatefulSet.Spec.Template.Spec.Containers[i].VolumeMounts[j].Name == "broker-secret-volume" {
				secretVolumeMountExists = true
				break
			}
		}
		if !secretVolumeMountExists {
			volumeMount := corev1.VolumeMount{
				Name:      "broker-secret-volume",
				MountPath: "/etc/amq-secret-volume",
				ReadOnly:  true,
			}
			currentStatefulSet.Spec.Template.Spec.Containers[i].VolumeMounts = append(currentStatefulSet.Spec.Template.Spec.Containers[i].VolumeMounts, volumeMount)
		}
	}
}

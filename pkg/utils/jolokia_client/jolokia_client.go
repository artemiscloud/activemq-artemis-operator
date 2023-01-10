/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package jolokia_client

import (
	"context"
	"os"
	"strconv"

	brokerv1beta1 "github.com/artemiscloud/activemq-artemis-operator/api/v1beta1"
	"github.com/artemiscloud/activemq-artemis-operator/pkg/resources"
	"github.com/artemiscloud/activemq-artemis-operator/pkg/resources/secrets"
	ss "github.com/artemiscloud/activemq-artemis-operator/pkg/resources/statefulsets"
	mgmt "github.com/artemiscloud/activemq-artemis-operator/pkg/utils/artemis"
	"github.com/artemiscloud/activemq-artemis-operator/pkg/utils/jolokia"
	"github.com/artemiscloud/activemq-artemis-operator/version"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	rtclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

const (
	DEFAULT_JOLOKIA_PROTOCOL = "http"
	BASIC_AUTH               = "Basic"
	TOKEN_AUTH               = "Token"
	TOKEN_AUTH_MIN_VERSION   = 2280
)

type JkInfo struct {
	Artemis *mgmt.Artemis
	IP      string
	Ordinal string
}

func SupportsTokenAuth(specVersion string) bool {
	reqLogger := ctrl.Log.WithValues("SupportsTokenAuth", specVersion)
	if jolokiaAuth, isSet := os.LookupEnv("JOLOKIA_AUTH"); isSet {
		reqLogger.V(5).Info("JOLOKIA_AUTH env var set", "value", jolokiaAuth)
		return jolokiaAuth == TOKEN_AUTH
	}
	compactVersion := version.CompactLatestVersion
	if specVersion != "" {
		compactVersion = version.CompactVersionFromVersion[specVersion]
		reqLogger.V(5).Info("retrieve supported authentication method from version", "spec.Version", specVersion, "compactVersion", compactVersion)
	} else {
		reqLogger.V(5).Info("spec.Version not provided, assuming latest supported", "spec.Version", compactVersion)
	}
	v, err := strconv.Atoi(compactVersion)
	if err != nil {
		reqLogger.Error(err, "unable to convert compactVersion to integer", "compactVersion", compactVersion)
		return true
	}
	if v >= TOKEN_AUTH_MIN_VERSION {
		reqLogger.V(5).Info("Token Authentication supported", "compactVersion", compactVersion)
		return true
	} else {
		reqLogger.V(5).Info("Token Authentication not supported using Basic Authentication", "compactVersion", compactVersion)
	}
	return false
}

func GetBrokers(resource types.NamespacedName, ssInfos []ss.StatefulSetInfo, client rtclient.Client) []*JkInfo {
	reqLogger := ctrl.Log.WithValues("Request.Namespace", resource.Namespace, "Request.Name", resource.Name)

	var artemisArray []*JkInfo = nil

	for n, info := range ssInfos {
		reqLogger.Info("Now retrieve ss", "order", n, "ssName", info.NamespacedName)
		statefulset, err := ss.RetrieveStatefulSet(info.NamespacedName.Name, info.NamespacedName, info.Labels, client)
		if nil != err {
			reqLogger.Error(err, "error retriving ss")
			reqLogger.Info("Statefulset: " + info.NamespacedName.Name + " not found")
		} else {
			reqLogger.Info("Statefulset: " + info.NamespacedName.Name + " found")
			pod := &corev1.Pod{}
			podNamespacedName := types.NamespacedName{
				Name:      statefulset.Name + "-0",
				Namespace: resource.Namespace,
			}

			// For each of the replicas
			var i int = 0
			var replicas int = int(*statefulset.Spec.Replicas)
			reqLogger.Info("finding pods in ss", "replicas", replicas)
			for i = 0; i < replicas; i++ {
				s := statefulset.Name + "-" + strconv.Itoa(i)
				podNamespacedName.Name = s
				reqLogger.Info("Trying finding pod " + s)
				if err = client.Get(context.TODO(), podNamespacedName, pod); err != nil {
					if errors.IsNotFound(err) {
						reqLogger.Error(err, "Pod IsNotFound", "Namespace", resource.Namespace, "Name", resource.Name)
					} else {
						reqLogger.Error(err, "Pod lookup error", "Namespace", resource.Namespace, "Name", resource.Name)
					}
				} else {
					reqLogger.Info("Pod found", "Namespace", resource.Namespace, "Name", resource.Name)
					containers := pod.Spec.Containers //get env from this

					jolokiaProtocol := resolveJolokiaProtocol(client, &containers, podNamespacedName, statefulset, info.Labels)

					jolokiaAuth, err := resolveJolokiaAuth(resource, client, &containers, podNamespacedName, statefulset, info.Labels)
					if err != nil {
						reqLogger.Error(err, "Unable to retrieve service account token")
						return nil
					}
					reqLogger.Info("New Jolokia with ", "Protocol: ", jolokiaProtocol, "broker ip", pod.Status.PodIP)

					artemis := mgmt.GetArtemis(pod.Status.PodIP, "8161", "amq-broker", jolokiaAuth, jolokiaProtocol)
					jkInfo := JkInfo{
						Artemis: artemis,
						IP:      pod.Status.PodIP,
						Ordinal: strconv.Itoa(i),
					}
					artemisArray = append(artemisArray, &jkInfo)
				}
			}
		}
	}

	reqLogger.Info("Finally we gathered some mgmt array", "size", len(artemisArray))
	return artemisArray
}

func resolveJolokiaProtocol(client rtclient.Client,
	containers *[]corev1.Container,
	podNamespacedName types.NamespacedName,
	statefulset *appsv1.StatefulSet,
	labels map[string]string) string {

	if len(*containers) == 1 {
		envVars := (*containers)[0].Env
		for _, oneVar := range envVars {
			if oneVar.Name == "AMQ_CONSOLE_ARGS" {
				jolokiaProtocol := getEnvVarValue(&oneVar, &podNamespacedName, statefulset, client, labels)
				if jolokiaProtocol == "https" {
					return jolokiaProtocol
				}
				return DEFAULT_JOLOKIA_PROTOCOL
			}
		}
	}
	return DEFAULT_JOLOKIA_PROTOCOL
}

func resolveJolokiaAuth(resource types.NamespacedName, client rtclient.Client,
	containers *[]corev1.Container,
	podNamespacedName types.NamespacedName,
	statefulset *appsv1.StatefulSet,
	labels map[string]string) (jolokia.Auth, error) {
	reqLogger := ctrl.Log.WithValues("Request.Namespace", resource.Namespace, "Request.Name", resource.Name)
	version := getVersionFromStatefulset(statefulset, client)
	// Token Auth
	if SupportsTokenAuth(version) {
		cfg, err := config.GetConfig()
		if err != nil {
			reqLogger.Error(err, "unable to read kubernetes configuration")
			return nil, err
		}

		jolokiaToken := cfg.BearerToken
		if jolokiaToken != "" {
			reqLogger.Info("retrieved service account token")
		} else {
			reqLogger.Info("empty service account token found, skipping...")
		}
		return &jolokia.TokenAuth{
			Token: jolokiaToken,
		}, nil
	}

	// Basic Auth
	jolokiaSecretName := resource.Name + "-jolokia-secret"
	var jolokiaUser, jolokiaPassword string
	userDefined := false
	jolokiaUserFromSecret := secrets.GetValueFromSecret(resource.Namespace, jolokiaSecretName, "jolokiaUser", labels, client)
	if jolokiaUserFromSecret != nil {
		userDefined = true
		jolokiaUser = *jolokiaUserFromSecret
	}
	if userDefined {
		jolokiaPasswordFromSecret := secrets.GetValueFromSecret(resource.Namespace, jolokiaSecretName, "jolokiaPassword", labels, client)
		if jolokiaPasswordFromSecret != nil {
			jolokiaPassword = *jolokiaPasswordFromSecret
		}
	}
	if len(*containers) == 1 {
		envVars := (*containers)[0].Env
		for _, oneVar := range envVars {
			if !userDefined && oneVar.Name == "AMQ_USER" {
				jolokiaUser = getEnvVarValue(&oneVar, &podNamespacedName, statefulset, client, labels)
			}
			if !userDefined && oneVar.Name == "AMQ_PASSWORD" {
				jolokiaPassword = getEnvVarValue(&oneVar, &podNamespacedName, statefulset, client, labels)
			}
			if jolokiaUser != "" && jolokiaPassword != "" {
				break
			}
		}
	}

	return &jolokia.BasicAuth{
		User:     jolokiaUser,
		Password: jolokiaPassword,
	}, nil

}

func getVersionFromStatefulset(sts *appsv1.StatefulSet, client rtclient.Client) string {
	reqLogger := ctrl.Log.WithValues("Namespace", sts.Namespace, "StatefulSet", sts.Name)
	reqLogger.V(5).Info("retrieving version from statefulset", "statefulset", sts.Name)
	for _, owner := range sts.OwnerReferences {
		if owner.Kind == "ActiveMQArtemis" {
			cr := &brokerv1beta1.ActiveMQArtemis{}
			crKey := types.NamespacedName{
				Name:      owner.Name,
				Namespace: sts.Namespace,
			}
			err := client.Get(context.TODO(), crKey, cr)
			if err == nil {
				reqLogger.V(5).Info("found ActiveMQArtemis from statefulset", "statefulset", sts.Name, "version", cr.Spec.Version)
				return cr.Spec.Version
			} else {
				reqLogger.V(5).Error(err, "unable to retrieve ActiveMQArtemis from statefulset", "statefulset", sts.Name)
			}
		}
	}
	reqLogger.V(5).Info("unable to determine resource version from statefulset", "statefulset", sts.Name)
	return ""
}

func getEnvVarValue(envVar *corev1.EnvVar, namespace *types.NamespacedName, statefulset *appsv1.StatefulSet, client rtclient.Client, labels map[string]string) string {
	var result string
	if envVar.Value == "" {
		result = getEnvVarValueFromSecret(envVar.Name, envVar.ValueFrom, namespace, statefulset, client, labels)
	} else {
		result = envVar.Value
	}
	return result
}

func getEnvVarValueFromSecret(envName string, varSource *corev1.EnvVarSource, namespace *types.NamespacedName, statefulset *appsv1.StatefulSet, client rtclient.Client, labels map[string]string) string {

	reqLogger := ctrl.Log.WithValues("Namespace", namespace.Name, "StatefulSet", statefulset.Name)

	var result string = ""

	secretName := varSource.SecretKeyRef.LocalObjectReference.Name
	secretKey := varSource.SecretKeyRef.Key

	namespacedName := types.NamespacedName{
		Name:      secretName,
		Namespace: statefulset.Namespace,
	}
	// Attempt to retrieve the secret
	stringDataMap := map[string]string{
		envName: "",
	}
	theSecret := secrets.NewSecret(namespacedName, secretName, stringDataMap, labels)
	var err error = nil
	if err = resources.Retrieve(namespacedName, client, theSecret); err != nil {
		if errors.IsNotFound(err) {
			reqLogger.Info("Secret IsNotFound.", "Secret Name", secretName, "Key", secretKey)
		}
	} else {
		elem, ok := theSecret.Data[envName]
		if ok {
			result = string(elem)
		}
	}
	return result
}

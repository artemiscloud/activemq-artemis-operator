/*
Copyright 2021.

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

package certutil

import (
	"context"
	"fmt"
	"strings"

	"github.com/artemiscloud/activemq-artemis-operator/api/v1beta1"
	"github.com/artemiscloud/activemq-artemis-operator/pkg/resources"
	cm "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	rtclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func GetPKCS12StorePasswordFromCert(cert *cm.Certificate, client rtclient.Client) ([]byte, error) {
	secretName := cert.Spec.Keystores.PKCS12.PasswordSecretRef.Name
	passwordKey := cert.Spec.Keystores.PKCS12.PasswordSecretRef.Key
	return getKeystorePasswordFromCert(secretName, passwordKey, cert, client)
}

func GetJKSStorePasswordFromCert(cert *cm.Certificate, client rtclient.Client) ([]byte, error) {
	secretName := cert.Spec.Keystores.JKS.PasswordSecretRef.Name
	passwordKey := cert.Spec.Keystores.JKS.PasswordSecretRef.Key
	return getKeystorePasswordFromCert(secretName, passwordKey, cert, client)
}

// cluster-issuer is cluster wise res, so if it is used
// the operator needs to be cluster wise to access it. need add a validation somewhere
func getKeystorePasswordFromCert(secretName string, passwordKey string, cert *cm.Certificate, client rtclient.Client) ([]byte, error) {
	secretNamespace := cert.Namespace
	secretKey := types.NamespacedName{Name: secretName, Namespace: secretNamespace}
	secret := corev1.Secret{}
	if err := client.Get(context.TODO(), secretKey, &secret); err != nil {
		return nil, err
	}
	return secret.Data[passwordKey], nil
}

// resolve namespaced name for a cert
// cert is a string of name and namespace separated by a colon
func GetCertKey(cert *string, defNamespace string) (*types.NamespacedName, error) {
	values := strings.Split(*cert, ":")
	if len(values) > 2 {
		return nil, fmt.Errorf("malformed cert string, expected name:namespace, but %s", *cert)
	}
	if len(values) == 2 {
		return &types.NamespacedName{Name: values[0], Namespace: values[1]}, nil
	}
	return &types.NamespacedName{Name: values[0], Namespace: defNamespace}, nil
}

func GetCertificate(brokerCert *string, cr *v1beta1.ActiveMQArtemis, client rtclient.Client) (*cm.Certificate, error) {
	key, err := GetCertKey(brokerCert, cr.Namespace)
	if err != nil {
		return nil, err
	}
	cert := cm.Certificate{}
	if err := resources.Retrieve(*key, client, &cert); err != nil {
		return nil, err
	}
	return &cert, nil
}

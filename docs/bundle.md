# Bunding A Bundle and Deploy it into the Operator Lifecycle Manager(OLM)

## About the Operator Lifecycle Manager (OLM)

The [Operator Lifecycle Manager](https://olm.operatorframework.io/) can help users to install and manage operators.
The ArtemisCloud operator can be built into a bundle and installed into OLM.

## Building the bundle

### Creating the bundle's manifests/metadata

Before you build the bundle image
```$xslt
make IMAGE_TAG_BASE=<bundle image registry> OPERATOR_IMAGE_REPO=<operator image registry> OPERATOR_VERSION=<operator tag> bundle
```

to build bundle image:

make IMAGE_TAG_BASE=quay.io/hgao/operator bundle-build

to push bundle image to remote repo:

make IMAGE_TAG_BASE=quay.io/hgao/operator bundle-push

to build catalog image (index image):

make IMAGE_TAG_BASE=quay.io/hgao/operator catalog-build

to push catalog image to repo:

make IMAGE_TAG_BASE=quay.io/hgao/operator catalog-push

to test bundle(kubernetes)

1. install olm

operator-sdk olm install

2. wait for all pods in olm namespace are up and running

3. create a catalog source (catalog-source.yaml):

```
  apiVersion: operators.coreos.com/v1alpha1
    kind: CatalogSource
    metadata:
      name: artemis-index
      namespace: olm
    spec:
      sourceType: grpc
      image: quay.io/hgao/operator-catalog:v0.0.1
      displayName: ArtemisCloud Index
      publisher: ArtemisCloud
      updateStrategy:
        registryPoll:
          interval: 10m
```

and deploy it:

  kubectl create -f catalog-source.yaml

4. create a subscription (subscription.yaml):

        apiVersion: operators.coreos.com/v1alpha1
        kind: Subscription
        metadata:
          name: my-subscription
          namespace: operators
        spec:
          channel: upstream
          name: activemq-artemis-operator
          source: artemis-index
          sourceNamespace: olm
          installPlanApproval: Automatic

and deploy it:

  kubectl create -f subscription.yaml

5. Watch the operator is up and running in **operators** namespace

(the operator watches all namespaces)

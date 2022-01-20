to create bundle manifests/metadata:

make IMAGE_TAG_BASE=quay.io/hgao/operator OPERATOR_IMAGE_REPO=quay.io/hgao/operator OPERATOR_VERSION=new0.20.1 bundle

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

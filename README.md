# ActiveMQ Artemis Operator

This project is a [Kubernetes](https://kubernetes.io/) [operator](https://coreos.com/blog/introducing-operators.html)
to manage the [Apache ActiveMQ Artemis](https://activemq.apache.org/artemis/) message broker.

## Status


## Building 

Currently the head of the code doesn't compile.

Checkout commit 953b0ff7d0b48ef964c243236b9d6e8cc64f13e3 in order to try building and deploying operator
with the following instructions:

test env:
go version 1.16
Minikube v1.21.0
operator-sdk v1.15.0


to build and push docker image:

make OPERATOR_IMAGE_REPO=quay.io/hgao/operator OPERATOR_VERSION=new0.20.1 docker-build
make OPERATOR_IMAGE_REPO=quay.io/hgao/operator OPERATOR_VERSION=new0.20.1 docker-push

(give OPERATOR_IMAGE_REPO and OPERATOR_VERSION proper values based on your env)

to test your operator image:

make OPERATOR_IMAGE_REPO=quay.io/hgao/operator OPERATOR_VERSION=new0.20.1 deploy
(for now this output to tmp/deploy.yaml instead of deploy directly to cluster)

kubectl create -f tmp/deploy.yaml

The operator will be deployed to namespace activemq-artemis-operator and watch all namespaces.

Now you can deploy a broker CR

apiVersion: broker.amq.io/v2alpha5
kind: ActiveMQArtemis
metadata:
  name: ex-v2alpha5
spec:
  deploymentPlan:
    size: 1
    image: placeholder
    persistenceEnabled: true

Check out the operator log to see some actions going. (basically just logs and doing nothing
because the reconciler code is empty at this point)

to undeploy:
make OPERATOR_IMAGE_REPO=quay.io/hgao/operator OPERATOR_VERSION=new0.20.1 undeploy

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

5. Watch the operator is up and running in olm namespace
(the operator watches all namespaces)




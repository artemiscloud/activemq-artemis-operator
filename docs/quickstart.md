# Overview

At the moment these instructions have been tested against OpenShift 3.11,
other kubernetes or OpenShift environments may require minor adjustment.

One important note about operators in general is that to get the operator
installed requires cluster-admin level privileges. Once installed, a regular
user should be able to install the AMQ Broker via the provided custom
resource.

## Quick Start

### Getting the code

To launch the operator you will need to clone the [activemq-artemis-operator](https://github.com/rh-messaging/activemq-artemis-operator)
and checkout the 0.2.0 tag as per

```$xslt
git clone https://github.com/rh-messaging/activemq-artemis-operator
git checkout 0.2.0
```

### Deploying the operator

In the activemq-artemis-operator/deploy directory you should see [operator.yaml](https://github.com/rh-messaging/activemq-artemis-operator/blob/0.2.0/deploy/operator.yaml)
within which you will want to update the [spec.containers.image](https://github.com/rh-messaging/activemq-artemis-operator/blob/0.2.0/deploy/operator.yaml#L18-L19)
with the correct location for the activemq-artemis-operator container image that you either pulled or [built](building.md).

As per the [operator-framework/operator-sdk](https://github.com/operator-framework/operator-sdk) Quick Start we first
deploy the operator via:

```$xslt
# Setup Service Account
$ kubectl create -f deploy/service_account.yaml
# Setup RBAC
$ kubectl create -f deploy/role.yaml
$ kubectl create -f deploy/role_binding.yaml
# Setup the CRD
$ kubectl create -f deploy/crds/broker_v1alpha1_amqbroker_crd.yaml
# Deploy the activemq-artemis-operator
$ kubectl create -f deploy/operator.yaml
```

At this point in your project namespace you should see the activemq-artemis-operator
starting up and if you check the logs you should see something like

```$xslt
Executing entrypoint
exec /activemq-artemis-operator/activemq-artemis-operator
{"level":"info","ts":1552055596.8029804,"logger":"cmd","msg":"Go Version: go1.11.5"}
{"level":"info","ts":1552055596.803102,"logger":"cmd","msg":"Go OS/Arch: linux/amd64"}
{"level":"info","ts":1552055596.8031216,"logger":"cmd","msg":"Version of operator-sdk: v0.5.0"}
{"level":"info","ts":1552055596.8044093,"logger":"leader","msg":"Trying to become the leader."}
{"level":"info","ts":1552055597.0208154,"logger":"leader","msg":"Found existing lock with my name. I was likely restarted."}
{"level":"info","ts":1552055597.0208917,"logger":"leader","msg":"Continuing as the leader."}
{"level":"info","ts":1552055597.159428,"logger":"cmd","msg":"Registering Components."}
{"level":"info","ts":1552055597.160093,"logger":"kubebuilder.controller","msg":"Starting EventSource","controller":"amqbroker-controller","source":"kind source: /, Kind="}
```

### Deploying the broker

Now that the operator is running and listening for changes related to our crd we can deploy our basic broker custom
resource instance for 'example-amqbroker' from [broker_v1alpha1_amqbroker_cr.yaml](https://github.com/rh-messaging/activemq-artemis-operator/blob/0.2.0/deploy/crds/broker_v1alpha1_amqbroker_cr.yaml)
which looks like

```$xslt
apiVersion: broker.amq.io/v1alpha1
kind: ActiveMQArtemis
metadata:
  name: example-amqbroker
spec:
  # Add fields here
  size: 4
  image: registry.access.redhat.com/amq-broker-7/amq-broker-72-openshift:latest
```  

Note in particular the [spec.image:](https://github.com/rh-messaging/activemq-artemis-operator/blob/0.2.0/deploy/crds/broker_v1alpha1_amqbroker_cr.yaml#L8)
which identifies the container image to use to launch the AMQ Broker.

To deploy the broker we simply execute

```$xslt
kubectl create -f activemq-artemis-operator/deploy/crds/broker_v1alpha1_amqbroker_cr.yaml
```

at which point you should now see the statefulset 'example-amqbroker-statefulset' under Overview in the web console.

```$xslt
amqbroker.broker.amq.io "example-amqbroker" created
```

If you expand the statefulset you should see that it has 1 pod running along with three services under 'Networking'; 
example-amqbroker-console-jolokia-service, example-amqbroker-mux-protocol-service, and headless-service.

#### exampleamq-broker-console-jolokia-service

The broker hosts its own web console at port 8161 and this service provides external access to it via a 
[NodePort](https://kubernetes.io/docs/concepts/services-networking/service/#nodeport)
that provides direct access to the web console. Note that the actual port number chosen is, by default, in the range
of 30000-32767 so you will want to click in here to get the actual port number needed to access the console.

Once you have the port number using http://any.ocp.node.ip:consoleNodePortNumber should allow you to access the running
brokers web console.


#### exampleamq-broker-mux-protocol-service

The broker has a multiplexed protocol port at 61616 supporting all protocols; AMQP, CORE, HornetQ, MQTT, OpenWire and STOMP.
This broker port is exposed via a [NodePort](https://kubernetes.io/docs/concepts/services-networking/service/#nodeport)
as is the web console. 

To produce or consume messages to the broker from outside the [OCP SDN](https://docs.openshift.com/container-platform/3.11/admin_guide/sdn_troubleshooting.html)
you should be able to produce or consume messages to protocol://://any.ocp.node.ip:muxProtocolNodePortNumber.


#### headless-service

The headless service internally exposes all broker ports: 
- 1883: MQTT
- 5672: AMQP
- 8161: Web Console / Jolokia
- 61613: STOMP
- 61616: All protocols as above 


#### Running pods

The example-amqbroker-statefulset-0 pod should now also be visible via the OpenShift console via the Applications -> Pods
sub menu, or by drilling down through the statefulset. The number of running pods can be adjusted via the 'oc scale' command as per:

```$xslt
oc scale --replicas 0 statefulset example-amqbroker-statefulset
```

which would scale down the number of pods to zero. Note that you can scale the number to greater than one, however for the moment
they share the same persistent volume claim and as such the first will be the master while subsequent pods will attempt
to gain the lock, fail, and remain in slave mode unless the master pod is killed. This will be addressed in the next,
0.3.0, release.
 
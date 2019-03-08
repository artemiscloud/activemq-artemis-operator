# Overview

At the moment these instructions have been tested against OpenShift 3.11,
other kubernetes or OpenShift environments may require minor adjustment.

One important note about operators in general is that to get the operator
installed requires cluster-admin level privileges. Once installed, a regular
user should be able to install the AMQ Broker via the provided custom
resource.

## Quick Start

### Getting the code

To launch the operator you will need to clone the [amq-broker-operator](https://github.com/rh-messaging/amq-broker-operator)
and checkout the 0.1.0 tag as per

```$xslt
git clone https://github.com/rh-messaging/amq-broker-operator
git checkout 0.1.0
```

### Deploying the operator

In the amq-broker-operator/deploy directory you should see [operator.yaml](https://github.com/rh-messaging/amq-broker-operator/blob/master/deploy/operator.yaml)
within which you will want to update the [spec.containers.image](https://github.com/rh-messaging/amq-broker-operator/blob/master/deploy/operator.yaml#L18-L19)
with the correct location for the amq-broker-operator container image that you either pulled or [built](building.md).

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
# Deploy the amq-broker-operator
$ kubectl create -f deploy/operator.yaml
```

At this point in your project namespace you should see the amq-broker-operator
starting up and if you check the logs you should see something like

```$xslt
Executing entrypoint
exec /amq-broker-operator/amq-broker-operator
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
resource instance for 'example-amqbroker' from [broker_v1alpha1_amqbroker_cr.yaml](https://github.com/rh-messaging/amq-broker-operator/blob/master/deploy/crds/broker_v1alpha1_amqbroker_cr.yaml)
which looks like

```$xslt
apiVersion: broker.amq.io/v1alpha1
kind: AMQBroker
metadata:
  name: example-amqbroker
spec:
  # Add fields here
  size: 4
  image: registry.access.redhat.com/amq-broker-7/amq-broker-72-openshift:latest
```  

Note in particular the [spec.image:](https://github.com/rh-messaging/amq-broker-operator/blob/master/deploy/crds/broker_v1alpha1_amqbroker_cr.yaml#L8)
which identifies the container image to use to launch the AMQ Broker.

To deploy the broker we simply execute

```$xslt
kubectl create -f amq-broker-operator/deploy/crds/broker_v1alpha1_amqbroker_cr.yaml
```

at which point you should see in the terminal

```$xslt
amqbroker.broker.amq.io "example-amqbroker" created
```
 
The example-amqbroker pod should now also be visible via the OpenShift console via the Applications -> Pods sub menu. At
this point the broker is running but not yet accessible.
 
### Enabling access

The broker hosts its own web console at port 8161, and has a multiplexed protocol port at 61616 supporting 
AMQP, CORE, HornetQ, MQTT, OpenWire and STOMP. Two separate yaml service definitions have been provided to
support access to each of these ports; [console_service.yaml](https://github.com/rh-messaging/amq-broker-operator/blob/master/deploy/console_service.yaml)
and [mux_protocol_service.yaml](https://github.com/rh-messaging/amq-broker-operator/blob/master/deploy/mux_protocol_service.yaml)
respectively.

Executing these to add nodeport backed services looks like:

```$xslt
$ kubectl create -f amq-broker-operator/deploy/console_service.yaml                      
service "example-amqbroker-amq-jolokia" create                                                                                                                                                                                                
$ kubectl create -f amq-broker-operator/deploy/mux_protocol_service.yaml                                                                                                                                         
service "example-amqbroker-amq-mux" created                                                                                                                                                                                                    
```

Once created the console should be available via http at port 31161 on any node in the cluster while the multiplexed multi-protocol
port should be available at 31616 to produce or consume from.
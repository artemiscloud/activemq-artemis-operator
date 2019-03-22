# Overview

At the moment these instructions have been tested against OpenShift 3.11,
other kubernetes or OpenShift environments may require minor adjustment.

One important note about operators in general is that to get the operator
installed requires cluster-admin level privileges. Once installed, a regular
user should be able to install ActiveMQ Artemis via the provided custom
resource.

## Quick Start

### Getting the code

To launch the operator you will need to clone the [activemq-artemis-operator](https://github.com/rh-messaging/activemq-artemis-operator)
and checkout the 0.3.1 tag as per

```$xslt
git clone https://github.com/rh-messaging/activemq-artemis-operator
git checkout 0.3.1
```

### Deploying the operator

In the activemq-artemis-operator/deploy directory you should see [operator.yaml](https://github.com/rh-messaging/activemq-artemis-operator/blob/0.3.1/deploy/operator.yaml)
within which you will want to update the [spec.containers.image](https://github.com/rh-messaging/activemq-artemis-operator/blob/0.3.1/deploy/operator.yaml#L18-L19)
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
$ kubectl create -f deploy/crds/broker_v1alpha1_activemqartemis_crd.yaml
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
{"level":"info","ts":1552055597.160093,"logger":"kubebuilder.controller","msg":"Starting EventSource","controller":"activemqartemis-controller","source":"kind source: /, Kind="}
```

### Deploying the broker

Now that the operator is running and listening for changes related to our crd we can deploy our basic broker custom
resource instance for 'example-activemqartemis' from [broker_v1alpha1_activemqartemis_cr.yaml](https://github.com/rh-messaging/activemq-artemis-operator/blob/0.3.1/deploy/crds/broker_v1alpha1_activemqartemis_cr.yaml)
which looks like

```$xslt
apiVersion: broker.amq.io/v1alpha1
kind: ActiveMQArtemis
metadata:
  name: example-activemqartemis
spec:
  # Add fields here
  size: 4
  image: registry.access.redhat.com/amq-broker-7/amq-broker-72-openshift:latest
```  

Note in particular the [spec.image:](https://github.com/rh-messaging/activemq-artemis-operator/blob/0.3.1/deploy/crds/broker_v1alpha1_activemqartemis_cr.yaml#L8)
which identifies the container image to use to launch the AMQ Broker.

To deploy the broker we simply execute

```$xslt
kubectl create -f deploy/crds/broker_v1alpha1_activemqartemis_cr.yaml
```

at which point you should now see the statefulset 'example-activemqartemis-ss' under Overview in the web console.

```$xslt
activemqartemis.broker.amq.io "example-activemqartemis" created
```

If you expand the example-activemqartemis-ss statefulset you should see that it has 1 pod running along with two services
under 'Networking'; hs which is short for headless service, and ping. The headless service provides access to all protocol
ports that the broker is listening on. The ping service is a service utilized by the brokers for discovery and allows them
to form a cluster properly within the OpenShift environment.

At the moment external access is limited in that the nodePort services, console-jolokia, and mux-protocol, acts as
load balancers meaning traffic sent to them gets round robined by default. To have at least basic functionality at
the moment the [session affinity](https://kubernetes.io/docs/concepts/services-networking/service/) is set to be
sticky to the client ip so that as you navigate through the console each click doesn't go to a separate broker.
This is an area that requires further development.

#### headless-service

The headless service internally exposes all broker ports: 
- 1883: MQTT
- 5672: AMQP
- 8161: Web Console / Jolokia
- 61613: STOMP
- 61616: All protocols as above 

#### ping

The ping service, used internally by the brokers for clustering themselves, internally exposes the 8888 port.

#### example-activemqartemis-console-jolokia-service

The broker hosts its own web console at port 8161 and this service provides external access to it via a 
[NodePort](https://kubernetes.io/docs/concepts/services-networking/service/#nodeport)
that provides direct access to the web console. Note that the actual port number chosen is, by default, in the range
of 30000-32767 so you will want to click in here to get the actual port number needed to access the console.

Once you have the port number using http://any.ocp.node.ip:consoleNodePortNumber should allow you to access the running
brokers web console.

#### example-activemqartemis-mux-protocol-service

The broker has a multiplexed protocol port at 61616 supporting all protocols; AMQP, CORE, HornetQ, MQTT, OpenWire and STOMP.
This broker port is exposed via a [NodePort](https://kubernetes.io/docs/concepts/services-networking/service/#nodeport)
as is the web console. 

To produce or consume messages to the broker from outside the [OCP SDN](https://docs.openshift.com/container-platform/3.11/admin_guide/sdn_troubleshooting.html)
you should be able to produce or consume messages to protocol://://any.ocp.node.ip:muxProtocolNodePortNumber.

### Scaling

The example-activemqartemis-ss-0 pod should now also be visible via the OpenShift console via the Applications -> Pods
sub menu, or by drilling down through the statefulset. The number of running pods can be adjusted via the 'oc scale' 
command as per:

```$xslt
oc scale --replicas 0 statefulset example-activemqartemis-ss
```

which would scale down the number of pods to zero. You can also scale the number to greater than one whereby the broker
pods will form a broker cluster with default settings: 

```$xslt
oc scale --replicas 2 statefulset example-activemqartemis-ss
```

One important caveat here is that upon scaledown the persistent
volume claim remains associated with the stable name of the broker pod that no longer exists. While that particular broker
pod is down the data in its journal files remains safe, but will only become accessible again once that particular broker
pod ordinal is up. 

#### Clustering

If broker pods are scaled to more than two then the broker pods form a broker
[cluster](https://activemq.apache.org/artemis/docs/2.6.0/clusters.html), meaning connect to each other and
redistribute messages 'ON_DEMAND'. With two or more broker pods running you can check the broker log of each
broker and you should see something similar to the following message regarding cluster formation:

broker 0

>
> 2019-03-22 16:59:46,406 INFO [org.apache.activemq.artemis.core.server] AMQ221027: Bridge ClusterConnectionBridge@6f13fb88 [name=$.artemis.internal.sf.my-cluster.e0709987-4cc3-11e9-b53b-0a580a810099, queue=QueueImpl[name=$.artemis.internal.sf.my-cluster.e0709987-4cc3-11e9-b53b-0a580a810099, postOffice=PostOfficeImpl [server=ActiveMQServerImpl::serverUUID=958a8ea9-4cc3-11e9-b153-0a580a820085], temp=false]@6478aa targetConnector=ServerLocatorImpl (identity=(Cluster-connection-bridge::ClusterConnectionBridge@6f13fb88 [name=$.artemis.internal.sf.my-cluster.e0709987-4cc3-11e9-b53b-0a580a810099, queue=QueueImpl[name=$.artemis.internal.sf.my-cluster.e0709987-4cc3-11e9-b53b-0a580a810099, postOffice=PostOfficeImpl [server=ActiveMQServerImpl::serverUUID=958a8ea9-4cc3-11e9-b153-0a580a820085], temp=false]@6478aa targetConnector=ServerLocatorImpl [initialConnectors=[TransportConfiguration(name=artemis, factory=org-apache-activemq-artemis-core-remoting-impl-netty-NettyConnectorFactory) ?port=61616&host=example-activemqartemis-ss-1-hs-abo-2-svc-cluster-local], discoveryGroupConfiguration=null]]::ClusterConnectionImpl@1533123860[nodeUUID=958a8ea9-4cc3-11e9-b153-0a580a820085, connector=TransportConfiguration(name=artemis, factory=org-apache-activemq-artemis-core-remoting-impl-netty-NettyConnectorFactory) ?port=61616&host=example-activemqartemis-ss-0-hs-abo-2-svc-cluster-local, address=, server=ActiveMQServerImpl::serverUUID=958a8ea9-4cc3-11e9-b153-0a580a820085])) [initialConnectors=[TransportConfiguration(name=artemis, factory=org-apache-activemq-artemis-core-remoting-impl-netty-NettyConnectorFactory) ?port=61616&host=example-activemqartemis-ss-1-hs-abo-2-svc-cluster-local], discoveryGroupConfiguration=null]] is connected

broker 1
>
> 2019-03-22 16:59:46,668 INFO [org.apache.activemq.artemis.core.server] AMQ221027: Bridge ClusterConnectionBridge@dcfadf [name=$.artemis.internal.sf.my-cluster.958a8ea9-4cc3-11e9-b153-0a580a820085, queue=QueueImpl[name=$.artemis.internal.sf.my-cluster.958a8ea9-4cc3-11e9-b153-0a580a820085, postOffice=PostOfficeImpl [server=ActiveMQServerImpl::serverUUID=e0709987-4cc3-11e9-b53b-0a580a810099], temp=false]@268b2b2c targetConnector=ServerLocatorImpl (identity=(Cluster-connection-bridge::ClusterConnectionBridge@dcfadf [name=$.artemis.internal.sf.my-cluster.958a8ea9-4cc3-11e9-b153-0a580a820085, queue=QueueImpl[name=$.artemis.internal.sf.my-cluster.958a8ea9-4cc3-11e9-b153-0a580a820085, postOffice=PostOfficeImpl [server=ActiveMQServerImpl::serverUUID=e0709987-4cc3-11e9-b53b-0a580a810099], temp=false]@268b2b2c targetConnector=ServerLocatorImpl [initialConnectors=[TransportConfiguration(name=artemis, factory=org-apache-activemq-artemis-core-remoting-impl-netty-NettyConnectorFactory) ?port=61616&host=example-activemqartemis-ss-0-hs-abo-2-svc-cluster-local], discoveryGroupConfiguration=null]]::ClusterConnectionImpl@104716441[nodeUUID=e0709987-4cc3-11e9-b53b-0a580a810099, connector=TransportConfiguration(name=artemis, factory=org-apache-activemq-artemis-core-remoting-impl-netty-NettyConnectorFactory) ?port=61616&host=example-activemqartemis-ss-1-hs-abo-2-svc-cluster-local, address=, server=ActiveMQServerImpl::serverUUID=e0709987-4cc3-11e9-b53b-0a580a810099])) [initialConnectors=[TransportConfiguration(name=artemis, factory=org-apache-activemq-artemis-core-remoting-impl-netty-NettyConnectorFactory) ?port=61616&host=example-activemqartemis-ss-0-hs-abo-2-svc-cluster-local], discoveryGroupConfiguration=null]] is connected

### Undeploying the broker

To undeploy the broker we simply execute

```$xslt
kubectl delete -f deploy/crds/broker_v1alpha1_activemqartemis_cr.yaml
```

at which point you should no longer see the statefulset 'example-activemqartemis-ss' under Overview in the web console.
On the command line you should see:

```$xslt
activemqartemis.broker.amq.io "example-activemqartemis" deleted
```

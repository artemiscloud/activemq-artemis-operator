# ArtemisCloud.io examples

This directory contains example YAML files to create and configure an Artemis broker on Kubernetes.

If you are new to Artemis on Kubernetes, start with [a basic deployment](artemis/artemis-basic-deployment.yaml):


1. Deploy the operator as described in the [Operator help](https://artemiscloud.io/docs/help/operator/#deploy-the-operator).

2. Create a basic broker deployment:

```bash
kubectl create -f examples/artemis/artemis_single.yaml -n <namespace>
```

See https://artemiscloud.io/ for tutorials and more information about Artemis.


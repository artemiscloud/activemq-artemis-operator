# ActiveMQ Artemis Operator

This project is a [Kubernetes](https://kubernetes.io/) [operator](https://coreos.com/blog/introducing-operators.html)
to manage the [Apache ActiveMQ Artemis](https://activemq.apache.org/artemis/) message broker.

## Status ##

The current api version of all main CRDs managed by the operator is **v1beta1**.

## Quickstart

The [quickstart.md](docs/getting-started/quick-start.md) provides simple steps to quickly get the operator up and running
as well as deploy/managing broker deployments.

## Building

The [building.md](docs/help/building.md) describes how to build operator and how to test your changes

## OLM integration

The [bundle.md](docs/help/bundle.md) contains instructions for how to build operator bundle images and integrate it into [Operator Liftcycle Manager](https://olm.operatorframework.io/) framework.

## Debugging operator inside a container

Install delve in the `builder` container, i.e. `RUN go install github.com/go-delve/delve/cmd/dlv@latest`
Disable build optimization, i.e. `go build -gcflags="all=-N -l"`
Copy delve to the `base-env` container, i.e. `COPY --from=builder /go/bin/dlv /bin`
Execute operator using delve, i.e. `/bin/dlv exec --listen=0.0.0.0:40000 --headless=true --api-version=2 --accept-multiclient ${OPERATOR} $@`

## Using cert-manager

In case that you want to leverage the cert-manager operator for creating the Webhook certificates, you
must have the operator installed on your cluster. Refer to the [cert-manager installation instructions](https://cert-manager.io/docs/installation/).

Then use the `USE_CERTMANAGER=true` flag during the local deployment.

```bash
make deploy USE_CERTMANAGER=true
```

Then check that the volume is properly mounted:

```bash
$ kubectl get deployment -n activemq-artemis-operator activemq-artemis-controller-manager -oyaml | yq .spec.template.spec.volumes
- name: cert
  secret:
    defaultMode: 420
    secretName: webhook-server-cert
```

By default a self-signed certificate is created and used instead:

```bash
$ kubectl get deployment -n activemq-artemis-operator activemq-artemis-controller-manager -oyaml | yq .spec.template.spec.volumes
- name: cert
  secret:
    defaultMode: 420
    secretName: activemq-artemis-webhook-server-cert
```
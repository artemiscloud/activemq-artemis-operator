#### :warning:This project is now part of the ArkMQ initiative. This repository has been archived as all activities are now happening in the [corresponding ArkMQ repository](https://github.com/arkmq-org/activemq-artemis-operator). See [here](https://artemiscloud.io/community/transition/) for the ArkMQ transition details.

---

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

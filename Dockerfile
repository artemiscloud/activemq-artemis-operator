# Build the manager binary
FROM golang:1.19 as builder
ARG TARGETOS
ARG TARGETARCH

ENV GOOS=linux
ENV CGO_ENABLED=0
ENV BROKER_NAME=activemq-artemis
ENV GO_MODULE=github.com/artemiscloud/activemq-artemis-operator


RUN mkdir -p /tmp/activemq-artemis-operator

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY .git/ .git/
COPY main.go main.go
COPY api/ api/
COPY controllers/ controllers/
COPY pkg/ pkg/
COPY version/ version/
COPY entrypoint/ entrypoint/

# Build
# the GOARCH has not a default value to allow the binary be built according to the host where the command
# was called. For example, if we call make docker-build in a local env which has the Apple Silicon M1 SO
# the docker BUILDPLATFORM arg will be linux/arm64 when for Apple x86 it will be linux/amd64. Therefore,
# by leaving it empty we can ensure that the container and binary shipped on it will have the same platform.
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} go build -a \
    -ldflags="-X '${GO_MODULE}/version.CommitHash=`git rev-parse --short HEAD`' \
    -X '${GO_MODULE}/version.BuildTimestamp=`date '+%Y-%m-%dT%H:%M:%S'`'" \
    -o /tmp/activemq-artemis-operator/${BROKER_NAME}-operator main.go

FROM registry.access.redhat.com/ubi8:8.6-855 AS base-env

ENV BROKER_NAME=activemq-artemis
ENV OPERATOR=/home/${BROKER_NAME}-operator/bin/${BROKER_NAME}-operator
ENV USER_UID=1000
ENV USER_NAME=${BROKER_NAME}-operator
ENV CGO_ENABLED=0
ENV GOPATH=/tmp/go
ENV JBOSS_IMAGE_NAME="amq7/amq-broker-rhel8-operator"
ENV JBOSS_IMAGE_VERSION="1.0"

WORKDIR /

COPY --from=builder /tmp/activemq-artemis-operator /home/${BROKER_NAME}-operator/bin
COPY --from=builder /workspace/entrypoint/entrypoint /home/${BROKER_NAME}-operator/bin

RUN useradd ${BROKER_NAME}-operator
RUN chown -R `id -u`:0 /home/${BROKER_NAME}-operator/bin && chmod -R 755 /home/${BROKER_NAME}-operator/bin

# Upgrade packages
RUN dnf update -y --setopt=install_weak_deps=0 && rm -rf /var/cache/yum

USER ${USER_UID}
ENTRYPOINT ["/home/${BROKER_NAME}-operator/bin/entrypoint"]

LABEL com.redhat.component="amq-broker-rhel8-operator-container"
LABEL com.redhat.delivery.appregistry="false"
LABEL description="ActiveMQ Artemis Broker Operator"
LABEL io.k8s.description="An associated operator that handles broker installation, updates and scaling."
LABEL io.k8s.display-name="ActiveMQ Artemis Broker Operator"
LABEL io.openshift.expose-services=""
LABEL io.openshift.s2i.scripts-url="image:///usr/local/s2i"
LABEL io.openshift.tags="messaging,amq,integration,operator,golang"
LABEL maintainer="Roddie Kieley <rkieley@redhat.com>"
LABEL name="amq7/amq-broker-rhel8-operator"
LABEL summary="ActiveMQ Artemis Broker Operator"
LABEL version="1.0"

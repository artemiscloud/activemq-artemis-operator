FROM registry.access.redhat.com/ubi7/go-toolset:1.13.4-23.1592916276 AS build-env

ENV OPERATOR=/tmp/activemq-artemis-operator/activemq-artemis-operator
ENV GOPATH=/tmp/go
ENV GOOS=linux
ENV CGO_ENABLED=0

RUN mkdir -p ${GOPATH}/src/github.com/rh-messaging/activemq-artemis-operator/bin
RUN mkdir -p /tmp/activemq-artemis-operator

ADD . ${GOPATH}/src/github.com/rh-messaging/activemq-artemis-operator
WORKDIR ${GOPATH}/src/github.com/rh-messaging/activemq-artemis-operator

RUN /opt/rh/go-toolset-1.13/root/usr/bin/go build -v -o ${OPERATOR} ${GOPATH}/src/github.com/rh-messaging/activemq-artemis-operator/cmd/manager

FROM alpine:3.8

ENV OPERATOR=/home/activemq-artemis-operator/bin/activemq-artemis-operator
ENV USER_UID=1001
ENV USER_NAME=activemq-artemis-operator
ENV CGO_ENABLED=0 
ENV GOPATH=/go
ENV HOME=/home/${USER_NAME}

COPY --from=build-env /tmp/activemq-artemis-operator /home/activemq-artemis-operator/bin
COPY --from=build-env /tmp/go/src/github.com/rh-messaging/activemq-artemis-operator/build/bin/entrypoint /home/activemq-artemis-operator/bin

RUN mkdir -p ${HOME} && \
    chown -R ${USER_UID}:0 ${HOME} && \
    chmod -R g=u ${HOME} /etc/passwd
USER ${USER_UID}
WORKDIR ${HOME}

ENTRYPOINT ["/home/activemq-artemis-operator/bin/entrypoint"]

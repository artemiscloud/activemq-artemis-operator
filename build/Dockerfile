FROM golang:1.11.5-alpine3.8 AS build-env

ENV OPERATOR=/amq-broker-operator/amq-broker-operator \
    CGO_ENABLED=0 \
    GOPATH=/go

# Allow Go to retrieve the dependencies for the build step
RUN apk add --no-cache git

RUN mkdir -p ${GOPATH}/src/github.com/rh-messaging/amq-broker-operator/bin
RUN mkdir -p /amq-broker-operator
ADD . ${GOPATH}/src/github.com/rh-messaging/amq-broker-operator
WORKDIR ${GOPATH}/src/github.com/rh-messaging/amq-broker-operator
RUN go build -v -o ${OPERATOR} ${GOPATH}/src/github.com/rh-messaging/amq-broker-operator/cmd/manager

FROM alpine:3.8

ENV OPERATOR=/amq-broker-operator/amq-broker-operator \
    USER_UID=1001 \
    USER_NAME=amq-broker-operator \
    CGO_ENABLED=0 \
    GOPATH=/go
ENV HOME=/home/${USER_NAME}

COPY --from=build-env ${OPERATOR} /amq-broker-operator/
COPY --from=build-env ${GOPATH}/src/github.com/rh-messaging/amq-broker-operator/build/bin /usr/local/bin

RUN mkdir -p ${HOME} && \
    chown -R ${USER_UID}:0 ${HOME} && \
    chmod -R g=u ${HOME} /etc/passwd
USER ${USER_UID}
WORKDIR ${HOME}

ENTRYPOINT ["/usr/local/bin/entrypoint"]
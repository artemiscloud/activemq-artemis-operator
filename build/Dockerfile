FROM golang:1.11.5-alpine3.8 AS build-env

ENV OPERATOR=/amq-broker-operator/amq-broker-operator \
    USER_UID=1001 \
    USER_NAME=amq-broker-operator \
    CGO_ENABLED=0 \
    GOPATH=/home/rkieley/go

# Allow Go to retrieve the dependencies for the build step
RUN apk add --no-cache git

ENV GOPATH /home/rkieley/go
RUN mkdir -p /home/rkieley/go/src/github.com/rh-messaging/amq-broker-operator/bin
RUN mkdir -p /amq-broker-operator
ADD .. /home/rkieley/go/src/github.com/rh-messaging/amq-broker-operator
WORKDIR /home/rkieley/go/src/github.com/rh-messaging/amq-broker-operator
RUN go build -v -o /amq-broker-operator/amq-broker-operator /home/rkieley/go/src/github.com/rh-messaging/amq-broker-operator/cmd/manager

# Get Delve from a GOPATH not from a Go Modules project
# as we aren't using Go Modules yet
WORKDIR /home/rkieley/go/src
RUN go get github.com/go-delve/delve/cmd/dlv


FROM alpine:3.8

ENV OPERATOR=/amq-broker-operator/amq-broker-operator \
    USER_UID=1001 \
    USER_NAME=amq-broker-operator \
    CGO_ENABLED=0 \
    GOPATH=/home/rkieley/go

WORKDIR /
COPY --from=build-env ${OPERATOR} /amq-broker-operator/
COPY --from=build-env /home/rkieley/go/bin/dlv /

USER ${USER_UID}
COPY bin /usr/local/bin
RUN  /usr/local/bin/user_setup

ENTRYPOINT ["/usr/local/bin/entrypoint"]
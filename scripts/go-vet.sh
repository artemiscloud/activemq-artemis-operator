#!/bin/sh

source ./scripts/go-mod-env.sh

if [[ -z ${CI} ]]; then
    ./scripts/go-mod.sh
    ./scripts/go-sdk-gen.sh
fi

#vet reports source code issues that may not
#be detected by compiler
go vet ./...

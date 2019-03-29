#!/bin/sh

./scripts/catalog-source.sh
kubectl apply -f deploy/catalog_resources/redhat/catalog-source.yaml
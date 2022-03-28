#!/bin/bash
unset RES_KIND
unset RES_NAME
rm -fr deploy
mkdir deploy
while IFS= read -r LINE; do
    if [ -z "${RES_KIND}" ]; then
        if [[ $LINE == "kind:"* ]]; then
            RES_KIND=$(echo ${LINE} | grep -oP "(?<=kind: ).*")
            if [ "$RES_KIND" != "CustomResourceDefinition" ]; then
                mkdir -p deploy/crds
                RES_NAME="crds/$(echo ${RES_KIND} | sed 's/\([A-Z]\)/_\L\1/g;s/^_//').yaml"
            fi
        fi
    fi
    if [ -z "${RES_NAME}" ]; then
        if [[ $LINE == "  name:"* ]]; then
            RES_NAME="broker_$(echo ${LINE} | grep -oP '(?<=name: )[^.]*')_crd.yaml"
        fi
    fi
    if [ "$LINE" = "---" ]; then
        mv deploy/tmp.yaml deploy/$RES_NAME
        unset RES_KIND
        unset RES_NAME
    else
        echo "$LINE" >> deploy/tmp.yaml
    fi
done
if [ -n "${RES_NAME}" ]; then
    mv deploy/tmp.yaml deploy/$RES_NAME
fi
cp -r config/crs deploy
cp -r examples deploy

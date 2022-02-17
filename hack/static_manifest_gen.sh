#!/bin/bash

mkdir -p $1

destdir=$1
file=()
resource_kind=""
resource_name=""

# write a single yaml.
# it takes one parameter:
# 1 array that contains total lines of the yaml
function writeFile() {
  array_name=$1[@]
  lines=("${!array_name}")
  file_path="$destdir/${resource_kind}_${resource_name}.yaml"
  rm -rf ${file_path}
  for value in "${file[@]}"
    do
      printf "%s\n" "${value}" >> ${file_path}
    done
}

function beginFile() {
  if [[ ${#file[@]} > 0 ]]
  then
    # it gathered a single file now
    writeFile file
  fi
  file=()
}

function processFile() {
  file+=("$1")
}

IFS=''
while read -r line
do
  if [[ $line =~ ^---$ ]]
  then
    beginFile
    resource_kind=""
    resource_name=""
  else
    processFile "$line"
    if [[ $line =~ ^(kind: ) ]]
    then
      resource_kind=$(echo $line | cut -b 7-)
      if [[ ${resource_kind} == "CustomResourceDefinition" ]]; then
        resource_kind="CRD"
      fi
    fi

    if [[ $line =~ ^(metadata:) ]]
    then
      meta_start="true"
    fi
    if [[ ${meta_start} == "true" && $line =~ ^(  name: ) ]]
    then
      resource_name=$(echo $line | cut -b 9-)
      meta_start="false"
    fi
  fi
done
# check the last one
beginFile


# Building the operator

- [imagebuilder](https://github.com/openshift/imagebuilder)

```$xslt
imagebuilder -t amq-broker-operator:latest -f amq-broker-operator/build/Dockerfile amq-broker-operator/build
```

or

```$xslt
imagebuilder -t amq-broker-operator:latest-debug -f amq-broker-operator/build/Dockerfile_debug amq-broker-operator/build
```
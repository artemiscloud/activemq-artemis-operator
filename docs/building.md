# Building the operator

- [imagebuilder](https://github.com/openshift/imagebuilder)

```$xslt
imagebuilder -t activemq-artemis-operator:latest -f activemq-artemis-operator/build/Dockerfile activemq-artemis-operator/build
```

or

```$xslt
imagebuilder -t activemq-artemis-operator:latest-debug -f activemq-artemis-operator/build/Dockerfile_debug activemq-artemis-operator/build
```
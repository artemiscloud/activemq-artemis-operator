# Building the operator

- [imagebuilder](https://github.com/openshift/imagebuilder)

```$xslt
imagebuilder -t activemq-artemis-operator:latest -f build/Dockerfile .
```

or

```$xslt
imagebuilder -t activemq-artemis-operator:latest-debug -f build/Dockerfile_debug .
```
to build docker file:

make OPERATOR_IMAGE_REPO=quay.io/hgao/operator OPERATOR_VERSION=new0.20.1 docker-build
make OPERATOR_IMAGE_REPO=quay.io/hgao/operator OPERATOR_VERSION=new0.20.1 docker-push

to test your image ovverride the image:

make OPERATOR_IMAGE_REPO=quay.io/hgao/operator OPERATOR_VERSION=new0.20.1 deploy

make OPERATOR_IMAGE_REPO=quay.io/hgao/operator OPERATOR_VERSION=new0.20.1 undeploy



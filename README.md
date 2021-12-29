to build and push docker image:

make OPERATOR_IMAGE_REPO=quay.io/hgao/operator OPERATOR_VERSION=new0.20.1 docker-build
make OPERATOR_IMAGE_REPO=quay.io/hgao/operator OPERATOR_VERSION=new0.20.1 docker-push

to test your image ovverride the image:

make OPERATOR_IMAGE_REPO=quay.io/hgao/operator OPERATOR_VERSION=new0.20.1 deploy

make OPERATOR_IMAGE_REPO=quay.io/hgao/operator OPERATOR_VERSION=new0.20.1 undeploy

to create bundle manifests/metadata:

make IMAGE_TAG_BASE=quay.io/hgao/operator bundle

to build bundle image:

make IMAGE_TAG_BASE=quay.io/hgao/operator bundle-build

to push bundle image to remote repo:

make IMAGE_TAG_BASE=quay.io/hgao/operator bundle-push

to build catalog image (index image):

make IMAGE_TAG_BASE=quay.io/hgao/operator catalog-build

to push catalog image to repo:

make IMAGE_TAG_BASE=quay.io/hgao/operator catalog-push


#!/usr/bin/env bash

# set -o errexit
set -o nounset
set -o pipefail

#KURYR_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
#echo "KURYR_ROOT:$KURYR_ROOT"

KURYR_ROOT=${GOPATH}/src/projectkuryr/kuryr
CONTAINER_WORKDIR=/go/src/projectkuryr/kuryr
IMAGE_NAME=registry-jinan-lab.insprcloud.cn/library/cke/kuryr/codegen:kubernetes-1.18.4

function docker_run() {
  docker pull ${IMAGE_NAME}
  docker run --rm \
		-w ${CONTAINER_WORKDIR} \
		-v ${KURYR_ROOT}:${CONTAINER_WORKDIR} \
		"${IMAGE_NAME}" "$@"
}

docker_run hack/update-codegen-dockerized.sh




# mount 本地目录到容器中，编译完成文件在本地
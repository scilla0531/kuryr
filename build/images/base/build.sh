#!/usr/bin/env bash

# This is a very simple script that builds the base image for Kuryr and pushes it to
# the Kuryr Dockerhub (https://hub.docker.com/u/kuryr). The image is tagged with the OVS version.

set -eo pipefail

function echoerr {
    >&2 echo "$@"
}

_usage="Usage: $0 [--pull] [--push] [--platform <PLATFORM>]
Build the kuryr/base-ubuntu:<OVS_VERSION> image.
        --pull                  Always attempt to pull a newer version of the base images
        --push                  Push the built image to the registry
        --platform <PLATFORM>   Target platform for the image if server is multi-platform capable"

function print_usage {
    echoerr "$_usage"
}

PULL=false
PUSH=false
PLATFORM=""

while [[ $# -gt 0 ]]
do
key="$1"

case $key in
    --push)
    PUSH=true
    shift
    ;;
    --pull)
    PULL=true
    shift
    ;;
    --platform)
    PLATFORM="$2"
    shift 2
    ;;
    -h|--help)
    print_usage
    exit 0
    ;;
    *)    # unknown option
    echoerr "Unknown option $1"
    exit 1
    ;;
esac
done

if [ "$PLATFORM" != "" ] && $PUSH; then
    echoerr "Cannot use --platform with --push"
    exit 1
fi

PLATFORM_ARG=""
if [ "$PLATFORM" != "" ]; then
    PLATFORM_ARG="--platform $PLATFORM"
fi

THIS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

pushd $THIS_DIR > /dev/null

OVS_VERSION=$(head -n 1 ../deps/ovs-version)
CNI_BINARIES_VERSION=$(head -n 1 ../deps/cni-binaries-version)

if $PULL; then
    docker pull $PLATFORM_ARG ubuntu:20.04
    docker pull $PLATFORM_ARG kuryr/openvswitch:$OVS_VERSION
    docker pull $PLATFORM_ARG kuryr/cni-binaries:$CNI_BINARIES_VERSION || true
    docker pull $PLATFORM_ARG kuryr/base-ubuntu:$OVS_VERSION || true
fi

docker build $PLATFORM_ARG --target cni-binaries \
       --cache-from kuryr/cni-binaries:$CNI_BINARIES_VERSION \
       -t kuryr/cni-binaries:$CNI_BINARIES_VERSION \
       --build-arg CNI_BINARIES_VERSION=$CNI_BINARIES_VERSION \
       --build-arg OVS_VERSION=$OVS_VERSION .

docker build $PLATFORM_ARG \
       --cache-from kuryr/cni-binaries:$CNI_BINARIES_VERSION \
       --cache-from kuryr/base-ubuntu:$OVS_VERSION \
       -t kuryr/base-ubuntu:$OVS_VERSION \
       --build-arg CNI_BINARIES_VERSION=$CNI_BINARIES_VERSION \
       --build-arg OVS_VERSION=$OVS_VERSION .

if $PUSH; then
    docker push kuryr/cni-binaries:$CNI_BINARIES_VERSION
    docker push kuryr/base-ubuntu:$OVS_VERSION
fi

popd > /dev/null

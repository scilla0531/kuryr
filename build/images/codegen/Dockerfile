FROM ubuntu:20.04 as protoc

RUN apt-get update && \
    apt-get install -y --no-install-recommends wget ca-certificates unzip

RUN PROTOBUF_VERSION=3.0.2; ZIPNAME="protoc-${PROTOBUF_VERSION}-linux-x86_64.zip"; \
  mkdir /tmp/protoc && cd /tmp/protoc && \
  wget "https://github.com/protocolbuffers/protobuf/releases/download/v${PROTOBUF_VERSION}/${ZIPNAME}" && \
  unzip "${ZIPNAME}" && \
  chmod -R +rX /tmp/protoc


FROM golang:1.15
WORKDIR /opt/kuryr

LABEL maintainer="Kuryr <scilla>"
LABEL description="A Docker image based on golang 1.15 which includes codegen tools."

ENV GO111MODULE=on
ENV GOPROXY="https://goproxy.cn,direct"

ARG K8S_VERSION=1.18.4
# The k8s.io/kube-openapi repo does not have tag, using a workable commit hash.
ARG KUBEOPENAPI_VERSION=v0.0.0-20200410145947-61e04a5be9a6

RUN go get k8s.io/code-generator/cmd/client-gen@kubernetes-$K8S_VERSION && \
    go get k8s.io/code-generator/cmd/deepcopy-gen@kubernetes-$K8S_VERSION && \
    go get k8s.io/code-generator/cmd/conversion-gen@kubernetes-$K8S_VERSION && \
    go get k8s.io/code-generator/cmd/lister-gen@kubernetes-$K8S_VERSION && \
    go get k8s.io/code-generator/cmd/informer-gen@kubernetes-$K8S_VERSION && \
    go get k8s.io/kube-openapi/cmd/openapi-gen@$KUBEOPENAPI_VERSION && \
    go get k8s.io/code-generator/cmd/go-to-protobuf@kubernetes-$K8S_VERSION && \
    go get k8s.io/code-generator/cmd/go-to-protobuf/protoc-gen-gogo@kubernetes-$K8S_VERSION && \
    go get github.com/golang/mock/mockgen@v1.4.4 && \
    go get github.com/golang/protobuf/protoc-gen-go@v1.3.2 && \
    go get golang.org/x/tools/cmd/goimports


COPY --from=protoc /tmp/protoc/bin /usr/local/bin
COPY --from=protoc /tmp/protoc/include /usr/local/include





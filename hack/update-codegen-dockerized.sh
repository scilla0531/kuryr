#!/usr/bin/env bash

#set -o errexit
set -o nounset
set -o pipefail
set -o xtrace

GOPATH=`go env GOPATH`
ANTREA_PKG="projectkuryr/kuryr"

# Generate protobuf code for CNI gRPC service with protoc.
protoc --go_out=plugins=grpc:. pkg/apis/cni/v1alpha1/cni.proto


if false; then

# Generate clientset and apis code with K8s codegen tools.
$GOPATH/bin/client-gen \
  --clientset-name versioned \
  --input-base "${ANTREA_PKG}/pkg/apis/" \
  --input "clusterinformation/v1beta1" \
  --input "controlplane/v1beta1" \
  --input "controlplane/v1beta2" \
  --input "system/v1beta1" \
  --input "security/v1alpha1" \
  --input "core/v1alpha2" \
  --input "ops/v1alpha1" \
  --input "stats/v1alpha1" \
  --output-package "${ANTREA_PKG}/pkg/client/clientset" \
  --plural-exceptions "NetworkPolicyStats:NetworkPolicyStats" \
  --plural-exceptions "AntreaNetworkPolicyStats:AntreaNetworkPolicyStats" \
  --plural-exceptions "AntreaClusterNetworkPolicyStats:AntreaClusterNetworkPolicyStats" \
  --go-header-file hack/boilerplate/license_header.go.txt

# Generate listers with K8s codegen tools.
$GOPATH/bin/lister-gen \
  --input-dirs "${ANTREA_PKG}/pkg/apis/security/v1alpha1,${ANTREA_PKG}/pkg/apis/core/v1alpha2" \
  --input-dirs "${ANTREA_PKG}/pkg/apis/ops/v1alpha1" \
  --input-dirs "${ANTREA_PKG}/pkg/apis/clusterinformation/v1beta1" \
  --output-package "${ANTREA_PKG}/pkg/client/listers" \
  --go-header-file hack/boilerplate/license_header.go.txt

# Generate informers with K8s codegen tools.
$GOPATH/bin/informer-gen \
  --input-dirs "${ANTREA_PKG}/pkg/apis/security/v1alpha1,${ANTREA_PKG}/pkg/apis/core/v1alpha2" \
  --input-dirs "${ANTREA_PKG}/pkg/apis/ops/v1alpha1" \
  --input-dirs "${ANTREA_PKG}/pkg/apis/clusterinformation/v1beta1" \
  --versioned-clientset-package "${ANTREA_PKG}/pkg/client/clientset/versioned" \
  --listers-package "${ANTREA_PKG}/pkg/client/listers" \
  --output-package "${ANTREA_PKG}/pkg/client/informers" \
  --go-header-file hack/boilerplate/license_header.go.txt

$GOPATH/bin/deepcopy-gen \
  --input-dirs "${ANTREA_PKG}/pkg/apis/clusterinformation/v1beta1" \
  --input-dirs "${ANTREA_PKG}/pkg/apis/controlplane" \
  --input-dirs "${ANTREA_PKG}/pkg/apis/controlplane/v1beta1" \
  --input-dirs "${ANTREA_PKG}/pkg/apis/controlplane/v1beta2" \
  --input-dirs "${ANTREA_PKG}/pkg/apis/system/v1beta1" \
  --input-dirs "${ANTREA_PKG}/pkg/apis/security/v1alpha1" \
  --input-dirs "${ANTREA_PKG}/pkg/apis/core/v1alpha2" \
  --input-dirs "${ANTREA_PKG}/pkg/apis/ops/v1alpha1" \
  --input-dirs "${ANTREA_PKG}/pkg/apis/stats" \
  --input-dirs "${ANTREA_PKG}/pkg/apis/stats/v1alpha1" \
  -O zz_generated.deepcopy \
  --go-header-file hack/boilerplate/license_header.go.txt

$GOPATH/bin/conversion-gen  \
  --input-dirs "${ANTREA_PKG}/pkg/apis/controlplane/v1beta2,${ANTREA_PKG}/pkg/apis/controlplane/v1beta1,${ANTREA_PKG}/pkg/apis/controlplane/" \
  --input-dirs "${ANTREA_PKG}/pkg/apis/stats/v1alpha1,${ANTREA_PKG}/pkg/apis/stats/" \
  -O zz_generated.conversion \
  --go-header-file hack/boilerplate/license_header.go.txt

$GOPATH/bin/openapi-gen  \
  --input-dirs "${ANTREA_PKG}/pkg/apis/controlplane/v1beta1" \
  --input-dirs "${ANTREA_PKG}/pkg/apis/controlplane/v1beta2" \
  --input-dirs "${ANTREA_PKG}/pkg/apis/clusterinformation/v1beta1" \
  --input-dirs "${ANTREA_PKG}/pkg/apis/system/v1beta1" \
  --input-dirs "${ANTREA_PKG}/pkg/apis/stats/v1alpha1" \
  --input-dirs "k8s.io/apimachinery/pkg/apis/meta/v1,k8s.io/apimachinery/pkg/runtime,k8s.io/apimachinery/pkg/util/intstr" \
  --input-dirs "k8s.io/api/core/v1" \
  --output-package "${ANTREA_PKG}/pkg/apiserver/openapi" \
  -O zz_generated.openapi \
  --go-header-file hack/boilerplate/license_header.go.txt

# Generate mocks for testing with mockgen.
MOCKGEN_TARGETS=(
  "pkg/agent/cniserver/ipam IPAMDriver testing"
  "pkg/agent/flowexporter/connections ConnectionStore,ConnTrackDumper,NetFilterConnTrack testing"
  "pkg/agent/interfacestore InterfaceStore testing"
  "pkg/agent/nodeportlocal/rules PodPortRules testing"
  "pkg/agent/openflow Client,OFEntryOperations testing"
  "pkg/agent/proxy Proxier testing"
  "pkg/agent/querier AgentQuerier testing"
  "pkg/agent/route Interface testing"
  "pkg/antctl AntctlClient ."
  "pkg/controller/networkpolicy EndpointQuerier testing"
  "pkg/controller/querier ControllerQuerier testing"
  "pkg/ipfix IPFIXExportingProcess,IPFIXSet,IPFIXRegistry,IPFIXCollectingProcess,IPFIXAggregationProcess testing"
  "pkg/ovs/openflow Bridge,Table,Flow,Action,CTAction,FlowBuilder testing"
  "pkg/ovs/ovsconfig OVSBridgeClient testing"
  "pkg/ovs/ovsctl OVSCtlClient testing"
  "pkg/querier AgentNetworkPolicyInfoQuerier testing"
  "third_party/proxy Provider testing"
)

# Command mockgen does not automatically replace variable YEAR with current year
# like others do, e.g. client-gen.
current_year=$(date +"%Y")
sed -i "s/YEAR/${current_year}/g" hack/boilerplate/license_header.raw.txt
for target in "${MOCKGEN_TARGETS[@]}"; do
  read -r package interfaces mock_package <<<"${target}"
  package_name=$(basename "${package}")
  if [[ "${mock_package}" == "." ]]; then # generate mocks in same package as src
      $GOPATH/bin/mockgen \
          -copyright_file hack/boilerplate/license_header.raw.txt \
          -destination "${package}/mock_${package_name}_test.go" \
          -package="${package_name}" \
          "${ANTREA_PKG}/${package}" "${interfaces}"
  else # generate mocks in subpackage
      $GOPATH/bin/mockgen \
          -copyright_file hack/boilerplate/license_header.raw.txt \
          -destination "${package}/${mock_package}/mock_${package_name}.go" \
          -package="${mock_package}" \
          "${ANTREA_PKG}/${package}" "${interfaces}"
  fi
done
git checkout HEAD -- hack/boilerplate/license_header.raw.txt

# Download vendored modules to the vendor directory so it's easier to
# specify the search path of required protobuf files.
go mod vendor
# In Go 1.14, vendoring changed (see release notes at
# https://golang.org/doc/go1.14), and the presence of a go.mod file specifying
# go 1.14 or higher causes the go command to default to -mod=vendor when a
# top-level vendor directory is present in the module. This causes the
# go-to-protobuf command below to complain about missing packages under vendor/,
# which were not downloaded by "go mod vendor". We can workaround this easily by
# renaming the vendor directory.
mv vendor /tmp/includes
$GOPATH/bin/go-to-protobuf \
  --proto-import /tmp/includes \
  --packages "${ANTREA_PKG}/pkg/apis/stats/v1alpha1,${ANTREA_PKG}/pkg/apis/controlplane/v1beta1,${ANTREA_PKG}/pkg/apis/controlplane/v1beta2" \
  --go-header-file hack/boilerplate/license_header.go.txt
rm -rf /tmp/includes

set +x

echo "=== Start resetting changes introduced by YEAR ==="
# The call to 'tac' ensures that we cannot have concurrent git processes, by
# waiting for the call to 'git diff  --numstat' to complete before iterating
# over the files and calling 'git diff ${file}'.
git diff  --numstat | awk '$1 == "1" && $2 == "1" {print $3}' | tac | while read file; do
  if [[ "$(git diff ${file})" == *"-// Copyright "*" Antrea Authors"* ]]; then
    git checkout HEAD -- "${file}"
    echo "=== ${file} is reset ==="
  fi
done



fi
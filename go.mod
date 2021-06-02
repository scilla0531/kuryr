module projectkuryr/kuryr

go 1.15

require (
	github.com/Microsoft/go-winio v0.4.16-0.20201130162521-d1ffc52c7331 // indirect
	github.com/TomCodeLV/OVSDB-golang-lib v0.0.0-20200116135253-9bbdfadcd881
	github.com/blang/semver v3.5.0+incompatible
	github.com/containernetworking/cni v0.8.0
	github.com/containernetworking/plugins v0.8.7
	github.com/contiv/libOpenflow v0.0.0-20210312221048-1d504242120d
	github.com/contiv/ofnet v0.0.0-00010101000000-000000000000
	github.com/gogo/protobuf v1.3.1
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/golang/mock v1.4.4
	github.com/golang/protobuf v1.3.2
	github.com/google/go-cmp v0.4.0 // indirect
	github.com/gophercloud/gophercloud v0.17.0
	github.com/kr/text v0.2.0 // indirect
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra v1.1.1
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.6.1 // indirect
	github.com/vishvananda/netlink v1.1.0
	github.com/vishvananda/netns v0.0.0-20191106174202-0a2b9b5464df
	golang.org/x/net v0.0.0-20201110031124-69a78807bb2b
	golang.org/x/sys v0.0.0-20201201145000-ef89a241ccb3 // indirect
	golang.org/x/tools v0.0.0-20191119224855-298f0cb1881e // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	google.golang.org/grpc v1.26.0
	gopkg.in/check.v1 v1.0.0-20200902074654-038fdea0a05b // indirect
	gopkg.in/yaml.v2 v2.3.0
	k8s.io/api v0.18.4
	k8s.io/apimachinery v0.18.4
	k8s.io/apiserver v0.18.4
	k8s.io/client-go v0.18.4
	k8s.io/code-generator v0.18.6-rc.0
	k8s.io/component-base v0.18.4
	k8s.io/klog v1.0.0
	k8s.io/kube-aggregator v0.18.4
	k8s.io/utils v0.0.0-20201110183641-67b214c5f920
)

// fake.NewSimpleClientset is quite slow when it's initialized with massive objects due to
// https://github.com/kubernetes/kubernetes/issues/89574. It takes more than tens of minutes to
// init a fake client with 200k objects, which makes it hard to run the NetworkPolicy scale test.
// There is an optimization https://github.com/kubernetes/kubernetes/pull/89575 but will only be
// available from 1.19.0 and later releases. Use this commit before Kuryr bumps up its K8s
// dependency version.
replace (
	github.com/contiv/ofnet => github.com/wenyingd/ofnet v0.0.0-20210318032909-171b6795a2da
	github.com/gogo/protobuf => github.com/gogo/protobuf v1.3.1
	k8s.io/client-go => github.com/tnqn/client-go v0.18.4-1

)

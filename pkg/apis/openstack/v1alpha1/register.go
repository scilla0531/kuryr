package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"projectkuryr/kuryr/pkg/apis/openstack"
)

var SchemeGroupVersion = schema.GroupVersion{
	Group:   openstack.GroupName,
	Version: "v1alpha1",
}

var (
	SchemeBuilder      runtime.SchemeBuilder
	localSchemeBuilder = &SchemeBuilder
	AddToScheme        = localSchemeBuilder.AddToScheme
)

func init() {
	localSchemeBuilder.Register(addKnownTypes)
}

func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}
/*
需要使客户端库知道新类型。允许客户端在与API服务器通信时自动处理新类型。
*/
func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(
		SchemeGroupVersion,
		&KuryrNetwork{},
		&KuryrNetworkList{},
		&KuryrPort{},
		&KuryrPortList{},
		&KuryrNetworkPolicy{},
		&KuryrNetworkPolicyList{},
	)

	metav1.AddToGroupVersion(
		scheme,
		SchemeGroupVersion,
	)
	return nil
}

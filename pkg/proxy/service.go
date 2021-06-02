package proxy

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sync"
)

// ServiceMap maps a service to its ServicePort.
type ServiceMap map[ServicePortName]ServicePort



// serviceChange contains all changes to services that happened since proxy rules were synced.  For a single object,
// changes are accumulated, i.e. previous is state from before applying the changes,
// current is state after applying all of the changes.
type serviceChange struct {
	previous ServiceMap
	current  ServiceMap
}

// ServiceChangeTracker carries state about uncommitted changes to an arbitrary number of
// Services, keyed by their namespace and name.
type ServiceChangeTracker struct {
	// lock protects items.
	lock sync.Mutex
	// items maps a service to its serviceChange.
	items map[types.NamespacedName]*serviceChange
	// makeServiceInfo allows proxier to inject customized information when processing service.
	//makeServiceInfo         makeServicePortFunc
	processServiceMapChange processServiceMapChangeFunc
	ipFamily                corev1.IPFamily

	recorder record.EventRecorder
}

// This handler is invoked by the apply function on every change. This function should not modify the
// ServiceMap's but just use the changes for any Proxier specific cleanup.
type processServiceMapChangeFunc func(previous, current ServiceMap)


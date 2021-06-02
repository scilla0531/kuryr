package k8s

// NamespacedName generates the conventional K8s resource name,
// which connects namespace and name with "/".
func NamespacedName(namespace, name string) string {
	if namespace == "" {
		// Cluster scoped resources will contain empty namespace.
		return name
	}
	return namespace + "/" + name
}

package app

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/record"
	aggregatorclientset "k8s.io/kube-aggregator/pkg/client/clientset_generated/clientset"
	"os"
	kuryrscheme "projectkuryr/kuryr/pkg/client/clientset/versioned/scheme"
	"strings"
	"time"

	"k8s.io/client-go/informers"
	"k8s.io/klog"

	"projectkuryr/kuryr/pkg/k8s"
	"projectkuryr/kuryr/pkg/version"

	//kuryrv1alpha1 "projectkuryr/kuryr/pkg/apis/openstack/v1alpha1"
	//kuryrclientset "projectkuryr/kuryr/pkg/client/clientset/versioned"
	//kuryrscheme "projectkuryr/kuryr/pkg/client/clientset/versioned/scheme"
	kuryrinformers "projectkuryr/kuryr/pkg/client/informers/externalversions"
	//kuryrlisters "projectkuryr/kuryr/pkg/client/listers/openstack/v1alpha1"
)

const (
	// informerDefaultResync is the default resync period if a handler doesn't specify one.
	// Use the same default value as kube-controller-manager:
	// https://github.com/kubernetes/kubernetes/blob/release-1.17/pkg/controller/apis/config/v1alpha1/defaults.go#L120
	informerDefaultResync = 12 * time.Hour
	//informerDefaultResync = time.Hour
	// serverMinWatchTimeout determines the timeout allocated to watches from Kuryr clients. Each watch will be allocated a random timeout between this value and twice this
	// value, to help randomly distribute reconnections over time.
	// This parameter corresponds to the MinRequestTimeout server config parameter in
	// https://godoc.org/k8s.io/apiserver/pkg/server#Config.
	// When the Kuryr client re-creates a watch, all relevant NetworkPolicy objects need to be
	// sent again by the controller. It may be a good idea to use a value which is larger than
	// the kube-apiserver default (1800s). The K8s documentation states that clients should be
	// able to handle watch timeouts gracefully but recommends using a large value in
	// production.
	serverMinWatchTimeout = 2 * time.Hour
)

var allowedPaths = []string{
	"/healthz",
	"/livez",
	"/readyz",
	"/mutate/acnp",
	"/mutate/anp",
	"/mutate/namespace",
	"/validate/tier",
	"/validate/acnp",
	"/validate/anp",
	"/validate/clustergroup",
}

const controllerName = "kuryr-controller"

func isNetworkResourceSpecifiedAndValid(annotations map[string]string) bool{
	subnet := annotations[AnnotationPodSubnet]
	router := annotations[AnnotationPodRouter]
	svc := annotations[AnnotationSvcSubnet] // 目前暂不强制要求
	sg := annotations[AnnotationPodSg] // podSgs: "" ，代表使用该租户的默认安全组
	return subnet != "" && router != "" && svc != "" && sg != ""
}

func IsKuryrCniType(annotations map[string]string) bool{
	cniType, exist := annotations[AnnotationCniType]
	return exist && cniType == AnnotationCniTypeKuryr
}

func IsHostNetworkPod(pod *corev1.Pod) bool{
	return pod.Spec.HostNetwork
}

func isFinalize(obj interface{}) bool{
	object := obj.(metav1.Object)
	return !object.GetDeletionTimestamp().IsZero()
}

func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func removeString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}

func createAPIServerConfig(kubeconfig string,
	client clientset.Interface,
	aggregatorClient aggregatorclientset.Interface,
	selfSignedCert bool,
	bindPort int,
	knsStore interface{},
	knsController interface{}){
	return
}

// GetHostname returns OS's hostname if 'hostnameOverride' is empty; otherwise, return 'hostnameOverride'.
func GetHostname(hostnameOverride string) (string, error) {
	hostName := hostnameOverride
	if len(hostName) == 0 {
		nodeName, err := os.Hostname()
		if err != nil {
			return "", fmt.Errorf("couldn't determine hostname: %v", err)
		}
		hostName = nodeName
	}

	// Trim whitespaces first to avoid getting an empty hostname
	// For linux, the hostname is read from file /proc/sys/kernel/hostname directly
	hostName = strings.TrimSpace(hostName)
	if len(hostName) == 0 {
		return "", fmt.Errorf("empty hostname is invalid")
	}
	return strings.ToLower(hostName), nil
}

// run starts Kuryr Controller with the given options and waits for termination signal.
func run(o *Options) error {
	klog.Infof("Starting Kuryr Controller (version %s)", version.GetFullVersion())

	LearnIndexer()

	// Aggregator Clientset is used to update the CABundle of the APIServices backed by kuryr-controller so that
	// the aggregator can verify its serving certificate.
	client, aggregatorClient, crdClient, err := k8s.CreateClientsCrd(o.config.ClientConnection, "")
	if err != nil {
		return fmt.Errorf("error creating k8s clients: %v", err)
	}

	osClient, err := geOsClient(o.config.Openstack)
	if err != nil {
		klog.Errorf("Invoke NewOSClient Error: %v\n", err)
		return err
	}
	informerFactory := informers.NewSharedInformerFactory(client, informerDefaultResync)
	crdInformerFactory := kuryrinformers.NewSharedInformerFactory(crdClient, informerDefaultResync)
	nsInformer := informerFactory.Core().V1().Namespaces()
	podInformer := informerFactory.Core().V1().Pods()
	//networkPolicyInformer := informerFactory.Networking().V1().NetworkPolicies()
	//serviceInformer := informerFactory.Core().V1().Services()
	knsInformer := crdInformerFactory.Openstack().V1alpha1().KuryrNetworks()
	kpInformer := crdInformerFactory.Openstack().V1alpha1().KuryrPorts()


	// Add kuryr-controller types to the default Kubernetes Scheme so Events can be logged for sample-controller types.
	utilruntime.Must(kuryrscheme.AddToScheme(scheme.Scheme)) //???????????????????????????????????????

	klog.Info("Creating event broadcaster")
	hostname, err := GetHostname("")
	if err != nil {
		return fmt.Errorf("error getting hostname: %v", err)
	}

	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: client.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerName, Host: hostname})

	// Create Kuryr object storage.
	knsStore := struct{}{}
	_ = aggregatorClient
	_ = knsStore





	nsController := NewNsController(o.config,
		client,
		crdClient,
		osClient,
		nsInformer,
		knsInformer,
		podInformer,
		kpInformer,
		recorder)

	stopCh := wait.NeverStop
	informerFactory.Start(stopCh)
	crdInformerFactory.Start(stopCh)

	go nsController.Run(1, stopCh)

	<-stopCh
	klog.Info("Stopping Kuryr controller")
	return nil
}

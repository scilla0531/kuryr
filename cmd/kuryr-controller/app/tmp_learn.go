package app

import (
	"context"
	"fmt"
	"github.com/spf13/pflag"
	"k8s.io/api/apps/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/informers"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	componentbaseconfig "k8s.io/component-base/config"
	"k8s.io/klog"
	"log"
	"net"
	"os"
	"projectkuryr/kuryr/pkg/ip"
	"strings"
)

var (
	_ pflag.Value = &IPVar{}
	_ pflag.Value = &IPPortVar{}
)

// IPVar is used for validating a command line option that represents an IP. It implements the pflag.Value interface
type IPVar struct {
	Val *string
}

// Set sets the flag value
func (v IPVar) Set(s string) error {
	if len(s) == 0 {
		v.Val = nil
		return nil
	}
	if net.ParseIP(s) == nil {
		return fmt.Errorf("%q is not a valid IP address", s)
	}
	if v.Val == nil {
		// it's okay to panic here since this is programmer error
		panic("the string pointer passed into IPVar should not be nil")
	}
	*v.Val = s
	return nil
}

// String returns the flag value
func (v IPVar) String() string {
	if v.Val == nil {
		return ""
	}
	return *v.Val
}

// Type gets the flag type
func (v IPVar) Type() string {
	return "ip"
}

// IPPortVar is used for validating a command line option that represents an IP and a port. It implements the pflag.Value interface
type IPPortVar struct {
	Val *string
}

// Set sets the flag value
func (v IPPortVar) Set(s string) error {
	if len(s) == 0 {
		v.Val = nil
		return nil
	}

	if v.Val == nil {
		// it's okay to panic here since this is programmer error
		panic("the string pointer passed into IPPortVar should not be nil")
	}

	// Both IP and IP:port are valid.
	// Attempt to parse into IP first.
	if net.ParseIP(s) != nil {
		*v.Val = s
		return nil
	}

	// Can not parse into IP, now assume IP:port.
	host, port, err := net.SplitHostPort(s)
	if err != nil {
		return fmt.Errorf("%q is not in a valid format (ip or ip:port): %v", s, err)
	}
	if net.ParseIP(host) == nil {
		return fmt.Errorf("%q is not a valid IP address", host)
	}
	if _, err := ip.ParsePort(port, true); err != nil {
		return fmt.Errorf("%q is not a valid number", port)
	}
	*v.Val = s
	return nil
}

// String returns the flag value
func (v IPPortVar) String() string {
	if v.Val == nil {
		return ""
	}
	return *v.Val
}

// Type gets the flag type
func (v IPPortVar) Type() string {
	return "ipport"
}

//func WatchSingle(namespace, name, resourceVersion string) (watch.Interface, error) {
//	return rest.RESTClient.Get().
//		NamespaceIfScoped(namespace, m.NamespaceScoped).
//		Resource(m.Resource).
//		VersionedParams(&metav1.ListOptions{
//			ResourceVersion: resourceVersion,
//			Watch:           true,
//			FieldSelector:   fields.OneTermEqualSelector("metadata.name", name).String(),
//		}, metav1.ParameterCodec).
//		Watch(context.TODO())
//}

func printline(title string) {
	if title == ""{
		//fmt.Printf("%s\n", strings.Repeat("=", 150))
		fmt.Printf("%s\n", strings.Repeat("*", 150))
	}else{
		//fmt.Printf("%s%-30v%s\n", strings.Repeat("=", 60) ,title, strings.Repeat("=", 60))
		fmt.Printf("%s%-30v%s\n", strings.Repeat("=", 60) ,title, strings.Repeat("=", 60))
	}
}

func checkerror(err error) {
	if err != nil {
		klog.Fatal("Error : %s\n", err)
		os.Exit(1)
	}
}

func learnRestClient(config componentbaseconfig.ClientConnectionConfiguration) {
	printline("\tRESTClient")

	kubeConfig, err := clientcmd.BuildConfigFromFlags("", config.Kubeconfig)
	checkerror(err)
	kubeConfig.APIPath = "api" //设置请求的HTTP路径
	kubeConfig.GroupVersion = &corev1.SchemeGroupVersion //设置请求的资源组/资源版本
	kubeConfig.NegotiatedSerializer = scheme.Codecs //设置数据的编解码器

	restClient, err := rest.RESTClientFor(kubeConfig)
	checkerror(err)

	result := &corev1.PodList{}
	err = restClient.Get().
		Namespace("default"). //Namespace函数设置请求的命名空间
		Resource("pods"). //Resource函数设置请求的资源名称
		VersionedParams(&metav1.ListOptions{Limit: 2}, scheme.ParameterCodec). //VersionedParams函数将一些查询选项（如limit、TimeoutSeconds等）添加到请求参数中
		Do(context.TODO()). //通过Do函数执行该请求
		Into(result) //将kube-apiserver返回的结果（Result对象）解析到corev1.PodList对象中
	checkerror(err)

	for _, d := range result.Items {
		fmt.Printf("NAMESPACE: %v\tNAME: %-50v\tSTATUS: %-20v\n", d.Namespace, d.Name, d.Status.Phase)
	}
}

func learnDiscoveryClient(config componentbaseconfig.ClientConnectionConfiguration) {
	kubeConfig, err := clientcmd.BuildConfigFromFlags("", config.Kubeconfig)
	checkerror(err)
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(kubeConfig)
	checkerror(err)
	printline("\tDiscoveryClient")
	// $ kubectl api-versions and api-resources by discoveryClient
	{
		apiGroups, apiResourceLists, err := discoveryClient.ServerGroupsAndResources()
		checkerror(err)

		fmt.Printf("\n$ kubectl api-versions\n")
		for _, apiGroup := range apiGroups{
			versions := apiGroup.Versions
			for _, version := range versions{
				fmt.Printf("%v\n", version.GroupVersion)
			}
		}

		fmt.Printf("\n$ kubectl api-resources\n")
		fmt.Printf("%-40v\t%-30v\t%-20v\t%-20v\n",
			"NAME", "APIGROUP", "NAMESPACED", "KIND")
		for _, apiResourceList := range apiResourceLists{ //此处包含了 subresource，命令行未显示子资源 subresource
			gv, err := schema.ParseGroupVersion(apiResourceList.GroupVersion)
			checkerror(err)
			for _, resource := range apiResourceList.APIResources{
				if strings.Count(resource.Name, "/") > 0{ //过滤掉 subresource
					continue
				}
				fmt.Printf("%-40v\t%-30v\t%-20v\t%-20v\n",
					resource.Name, gv.Group, resource.Namespaced, resource.Kind)
			}
		}
	}
}

func learnDynamicClient(config componentbaseconfig.ClientConnectionConfiguration) {
	kubeConfig, err := clientcmd.BuildConfigFromFlags("", config.Kubeconfig)
	checkerror(err)
	dynamicClient, err := dynamic.NewForConfig(kubeConfig)
	checkerror(err)

	printline("\tdynamicClient")
	{
		/*
			GroupVersionResource 格式： {APIGROUP}/{api-versions}/{api-resources-name}
			URL 前面添加 /api 或者 /apis
			$ kubectl get --raw /apis/extensions/v1beta1/deployments/
			$ kubectl get --raw /apis/apps/v1beta1/deployments/
			apiVersion: extensions/v1beta1
			kind: Deployment

			对应的代码获取方式：
			gvr := schema.GroupVersionResource{Group: "extensions", Version: "v1beta1", Resource: "deployments"}
			gvr := schema.GroupVersionResource{Group: "apps", Version: "v1beta1", Resource: "deployments"}
		*/

		gvr := schema.GroupVersionResource{Group: "apps", Version: "v1beta1", Resource: "deployments"}
		unstructobj, err := dynamicClient.Resource(gvr).Namespace(corev1.NamespaceDefault).List(context.TODO(), metav1.ListOptions{Limit: 50})

		deploys := &v1beta1.DeploymentList{}
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(unstructobj.UnstructuredContent(), deploys)
		checkerror(err)
		printline("get deploys")
		for _, d := range deploys.Items {
			fmt.Printf("NAMESPACE: %v\tNAME: %-50v\tReplicas: %-20v\n", d.Namespace, d.Name, d.Status.Replicas)
		}

		gvr = schema.GroupVersionResource{Group: "apiextensions.k8s.io", Version: "v1beta1", Resource: "customresourcedefinitions"}
		unstructobj, err = dynamicClient.Resource(gvr).Namespace(corev1.NamespaceAll).List(context.TODO(), metav1.ListOptions{Limit: 50})
		checkerror(err)
		printline("get customresourcedefinitions")
		fmt.Printf("%s\n\n", unstructobj)

		/*
			$ kubectl api-resources|grep openstack
			NAME                              SHORTNAMES   APIGROUP                       NAMESPACED   KIND
			kuryrports                        kp           openstack.org                  true         KuryrPort
			$ kubectl api-versions|grep opensta
			openstack.org/v1

			selfLink: /apis/openstack.org/v1/kuryrports/
		*/
		gvr = schema.GroupVersionResource{Group: "openstack.org", Version: "v1", Resource: "kuryrports"}
		unstructobj, err = dynamicClient.Resource(gvr).Namespace("ljx").List(context.TODO(), metav1.ListOptions{Limit: 50})
		checkerror(err)
		printline("get kuryrports")
		fmt.Printf("%s\n\n", unstructobj)
		// 此处如果需要解析，需要自己定义 kuryrport 的struct
	}
}

func learnClientSet(config componentbaseconfig.ClientConnectionConfiguration)  {
	kubeConfig, err := clientcmd.BuildConfigFromFlags("", config.Kubeconfig)
	checkerror(err)
	//kubeConfig.AcceptContentTypes = ""//config.AcceptContentTypes
	//kubeConfig.ContentType = ""//config.ContentType
	//kubeConfig.QPS = config.QPS
	//kubeConfig.Burst = int(config.Burst)

	clientSetClient, err := clientset.NewForConfig(kubeConfig)
	checkerror(err)


	printline("\tClientset")
	{
		pods, err := clientSetClient.CoreV1().Pods(corev1.NamespaceDefault).List(context.TODO(), metav1.ListOptions{Limit: 50})
		checkerror(err)
		printline("get Pods ")

		for _, d := range pods.Items {
			fmt.Printf("NAMESPACE: %v\tNAME: %-50v\tSTATUS: %-20v\n", d.Namespace, d.Name, d.Status.Phase)
		}

		printline("get Deployments ")
		deployments, err := clientSetClient.ExtensionsV1beta1().Deployments(corev1.NamespaceDefault).List(context.TODO(), metav1.ListOptions{Limit: 50})
		checkerror(err)
		for _, d := range deployments.Items {
			fmt.Printf("NAMESPACE: %v\tNAME: %-50v\tAVAILABLE: %v\n", d.Namespace, d.Name, d.Status.AvailableReplicas)
		}
	}
}

func learnInformer(config componentbaseconfig.ClientConnectionConfiguration)  {
	printline("\tInformer")
	kubeConfig, err := clientcmd.BuildConfigFromFlags("", config.Kubeconfig)
	checkerror(err)

	clientSetClient, err := clientset.NewForConfig(kubeConfig)
	checkerror(err)
	/*
		在正常的情况下，Kubernetes的其他组件在使用Informer机制时触发资源事件回调方法，将资源对象推送到WorkQueue或其他队列中。
		在该例子中是直接输出触发的资源事件。最后通过informer.Run函数运行当前的Informer，内部为Pod资源类型创建Informer。
	*/

	informerFactory := informers.NewSharedInformerFactory(clientSetClient, informerDefaultResync) // cmd/kube-proxy/app/server.go#733

	{
		podInformer := informerFactory.Core().V1().Pods().Informer() //得到具体Pod资源的informer 对象
		podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{ //为Pod资源添加资源事件回调方法
			AddFunc: func(obj interface{}){ //当创建Pod资源对象时触发的事件回调方法。
			},
			UpdateFunc: func(oldObj, newObj interface{}){ //当更新Pod资源对象时触发的事件回调方法。
				oObj := oldObj.(metav1.Object)
				nObj := newObj.(metav1.Object)
				log.Printf("Informer-> %s Pod Update to %s\n", oObj.GetName(), nObj.GetName())
			},
			DeleteFunc: func(obj interface{}){//当删除Pod资源对象时触发的事件回调方法。
			},
		})
		informerFactory.Core().V1().Pods().Lister() //这行的作用是？
		go podInformer.Run(wait.NeverStop)
	}

	{ //svcInformer
		stopCh := make(chan struct{})
		defer close(stopCh)

		svcInformer := informerFactory.Core().V1().Services().Informer()

		svcInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}){
				mObj := obj.(metav1.Object)
				log.Printf("Informer-> New Service Added to Store : %s\n", mObj.GetName())
			},
			UpdateFunc: func(oldObj, newObj interface{}){
				oObj := oldObj.(metav1.Object)
				nObj := newObj.(metav1.Object)
				log.Printf("Informer-> %s Service Update to %s\n", oObj.GetName(), nObj.GetName())
			},
			DeleteFunc: func(obj interface{}){
				mObj := obj.(metav1.Object)
				log.Printf("Informer-> Service Delete from Store : %s\n", mObj.GetName())
			},
		})

		svcInformer.Run(stopCh)
	}

	informerFactory.Start(wait.NeverStop) //非阻塞运行，但是该代码中主线程会退出，所以暂时只能以阻塞形式运行

}

func userIndexFunc(obj interface{}) ([]string, error){
	pod := obj.(*corev1.Pod)
	klog.Infof("users: %v\n", pod.Annotations["users"])
	return strings.Split(pod.Annotations["users"], ","), nil
}

func LearnIndexer(){
	index := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{"byUser": userIndexFunc})

	pod1 := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "one", Annotations: map[string]string{"users": "ernie,bert"}}}
	pod2 := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "two", Annotations: map[string]string{"users": "bert,oscar,ernie"}}}

	index.Add(pod1)
	index.Add(pod2)

	erniePods, err := index.ByIndex("byUser", "ernie")
	if err != nil {
		klog.Errorf("Error: %s", err)
	}
	for _, erniePod := range erniePods {
		klog.Infof("erniePod: %s\n", erniePod.(*corev1.Pod).Name)
	}
}


func LearnClientGo(config *ControllerConfig) {
	learnRestClient(config.ClientConnection)
	learnDiscoveryClient(config.ClientConnection)
	learnDynamicClient(config.ClientConnection)
	learnClientSet(config.ClientConnection)
	learnInformer(config.ClientConnection)
}

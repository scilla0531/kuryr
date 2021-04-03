# Manual Installation
[antrea ref](docs/contributors/manual-installation.md)

## Overview
There are four components which need to be deployed in order to run Kuryr:
* The OpenVSwitch daemons `ovs-vswitchd` and `ovsdb-server`

* The controller `kuryr-controller`

* The agent `kuryr-agent`

* The CNI plugin `kuryr-cni`

## Instructions
### kuryr-controller

`kuryr-controller` is required to implement Kubernetes Network Policies. At any time, there should be only a single
active replica of `kuryr-controller`.

1. Grant the `kuryr-controller` ServiceAccount necessary permissions to Kubernetes APIs. You can apply
[controller-rbac.yaml](/build/yamls/base/controller-rbac.yml) to do it.

    ```bash
    kubectl apply -f build/yamls/base/controller-rbac.yml
    ```

2. Create the kubeconfig file that contains the K8s APIServer endpoint and the token of ServiceAccount created in the
above step. See [Configure Access to Multiple Clusters](
https://kubernetes.io/docs/tasks/access-application-cluster/configure-access-multiple-clusters/) for more information.

    ```bash
    APISERVER=$(kubectl config view --minify -o jsonpath='{.clusters[0].cluster.server}')
    TOKEN=$(kubectl get secrets -n kube-system -o jsonpath="{.items[?(@.metadata.annotations['kubernetes\.io/service-account\.name']=='kuryr-controller')].data.token}"|base64 --decode)
    kubectl config --kubeconfig=kuryr-controller.kubeconfig set-cluster kubernetes --server=$APISERVER --insecure-skip-tls-verify
    kubectl config --kubeconfig=kuryr-controller.kubeconfig set-credentials kuryr-controller --token=$TOKEN
    kubectl config --kubeconfig=kuryr-controller.kubeconfig set-context kuryr-controller@kubernetes --cluster=kubernetes --user=kuryr-controller
    kubectl config --kubeconfig=kuryr-controller.kubeconfig use-context kuryr-controller@kubernetes
    ```

3. Create the `kuryr-controller` config file, see [Configuration](../configuration.md) for details.
    网络资源使用的两种模式：
        配置中指定了共享网络资源（project-admin、shared-network、subnet-pool），那么在创建ns的时候不指定 cnitype 也会创建 kns
        配置中没有指定网络资源，ns变化时判断cnitype，只有指定了cni的ns才创建kns，pod等其他资源管理同理
    projectId为必须指定的资源

    ```bash
最后再写，暂时先以run-controller.sh 进行调试
    ```

4. Start `kuryr-controller`.

    ```bash
    /bin/kuryr-controller --config build/yamls/base/conf/kuryr-controller.conf
   go build -o ./kuryr-controller ./cmd/kuryr-controller; ./kuryr-controller --config build/yamls/base/conf/kuryr-controller.conf
    ```

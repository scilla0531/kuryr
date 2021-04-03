## kuryr-agent

2. Create the kubeconfig file that contains the K8s APIServer endpoint and the token of ServiceAccount created in the
above step. See [Configure Access to Multiple Clusters](
https://kubernetes.io/docs/tasks/access-application-cluster/configure-access-multiple-clusters/) for more information.

    ```bash
    APISERVER=$(kubectl config view --minify -o jsonpath='{.clusters[0].cluster.server}')
    TOKEN=$(kubectl get secrets -n kube-system -o jsonpath="{.items[?(@.metadata.annotations['kubernetes\.io/service-account\.name']=='kuryr-agent')].data.token}"|base64 --decode)
    TOKEN=$(kubectl get secrets -n kube-system -o jsonpath="{.items[?(@.metadata.annotations['kubernetes\.io/service-account\.name']=='default')].data.token}"|base64 --decode)
   
    kubectl config --kubeconfig=kuryr-agent.kubeconfig set-cluster kubernetes --server=$APISERVER --insecure-skip-tls-verify
    kubectl config --kubeconfig=kuryr-agent.kubeconfig set-credentials kuryr-agent --token=$TOKEN
    kubectl config --kubeconfig=kuryr-agent.kubeconfig set-context kuryr-agent@kubernetes --cluster=kubernetes --user=kuryr-agent
    kubectl config --kubeconfig=kuryr-agent.kubeconfig use-context kuryr-agent@kubernetes
    ```
   
4. Create the `kuryr-agent` config file, see [Configuration](../configuration.md) for details.
    ```bash
    cat >kuryr-agent.conf <<EOF
    clientConnection:
      kubeconfig: kuryr-agent.kubeconfig
    cniSocket: "/"
    EOF
    ```
   
bulild:  
`go build -o ./kuryr-agent ./cmd/kuryr-agent`

run:  
`go run ./cmd/kuryr-agent/ --config  ./conf/kuryr-agent.confls`

# 创建认证信息
`kubectl create -f conf/service_account.yml`
# 创建 crd
```bash
> kubectl apply -f crds.yaml
customresourcedefinition.apiextensions.k8s.io/kuryrnetworks.openstack.org created
```
创建该资源的实例：

使用client-go包来访问这些自定义资源。

# token 信息








---
---

## kuryr-cni
go build -o /go/bin/kuryr-cni ./cmd/kuryr-cni
## kuryr-controller
go build -o /go/bin/kuryr-controller ./cmd/kuryr-controller
## kuryr-proxy
go build -o /go/bin/kuryr-proxy ./cmd/kuryr-proxy


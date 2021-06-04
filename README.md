# Manual Installation
## 镜像制作
[protoc 镜像制作](build/images/codegen/README.md)

## grpc codegen
```bash
./hack/update-codegen.sh  //需要更改为自己的镜像仓库
```

## config&token init
```bash
./run-kuryr.sh init
```

## kuryr-cni
```bash
./run-kuryr.sh cni
```

## kuryr-controller
```bash
./run-kuryr.sh con
```

## kuryr-agent
```bash
./run-kuryr.sh agent
```

# 未明
2. Create the kubeconfig file that contains the K8s APIServer endpoint and the token of ServiceAccount created in the
above step. See [Configure Access to Multiple Clusters](
https://kubernetes.io/docs/tasks/access-application-cluster/configure-access-multiple-clusters/) for more information.



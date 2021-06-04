2. Create the kubeconfig file that contains the K8s APIServer endpoint and the token of ServiceAccount created in the
above step. See [Configure Access to Multiple Clusters](
https://kubernetes.io/docs/tasks/access-application-cluster/configure-access-multiple-clusters/) for more information.

# 调试运行命令
## 镜像制作
[protoc 镜像制作](build/images/codegen/README.md)

## grpc codegen
```bash
./hack/update-codegen.sh
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

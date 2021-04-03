# 代码生成器
CustomResourceDefinition（CRD） 在很多k8s周边开源项目中有使用，比如ingress-controller和众多的operator。

编写 CRD controller之前，一定要使用k8s官方提供的代码生成工具 http://k8s.io/code-generator 去生成 client, informers, listers and deep-copy函数。不仅代码风格符合k8s，而且减少出错和减少工作量都是有很大的帮助。

## 官方提供的示例项目 [sample-controller](https://github.com/kubernetes/sample-controller)
这个例子主要讲述了以下几个方面：
- 如何使用自定义资源定义注册 Foo 类型的新自定义资源（自定义资源类型）
- 如何创建/获取/列出新资源类型 Foo 实例
- 如何在资源处理创建/更新/删除事件上设置控制器
### 使用方式
- client 代码生成
```bash
cd $GOPATH/src/k8s.io/
git clone https://github.com/kubernetes/sample-controller.git
    cd sample-controller; git checkout release-1.20; git pull; cd -
git clone https://github.com/kubernetes/code-generator.git
    cd code-generator; git checkout release-1.20; git pull; cd -

cd sample-controller
go mod vendor
./hack/update-codegen.sh
```
执行完之后新增文件：

    pkg/apis/${group}/${version}/zz_generated.deepcopy.go
    pkg/client/

- 编译运行程序
```bash
go build -o sample-controller .
./sample-controller -kubeconfig=$HOME/.kube/config

kubectl create -f artifacts/examples/crd.yaml
kubectl create -f artifacts/examples/example-foo.yaml
kubectl get deployments
```

### 编写 CRD controller
- 定义CRD，最重要的是group以及version
- 需要自己写的文件
```
├── hack
│   ├── boilerplate.go.txt
│   ├── update-codegen.sh
│   └── verify-codegen.sh
└── pkg
    ├── apis
        └── mycontroller
            ├── register.go
            └── v1alpha1
                ├── doc.go
                ├── register.go
                ├── types.go
```
```bash 
mkdir -p pkg/apis/${group}/${version}
touch pkg/apis/${group}/${version}/types.go
touch pkg/apis/${group}/${version}/register.go
```
- 执行 `./hack/update-codegen.sh` 

### update-codegen.sh 文件详解：[:-变量置换](https://blog.csdn.net/Zheng__Huang/article/details/107902325)

SCRIPT_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
    ${BASH_SOURCE[0]} 即 source/bash 后面跟的第一个参数，比如：bash ./gen.sh 就是 ./gen.sh， dirname命令取目录名字，不包含/，即.)
    所以 SCRIPT_ROOT 为 ./..

CODEGEN_PKG=${CODEGEN_PKG:-$(cd "${SCRIPT_ROOT}"; ls -d -1 ./vendor/k8s.io/code-generator 2>/dev/null || echo ../code-generator)}
    默认为 ../code-generator（可以手动设置：CODEGEN_PKG="$GOPATH/src/k8s.io/code-generator"）

bash "${CODEGEN_PKG}"/generate-groups.sh "deepcopy,client,informer,lister" \
  k8s.io/sample-controller/pkg/generated k8s.io/sample-controller/pkg/apis \
  samplecontroller:v1alpha1 \
  --output-base "$(dirname "${BASH_SOURCE[0]}")/../../.." \
  --go-header-file "${SCRIPT_ROOT}"/hack/boilerplate.go.txt
    该命令的当前目录为 SCRIPT_ROOT， 即执行命令的上级目录

`code-generator/generate-groups.sh` `Usage: <generators> <output-package> <apis-package> <groups-versions> --output-base ...`
    --output-base //输出基础路径
    <output-package> <apis-package> //表示个是你要生成代码的目录，目录的名称是 generated(一般为client)

### controller.go 解析

创建：
    kubeClient
    crdClient
    kubeInformerFactory
    crdInformerFactory
    创建 controller(config, kubeClient, crdClient, kubeInformerFactory, crdInformerFactory) // 也可以在controller内创建
        controller中
    kubeInformerFactory.Start(stopCh)
    crdInformerFactory.Start(stopCh)
    controller.Run

---
### projectkuryr:
```bash
export CODEGEN_PKG="$GOPATH/src/k8s.io/code-generator"
cd ${GOPATH/src/projectkuryr/kuryr}
./hack/update-codegen.sh


go build -o ./kuryr-controller ./cmd/kuryr-controller;
./kuryr-controller --config /home/ljx/conf/tmp.conf
```







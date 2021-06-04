# images/codegen

This Docker image is a very lightweight image based on golang 1.15 which
includes codegen tools.

If you need to build a new version of the image and push it to Dockerhub, you
can run the following:

```bash
cd build/images/codegen
docker build -t kuryr/codegen:<TAG> .
docker push kuryr/codegen:<TAG>
```

```bash
IMAGE_NAME=registry-jinan-lab.insprcloud.cn/library/cke/kuryr/codegen:kubernetes-1.18.4  // 跟 hack中pull的镜像路径一致
docker build -t ${IMAGE_NAME} -f ./build/images/codegen/Dockerfile .
docker push ${IMAGE_NAME}
```


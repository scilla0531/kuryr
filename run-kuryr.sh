#!/bin/bash
conf_dir=/home/ljx/conf

initKubeConfig(){
  PROJECT=admin
  OPENSTACK_AUTH_URL=$(openstack endpoint list | grep public| grep keystone |awk '{print $(NF-1)}')
  OPENSTACK_PROJECT_SG_ID=$(openstack security group list --project ${PROJECT} | grep "Default security group" | awk '{print $2}')
  OPENSTACK_KURYR_POD_SUBNET_POOL=$(openstack subnet pool list | grep kuryr | awk '{print $2}') # and OPENSTACK_KURYR_POD_SUBNETID either one
  OPENSTACK_KURYR_POD_SUBNETID=$(openstack network list |grep kuryr-pod | awk '{print $(NF-1)}') #

  OPENSTACK_KURYR_ROUTER_ID=$(openstack router list | grep kuryr | awk '{print $2}')
  OPENSTACK_KURYR_SVC_SUBNETID=$(openstack network list |grep kuryr-service | awk '{print $(NF-1)}')

  echo ${OPENSTACK_AUTH_URL}
  APISERVER=$(kubectl config view --minify -o jsonpath='{.clusters[0].cluster.server}')
  TOKEN=$(kubectl get secrets -n kube-system -o jsonpath="{.items[?(@.metadata.annotations['kubernetes\.io/service-account\.name']=='kuryr-controller')].data.token}"|base64 --decode)
  kubectl config --kubeconfig=${conf_dir}/kuryr.kubeconfig set-cluster kubernetes --server=$APISERVER --insecure-skip-tls-verify
  kubectl config --kubeconfig=${conf_dir}/kuryr.kubeconfig set-credentials kuryr-controller --token=$TOKEN
  kubectl config --kubeconfig=${conf_dir}/kuryr.kubeconfig set-context kuryr-controller@kubernetes --cluster=kubernetes --user=kuryr-controller
  kubectl config --kubeconfig=${conf_dir}/kuryr.kubeconfig use-context kuryr-controller@kubernetes

cat >${conf_dir}/kuryr-controller.conf <<EOF
# Required Configuration
clientConnection:
    kubeconfig: ${conf_dir}/kuryr.kubeconfig
openstack:
    authUrl : ${OPENSTACK_AUTH_URL}
    authType : password
    projectDomainName : Default
    userDomainName : Default
    userName : admin
    passWord : password
    projectName : admin

    podSubnet : ${OPENSTACK_KURYR_POD_SUBNETID}
#   podSubnetPool : ${OPENSTACK_KURYR_POD_SUBNET_POOL}
    podRouter : ${OPENSTACK_KURYR_ROUTER_ID}
    podSecurityGroups :
    - ${OPENSTACK_PROJECT_SG_ID}
    svcSubnet : ${OPENSTACK_KURYR_SVC_SUBNETID}
    linkIface : eth0
    ovsBridge : br-int
EOF

cat >${conf_dir}/kuryr-agent.conf <<EOF
# Required Configuration
clientConnection:
    kubeconfig: ${conf_dir}/kuryr.kubeconfig

ovsBridge : br-int
healthzBindAddress :

# linkIface : eth0
# dockerMode = true
# netnsProcDir = /host_proc
EOF
}


mock_kubelet_invoke_cni(){
  echo "Namespace Name: $1"
  export K8S_POD_NAMESPACE=$1
  export K8S_POD_NAME=$(kubectl get pod -nljx | grep -v "NAME" |  awk '{print $1}')
  export PAUSE_CID=$(docker ps |grep POD_$K8S_POD_NAME |awk '{print $1}')
  export PAUSE_PID=`docker inspect $PAUSE_CID -f {{.State.Pid}}`
  export PAUSE_ID=`docker inspect $PAUSE_CID -f {{.Id}}`

  echo "PAUSE_CID "${PAUSE_CID}
  echo "PAUSE_ID: "${PAUSE_ID}
  echo "PAUSE_PID: "${PAUSE_PID}

  cat <<EOF | sudo tee /etc/cni/net.d/10-kuryr.conf
{
  "cniVersion": "0.3.1",
  "name": "kuryr",
  "type": "kuryr-cni",
  "kuryr_conf": "/etc/kuryr/xxx.conf",
  "debug": true
}
EOF


  cat /etc/cni/net.d/10-kuryr.conf | sudo CNI_COMMAND=ADD CNI_NETNS="/proc/${PAUSE_PID}/ns/net" \
    CNI_PATH=${conf_dir} \
    CNI_IFNAME=eth1 \
    CNI_CONTAINERID=${PAUSE_ID} \
    CNI_ARGS="IgnoreUnknown=1;K8S_POD_NAMESPACE=${K8S_POD_NAMESPACE};K8S_POD_NAME=${K8S_POD_NAME};K8S_POD_INFRA_CONTAINER_ID=${PAUSE_ID}"  ${conf_dir}/kuryr-cni
}

echo ""

if [ "$1" = "init" ];then
  if [ "" == "$2" ];then
    echo "Need <openrcFile>"
    return 100
  else
    source $2
    echo "Initialize kubeconfig!"
    initKubeConfig
  fi
elif [ "$1" = "agent" ];then
  echo "build & run kuryr-agent!"
  go build -o ${conf_dir}/kuryr-agent ./cmd/kuryr-agent
  if [ $? -ne 0 ]; then
    return 100
  fi
  ${conf_dir}/kuryr-agent --config ${conf_dir}/kuryr-agent.conf
elif [ "$1" = "cni" ];then
  ComponentName=$1
  echo "build ${ComponentName}!"
  go build -o ${conf_dir}/kuryr-${ComponentName} ./cmd/kuryr-${ComponentName}
  if [ $? -ne 0 ]; then
    return 100
  fi
  mock_kubelet_invoke_cni "ljx"
else
  echo "build & run kuryr-controller!"
  go build -o ${conf_dir}/kuryr-controller ./cmd/kuryr-controller
  if [ $? -ne 0 ]; then
    return 100
  fi
  ${conf_dir}/kuryr-controller --config ${conf_dir}/kuryr-controller.conf
fi

echo ""










#!/bin/bash
conf_dir=/home/ljx

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
  kubectl config --kubeconfig=${conf_dir}/kuryr-controller.kubeconfig set-cluster kubernetes --server=$APISERVER --insecure-skip-tls-verify
  kubectl config --kubeconfig=${conf_dir}/kuryr-controller.kubeconfig set-credentials kuryr-controller --token=$TOKEN
  kubectl config --kubeconfig=${conf_dir}/kuryr-controller.kubeconfig set-context kuryr-controller@kubernetes --cluster=kubernetes --user=kuryr-controller
  kubectl config --kubeconfig=${conf_dir}/kuryr-controller.kubeconfig use-context kuryr-controller@kubernetes


cat >${conf_dir}/kuryr-controller.conf <<EOF
# Required Configuration
clientConnection:
    kubeconfig: ${conf_dir}/kuryr-controller.kubeconfig
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

# 再生成kuryr-agent.conf

}

if [ "$1" = "init" ];then
  if [ "" == "$2" ];then
    echo "Need <openrcFile>"
    exit 1
  else
    source $2
    echo "Initialize kubeconfig!"
    initKubeConfig
  fi
elif [ "$1" = "agent" ];then
  echo "run kuryr-agent!"
  go build -o ./kuryr-controller ./cmd/kuryr-controller
  ./kuryr-controller --config ${conf_dir}kuryr-agent.conf
else
  echo "run kuryr-controller!"
  go build -o ${conf_dir}/kuryr-controller ./cmd/kuryr-controller
  if [ $? -ne 0 ]; then
    exit 1
  fi
  ${conf_dir}/kuryr-controller --config ${conf_dir}/kuryr-controller.conf
fi






package main

import (
	"k8s.io/component-base/logs"
	"os"
	"projectkuryr/kuryr/cmd/kuryr-agent/app"
)

/*
配置解析：
	解析配置文件 -> 设置必备缺省参数 -> 检查参数合法性（比如ip:port）
线程：
	// 线程1： 轮询监视 apiserver （还没整明白 clientgo）
ListenAndServe
	// 线程2:  监听端口 响应 kubelet 的 addNetwork 和 delNetwork -> 查询到kp数据 -> 下发ovs操作 -> 返回状态
	// 线程3： 健康检查
	// 线程4： metric
*/

func main() {
	logs.InitLogs()
	defer logs.FlushLogs()

	command := app.NewAgentCommand()

	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}



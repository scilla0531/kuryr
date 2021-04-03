package main

import (
	"k8s.io/component-base/logs"
	"os"
	"projectkuryr/kuryr/cmd/kuryr-controller/app"
)

/*
配置解析：
	解析配置文件 -> 设置必备缺省参数 -> 检查参数合法性（比如ip:port）

watch (if specify kuryr:)
    ns      -> create kns
    pod     -> create kp
    svc,ep     -> create lb
healthcheck
metric
*/


func main() {
	logs.InitLogs()
	defer logs.FlushLogs()

	command := app.NewControllerCommand()

	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}


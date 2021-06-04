package main

import (
	"fmt"
	"github.com/containernetworking/cni/pkg/skel"
	cniversion "github.com/containernetworking/cni/pkg/version"
	"projectkuryr/kuryr/pkg/cni"
	"projectkuryr/kuryr/pkg/version"
)

func main() {
	skel.PluginMain(
		cni.ActionAdd.Request,
		cni.ActionCheck.Request,
		cni.ActionDel.Request,
		cniversion.All,
		fmt.Sprintf("Kuryr CNI %s", version.GetFullVersionWithRuntimeInfo()),
	)
}

// +build !windows

package ovsctl

import "os/exec"

// File path of the ovs-vswitchd control UNIX domain socket.
const ovsVSwitchdUDS = "/var/run/openvswitch/ovs-vswitchd.*.ctl"

func getOVSCommand(cmdStr string) *exec.Cmd {
	return exec.Command("/bin/sh", "-c", cmdStr)
}

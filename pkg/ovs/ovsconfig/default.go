package ovsconfig

import (
	"path"
	"time"
)

const (
	DefaultOVSRunDir = "/var/run/openvswitch"

	defaultConnNetwork = "unix"
	// Wait up to 5 seconds when getting port.
	defaultGetPortTimeout    = 5 * time.Second
	defaultOvsVersionMessage = "OVS version not found in ovsdb. Please configure your OVS (ovsdb) to provide version information."
)

func GetConnAddress(ovsRunDir string) string {
	return path.Join(ovsRunDir, defaultOVSDBFile)
}

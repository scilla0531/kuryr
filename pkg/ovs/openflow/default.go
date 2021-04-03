
// +build !windows

package openflow

import (
	"path"
)

func GetMgmtAddress(ovsRunDir, brName string) string {
	return path.Join(ovsRunDir, brName+".mgmt")
}

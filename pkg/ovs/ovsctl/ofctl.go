package ovsctl

import (
	"bufio"
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

func (c *ovsCtlClient) DumpFlows(args ...string) ([]string, error) {
	// Print table and port names.
	flowDump, err := c.RunOfctlCmd("dump-flows", append(args, "--names")...)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(strings.NewReader(string(flowDump)))
	scanner.Split(bufio.ScanLines)
	flowList := []string{}
	for scanner.Scan() {
		flowList = append(flowList, trimFlowStr(scanner.Text()))
	}
	return flowList, nil

}

func (c *ovsCtlClient) DumpMatchedFlow(matchStr string) (string, error) {
	flowDump, err := c.RunOfctlCmd("dump-flows", matchStr, "--names")
	if err != nil {
		return "", err
	}
	scanner := bufio.NewScanner(strings.NewReader(string(flowDump)))
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		flowStr := trimFlowStr(scanner.Text())
		// ovs-ofctl dump-flows can return multiple flows that match matchStr, here we
		// check and return only the one that exactly matches matchStr (no extra match
		// conditions).
		if flowExactMatch(matchStr, flowStr) {
			return flowStr, nil
		}
	}

	// No exactly matched flow found.
	return "", nil
}

func (c *ovsCtlClient) DumpTableFlows(table uint8) ([]string, error) {
	return c.DumpFlows(fmt.Sprintf("table=%d", table))
}

func (c *ovsCtlClient) DumpGroup(groupID int) (string, error) {
	// There seems a bug in ovs-ofctl that dump-groups always returns all
	// the groups when using Openflow13, even when the group ID is provided.
	// As a workaround, we do not specify Openflow13 to run the command.
	groupDump, err := c.runOfctlCmd(false, "dump-groups", strconv.Itoa(groupID))
	if err != nil {
		return "", err
	}

	scanner := bufio.NewScanner(strings.NewReader(string(groupDump)))
	scanner.Split(bufio.ScanLines)
	// Skip the first line.
	scanner.Scan()
	if !scanner.Scan() {
		// No group found.
		return "", nil
	}
	// Should have at most one line (group) returned.
	return strings.TrimSpace(scanner.Text()), nil
}

func (c *ovsCtlClient) DumpGroups(args ...string) ([][]string, error) {
	groupsDump, err := c.RunOfctlCmd("dump-groups", args...)
	if err != nil {
		return nil, err
	}
	groupsDumpStr := strings.TrimSpace(string(groupsDump))

	scanner := bufio.NewScanner(strings.NewReader(groupsDumpStr))
	scanner.Split(bufio.ScanLines)
	// Skip the first line.
	scanner.Scan()
	rawGroupItems := []string{}
	for scanner.Scan() {
		rawGroupItems = append(rawGroupItems, scanner.Text())
	}

	var groupList [][]string
	for _, rawGroupItem := range rawGroupItems {
		rawGroupItem = strings.TrimSpace(rawGroupItem)
		elems := strings.Split(rawGroupItem, ",bucket=")
		groupList = append(groupList, elems)
	}
	return groupList, nil
}

func (c *ovsCtlClient) DumpPortsDesc() ([][]string, error) {
	portsDescDump, err := c.RunOfctlCmd("dump-ports-desc")
	if err != nil {
		return nil, err
	}
	portsDescStr := strings.TrimSpace(string(portsDescDump))
	scanner := bufio.NewScanner(strings.NewReader(portsDescStr))
	scanner.Split(bufio.ScanLines)
	// Skip the first line.
	scanner.Scan()

	rawPortDescItems := make([][]string, 0)
	var portItem []string
	for scanner.Scan() {
		str := scanner.Text()
		// If the line starts with a port number, it should be the first line of an OF port. There should be some
		// subsequent lines to describe the status of the current port, which start with multiple while-spaces.
		if len(str) > 2 && string(str[1]) != " " {
			if len(portItem) > 0 {
				rawPortDescItems = append(rawPortDescItems, portItem)
			}
			portItem = nil
		}
		portItem = append(portItem, scanner.Text())
	}
	if len(portItem) > 0 {
		rawPortDescItems = append(rawPortDescItems, portItem)
	}
	return rawPortDescItems, nil
}

func (c *ovsCtlClient) SetPortNoFlood(ofport int) error {
	cmdStr := fmt.Sprintf("ovs-ofctl mod-port %s %d no-flood", c.bridge, ofport)
	cmd := getOVSCommand(cmdStr)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("fail to set no-food config for port %d on bridge %s: %v, stderr: %s", ofport, c.bridge, err, string(stderr.Bytes()))
	}
	return nil
}

func (c *ovsCtlClient) runOfctlCmd(openflow13 bool, cmd string, args ...string) ([]byte, error) {
	cmdStr := fmt.Sprintf("ovs-ofctl %s %s", cmd, c.bridge)
	cmdStr = cmdStr + " " + strings.Join(args, " ")
	if openflow13 {
		cmdStr += " -O Openflow13"
	}
	out, err := getOVSCommand(cmdStr).Output()
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *ovsCtlClient) RunOfctlCmd(cmd string, args ...string) ([]byte, error) {
	// Default to use Openflow13.
	return c.runOfctlCmd(true, cmd, args...)
}

// trimFlowStr removes undesirable fields from the flow string.
func trimFlowStr(flowStr string) string {
	return flowStr[strings.Index(flowStr, " table")+1:]
}

func flowExactMatch(matchStr, flowStr string) bool {
	// Get the match string which starts with "priority=".
	flowStr = flowStr[strings.Index(flowStr, " priority")+1 : strings.LastIndexByte(flowStr, ' ')]
	matches := strings.Split(flowStr, ",")
	for i, m := range matches {
		// Skip "priority=".
		if i == 0 {
			continue
		}
		if i := strings.Index(m, "="); i != -1 {
			m = m[:i]
		}
		if !strings.Contains(matchStr, m) {
			// The match condition is not included in matchStr.
			return false
		}
	}
	return true
}

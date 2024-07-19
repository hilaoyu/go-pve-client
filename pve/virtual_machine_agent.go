package pve

import (
	"fmt"
	"github.com/hilaoyu/go-utils/utils"
	"math"
	"time"
)

type AgentNetworkIpAddress struct {
	IpAddressType string `json:"ip-address-type"` //ipv4 ipv6
	IpAddress     string `json:"ip-address"`
	Prefix        int    `json:"prefix"`
	MacAddress    string `json:"mac-address"`
}

type AgentNetworkIface struct {
	Name            string                   `json:"name"`
	HardwareAddress string                   `json:"hardware-address"`
	IpAddresses     []*AgentNetworkIpAddress `json:"ip-addresses"`
}

type AgentOsInfo struct {
	Version       string `json:"version"`
	VersionId     string `json:"version-id"`
	Id            string `json:"id"`
	Machine       string `json:"machine"`
	PrettyName    string `json:"pretty-name"`
	Name          string `json:"name"`
	KernelRelease string `json:"kernel-release"`
	KernelVersion string `json:"kernel-version"`
}

type AgentFileReadResult struct {
	Content   string         `json:"content,omitempty"`
	BytesRead StringOrUint64 `json:"bytes-read,omitempty"`
}

type AgentExecCommand struct {
	Command string
	Args    []string
}
type AgentExecResult struct {
	Pid int64 `json:"pid,omitempty"`
}
type AgentExecStatusResult struct {
	Exited   int    `json:"exited,omitempty"`
	OutData  string `json:"out-data,omitempty"`
	Exitcode int    `json:"exitcode,omitempty"`
}

func (v *VirtualMachine) AgentGetNetworkIFaces() (iFaces []*AgentNetworkIface, err error) {
	node, err := v.client.Node(v.Node)
	if err != nil {
		return
	}

	networks := map[string][]*AgentNetworkIface{}
	err = v.client.Get(fmt.Sprintf("/nodes/%s/qemu/%d/agent/network-get-interfaces", node.Name, v.VMID), &networks)
	if err != nil {
		return
	}
	if result, ok := networks["result"]; ok {
		for _, iface := range result {
			if "lo" == iface.Name {
				continue
			}
			iFaces = append(iFaces, iface)
		}
	}

	return

}

func (v *VirtualMachine) AgentOsInfo() (info *AgentOsInfo, err error) {
	node, err := v.client.Node(v.Node)
	if err != nil {
		return
	}
	results := map[string]*AgentOsInfo{}
	err = v.client.Get(fmt.Sprintf("/nodes/%s/qemu/%d/agent/get-osinfo", node.Name, v.VMID), &results)

	if err != nil {
		return
	}
	info, ok := results["result"]
	if !ok {
		err = fmt.Errorf("result is empty")
	}
	return

}
func (v *VirtualMachine) AgentSetUserPassword(password string, username string) (err error) {
	node, err := v.client.Node(v.Node)
	if err != nil {
		return
	}

	err = v.client.Post(fmt.Sprintf("/nodes/%s/qemu/%d/agent/set-user-password", node.Name, v.VMID), map[string]string{"password": password, "username": username}, nil)

	return

}
func (v *VirtualMachine) AgentFileRead(file string) (content string, err error) {
	node, err := v.client.Node(v.Node)
	if err != nil {
		return
	}

	result := &AgentFileReadResult{}
	err = v.client.Get(fmt.Sprintf("/nodes/%s/qemu/%d/agent/file-read?file=%s", node.Name, v.VMID, file), &result)

	if nil != err {
		return
	}

	content = result.Content

	return

}
func (v *VirtualMachine) AgentFileWrite(file string, content string) (err error) {
	node, err := v.client.Node(v.Node)
	if err != nil {
		return
	}

	//content = base64.StdEncoding.EncodeToString([]byte(content))
	err = v.client.Post(fmt.Sprintf("/nodes/%s/qemu/%d/agent/file-write", node.Name, v.VMID), map[string]interface{}{"file": file, "content": content, "encode": true}, nil)

	return

}
func (v *VirtualMachine) AgentExec(command *AgentExecCommand) (pid int64, err error) {
	node, err := v.client.Node(v.Node)
	if err != nil {
		return
	}

	params := []string{command.Command}
	params = append(params, command.Args...)
	result := &AgentExecResult{}
	err = v.client.Post(fmt.Sprintf("/nodes/%s/qemu/%d/agent/exec", node.Name, v.VMID), map[string][]string{"command": params}, result)
	if nil != err {
		return
	}

	pid = result.Pid

	return

}
func (v *VirtualMachine) AgentExecStatus(pid int64) (result *AgentExecStatusResult, err error) {
	node, err := v.client.Node(v.Node)
	if err != nil {
		return
	}
	result = &AgentExecStatusResult{}
	err = v.client.Get(fmt.Sprintf("/nodes/%s/qemu/%d/agent/exec-status?pid=%d", node.Name, v.VMID, pid), result)
	if nil != err {
		return
	}

	return

}

func (v *VirtualMachine) AgentExecSync(command *AgentExecCommand, timeoutSeconds int) (output string, err error) {
	pid, err := v.AgentExec(command)
	if nil != err {
		return
	}

	step := 5
	times := 6
	if timeoutSeconds <= 0 {
		times = timeoutSeconds
	} else {
		if timeoutSeconds < step {
			step = timeoutSeconds
		}
		times = int(math.Ceil(float64(timeoutSeconds/step))) + 1
	}

	result := &AgentExecStatusResult{}
	utils.ReTry(func() bool {
		result, _ = v.AgentExecStatus(pid)
		if 1 == result.Exited {
			output = result.OutData
			if 0 != result.Exitcode {
				err = fmt.Errorf("执行状态有错误,code: %d", result.Exitcode)
			}
			return true
		}

		return false
	}, times, time.Duration(step)*time.Second)

	if 1 != result.Exited {
		err = fmt.Errorf("执行超过时设定的时间: %d秒", timeoutSeconds)
	}

	return

}

package pve

import "fmt"

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

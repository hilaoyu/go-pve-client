package pve

import (
	"fmt"
	"net/url"
)

type NodeStatuses []*NodeStatus
type NodeStatus struct {
	// shared
	Status string `json:",omitempty"`
	Level  string `json:",omitempty"`
	ID     string `json:",omitempty"` // format "node/<name>"

	// from /nodes endpoint
	Node           string  `json:",omitempty"`
	MaxCPU         int     `json:",omitempty"`
	MaxMem         uint64  `json:",omitempty"`
	Disk           uint64  `json:",omitempty"`
	SSLFingerprint string  `json:"ssl_fingerprint,omitempty"`
	MaxDisk        uint64  `json:",omitempty"`
	Mem            uint64  `json:",omitempty"`
	CPU            float64 `json:",omitempty"`

	// from /cluster endpoint
	NodeID int    `json:",omitempty"` // the internal id of the node
	Name   string `json:",omitempty"`
	IP     string `json:",omitempty"`
	Online int    `json:",omitempty"`
	Local  int    `json:",omitempty"`
}

type Node struct {
	Name       string
	client     *Client
	Kversion   string
	LoadAvg    []string
	CPU        float64
	RootFS     RootFS
	PVEVersion string
	CPUInfo    CPUInfo
	Swap       Memory
	Idle       int
	Memory     Memory
	Ksm        Ksm
	Uptime     uint64
	Wait       float64
}

func (c *Client) Nodes() (ns NodeStatuses, err error) {
	return ns, c.Get("/nodes", &ns)
}

func (c *Client) Node(name string) (*Node, error) {
	var node Node
	if err := c.Get(fmt.Sprintf("/nodes/%s/status", name), &node); err != nil {
		return nil, err
	}
	node.Name = name
	node.client = c

	return &node, nil
}

func (n *Node) Version() (version *Version, err error) {
	return version, n.client.Get("/nodes/%s/version", &version)
}

func (n *Node) TermProxy() (vnc *VNC, err error) {
	return vnc, n.client.Post(fmt.Sprintf("/nodes/%s/termproxy", n.Name), nil, &vnc)
}

// VNCWebSocket send, recv, errors, closer, error
func (n *Node) VNCWebSocket(vnc *VNC) (chan string, chan string, chan error, func() error, error) {
	p := fmt.Sprintf("/nodes/%s/vncwebsocket?port=%d&vncticket=%s",
		n.Name, vnc.Port, url.QueryEscape(vnc.Ticket))

	return n.client.VNCWebSocket(p, vnc)
}

func (n *Node) VirtualMachines() (vms VirtualMachines, err error) {
	if err := n.client.Get(fmt.Sprintf("/nodes/%s/qemu", n.Name), &vms); err != nil {
		return nil, err
	}

	for _, v := range vms {
		v.client = n.client
		v.Node = n.Name
	}

	return vms, nil
}

func (n *Node) NewVirtualMachine(id int, options ...VirtualMachineOption) (*Task, error) {
	var upid string
	data := make(map[string]interface{})
	data["vmid"] = id

	for _, option := range options {
		data[option.Name] = option.Value
	}

	err := n.client.Post(fmt.Sprintf("/nodes/%s/qemu", n.Name), data, &upid)
	return NewTask(upid, n.client), err
}

func (n *Node) VirtualMachine(id int) (*VirtualMachine, error) {
	vm := &VirtualMachine{
		client: n.client,
		Node:   n.Name,
	}

	if err := n.client.Get(fmt.Sprintf("/nodes/%s/qemu/%d/status/current", n.Name, id), &vm); nil != err {
		return nil, err
	}

	//var vmconf VirtualMachineConfig
	if err := n.client.Get(fmt.Sprintf("/nodes/%s/qemu/%d/config", n.Name, id), &vm.VirtualMachineConfig); err != nil {
		return nil, err
	}

	//vm.VirtualMachineConfig = &vmconf

	return vm, nil
}

func (n *Node) LxcContainers() (c LxcContainers, err error) {
	if err := n.client.Get(fmt.Sprintf("/nodes/%s/lxc", n.Name), &c); err != nil {
		return nil, err
	}

	for _, container := range c {
		container.client = n.client
		container.Node = n.Name
	}

	return c, nil
}

func (n *Node) LxcContainer(id int) (*LxcContainer, error) {
	var c LxcContainer
	if err := n.client.Get(fmt.Sprintf("/nodes/%s/lxc/%d/status/current", n.Name, id), &c); err != nil {
		return nil, err
	}
	c.client = n.client
	c.Node = n.Name

	return &c, nil
}

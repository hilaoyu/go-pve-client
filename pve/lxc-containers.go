package pve

import (
	"fmt"
	"net/url"
)

type LxcContainers []*LxcContainer
type LxcContainer struct {
	Name    string
	Node    string
	client  *Client
	CPUs    int
	Status  string
	VMID    StringOrUint64
	Uptime  uint64
	MaxMem  uint64
	MaxDisk uint64
	MaxSwap uint64
}

type LxcContainerStatuses []*LxcContainerStatus
type LxcContainerStatus struct {
	Data string `json:",omitempty"`
}

func (c *LxcContainer) Start() (status string, err error) {
	return status, c.client.Post(fmt.Sprintf("/nodes/%s/lxc/%d/status/start", c.Node, c.VMID), nil, &status)
}

func (c *LxcContainer) Stop() (status *LxcContainerStatus, err error) {
	return status, c.client.Post(fmt.Sprintf("/nodes/%s/lxc/%d/status/stop", c.Node, c.VMID), nil, &status)
}

func (c *LxcContainer) Suspend() (status *LxcContainerStatus, err error) {
	return status, c.client.Post(fmt.Sprintf("/nodes/%s/lxc/%d/status/suspend", c.Node, c.VMID), nil, &status)
}

func (c *LxcContainer) Reboot() (status *LxcContainerStatus, err error) {
	return status, c.client.Post(fmt.Sprintf("/nodes/%s/lxc/%d/status/reboot", c.Node, c.VMID), nil, &status)
}

func (c *LxcContainer) Resume() (status *LxcContainerStatus, err error) {
	return status, c.client.Post(fmt.Sprintf("/nodes/%s/lxc/%d/status/resume", c.Node, c.VMID), nil, &status)
}

func (c *LxcContainer) TermProxy() (vnc *VNC, err error) {
	return vnc, c.client.Post(fmt.Sprintf("/nodes/%s/lxk/%d/termproxy", c.Node, c.VMID), nil, &vnc)
}

func (c *LxcContainer) VNCWebSocket(vnc *VNC) (chan string, chan string, chan error, func() error, error) {
	p := fmt.Sprintf("/nodes/%s/lxc/%d/vncwebsocket?port=%d&vncticket=%s",
		c.Node, c.VMID, vnc.Port, url.QueryEscape(vnc.Ticket))

	return c.client.VNCWebSocket(p, vnc)
}

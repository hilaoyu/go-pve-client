package pve

import (
	"fmt"
)

type NodeNetworks []*NodeNetwork
type NodeNetwork struct {
	client  *Client `json:"-"`
	Node    string  `json:"-"`
	NodeApi *Node   `json:"-"`

	Iface    string `json:"iface,omitempty"`
	BondMode string `json:"bond_mode,omitempty"`

	Autostart int `json:"autostart,omitempty"`

	CIDR            string `json:"cidr,omitempty"`
	CIDR6           string `json:"cidr6,omitempty"`
	Gateway         string `json:"gateway,omitempty"`
	Gateway6        string `json:"gateway6,omitempty"`
	Netmask         string `json:"netmask,omitempty"`
	Netmask6        string `json:"netmask6,omitempty"`
	BridgeVlanAware bool   `json:"bridge_vlan_aware,omitempty"`
	BridgePorts     string `json:"bridge_ports,omitempty"`
	Comments        string `json:"comments,omitempty"`
	Comments6       string `json:"comments6,omitempty"`
	BridgeStp       string `json:"bridge_stp,omitempty"`
	BridgeFd        string `json:"bridge_fd,omitempty"`
	BondPrimary     string `json:"bond-primary,omitempty"`

	Address  string `json:"address,omitempty"`
	Address6 string `json:"address6,omitempty"`
	Type     string `json:"type,omitempty"`
	Active   int    `json:"active,omitempty"`
	Method   string `json:"method,omitempty"`
	Method6  string `json:"method6,omitempty"`
	Priority int    `json:"priority,omitempty"`
}

func (nw *NodeNetwork) Update() (task *Task, err error) {
	var upid string
	if "" == nw.Iface {
		return
	}
	err = nw.client.Put(fmt.Sprintf("/nodes/%s/network/%s", nw.Node, nw.Iface), nw, &upid)
	if err != nil {
		return
	}

	return nw.NodeApi.NetworkReload()
}
func (nw *NodeNetwork) Delete() (task *Task, err error) {
	var upid string
	if "" == nw.Iface {
		return
	}
	err = nw.client.Delete(fmt.Sprintf("/nodes/%s/network/%s", nw.Node, nw.Iface), &upid)
	if err != nil {
		return
	}

	return nw.NodeApi.NetworkReload()
}

func (n *Node) Networks() (networks NodeNetworks, err error) {
	err = n.client.Get(fmt.Sprintf("/nodes/%s/network", n.Name), &networks)
	if err != nil {
		return nil, err
	}

	for _, v := range networks {
		v.client = n.client
		v.Node = n.Name
		v.NodeApi = n
	}

	return
}
func (n *Node) Network(iface string) (network *NodeNetwork, err error) {

	err = n.client.Get(fmt.Sprintf("/nodes/%s/network/%s", n.Name, iface), &network)
	if err != nil {
		return nil, err
	}

	if nil != network {
		network.client = n.client
		network.Node = n.Name
		network.NodeApi = n
		network.Iface = iface
	}

	return network, nil
}

func (n *Node) NewNetwork(network *NodeNetwork) (task *Task, err error) {

	err = n.client.Post(fmt.Sprintf("/nodes/%s/network", n.Name), network, network)
	if nil != err {
		return
	}

	network.client = n.client
	network.Node = n.Name
	network.NodeApi = n
	return n.NetworkReload()
}
func (n *Node) NetworkReload() (*Task, error) {
	var upid string
	err := n.client.Put(fmt.Sprintf("/nodes/%s/network", n.Name), nil, &upid)
	if err != nil {
		return nil, err
	}

	return NewTask(upid, n.client), nil
}

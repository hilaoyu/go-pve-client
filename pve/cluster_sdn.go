package pve

import "fmt"

type SdnZone struct {
	client    *Client
	Zone      string `json:"zone,omitempty"`
	Type      string `json:"type,omitempty"`
	Mtu       int    `json:"mtu,omitempty"`
	Ipam      string `json:"ipam,omitempty"`
	Bridge    string `json:"bridge,omitempty"`
	VXLanPort string `json:"vxlan-port,omitempty"`
}
type SdnVNet struct {
	client    *Client
	Tag       int    `json:"tag,omitempty"`
	Zone      string `json:"zone,omitempty"`
	Type      string `json:"type,omitempty"`
	Vnet      string `json:"vnet,omitempty"`
	Alias     string `json:"alias,omitempty"`
	Vlanaware int    `json:"vlanaware,omitempty"`
}

func (cl *Cluster) SdnApply() (err error) {
	err = cl.client.Put("/cluster/sdn", nil, nil)
	return
}

func (cl *Cluster) SdnZonesGet() (zones []*SdnZone, err error) {
	err = cl.client.Get("/cluster/sdn/zones", &zones)

	for _, zone := range zones {
		zone.client = cl.client
	}
	return
}
func (cl *Cluster) SdnVNetsGet() (vNets []*SdnVNet, err error) {
	err = cl.client.Get("/cluster/sdn/vnets", &vNets)

	for _, vNet := range vNets {
		vNet.client = cl.client
	}
	return
}
func (cl *Cluster) SdnVNetAdd(vNet *SdnVNet) (err error) {
	err = cl.client.Post("/cluster/sdn/vnets", vNet, &vNet)

	if nil != err {
		return
	}
	vNet.client = cl.client
	return
}
func (cl *Cluster) SdnVNetGet(vNet *SdnVNet) (err error) {
	err = cl.client.Get(fmt.Sprintf("/cluster/sdn/vnets/%s", vNet.Vnet), &vNet)

	if nil != err {
		return
	}
	vNet.client = cl.client
	return
}
func (cl *Cluster) SdnVNetUpdate(vNet *SdnVNet) (err error) {
	err = cl.client.Put(fmt.Sprintf("/cluster/sdn/vnets/%s", vNet.Vnet), vNet, &vNet)

	if nil != err {
		return
	}
	vNet.client = cl.client
	return
}
func (cl *Cluster) SdnVNetDelete(vNet *SdnVNet) (err error) {
	err = cl.client.Delete(fmt.Sprintf("/cluster/sdn/vnets/%s", vNet.Vnet), &vNet)

	if nil != err {
		return
	}
	vNet.client = cl.client
	return
}

package pve

import (
	"fmt"
	"github.com/hilaoyu/go-utils/utilStr"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

const (
	StatusVirtualMachineRunning = "running"
	StatusVirtualMachineStopped = "stopped"
	StatusVirtualMachinePaused  = "paused"
)

var (
	vmConfigRegexpIDE    *regexp.Regexp
	vmConfigRegexpSCSI   *regexp.Regexp
	vmConfigRegexpSATA   *regexp.Regexp
	vmConfigRegexpNet    *regexp.Regexp
	vmConfigRegexpUnused *regexp.Regexp
)

func init() {
	vmConfigRegexpIDE, _ = regexp.Compile("^IDE[\\d]+$")
	vmConfigRegexpSCSI, _ = regexp.Compile("^SCSI[\\d]+$")
	vmConfigRegexpSATA, _ = regexp.Compile("^SATAIDE[\\d]+$")
	vmConfigRegexpNet, _ = regexp.Compile("^Net[\\d]+$")
	vmConfigRegexpUnused, _ = regexp.Compile("^Unused[\\d]+$")
}

type IsTemplate bool

func (it *IsTemplate) UnmarshalJSON(b []byte) error {
	*it = true
	if string(b) == "\"\"" {
		*it = false
	}

	return nil
}

type HA struct {
	Managed int
}

type VirtualMachines []*VirtualMachine
type VirtualMachine struct {
	client               *Client
	VirtualMachineConfig *VirtualMachineConfig

	Name      string
	Node      string
	NetIn     uint64
	CPUs      int
	DiskWrite uint64
	Status    string
	Lock      string `json:",omitempty"`
	VMID      StringOrUint64
	PID       StringOrUint64
	Netout    uint64
	Disk      uint64
	Uptime    uint64
	Mem       uint64
	CPU       float64
	MaxMem    uint64
	MaxDisk   uint64
	DiskRead  uint64
	QMPStatus string     `json:"qmpstatus,omitempty"`
	Template  IsTemplate // empty str if a vm, int 1 if a template
	HA        HA         `json:",omitempty"`
}

type VirtualMachineConfig struct {
	Cores   int
	Numa    int
	Memory  StringOrUint64
	Sockets int
	IDE2    string
	OSType  string
	SMBios1 string
	SCSIHW  string

	Digest  string
	Meta    string
	SCSI0   string
	Boot    string
	VMGenID string
	Name    string

	IDEs map[string]string
	IDE0 string
	IDE1 string
	IDE3 string
	IDE4 string
	IDE5 string
	IDE6 string
	IDE7 string
	IDE8 string
	IDE9 string

	SCSIs map[string]string
	SCSI1 string
	SCSI2 string
	SCSI3 string
	SCSI4 string
	SCSI5 string
	SCSI6 string
	SCSI7 string
	SCSI8 string
	SCSI9 string

	SATAs map[string]string
	SATA0 string
	SATA1 string
	SATA2 string
	SATA3 string
	SATA4 string
	SATA5 string
	SATA6 string
	SATA7 string
	SATA8 string
	SATA9 string

	Nets map[string]*VirtualMachineNetwork
	Net0 *VirtualMachineNetwork
	Net1 *VirtualMachineNetwork
	Net2 *VirtualMachineNetwork
	Net3 *VirtualMachineNetwork
	Net4 *VirtualMachineNetwork
	Net5 *VirtualMachineNetwork
	Net6 *VirtualMachineNetwork
	Net7 *VirtualMachineNetwork
	Net8 *VirtualMachineNetwork
	Net9 *VirtualMachineNetwork

	Unuseds map[string]string
	Unused0 string
	Unused1 string
	Unused2 string
	Unused3 string
	Unused4 string
	Unused5 string
	Unused6 string
	Unused7 string
	Unused8 string
	Unused9 string
}

type VirtualMachineOptions []*VirtualMachineOption
type VirtualMachineOption struct {
	Name  string
	Value interface{}
}

type VirtualMachineNetwork struct {
	Name     string `json:"name,omitempty"`
	Type     string `json:"type,omitempty"`
	Mac      string `json:"mac,omitempty"`
	Bridge   string `json:"bridge,omitempty"`
	Firewall int    `json:"firewall,omitempty"`
	LinkDown int    `json:"link_down,omitempty"`
	Mtu      int    `json:"mtu,omitempty"`
	Queues   int    `json:"queues,omitempty"`
	Rate     int    `json:"rate,omitempty"`
}

func (vmn *VirtualMachineNetwork) UnmarshalJSON(b []byte) error {

	netConfig := string(b)
	netConfig = strings.Trim(netConfig, "\"")
	if "" == netConfig {
		return nil
	}
	regMacAddress, _ := regexp.Compile("([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})")
	macOrg := regMacAddress.FindString(netConfig)
	mac := strings.ToUpper(strings.ReplaceAll(macOrg, "-", ":"))

	vmn.Mac = mac
	items := strings.Split(netConfig, ",")
	for _, item := range items {

		switch {
		case strings.HasSuffix(item, "="+macOrg):
			vmn.Type = utilStr.Before(item, "="+macOrg)
			break
		case strings.HasPrefix(item, "bridge="):
			vmn.Bridge = utilStr.After(item, "bridge=")
			break

		case strings.HasPrefix(item, "firewall="):
			firewall, _ := strconv.Atoi(utilStr.After(item, "firewall="))
			vmn.Firewall = firewall
			break
		case strings.HasPrefix(item, "link_down="):
			linkDown, _ := strconv.Atoi(utilStr.After(item, "link_down="))
			vmn.LinkDown = linkDown
			break
		case strings.HasPrefix(item, "mtu="):
			mtu, _ := strconv.Atoi(utilStr.After(item, "mtu="))
			vmn.LinkDown = mtu
			break
		case strings.HasPrefix(item, "queues="):
			queues, _ := strconv.Atoi(utilStr.After(item, "queues="))
			vmn.Queues = queues
			break
		case strings.HasPrefix(item, "rate="):
			rate, _ := strconv.Atoi(utilStr.After(item, "rate="))
			vmn.Rate = rate
			break

		default:
			break
		}
	}
	return nil
}

func (vmc *VirtualMachineConfig) MergeIDEs() map[string]string {
	if nil == vmc.IDEs {
		vmc.IDEs = map[string]string{}
		t := reflect.TypeOf(*vmc)
		v := reflect.ValueOf(*vmc)
		count := v.NumField()

		for i := 0; i < count; i++ {
			fn := t.Field(i).Name
			fv := v.Field(i).String()
			//fmt.Println(fn, fv)
			if "" == fv {
				continue
			}
			if vmConfigRegexpIDE.MatchString(fn) {
				vmc.IDEs[strings.ToLower(fn)] = fv
			}
		}
	}
	return vmc.IDEs
}
func (vmc *VirtualMachineConfig) MergeSCSIs() map[string]string {
	if nil == vmc.SCSIs {
		vmc.SCSIs = map[string]string{}
		t := reflect.TypeOf(*vmc)
		v := reflect.ValueOf(*vmc)
		count := v.NumField()

		for i := 0; i < count; i++ {
			fn := t.Field(i).Name
			fv := v.Field(i).String()
			//fmt.Println(fn, fv)
			if "" == fv {
				continue
			}
			if vmConfigRegexpSCSI.MatchString(fn) {
				vmc.SCSIs[strings.ToLower(fn)] = fv
			}
		}
	}
	return vmc.SCSIs
}

func (vmc *VirtualMachineConfig) MergeSATAs() map[string]string {
	if nil == vmc.SATAs {
		vmc.SATAs = map[string]string{}
		t := reflect.TypeOf(*vmc)
		v := reflect.ValueOf(*vmc)
		count := v.NumField()

		for i := 0; i < count; i++ {
			fn := t.Field(i).Name
			fv := v.Field(i).String()
			//fmt.Println(fn, fv)
			if "" == fv {
				continue
			}
			if vmConfigRegexpSATA.MatchString(fn) {
				vmc.SATAs[strings.ToLower(fn)] = fv
			}
		}
	}
	return vmc.SATAs
}
func (vmc *VirtualMachineConfig) MergeNets() map[string]*VirtualMachineNetwork {
	if nil == vmc.Nets {
		vmc.Nets = map[string]*VirtualMachineNetwork{}
		t := reflect.TypeOf(*vmc)
		v := reflect.ValueOf(*vmc)
		count := v.NumField()

		for i := 0; i < count; i++ {
			fn := t.Field(i).Name

			if vmConfigRegexpNet.MatchString(fn) {
				fv, _ := v.Field(i).Interface().(*VirtualMachineNetwork)
				//fmt.Println(fn, fv)
				if nil == fv {
					continue
				}
				fv.Name = strings.ToLower(fn)
				vmc.Nets[strings.ToLower(fn)] = fv
			}
		}
	}
	return vmc.Nets
}
func (vmc *VirtualMachineConfig) MergeUnuseds() map[string]string {
	if nil == vmc.Unuseds {
		vmc.Unuseds = map[string]string{}
		t := reflect.TypeOf(*vmc)
		v := reflect.ValueOf(*vmc)
		count := v.NumField()

		for i := 0; i < count; i++ {
			fn := t.Field(i).Name
			fv := v.Field(i).String()
			//fmt.Println(fn, fv)
			if "" == fv {
				continue
			}
			if vmConfigRegexpUnused.MatchString(fn) {
				vmc.Unuseds[strings.ToLower(fn)] = fv
			}
		}
	}
	return vmc.Unuseds
}

func (v *VirtualMachine) Ping() error {
	return v.client.Get(fmt.Sprintf("/nodes/%s/qemu/%d/status/current", v.Node, v.VMID), &v)
}

func (v *VirtualMachine) Config(options ...VirtualMachineOption) (*Task, error) {
	var upid string
	data := make(map[string]interface{})
	for _, opt := range options {
		data[opt.Name] = opt.Value
	}
	err := v.client.Post(fmt.Sprintf("/nodes/%s/qemu/%d/config", v.Node, v.VMID), data, &upid)
	return NewTask(upid, v.client), err
}
func (v *VirtualMachine) ConfigLoad() (err error) {
	if nil == v.VirtualMachineConfig {
		err = v.client.Get(fmt.Sprintf("/nodes/%s/qemu/%d/config", v.Node, v.VMID), &v.VirtualMachineConfig)
	}

	return
}

func (v *VirtualMachine) TermProxy() (vnc *VNC, err error) {
	return vnc, v.client.Post(fmt.Sprintf("/nodes/%s/qemu/%d/termproxy", v.Node, v.VMID), nil, &vnc)
}

func (v *VirtualMachine) TermProxyWebsocketServeHTTP(w http.ResponseWriter, r *http.Request, responseHeader http.Header) (err error) {
	vnc, err := v.TermProxy()
	if nil != err {
		return
	}

	path := fmt.Sprintf("/nodes/%s/qemu/%d/vncwebsocket?port=%d&vncticket=%s",
		v.Node, v.VMID, vnc.Port, url.QueryEscape(vnc.Ticket))

	return v.client.TermProxyWebsocketServeHTTP(path, vnc, w, r, responseHeader)
}

func (v *VirtualMachine) VncProxy() (vnc *VNC, err error) {
	return vnc, v.client.Post(fmt.Sprintf("/nodes/%s/qemu/%d/vncproxy", v.Node, v.VMID), map[string]interface{}{"websocket": true, "generate-password": false}, &vnc)
}

func (v *VirtualMachine) VNCProxyWebsocketServeHTTP(w http.ResponseWriter, r *http.Request, responseHeader http.Header) (err error) {
	vnc, err := v.VncProxy()
	if nil != err {
		return
	}

	path := fmt.Sprintf("/nodes/%s/qemu/%d/vncwebsocket?port=%d&vncticket=%s",
		v.Node, v.VMID, vnc.Port, url.QueryEscape(vnc.Ticket))

	return v.client.VNCProxyWebsocketServeHTTP(path, vnc, w, r, responseHeader)
}

// VNCWebSocket copy/paste when calling to get the channel names right
// send, recv, errors, closer, errors := vm.VNCWebSocket(vnc)
// for this to work you need to first setup a serial terminal on your vm https://pve.proxmox.com/wiki/Serial_Terminal
func (v *VirtualMachine) VNCWebSocket(vnc *VNC) (chan string, chan string, chan error, func() error, error) {
	p := fmt.Sprintf("/nodes/%s/qemu/%d/vncwebsocket?port=%d&vncticket=%s",
		v.Node, v.VMID, vnc.Port, url.QueryEscape(vnc.Ticket))

	return v.client.VNCWebSocket(p, vnc)
}

func (v *VirtualMachine) IsRunning() bool {
	status := v.Status == StatusVirtualMachineRunning
	if "" != v.QMPStatus {
		status = status && v.QMPStatus == StatusVirtualMachineRunning
	}
	return status
}

func (v *VirtualMachine) Start() (*Task, error) {
	var upid string
	if err := v.client.Post(fmt.Sprintf("/nodes/%s/qemu/%d/status/start", v.Node, v.VMID), nil, &upid); err != nil {
		return nil, err
	}

	return NewTask(upid, v.client), nil
}

func (v *VirtualMachine) IsStopped() bool {
	status := v.Status == StatusVirtualMachineStopped
	if "" != v.QMPStatus {
		status = status && v.QMPStatus == StatusVirtualMachineStopped
	}
	return status
}

func (v *VirtualMachine) Reset() (task *Task, err error) {
	var upid string
	if err := v.client.Post(fmt.Sprintf("/nodes/%s/qemu/%d/status/reset", v.Node, v.VMID), nil, &upid); err != nil {
		return nil, err
	}

	return NewTask(upid, v.client), nil
}

func (v *VirtualMachine) Shutdown() (task *Task, err error) {
	var upid string
	if err := v.client.Post(fmt.Sprintf("/nodes/%s/qemu/%d/status/shutdown", v.Node, v.VMID), nil, &upid); err != nil {
		return nil, err
	}

	return NewTask(upid, v.client), nil
}

func (v *VirtualMachine) Stop() (task *Task, err error) {
	var upid string
	if err := v.client.Post(fmt.Sprintf("/nodes/%s/qemu/%d/status/stop", v.Node, v.VMID), nil, &upid); err != nil {
		return nil, err
	}

	return NewTask(upid, v.client), nil
}

func (v *VirtualMachine) IsPaused() bool {
	status := v.Status == StatusVirtualMachineRunning
	if "" != v.QMPStatus {
		status = status && v.QMPStatus == StatusVirtualMachinePaused
	}
	return status
}

func (v *VirtualMachine) Pause() (task *Task, err error) {
	var upid string
	if err := v.client.Post(fmt.Sprintf("/nodes/%s/qemu/%d/status/suspend", v.Node, v.VMID), nil, &upid); err != nil {
		return nil, err
	}

	return NewTask(upid, v.client), nil
}

func (v *VirtualMachine) IsHibernated() bool {
	return v.Status == StatusVirtualMachineStopped && v.QMPStatus == StatusVirtualMachineStopped && v.Lock == "suspended"
}

func (v *VirtualMachine) Hibernate() (task *Task, err error) {
	var upid string
	if err := v.client.Post(fmt.Sprintf("/nodes/%s/qemu/%d/status/suspend", v.Node, v.VMID), map[string]string{"todisk": "1"}, &upid); err != nil {
		return nil, err
	}

	return NewTask(upid, v.client), nil
}

func (v *VirtualMachine) Resume() (task *Task, err error) {
	var upid string
	if err := v.client.Post(fmt.Sprintf("/nodes/%s/qemu/%d/status/resume", v.Node, v.VMID), nil, &upid); err != nil {
		return nil, err
	}

	return NewTask(upid, v.client), nil
}

func (v *VirtualMachine) Reboot() (task *Task, err error) {
	var upid string
	if err := v.client.Post(fmt.Sprintf("/nodes/%s/qemu/%d/status/reboot", v.Node, v.VMID), nil, &upid); err != nil {
		return nil, err
	}

	return NewTask(upid, v.client), nil
}

func (v *VirtualMachine) Delete() (task *Task, err error) {
	var upid string
	if err := v.client.Delete(fmt.Sprintf("/nodes/%s/qemu/%d", v.Node, v.VMID), &upid); err != nil {
		return nil, err
	}

	return NewTask(upid, v.client), nil
}

func (v *VirtualMachine) Clone(name, target string) (newid int, task *Task, err error) {
	var upid string
	cluster, err := v.client.Cluster()
	if err != nil {
		return newid, nil, err
	}

	newid, err = cluster.NextID()
	if err != nil {
		return newid, nil, err
	}

	if err := v.client.Post(fmt.Sprintf("/nodes/%s/qemu/%d/clone", v.Node, v.VMID), map[string]string{
		"newid":  strconv.Itoa(newid),
		"name":   name,
		"target": target,
	}, &upid); err != nil {
		return newid, nil, err
	}

	return newid, NewTask(upid, v.client), nil
}
func (v *VirtualMachine) MoveDisk(disk, storage string) (task *Task, err error) {
	var upid string

	err = v.client.Post(fmt.Sprintf("/nodes/%s/qemu/%d/move_disk", v.Node, v.VMID), map[string]string{
		"disk":    disk,
		"storage": storage,
	}, &upid)
	if err != nil {
		return
	}

	return NewTask(upid, v.client), nil
}

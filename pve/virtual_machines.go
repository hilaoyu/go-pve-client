package pve

import (
	"fmt"
	"github.com/hilaoyu/go-utils/utilFile"
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
	vmConfigRegexpIDE, _ = regexp.Compile("(?i:^IDE)[\\d]+$")
	vmConfigRegexpSCSI, _ = regexp.Compile("(?i:^SCSI)[\\d]+$")
	vmConfigRegexpSATA, _ = regexp.Compile("(?i:^SATA)[\\d]+$")
	vmConfigRegexpNet, _ = regexp.Compile("(?i:^Net)[\\d]+$")
	vmConfigRegexpUnused, _ = regexp.Compile("(?i:^Unused)[\\d]+$")
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
	OSType  string
	SMBios1 string
	SCSIHW  string

	Digest  string
	Meta    string
	Boot    string
	VMGenID string
	Name    string

	IDEs map[string]*VirtualMachineDisk
	IDE0 *VirtualMachineDisk
	IDE1 *VirtualMachineDisk
	IDE2 *VirtualMachineDisk
	IDE3 *VirtualMachineDisk
	IDE4 *VirtualMachineDisk
	IDE5 *VirtualMachineDisk
	IDE6 *VirtualMachineDisk
	IDE7 *VirtualMachineDisk
	IDE8 *VirtualMachineDisk
	IDE9 *VirtualMachineDisk

	SCSIs map[string]*VirtualMachineDisk
	SCSI0 *VirtualMachineDisk
	SCSI1 *VirtualMachineDisk
	SCSI2 *VirtualMachineDisk
	SCSI3 *VirtualMachineDisk
	SCSI4 *VirtualMachineDisk
	SCSI5 *VirtualMachineDisk
	SCSI6 *VirtualMachineDisk
	SCSI7 *VirtualMachineDisk
	SCSI8 *VirtualMachineDisk
	SCSI9 *VirtualMachineDisk

	SATAs map[string]*VirtualMachineDisk
	SATA0 *VirtualMachineDisk
	SATA1 *VirtualMachineDisk
	SATA2 *VirtualMachineDisk
	SATA3 *VirtualMachineDisk
	SATA4 *VirtualMachineDisk
	SATA5 *VirtualMachineDisk
	SATA6 *VirtualMachineDisk
	SATA7 *VirtualMachineDisk
	SATA8 *VirtualMachineDisk
	SATA9 *VirtualMachineDisk

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

	Unuseds map[string]*VirtualMachineDisk
	Unused0 *VirtualMachineDisk
	Unused1 *VirtualMachineDisk
	Unused2 *VirtualMachineDisk
	Unused3 *VirtualMachineDisk
	Unused4 *VirtualMachineDisk
	Unused5 *VirtualMachineDisk
	Unused6 *VirtualMachineDisk
	Unused7 *VirtualMachineDisk
	Unused8 *VirtualMachineDisk
	Unused9 *VirtualMachineDisk
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
	Tag      int    `json:"tag,omitempty"`
	Trunks   string `json:"trunks,omitempty"`
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
		case strings.HasPrefix(item, "tag="):
			tag, _ := strconv.Atoi(utilStr.After(item, "rate="))
			vmn.Tag = tag
			break

		case strings.HasPrefix(item, "trunks="):
			vmn.Trunks = utilStr.After(item, "trunks=")
			break

		default:
			break
		}
	}
	return nil
}

func (vmc *VirtualMachineConfig) MergeIDEs() map[string]*VirtualMachineDisk {
	if nil == vmc.IDEs {
		vmc.IDEs = map[string]*VirtualMachineDisk{}
		t := reflect.TypeOf(*vmc)
		v := reflect.ValueOf(*vmc)
		count := v.NumField()

		for i := 0; i < count; i++ {
			fn := t.Field(i).Name

			if vmConfigRegexpIDE.MatchString(fn) {
				fv, _ := v.Field(i).Interface().(*VirtualMachineDisk)
				//fmt.Println(fn, fv)
				if nil == fv {
					continue
				}
				fv.Name = strings.ToLower(fn)
				vmc.IDEs[strings.ToLower(fn)] = fv
			}
		}
	}
	return vmc.IDEs
}
func (vmc *VirtualMachineConfig) MergeSCSIs() map[string]*VirtualMachineDisk {
	if nil == vmc.SCSIs {
		vmc.SCSIs = map[string]*VirtualMachineDisk{}
		t := reflect.TypeOf(*vmc)
		v := reflect.ValueOf(*vmc)
		count := v.NumField()

		for i := 0; i < count; i++ {
			fn := t.Field(i).Name

			if vmConfigRegexpSCSI.MatchString(fn) {
				fv, _ := v.Field(i).Interface().(*VirtualMachineDisk)
				//fmt.Println(fn, fv)
				if nil == fv {
					continue
				}
				fv.Name = strings.ToLower(fn)
				vmc.SCSIs[strings.ToLower(fn)] = fv
			}
		}
	}
	return vmc.SCSIs
}

func (vmc *VirtualMachineConfig) MergeSATAs() map[string]*VirtualMachineDisk {
	if nil == vmc.SATAs {
		vmc.SATAs = map[string]*VirtualMachineDisk{}
		t := reflect.TypeOf(*vmc)
		v := reflect.ValueOf(*vmc)
		count := v.NumField()

		for i := 0; i < count; i++ {
			fn := t.Field(i).Name

			if vmConfigRegexpSATA.MatchString(fn) {
				fv, _ := v.Field(i).Interface().(*VirtualMachineDisk)
				//fmt.Println(fn, fv)
				if nil == fv {
					continue
				}
				fv.Name = strings.ToLower(fn)
				vmc.SATAs[strings.ToLower(fn)] = fv
			}
		}
	}
	return vmc.SATAs
}
func (vmc *VirtualMachineConfig) MergeUnuseds() map[string]*VirtualMachineDisk {
	if nil == vmc.Unuseds {
		vmc.Unuseds = map[string]*VirtualMachineDisk{}
		t := reflect.TypeOf(*vmc)
		v := reflect.ValueOf(*vmc)
		count := v.NumField()

		for i := 0; i < count; i++ {
			fn := t.Field(i).Name

			if vmConfigRegexpUnused.MatchString(fn) {
				fv, _ := v.Field(i).Interface().(*VirtualMachineDisk)
				//fmt.Println(fn, fv)
				if nil == fv {
					continue
				}
				fv.Name = strings.ToLower(fn)
				vmc.Unuseds[strings.ToLower(fn)] = fv
			}
		}
	}
	return vmc.Unuseds
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

type VirtualMachineDisk struct {
	Name      string `json:"name,omitempty"`
	Storage   string `json:"storage,omitempty"`
	File      string `json:"file,omitempty"`
	Aio       string `json:"aio,omitempty"`
	Cache     string `json:"cache,omitempty"`
	Discard   string `json:"discard,omitempty"`
	IoThread  int    `json:"iothread,omitempty"`
	Replicate int    `json:"replicate,omitempty"`
	Ro        int    `json:"ro,omitempty"`
	Backup    int    `json:"backup,omitempty"`
	SizeGb    int64  `json:"size_gb,omitempty"`
	Ssd       int    `json:"ssd,omitempty"`
	Media     string `json:"media,omitempty"`
	Format    string `json:"format,omitempty"`
}

func (vmd *VirtualMachineDisk) UnmarshalJSON(b []byte) error {

	conf := string(b)
	conf = strings.Trim(conf, "\"")
	if "" == conf {
		return nil
	}
	regDiskFile := regexp.MustCompile(`([0-9A-Za-z_-]+):([0-9A-Za-z_\.\\/-]+)`)
	ret := regDiskFile.FindStringSubmatch(conf)
	if len(ret) >= 3 {
		vmd.Storage = ret[1]
		vmd.File = ret[2]
	}

	items := strings.Split(conf, ",")
	for _, item := range items {

		switch {
		case strings.HasPrefix(item, "aio="):
			vmd.Aio = utilStr.After(item, "aio=")
			break
		case strings.HasPrefix(item, "cache="):
			vmd.Cache = utilStr.After(item, "cache=")
			break
		case strings.HasPrefix(item, "discard="):
			vmd.Discard = utilStr.After(item, "discard=")
			break
		case strings.HasPrefix(item, "media="):
			vmd.Media = utilStr.After(item, "media=")
			break
		case strings.HasPrefix(item, "format="):
			vmd.Format = utilStr.After(item, "format=")
			break

		case strings.HasPrefix(item, "iothread="):
			thread, _ := strconv.Atoi(utilStr.After(item, "iothread="))
			vmd.IoThread = thread
			break
		case strings.HasPrefix(item, "replicate="):
			replicate, _ := strconv.Atoi(utilStr.After(item, "replicate="))
			vmd.Replicate = replicate
			break
		case strings.HasPrefix(item, "ro="):
			ro, _ := strconv.Atoi(utilStr.After(item, "ro="))
			vmd.Ro = ro
			break
		case strings.HasPrefix(item, "backup="):
			backup, _ := strconv.Atoi(utilStr.After(item, "backup="))
			vmd.Backup = backup
			break
		case strings.HasPrefix(item, "ssd="):
			ssd, _ := strconv.Atoi(utilStr.After(item, "ssd="))
			vmd.Ssd = ssd
			break
		case strings.HasPrefix(item, "size="):
			size, _ := utilFile.SizeStringToNumber(utilStr.After(item, "size="), "g")
			vmd.SizeGb = size
			break

		default:
			break
		}
	}
	return nil
}
func (vmd *VirtualMachineDisk) SourcePath() string {
	return fmt.Sprintf("%s:%s", vmd.Storage, vmd.File)
}
func (vmd *VirtualMachineDisk) ToConfigOptions() (options []string) {
	if "" != vmd.Aio {
		options = append(options, fmt.Sprintf("aio=%s", vmd.Aio))
	}
	if "" != vmd.Cache {
		options = append(options, fmt.Sprintf("cache=%s", vmd.Cache))
	}
	if "" != vmd.Discard {
		options = append(options, fmt.Sprintf("discard=%s", vmd.Discard))
	}
	if "" != vmd.Media {
		options = append(options, fmt.Sprintf("media=%s", vmd.Media))
	}
	if vmd.IoThread > 0 {
		options = append(options, fmt.Sprintf("iothread=%d", vmd.IoThread))
	}
	if vmd.Replicate > 0 {
		options = append(options, fmt.Sprintf("replicate=%d", vmd.Replicate))
	}
	if vmd.Ro > 0 {
		options = append(options, fmt.Sprintf("ro=%d", vmd.Ro))
	}
	if vmd.Backup > 0 {
		options = append(options, fmt.Sprintf("backup=%d", vmd.Backup))
	}
	if vmd.Ssd > 0 {
		options = append(options, fmt.Sprintf("ssd=%d", vmd.Ssd))
	}

	return
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
func (v *VirtualMachine) ConfigLoad(force ...bool) (err error) {
	if nil == v.VirtualMachineConfig || (len(force) > 0 && force[0]) {
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
func (v *VirtualMachine) MoveDisk(diskName, storage string, format ...string) (task *Task, err error) {
	var upid string

	deskFormat := "qcow2"
	if len(format) > 0 && "" != format[0] {
		deskFormat = format[0]
	}

	err = v.client.Post(fmt.Sprintf("/nodes/%s/qemu/%d/move_disk", v.Node, v.VMID), map[string]interface{}{
		"disk":    diskName,
		"storage": storage,
		"format":  deskFormat,
	}, &upid)
	if err != nil {
		return
	}

	return NewTask(upid, v.client), nil
}
func (v *VirtualMachine) Resize(diskName string, sizeGb int64) (task *Task, err error) {
	var upid string

	err = v.client.Put(fmt.Sprintf("/nodes/%s/qemu/%d/resize", v.Node, v.VMID), map[string]interface{}{
		"disk": diskName,
		"size": fmt.Sprintf("%dG", sizeGb),
	}, &upid)
	if err != nil {
		return
	}

	return NewTask(upid, v.client), nil
}
func (v *VirtualMachine) GetDisk(diskName string) (disk *VirtualMachineDisk, err error) {
	err = v.ConfigLoad(true)
	if nil != err {
		return
	}
	switch {
	case vmConfigRegexpSCSI.MatchString(diskName):
		disks := v.VirtualMachineConfig.MergeSCSIs()
		if tmp, ok := disks[diskName]; ok {
			disk = tmp
		}
		break
	case vmConfigRegexpSATA.MatchString(diskName):
		disks := v.VirtualMachineConfig.MergeSATAs()
		if tmp, ok := disks[diskName]; ok {
			disk = tmp
		}
		break
	case vmConfigRegexpIDE.MatchString(diskName):
		disks := v.VirtualMachineConfig.MergeIDEs()
		if tmp, ok := disks[diskName]; ok {
			disk = tmp
		}
		break
	case vmConfigRegexpUnused.MatchString(diskName):
		disks := v.VirtualMachineConfig.MergeUnuseds()
		if tmp, ok := disks[diskName]; ok {
			disk = tmp
		}
		break

	}

	return
}
func (v *VirtualMachine) ChangeDisk(diskName string, disk *VirtualMachineDisk, storage string) (err error) {
	if nil == disk {
		err = fmt.Errorf("disk can not be nil")
		return
	}
	needStart := false
	var task *Task
	if !v.IsStopped() {
		needStart = true
		task, err = v.Stop()
		if nil != err {
			return
		}
		err = task.WaitForComplete(36, 5)
		if nil != err {
			err = fmt.Errorf("vm stop faild: %v", err)
			return
		}
	}

	diskSizeGb := disk.SizeGb

	existDisk, err := v.GetDisk(diskName)
	if nil != err {
		return
	}
	if nil != existDisk {
		if "" == storage {
			storage = existDisk.Storage
		}
		if existDisk.SizeGb > diskSizeGb {
			diskSizeGb = existDisk.SizeGb
		}
	}
	if "" == storage {
		err = fmt.Errorf("storage is empty")
		return
	}

	options := disk.ToConfigOptions()
	options = append(options, fmt.Sprintf("%s:0", storage), fmt.Sprintf("import-from=%s", disk.SourcePath()))
	task, err = v.Config(VirtualMachineOption{
		Name:  diskName,
		Value: strings.Join(options, ","),
	})
	if nil != err {
		return
	}
	err = task.WaitForComplete(10, 3)
	if nil != err {
		err = fmt.Errorf("vm config disk faild: %v", err)
		return
	}

	newDisk, err := v.GetDisk(diskName)
	if nil != err {
		return
	}
	if diskSizeGb > newDisk.SizeGb {
		task, err = v.Resize(diskName, diskSizeGb)
		err = task.WaitForComplete(10, 3)
		if nil != err {
			err = fmt.Errorf("vm disk %s resize faild: %v", diskName, err)
			return
		}
	}

	if needStart {
		task, err = v.Start()
		err = task.WaitForComplete(36, 5)
		if nil != err {
			err = fmt.Errorf("vm start faild: %v", err)
			return
		}
	}

	return
}

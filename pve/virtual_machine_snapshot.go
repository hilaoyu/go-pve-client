package pve

import "fmt"

type Snapshot struct {
	Name        string
	Vmstate     int
	Description string
	Snaptime    int64
	Parent      string
	Snapstate   string
}

func (v *VirtualMachine) NewSnapshot(name string) (task *Task, err error) {
	var upid string
	if err = v.client.Post(fmt.Sprintf("/nodes/%s/qemu/%d/snapshot", v.Node, v.VMID), map[string]string{"snapname": name}, &upid); err != nil {
		return nil, err
	}

	return NewTask(upid, v.client), nil
}
func (v *VirtualMachine) Snapshots() (snapshots []*Snapshot, err error) {
	err = v.client.Get(fmt.Sprintf("/nodes/%s/qemu/%d/snapshot", v.Node, v.VMID), &snapshots)
	return
}

func (v *VirtualMachine) SnapshotRollback(name string) (task *Task, err error) {
	var upid string
	if err = v.client.Post(fmt.Sprintf("/nodes/%s/qemu/%d/snapshot/%s/rollback", v.Node, v.VMID, name), nil, &upid); err != nil {
		return nil, err
	}

	return NewTask(upid, v.client), nil
}
func (v *VirtualMachine) SnapshotDelete(name string) (task *Task, err error) {
	var upid string
	if err = v.client.Delete(fmt.Sprintf("/nodes/%s/qemu/%d/snapshot/%s", v.Node, v.VMID, name), &upid); err != nil {
		return nil, err
	}

	return NewTask(upid, v.client), nil
}

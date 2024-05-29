package pve

import "fmt"

type FirewallVirtualMachineOption struct {
	Enable      bool   `json:"enable,omitempty"`
	Dhcp        bool   `json:"dhcp,omitempty"`
	Ipfilter    bool   `json:"ipfilter,omitempty"`
	LogLevelIn  string `json:"log_level_in,omitempty"`
	LogLevelOut string `json:"log_level_out,omitempty"`
	Macfilter   bool   `json:"macfilter,omitempty"`
	Ndp         bool   `json:"ndp,omitempty"`
	PolicyIn    string `json:"policy_in,omitempty"`
	PolicyOut   string `json:"policy_out,omitempty"`
	Radv        bool   `json:"radv,omitempty"`
}

func (v *VirtualMachine) FirewallOptionGet() (firewallOption *FirewallVirtualMachineOption, err error) {
	err = v.client.Get(fmt.Sprintf("/nodes/%s/qemu/%d/firewall/options", v.Node, v.VMID), firewallOption)
	return
}
func (v *VirtualMachine) FirewallOptionSet(firewallOption *FirewallVirtualMachineOption) (err error) {
	err = v.client.Put(fmt.Sprintf("/nodes/%s/qemu/%d/firewall/options", v.Node, v.VMID), firewallOption, nil)
	return
}

func (v *VirtualMachine) FirewallGetRules() (rules []*FirewallRule, err error) {
	err = v.client.Get(fmt.Sprintf("/nodes/%s/qemu/%d/firewall/rules", v.Node, v.VMID), &rules)
	return
}

func (v *VirtualMachine) FirewallRulesCreate(rule *FirewallRule) (err error) {
	err = v.client.Post(fmt.Sprintf("/nodes/%s/qemu/%d/firewall/rules", v.Node, v.VMID), rule, nil)
	return
}
func (v *VirtualMachine) FirewallRulesUpdate(rule *FirewallRule) (err error) {
	err = v.client.Put(fmt.Sprintf("/nodes/%s/qemu/%d/firewall/rules/%d", v.Node, v.VMID, rule.Pos), rule, nil)
	return
}
func (v *VirtualMachine) FirewallRulesDelete(rulePos int) (err error) {
	err = v.client.Delete(fmt.Sprintf("/nodes/%s/qemu/%d/firewall/rules/%d", v.Node, v.VMID, rulePos), nil)
	return
}

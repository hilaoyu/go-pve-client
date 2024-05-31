package pve

import (
	"fmt"
)

type FirewallSecurityGroup struct {
	client  *Client
	Group   string          `json:"group,omitempty"`
	Comment string          `json:"comment,omitempty"`
	Rules   []*FirewallRule `json:"rules,omitempty"`
}

func (cl *Cluster) FWGroups() (groups []*FirewallSecurityGroup, err error) {
	err = cl.client.Get("/cluster/firewall/groups", &groups)

	if nil == err {
		for _, g := range groups {
			g.client = cl.client
		}
	}
	return
}

func (cl *Cluster) FWGroup(name string) (group *FirewallSecurityGroup, err error) {
	group = &FirewallSecurityGroup{}
	err = cl.client.Get(fmt.Sprintf("/cluster/firewall/groups/%s", name), &group.Rules)
	if nil == err {
		group.Group = name
		group.client = cl.client
	}
	return
}

func (cl *Cluster) NewFWGroup(group *FirewallSecurityGroup) (err error) {
	err = cl.client.Post(fmt.Sprintf("/cluster/firewall/groups"), group, &group)
	return
}

func (g *FirewallSecurityGroup) GetRules() (rules []*FirewallRule, err error) {
	err = g.client.Get(fmt.Sprintf("/cluster/firewall/groups/%s", g.Group), &g.Rules)
	rules = g.Rules
	return
}
func (g *FirewallSecurityGroup) Delete() (err error) {
	err = g.client.Delete(fmt.Sprintf("/cluster/firewall/groups/%s", g.Group), nil)
	return
}
func (g *FirewallSecurityGroup) RuleCreate(rule *FirewallRule) (err error) {
	err = g.client.Post(fmt.Sprintf("/cluster/firewall/groups/%s", g.Group), rule, nil)
	return
}
func (g *FirewallSecurityGroup) RuleUpdate(rule *FirewallRule) (err error) {
	err = g.client.Put(fmt.Sprintf("/cluster/firewall/groups/%s/%d", g.Group, rule.Pos), rule, nil)
	return
}
func (g *FirewallSecurityGroup) RuleDelete(rulePos int) (err error) {
	err = g.client.Delete(fmt.Sprintf("/cluster/firewall/groups/%s/%d", g.Group, rulePos), nil)
	return
}

type FirewallRule struct {
	Type     string `json:"type,omitempty"`
	Action   string `json:"action,omitempty"`
	Pos      int    `json:"pos,omitempty"`
	Comment  string `json:"comment,omitempty"`
	Dest     string `json:"dest,omitempty"`
	Dport    string `json:"dport,omitempty"`
	Enable   int    `json:"enable,omitempty"`
	IcmpType string `json:"icmp_type,omitempty"`
	Iface    string `json:"iface,omitempty"`
	Log      string `json:"log,omitempty"`
	Macro    string `json:"macro,omitempty"`
	Proto    string `json:"proto,omitempty"`
	Source   string `json:"source,omitempty"`
	Sport    string `json:"sport,omitempty"`
}

func (r *FirewallRule) IsEnable() bool {
	return 1 == r.Enable
}

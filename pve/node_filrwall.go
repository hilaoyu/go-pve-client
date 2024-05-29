package pve

import "fmt"

type FirewallNodeOption struct {
	Enable                           bool   `json:"enable,omitempty"`
	LogLevelIn                       string `json:"log_level_in,omitempty"`
	LogLevelOut                      string `json:"log_level_out,omitempty"`
	LogNfConntrack                   bool   `json:"log_nf_conntrack,omitempty"`
	Ntp                              bool   `json:"ntp,omitempty"`
	NfConntrackAllowInvalid          bool   `json:"nf_conntrack_allow_invalid,omitempty"`
	NfConntrackMax                   int    `json:"nf_conntrack_max,omitempty"`
	NfConntrackTcpTimeoutEstablished int    `json:"nf_conntrack_tcp_timeout_established,omitempty"`
	NfConntrackTcpTimeoutSynRecv     int    `json:"nf_conntrack_tcp_timeout_syn_recv,omitempty"`
	Nosmurfs                         bool   `json:"nosmurfs,omitempty"`
	ProtectionSynflood               bool   `json:"protection_synflood,omitempty"`
	ProtectionSynfloodBurst          int    `json:"protection_synflood_burst,omitempty"`
	ProtectionSynfloodRate           int    `json:"protection_synflood_rate,omitempty"`
	SmurfLogLevel                    string `json:"smurf_log_level,omitempty"`
	TcpFlagsLogLevel                 string `json:"tcp_flags_log_level,omitempty"`
	Tcpflags                         bool   `json:"tcpflags,omitempty"`
}

func (n *Node) FirewallOptionGet() (firewallOption *FirewallNodeOption, err error) {
	err = n.client.Get(fmt.Sprintf("/nodes/%s/firewall/options", n.Name), firewallOption)
	return
}
func (n *Node) FirewallOptionSet(firewallOption *FirewallNodeOption) (err error) {
	err = n.client.Put(fmt.Sprintf("/nodes/%s/firewall/options", n.Name), firewallOption, nil)
	return
}

func (n *Node) FirewallGetRules() (rules []*FirewallRule, err error) {
	err = n.client.Get(fmt.Sprintf("/nodes/%s/firewall/rules", n.Name), &rules)
	return
}

func (n *Node) FirewallRulesCreate(rule *FirewallRule) (err error) {
	err = n.client.Post(fmt.Sprintf("/nodes/%s/firewall/rules", n.Name), rule, nil)
	return
}
func (n *Node) FirewallRulesUpdate(rule *FirewallRule) (err error) {
	err = n.client.Put(fmt.Sprintf("/nodes/%s/firewall/rules/%d", n.Name, rule.Pos), rule, nil)
	return
}
func (n *Node) FirewallRulesDelete(rulePos int) (err error) {
	err = n.client.Delete(fmt.Sprintf("/nodes/%s/firewall/rules/%d", n.Name, rulePos), nil)
	return
}

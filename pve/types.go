package pve

import (
	"strconv"
	"strings"
)

type StringOrUint64 uint64

func (d *StringOrUint64) UnmarshalJSON(b []byte) error {
	str := strings.Replace(string(b), "\"", "", -1)
	parsed, err := strconv.ParseUint(str, 0, 64)
	if err != nil {
		return err
	}
	*d = StringOrUint64(parsed)
	return nil
}

type VNC struct {
	Cert     string
	Port     StringOrUint64
	Ticket   string
	UPID     string
	User     string
	Password string
}

type RootFS struct {
	Avail uint64
	Total uint64
	Free  uint64
	Used  uint64
}

type CPUInfo struct {
	UserHz  int `json:"user_hz"`
	MHZ     string
	Mode    string
	Cores   int
	Sockets int
	Flags   string
	CPUs    int
	HVM     string
}

type Memory struct {
	Used  uint64
	Free  uint64
	Total uint64
}

type Ksm struct {
	Shared int64
}

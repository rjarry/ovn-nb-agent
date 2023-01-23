// SPDX-License-Identifier: MIT
// Copyright (c) 2023 Robin Jarry

package linux

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io"
	"os"
	"os/exec"
	"strconv"
	"syscall"

	"github.com/rjarry/ovn-nb-agent/logger"
	"github.com/rjarry/ovn-nb-agent/models"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"golang.org/x/sys/unix"
)

const (
	DHCP_SERVER_PORT = 67
	DHCP_IFNAME      = "dhcp0"
)

func PrepareDhcpNetns(
	zone *models.DHCPZone, name string,
) (veth *netlink.Veth, err error) {
	var peer, lo netlink.Link
	var dhnl *netlink.Handle
	var stdin io.WriteCloser
	var dhns netns.NsHandle
	var addr *netlink.Addr
	var buf bytes.Buffer
	var pidfile string
	var cmd *exec.Cmd
	var f *os.File

	dhns, dhnl, err = EnsureNetns(name)
	if err != nil {
		goto out
	}
	defer dhnl.Close()
	defer dhns.Close()

	veth = new(netlink.Veth)
	veth.TxQLen = -1
	veth.Name = name
	veth.Alias = fmt.Sprintf("dhcp %s", zone.Opts.CIDR())
	veth.PeerName = DHCP_IFNAME
	veth.PeerNamespace = netlink.NsFd(dhns)
	veth.PeerHardwareAddr = *zone.Opts.ServerMac()

	logger.Debugf("add %#v", veth)
	if err = netlink.LinkAdd(veth); err != nil && !errors.Is(err, syscall.EEXIST) {
		goto out
	}
	logger.Debugf("set veth %s up", name)
	if err = netlink.LinkSetUp(veth); err != nil {
		goto out
	}
	logger.Debugf("netns %s set lo up", name)
	lo, err = dhnl.LinkByName("lo")
	if err != nil {
		goto out
	}
	if err = dhnl.LinkSetUp(lo); err != nil {
		goto out
	}

	peer, err = dhnl.LinkByName(DHCP_IFNAME)
	if err != nil {
		goto out
	}
	logger.Debugf("netns %s set %s up", name, DHCP_IFNAME)
	if err = dhnl.LinkSetUp(peer); err != nil {
		goto out
	}
	addr, err = netlink.ParseAddr(
		fmt.Sprintf("%s/%d", zone.Opts.ServerIP(), zone.Opts.CIDR().Bits()))
	if err != nil {
		goto out
	}
	logger.Debugf("netns %s addr add %s dev %s", name, addr.IPNet, DHCP_IFNAME)
	err = dhnl.AddrAdd(peer, addr)
	if err != nil && !errors.Is(err, syscall.EEXIST) {
		goto out
	}

	err = os.MkdirAll("/run/dnsmasq", 0o755)
	if err != nil && !errors.Is(err, syscall.EEXIST) {
		goto out
	}
	pidfile = fmt.Sprintf("/run/dnsmasq/%s.pid", name)
	f, err = os.Open(pidfile)
	if err == nil {
		var pid uint64
		buf, err := io.ReadAll(f)
		f.Close()
		if err == nil {
			pid, err = strconv.ParseUint(string(buf), 10, 0)
		}
		if err == nil {
			logger.Debugf("killing %s -> %d", pidfile, pid)
			unix.Kill(int(pid), syscall.SIGTERM)
		}
	}

	cmd = exec.Command(
		"ip", "netns", "exec", name, "dnsmasq", "-C", "-", "-x", pidfile)
	stdin, err = cmd.StdinPipe()
	if err != nil {
		goto out
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Start()
	if err != nil {
		stdin.Close()
		goto out
	}
	dnsmasqConf.Execute(&buf, zone)
	logger.Debugf("dnsmasq.conf: %s", buf.String())
	err = dnsmasqConf.Execute(stdin, zone)
	if err != nil {
		stdin.Close()
		goto out
	}
	stdin.Close()
	err = cmd.Wait()
	if err != nil {
		goto out
	}

out:
	return veth, err
}

var dnsmasqConf, _ = template.New("dnsmasq.conf").Parse(`port=0
interface=dhcp0
dhcp-range={{.Opts.CIDR.Masked.Addr}},static
{{range .Addrs}}dhcp-host={{.Mac}},{{.Ip}}
{{end}}
dhcp-option=option:router,{{.Opts.Router}}
dhcp-option=option:mtu,{{.Opts.Mtu}}
dhcp-option=option:dns-server{{range .Opts.DnsServers}},{{.}}{{end}}
dhcp-option=option:classless-static-route{{range .Opts.Routes}},{{.Prefix}},{{.NextHop}}{{end}}
dhcp-authoritative
`)

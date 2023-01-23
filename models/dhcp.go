// SPDX-License-Identifier: MIT
// Copyright (c) 2023 Robin Jarry

package models

import (
	"net"
	"net/netip"
	"strconv"
	"strings"

	"github.com/rjarry/ovn-nb-agent/logger"
	"github.com/rjarry/ovn-nb-agent/schema/ovn"
)

type DHCPOptions struct {
	uuid       string
	cidr       netip.Prefix
	router     netip.Addr
	mtu        uint
	leaseTime  uint
	dnsServers []netip.Addr
	routes     []Route
	serverIP   netip.Addr
	serverMac  net.HardwareAddr

	fetched bool
	db      ovn.DB
}

type DHCPZone struct {
	Opts  *DHCPOptions
	Addrs []*PortAddress
}

func NewDHCPOptions(uuid string, db ovn.DB) *DHCPOptions {
	return &DHCPOptions{uuid: uuid, db: db}
}

func (d *DHCPOptions) CIDR() *netip.Prefix {
	if !d.fetched {
		d.fetch()
	}
	return &d.cidr
}

func (d *DHCPOptions) Router() *netip.Addr {
	if !d.fetched {
		d.fetch()
	}
	return &d.router
}

func (d *DHCPOptions) Mtu() uint {
	if !d.fetched {
		d.fetch()
	}
	return d.mtu
}

func (d *DHCPOptions) LeaseTime() uint {
	if !d.fetched {
		d.fetch()
	}
	return d.leaseTime
}

func (d *DHCPOptions) DnsServers() []netip.Addr {
	if !d.fetched {
		d.fetch()
	}
	return d.dnsServers
}

func (d *DHCPOptions) Routes() []Route {
	if !d.fetched {
		d.fetch()
	}
	return d.routes
}

func (d *DHCPOptions) ServerIP() *netip.Addr {
	if !d.fetched {
		d.fetch()
	}
	return &d.serverIP
}

func (d *DHCPOptions) ServerMac() *net.HardwareAddr {
	if !d.fetched {
		d.fetch()
	}
	return &d.serverMac
}

func (d *DHCPOptions) fetch() {
	var dho *ovn.DHCPOptions
	var err error

	d.mtu = 0
	d.leaseTime = 0
	d.router = netip.IPv4Unspecified()
	d.serverIP = netip.IPv4Unspecified()
	d.dnsServers = nil
	d.routes = nil
	d.serverMac = nil

	dho, err = d.db.GetDHCPOptions(d.uuid)
	if err != nil {
		goto end
	}

	d.cidr, err = netip.ParsePrefix(dho.Cidr)

	if m, ok := dho.Options["mtu"]; ok {
		mtu, e := strconv.ParseUint(m, 10, 32)
		if e != nil {
			err = e
			goto end
		}
		d.mtu = uint(mtu)
	}
	if l, ok := dho.Options["lease_time"]; ok {
		lease, e := strconv.ParseUint(l, 10, 32)
		if e != nil {
			err = e
			goto end
		}
		d.leaseTime = uint(lease)
	}
	if r, ok := dho.Options["router"]; ok {
		gw, e := netip.ParseAddr(r)
		if e != nil {
			err = e
			goto end
		}
		d.router = gw
	}
	if i, ok := dho.Options["server_id"]; ok {
		ip, e := netip.ParseAddr(i)
		if e != nil {
			err = e
			goto end
		}
		d.serverIP = ip
	}
	if m, ok := dho.Options["server_mac"]; ok {
		mac, e := net.ParseMAC(m)
		if e != nil {
			err = e
			goto end
		}
		d.serverMac = mac
	}
	if s, ok := dho.Options["dns_server"]; ok {
		for _, ns := range strings.Split(strings.Trim(s, "{}"), ", ") {
			addr, e := netip.ParseAddr(ns)
			if e != nil {
				err = e
				goto end
			}
			d.dnsServers = append(d.dnsServers, addr)
		}
	}
	if r, ok := dho.Options["classless_static_route"]; ok {
		for _, rt := range strings.Split(strings.Trim(r, "{}"), ", ") {
			route, e := ParseRoute(rt)
			if e != nil {
				err = e
				goto end
			}
			d.routes = append(d.routes, route)
		}
	}
end:
	if err != nil {
		logger.Errorf("dhcpoptions fetch: %s", err)
	}
	d.fetched = true
}

func (d *DHCPOptions) Invalidate() {
	d.fetched = false
}

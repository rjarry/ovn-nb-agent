// SPDX-License-Identifier: MIT
// Copyright (c) 2023 Robin Jarry

package models

import (
	"github.com/rjarry/ovn-nb-agent/logger"
	"github.com/rjarry/ovn-nb-agent/schema/ovn"
)

type SwitchPort struct {
	uuid        string
	name        string
	type_       PortType
	addresses   []*PortAddress
	tag         uint
	networkName string
	dhcpOpts    []*DHCPOptions

	fetched bool
	db      ovn.DB
}

func NewSwitchPort(uuid string, db ovn.DB) *SwitchPort {
	return &SwitchPort{uuid: uuid, db: db}
}

func (p *SwitchPort) Name() string {
	if !p.fetched {
		p.fetch()
	}
	return p.name
}

func (p *SwitchPort) Type() PortType {
	if !p.fetched {
		p.fetch()
	}
	return p.type_
}

func (p *SwitchPort) NetworkName() string {
	if !p.fetched {
		p.fetch()
	}
	if p.networkName != "" {
		return p.networkName
	}
	return p.name
}

func (p *SwitchPort) Tag() uint {
	if !p.fetched {
		p.fetch()
	}
	return p.tag
}

func (p *SwitchPort) Addresses() []*PortAddress {
	if !p.fetched {
		p.fetch()
	}
	return p.addresses
}

func (p *SwitchPort) DHCPOptions() []*DHCPOptions {
	if !p.fetched {
		p.fetch()
	}
	return p.dhcpOpts
}

func (p *SwitchPort) fetch() {
	var lsp *ovn.LogicalSwitchPort
	var addr *PortAddress
	var err error

	p.addresses = nil
	p.dhcpOpts = nil

	lsp, err = p.db.GetSwitchPort(p.uuid)
	if err != nil {
		goto end
	}

	p.name = lsp.Name
	p.type_, err = ParsePortType(lsp.Type)
	if err != nil {
		goto end
	}
	for _, tag := range lsp.TagRequest {
		p.tag = uint(tag)
		break
	}
	for _, a := range lsp.Addresses {
		switch a {
		case "unknown":
			// XXX: wtf?
			continue
		case "router":
			// XXX: wtf?
			continue
		}
		addr, err = ParsePortAddress(a)
		if err != nil {
			goto end
		}
		p.addresses = append(p.addresses, addr)
	}
	p.networkName = lsp.Options["network_name"]

	for _, o := range lsp.Dhcpv4Options {
		p.dhcpOpts = append(p.dhcpOpts, NewDHCPOptions(o, p.db))
	}
	for _, o := range lsp.Dhcpv6Options {
		p.dhcpOpts = append(p.dhcpOpts, NewDHCPOptions(o, p.db))
	}

end:
	if err != nil {
		logger.Errorf("switchport fetch: %s", err)
	}
	p.fetched = true
}

func (p *SwitchPort) Invalidate() {
	p.fetched = false
}

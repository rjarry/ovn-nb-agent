// SPDX-License-Identifier: MIT
// Copyright (c) 2023 Robin Jarry

package models

import (
	"strconv"

	"github.com/rjarry/ovn-nb-agent/logger"
	"github.com/rjarry/ovn-nb-agent/schema/ovn"
)

type Switch struct {
	uuid         string
	name         string
	mcastSnoop   bool
	vlanPassthru bool
	mtu          uint
	ports        []*SwitchPort

	fetched bool
	db      ovn.DB
}

func NewSwitch(uuid string, db ovn.DB) *Switch {
	return &Switch{uuid: uuid, db: db}
}

func (sw *Switch) Name() string {
	if !sw.fetched {
		sw.fetch()
	}
	return sw.name
}

func (sw *Switch) McastSnoop() bool {
	if !sw.fetched {
		sw.fetch()
	}
	return sw.mcastSnoop
}

func (sw *Switch) VlanPassthru() bool {
	if !sw.fetched {
		sw.fetch()
	}
	return sw.vlanPassthru
}

func (sw *Switch) MTU() uint {
	if !sw.fetched {
		sw.fetch()
	}
	return sw.mtu
}

func (sw *Switch) Ports() []*SwitchPort {
	if !sw.fetched {
		sw.fetch()
	}
	return sw.ports
}

func (sw *Switch) fetch() {
	var err error
	var ls *ovn.LogicalSwitch

	sw.mtu = 0
	sw.ports = nil

	ls, err = sw.db.GetSwitch(sw.uuid)
	if err != nil {
		goto end
	}

	sw.name = ls.Name
	sw.mcastSnoop = ls.OtherConfig["mcast_snoop"] == "true"
	sw.vlanPassthru = ls.OtherConfig["vlan-passthru"] == "true"
	if m, ok := ls.ExternalIDs["neutron:mtu"]; ok {
		mtu, e := strconv.ParseUint(m, 10, 32)
		if e != nil {
			err = e
			goto end
		}
		sw.mtu = uint(mtu)
	}
	for _, uuid := range ls.Ports {
		sw.ports = append(sw.ports, NewSwitchPort(uuid, sw.db))
	}

end:
	if err != nil {
		logger.Errorf("switch fetch: %s", err)
	}
	sw.fetched = true
}

func (sw *Switch) Invalidate() {
	sw.fetched = false
}

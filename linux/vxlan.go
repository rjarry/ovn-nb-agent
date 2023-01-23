// SPDX-License-Identifier: MIT
// Copyright (c) 2023 Robin Jarry

package linux

import (
	"errors"
	"hash/fnv"
	"net/netip"
	"syscall"

	"github.com/rjarry/ovn-nb-agent/logger"
	"github.com/rjarry/ovn-nb-agent/models"
	"github.com/vishvananda/netlink"
)

const (
	VXLAN_PORT = 4789
	VXLAN_TTL  = 5
)

// Create and configure a Linux VXLAN interface using a multicast group address
// and VNI derived from the logical switch name.
func ConfigureVtep(
	sw *models.Switch, port *models.SwitchPort,
	encapIfindex int, encapAddr netip.Addr,
) (vxlan *netlink.Vxlan, err error) {
	var mcastGroup netip.Addr
	var vni uint32

	h := fnv.New32a()
	h.Write([]byte(sw.Name()))
	// mask to get a 24bit VXLAN identifier
	vni = h.Sum32() & 0xffffff

	// Multicast group address. Last 3 bytes == vni.
	if encapAddr.Is4() {
		var buf [4]byte
		buf[0] = 239 // For use in private multicast domains (RFC 5771)
		buf[1] = byte(vni >> 16)
		buf[2] = byte(vni >> 8)
		buf[3] = byte(vni)
		mcastGroup = netip.AddrFrom4(buf)
	} else {
		var buf [16]byte
		buf[0] = 0xff
		buf[1] = 0x08 // Organization-Local scope (RFC 4291)
		buf[13] = byte(vni >> 16)
		buf[14] = byte(vni >> 8)
		buf[15] = byte(vni)
		mcastGroup = netip.AddrFrom16(buf)
	}

	vxlan = new(netlink.Vxlan)
	vxlan.Name = CompactIfName(port.Name(), "lsp-", "")
	vxlan.Alias = "ovn-ls " + sw.Name()
	vxlan.MTU = int(sw.MTU())
	vxlan.TxQLen = -1 // use the kernel default
	vxlan.SrcAddr = encapAddr.AsSlice()
	vxlan.VxlanId = int(vni)
	vxlan.Group = mcastGroup.AsSlice()
	vxlan.VtepDevIndex = encapIfindex
	// Enable learning inner source mac and outer IP addresses into the
	// VXLAN device forwarding database.
	// When an inner destination mac address is present in the VXLAN FDB
	// and has a corresponding outer IP address, the outer header will use
	// that IP address instead of the multicast group address.
	vxlan.Learning = true
	vxlan.Port = VXLAN_PORT
	vxlan.TTL = VXLAN_TTL

	logger.Tracef("vxlan add: %#v", vxlan)
	err = netlink.LinkAdd(vxlan)
	if err != nil && !errors.Is(err, syscall.EEXIST) {
		goto out
	}

	logger.Tracef("set %s up", vxlan.Name)
	err = netlink.LinkSetUp(vxlan)

out:
	return
}

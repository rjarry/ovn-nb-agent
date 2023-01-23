package linux

import (
	"context"
	"errors"
	"fmt"
	"net/netip"
	"syscall"
	"time"

	"github.com/rjarry/ovn-nb-agent/linux"
	"github.com/rjarry/ovn-nb-agent/logger"
	"github.com/rjarry/ovn-nb-agent/models"
	"github.com/vishvananda/netlink"
)

func (b *LinuxBackend) configureBridge(
	sw *models.Switch,
) (*netlink.Bridge, error) {
	var err error
	bridge := new(netlink.Bridge)

	bridge.Name = linux.CompactIfName(sw.Name(), "ls-", "")
	bridge.Alias = "ovn-ls " + sw.Name()
	bridge.TxQLen = -1 // use default
	mcastSnoop := sw.McastSnoop()
	bridge.MulticastSnooping = &mcastSnoop
	vlanFiltering := !sw.VlanPassthru()
	// ensure proper isolation of the bridge: discard untagged
	// traffic by default
	bridge.VlanFiltering = &vlanFiltering

	logger.Tracef("bridge add: %#v", bridge)
	err = netlink.LinkAdd(bridge)
	if err != nil && !errors.Is(err, syscall.EEXIST) {
		return nil, err
	}
	timeout, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	link, err := b.linkCache.WaitName(timeout, bridge.Name)
	if err != nil {
		return nil, err
	}
	bridge, _ = link.(*netlink.Bridge)
	if bridge == nil {
		return nil, fmt.Errorf("not a bridge: %#v", link)
	}

	// ensure proper isolation of the bridge: prevent packets from
	// leaking out on the bridge interface
	logger.Tracef("bridge vlan del 1: %s", bridge.Name)
	err = netlink.BridgeVlanDel(bridge, 1, false, false, true, false)
	if err != nil && !errors.Is(err, syscall.ENOENT) {
		return nil, err
	}

	logger.Tracef("set %s up", bridge.Name)
	if err = netlink.LinkSetUp(bridge); err != nil {
		return nil, err
	}

	zones := make(map[netip.Prefix]*models.DHCPZone)
	var addrs []*models.PortAddress

	for _, port := range sw.Ports() {
		switch port.Type() {
		case models.LOCALNET:
			// direct access via vlan
			err = b.plugPhysical(bridge, sw, port)
		case models.LOCALPORT:
			// tunnel access
			err = b.plugOverlay(bridge, sw, port)
		case models.VIF:
			for _, dho := range port.DHCPOptions() {
				zones[*dho.CIDR()] = &models.DHCPZone{Opts: dho}
			}
			addrs = append(addrs, port.Addresses()...)
		}
		if err != nil {
			return nil, err
		}
	}
	for cidr, zone := range zones {
		for _, addr := range addrs {
			if cidr.Contains(addr.Ip) {
				zone.Addrs = append(zone.Addrs, addr)
			}
		}
		err = b.configureDHCP(bridge, zone)
		if err != nil {
			return nil, err
		}
	}

	return bridge, nil
}

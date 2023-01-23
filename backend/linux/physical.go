package linux

import (
	"errors"
	"fmt"
	"syscall"

	"github.com/rjarry/ovn-nb-agent/linux"
	"github.com/rjarry/ovn-nb-agent/logger"
	"github.com/rjarry/ovn-nb-agent/models"
	"github.com/vishvananda/netlink"
)

func (b *LinuxBackend) plugPhysical(
	bridge *netlink.Bridge, sw *models.Switch, port *models.SwitchPort,
) error {
	logger.Debugf("bridge plug physical: %s -> %s", bridge.Name, port.NetworkName())
	ifname, ok := b.bridgeMappings[port.NetworkName()]
	if !ok {
		return fmt.Errorf("no bridge mapping for network %q",
			port.NetworkName())
	}
	link := b.linkCache.FromName(ifname)
	if link == nil {
		return fmt.Errorf("network %q no such interface %q",
			port.NetworkName(), ifname)
	}

	logger.Debugf("network %s -> link %s (vlan: %d)",
		port.NetworkName(), link.Attrs().Name, port.Tag())

	if port.Tag() != 0 {
		vlan := new(netlink.Vlan)
		vlan.Name = linux.CompactIfName(ifname, "", fmt.Sprintf(".%d", port.Tag()))
		vlan.Alias = port.NetworkName() + fmt.Sprintf(" (VLAN %d)", port.Tag())
		vlan.TxQLen = -1 // use default
		vlan.ParentIndex = link.Attrs().Index
		vlan.VlanId = int(port.Tag())
		vlan.VlanProtocol = netlink.VLAN_PROTOCOL_8021Q
		logger.Tracef("vlan add: %#v", vlan)
		err := netlink.LinkAdd(vlan)
		if err != nil && !errors.Is(err, syscall.EEXIST) {
			return err
		}
		link = vlan
	}

	logger.Tracef("set %s master %s", link.Attrs().Name, bridge.Name)
	err := netlink.LinkSetMaster(link, bridge)
	if err != nil {
		return err
	}

	logger.Tracef("set %s isolated", link.Attrs().Name)
	err = netlink.LinkSetIsolated(link, true)
	if err != nil {
		return err
	}

	logger.Tracef("set %s up", link.Attrs().Name)
	return netlink.LinkSetUp(link)
}

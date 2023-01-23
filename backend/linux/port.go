package linux

import (
	"github.com/rjarry/ovn-nb-agent/logger"
	"github.com/rjarry/ovn-nb-agent/models"
	"github.com/vishvananda/netlink"
)

func (b *LinuxBackend) configurePort(
	link netlink.Link, bridge *netlink.Bridge, port *models.SwitchPort,
) error {
	for _, dho := range port.DHCPOptions() {
		err := b.redirectDhcpRequests(link, bridge, dho)
		if err != nil {
			return err
		}
	}

	logger.Tracef("set %s master %s", link.Attrs().Name, bridge.Name)
	if err := netlink.LinkSetMaster(link, bridge); err != nil {
		return err
	}

	logger.Tracef("set %s up", link.Attrs().Name)
	if err := netlink.LinkSetUp(link); err != nil {
		return err
	}

	return nil
}

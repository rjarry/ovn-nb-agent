package linux

import (
	"fmt"

	"github.com/rjarry/ovn-nb-agent/linux"
	"github.com/rjarry/ovn-nb-agent/logger"
	"github.com/rjarry/ovn-nb-agent/models"
	"github.com/vishvananda/netlink"
)

func (b *LinuxBackend) plugOverlay(
	bridge *netlink.Bridge, sw *models.Switch, port *models.SwitchPort,
) error {
	logger.Tracef("%s: getting vtep link index", b.encapAddress)
	linkIndex, ok := b.addrCache.GetLinkIndex(b.encapAddress)
	if !ok {
		return fmt.Errorf("no link with encap address: %s", b.encapAddress)
	}

	vxlan, err := linux.ConfigureVtep(sw, port, linkIndex, b.encapAddress)
	if err != nil {
		return err
	}
	err = netlink.LinkSetMaster(vxlan, bridge)
	if err != nil {
		return err
	}

	logger.Tracef("set %s isolated", vxlan.Name)
	return netlink.LinkSetIsolated(vxlan, true)
}

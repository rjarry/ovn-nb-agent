package ovs

import (
	"errors"

	"github.com/rjarry/ovn-nb-agent/models"
	"github.com/vishvananda/netlink"
)

func (b *OvsBackend) plugOverlay(
	bridge *netlink.Bridge, sw *models.Switch, port *models.SwitchPort,
) error {
	return errors.New("not implemented")
}

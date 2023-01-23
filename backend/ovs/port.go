package ovs

import (
	"errors"

	"github.com/rjarry/ovn-nb-agent/models"
	"github.com/rjarry/ovn-nb-agent/schema/ovs"
)

func (b *OvsBackend) configurePort(
	iface *ovs.Interface, bridge *ovs.Bridge, port *models.SwitchPort,
) error {
	return errors.New("not implemented")
}

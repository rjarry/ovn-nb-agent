package ovs

import (
	"errors"

	"github.com/rjarry/ovn-nb-agent/models"
	"github.com/rjarry/ovn-nb-agent/schema/ovs"
	"github.com/vishvananda/netlink"
)

func (b *OvsBackend) configureDHCP(bridge *ovs.Bridge, zone *models.DHCPZone) error {
	return errors.New("not implemented")
}

func (b *OvsBackend) redirectDhcpRequests(
	port netlink.Link, bridge *netlink.Bridge, opts *models.DHCPOptions,
) error {
	return errors.New("not implemented")
}

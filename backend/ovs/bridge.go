package ovs

import (
	"github.com/rjarry/ovn-nb-agent/models"
	"github.com/rjarry/ovn-nb-agent/schema/ovs"
)

func (b *OvsBackend) configureBridge(
	sw *models.Switch,
) (*ovs.Bridge, error) {
	bridge := &ovs.Bridge{Name: sw.Name()}

	err := b.vs.ovs.Get(bridge)
	if err != nil {
		bridge, err = b.vs.AddBridge(sw.Name())
	}

	return bridge, err
}

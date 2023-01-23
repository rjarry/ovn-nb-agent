// SPDX-License-Identifier: MIT
// Copyright (c) 2023 Robin Jarry
package ovs

import (
	"net/netip"
	"os"
	"path"

	"github.com/akamensky/argparse"
	"github.com/rjarry/ovn-nb-agent/linux/cache"
	"github.com/rjarry/ovn-nb-agent/logger"
	"github.com/rjarry/ovn-nb-agent/models"
	"github.com/rjarry/ovn-nb-agent/schema/ovs"
)

type OvsBackend struct {
	endpoint  *string
	datapath  *string
	vs        Vswitchd
	addrCache cache.AddrCache
}

var Backend OvsBackend

func (b *OvsBackend) Name() string {
	return "ovs"
}

func (b *OvsBackend) Arguments(parser *argparse.Parser) {
	b.endpoint = parser.String(
		"o", "ovs-endpoint",
		&argparse.Options{
			Help:    "Open vSwitch database socket endpoint",
			Default: defaultEndpoint(),
		},
	)
	b.datapath = parser.Selector(
		"d", "ovs-datapath",
		[]string{string(DATAPATH_SYSTEM), string(DATAPATH_NETDEV)},
		&argparse.Options{
			Help: "Open vSwitch datapath for created bridges",
		},
	)
}

func defaultEndpoint() string {
	var runDir string = os.Getenv("OVS_RUNDIR")
	if runDir == "" {
		runDir = "/run/openvswitch"
	}
	return "unix:" + path.Join(runDir, "db.sock")
}

func (b *OvsBackend) Init(
	bridgeMappings map[string]string, encapAddress netip.Addr,
) error {
	err := b.vs.Connect(*b.endpoint)
	if err != nil {
		return err
	}
	if b.datapath != nil && *b.datapath != "" {
		b.vs.datapath = *b.datapath
	}
	if bridgeMappings != nil {
		b.vs.bridgeMappings = bridgeMappings
	}
	if encapAddress.IsGlobalUnicast() {
		b.vs.encapAddress = encapAddress
	}

	err = b.addrCache.Start()
	if err != nil {
		return err
	}
	logger.Infof("monitoring linux addresses")

	return nil
}

func (b *OvsBackend) Shutdown() error {
	b.vs.Disconnect()
	b.addrCache.Stop()
	return nil
}

func (b *OvsBackend) PortAdd(p *models.SwitchPort, sw *models.Switch) error {
	name := p.Name()
	p.Invalidate()
	b.vs.Cache().IfaceIdCallback(name, func(iface *ovs.Interface) {
		b.configure(p, sw, iface)
	})
	return nil
}

func (b *OvsBackend) PortUpdate(p *models.SwitchPort, sw *models.Switch) error {
	name := p.Name()
	p.Invalidate()
	b.vs.Cache().IfaceIdCallback(name, func(iface *ovs.Interface) {
		b.configure(p, sw, iface)
	})
	return nil
}

func (b *OvsBackend) PortDelete(p *models.SwitchPort, sw *models.Switch) error {
	// TODO: unconfigure stuff
	return nil
}

func (b *OvsBackend) configure(
	port *models.SwitchPort, sw *models.Switch, iface *ovs.Interface,
) {
	var err error
	var br *ovs.Bridge

	if br, err = b.configureBridge(sw); err != nil {
		goto end
	}
	if err = b.configurePort(iface, br, port); err != nil {
		goto end
	}
end:
	if err != nil {
		logger.Errorf("%s", err)
	}
}

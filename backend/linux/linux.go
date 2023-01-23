// SPDX-License-Identifier: MIT
// Copyright (c) 2023 Robin Jarry
package linux

import (
	"fmt"
	"net/netip"

	"github.com/akamensky/argparse"
	"github.com/rjarry/ovn-nb-agent/linux/cache"
	"github.com/rjarry/ovn-nb-agent/logger"
	"github.com/rjarry/ovn-nb-agent/models"
	"github.com/vishvananda/netlink"
)

type LinuxBackend struct {
	bridgeMappings map[string]string
	encapAddress   netip.Addr
	encapLinkIndex int
	linkCache      cache.LinkCache
	addrCache      cache.AddrCache
}

var Backend LinuxBackend

func (b *LinuxBackend) Name() string {
	return "linux"
}

func (b *LinuxBackend) Arguments(parser *argparse.Parser) {
}

func (b *LinuxBackend) Init(
	bridgeMappings map[string]string, encapAddress netip.Addr,
) error {
	var err error
	if len(bridgeMappings) == 0 {
		return fmt.Errorf("-m|--bridge-mappings argument is required")
	}
	if encapAddress.IsUnspecified() {
		return fmt.Errorf("-e|--encap-address argument is required")
	}
	b.bridgeMappings = bridgeMappings
	b.encapAddress = encapAddress
	if err != nil {
		return err
	}
	err = b.linkCache.Start()
	if err != nil {
		return err
	}
	err = b.addrCache.Start()
	if err != nil {
		return err
	}

	return nil
}

func (b *LinuxBackend) Shutdown() error {
	b.linkCache.Stop()
	b.addrCache.Stop()
	return nil
}

func (b *LinuxBackend) PortAdd(
	port *models.SwitchPort, sw *models.Switch,
) error {
	name := port.Name()
	port.Invalidate()
	b.linkCache.AliasCallback(name, func(link netlink.Link) {
		b.configure(port, sw, link)
	})
	return nil
}

func (b *LinuxBackend) PortUpdate(
	port *models.SwitchPort, sw *models.Switch,
) error {
	name := port.Name()
	port.Invalidate()
	b.linkCache.AliasCallback(name, func(link netlink.Link) {
		b.configure(port, sw, link)
	})
	return nil
}

func (b *LinuxBackend) PortDelete(
	port *models.SwitchPort, sw *models.Switch,
) error {
	name := port.Name()
	port.Invalidate()
	b.linkCache.AliasCallback(name, func(link netlink.Link) {
		// TODO: delete port
	})
	return nil
}

func (b *LinuxBackend) configure(
	port *models.SwitchPort, sw *models.Switch, link netlink.Link,
) {
	var err error
	var br *netlink.Bridge

	if br, err = b.configureBridge(sw); err != nil {
		goto end
	}
	if err = b.configurePort(link, br, port); err != nil {
		goto end
	}
end:
	if err != nil {
		logger.Errorf("%s", err)
	}
}

// SPDX-License-Identifier: MIT
// Copyright (c) 2023 Robin Jarry
package backend

import (
	"net/netip"

	"github.com/akamensky/argparse"
	"github.com/rjarry/ovn-nb-agent/backend/linux"
	"github.com/rjarry/ovn-nb-agent/backend/ovs"
	"github.com/rjarry/ovn-nb-agent/models"
)

type Backend interface {
	Name() string
	Arguments(*argparse.Parser)

	Init(bridgeMappings map[string]string, encapAddress netip.Addr) error
	Shutdown() error

	PortAdd(*models.SwitchPort, *models.Switch) error
	PortUpdate(*models.SwitchPort, *models.Switch) error
	PortDelete(*models.SwitchPort, *models.Switch) error
}

var Backends = []Backend{
	&linux.Backend,
	&ovs.Backend,
}

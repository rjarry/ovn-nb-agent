// SPDX-License-Identifier: MIT
// Copyright (c) 2023 Robin Jarry
package nm

import (
	"github.com/Wifx/gonetworkmanager/v2"
	"github.com/ovn-org/libovsdb/client"
	"github.com/rjarry/ovn-nb-agent/logger"
)

type NmBackend struct {
	nm    gonetworkmanager.NetworkManager
	north client.Client
}

var Backend NmBackend

func (n *NmBackend) Init(north client.Client, args []string) error {
	nm, err := gonetworkmanager.NewNetworkManager()
	if err != nil {
		return err
	}
	version, err := nm.GetPropertyVersion()
	if err != nil {
		return err
	}
	logger.Infof("connected to network manager %s", version)
	n.nm = nm
	n.north = north
	return nil
}

// SPDX-License-Identifier: MIT
// Copyright (c) 2023 Robin Jarry

package ovn

import (
	"context"
	"fmt"
	"time"

	"github.com/ovn-org/libovsdb/cache"
	"github.com/ovn-org/libovsdb/client"
	"github.com/ovn-org/libovsdb/model"
	"github.com/rjarry/ovn-nb-agent/backend"
	"github.com/rjarry/ovn-nb-agent/logger"
	"github.com/rjarry/ovn-nb-agent/models"
	"github.com/rjarry/ovn-nb-agent/schema/ovn"
)

type NorthboundDB struct {
	db      client.Client // ovsdb connection
	monitor string        // from ovn monitor request
	back    backend.Backend
}

func Connect(
	endpoint string, back backend.Backend,
) (nb *NorthboundDB, err error) {
	logger.Debugf("connecting to ovn northbound db: %s", endpoint)
	dbModel, err := ovn.FullDatabaseModel()
	if err != nil {
		return
	}
	db, err := client.NewOVSDBClient(dbModel, client.WithEndpoint(endpoint))
	if err != nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = db.Connect(ctx)
	if err != nil {
		return
	}
	nb = &NorthboundDB{db: db, back: back}
	// init monitoring
	db.Cache().AddEventHandler(&cache.EventHandlerFuncs{
		AddFunc:    nb.onAdd,
		UpdateFunc: nb.onUpdate,
		DeleteFunc: nb.onDelete,
	})
	nb.monitor, err = db.MonitorAll()
	return
}

func (nb *NorthboundDB) Disconnect() {
	if nb.db != nil && nb.db.Connected() {
		nb.db.MonitorCancel(nb.monitor)
		nb.db.Disconnect()
	}
}

func (nb *NorthboundDB) onAdd(table string, n model.Model) {
	var err error
	logger.Tracef("add %s %#v", table, n)
	switch n := n.(type) {
	case *ovn.LogicalSwitchPort:
		err = nb.port(n, nb.back.PortAdd)
		break
	}
	if err != nil {
		logger.Errorf("%s", err)
	}
}

func (nb *NorthboundDB) onUpdate(table string, o, n model.Model) {
	var err error
	logger.Tracef("update %s %#v -> %#v", table, o, n)
	switch n := n.(type) {
	case *ovn.LogicalSwitchPort:
		err = nb.port(n, nb.back.PortUpdate)
		break
	}
	if err != nil {
		logger.Errorf("%s", err)
	}
}

func (nb *NorthboundDB) onDelete(table string, o model.Model) {
	var err error
	logger.Tracef("delete %s %#v", table, o)
	switch n := o.(type) {
	case *ovn.LogicalSwitchPort:
		err = nb.port(n, nb.back.PortDelete)
		break
	}
	if err != nil {
		logger.Errorf("%s", err)
	}
}

func (nb *NorthboundDB) port(
	port *ovn.LogicalSwitchPort,
	cb func(*models.SwitchPort, *models.Switch) error,
) error {
	if port.Type != "" {
		// not a VIF port, ignore
		return nil
	}
	ls, err := nb.parentSwitch(port.UUID)
	if err != nil {
		return err
	}

	p := models.NewSwitchPort(port.UUID, nb)
	sw := models.NewSwitch(ls.UUID, nb)

	return cb(p, sw)
}

func (nb *NorthboundDB) GetSwitchPort(uuid string) (*ovn.LogicalSwitchPort, error) {
	api := nb.db.WhereCache(
		func(lsp *ovn.LogicalSwitchPort) bool {
			return lsp.UUID == uuid
		},
	)
	var results []ovn.LogicalSwitchPort
	err := api.List(&results)
	if err != nil || len(results) == 0 {
		return nil, fmt.Errorf("no such port %v", uuid)
	}
	return &results[0], nil
}

func (nb *NorthboundDB) GetSwitch(uuid string) (*ovn.LogicalSwitch, error) {
	api := nb.db.WhereCache(
		func(ls *ovn.LogicalSwitch) bool {
			return ls.UUID == uuid
		},
	)
	var results []ovn.LogicalSwitch
	err := api.List(&results)
	if err != nil || len(results) == 0 {
		return nil, fmt.Errorf("no such switch %v", uuid)
	}
	return &results[0], nil
}

func (nb *NorthboundDB) GetDHCPOptions(uuid string) (*ovn.DHCPOptions, error) {
	api := nb.db.WhereCache(
		func(d *ovn.DHCPOptions) bool {
			return d.UUID == uuid
		},
	)
	var results []ovn.DHCPOptions
	err := api.List(&results)
	if err != nil || len(results) == 0 {
		return nil, fmt.Errorf("no such dhcpoptions %v", uuid)
	}
	return &results[0], nil
}

func (nb *NorthboundDB) parentSwitch(port string) (*ovn.LogicalSwitch, error) {
	api := nb.db.WhereCache(
		func(ls *ovn.LogicalSwitch) bool {
			for _, p := range ls.Ports {
				if p == port {
					return true
				}
			}
			return false
		},
	)
	var results []ovn.LogicalSwitch
	err := api.List(&results)
	if err != nil || len(results) == 0 {
		return nil, fmt.Errorf("no switch with port %v", port)
	}
	return &results[0], nil
}

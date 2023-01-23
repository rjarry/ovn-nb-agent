// SPDX-License-Identifier: MIT
// Copyright (c) 2023 Robin Jarry
package backend

import (
	"fmt"

	"github.com/ovn-org/libovsdb/client"
	"github.com/ovn-org/libovsdb/model"
	"github.com/rjarry/ovn-nb-agent/backend/nm"
	"github.com/rjarry/ovn-nb-agent/logger"
	nb "github.com/rjarry/ovn-nb-agent/schema/northbound"
)

type Backend interface {
	Init(client.Client) error

	LogicalRouterAdd(*nb.LogicalRouter) error
	LogicalRouterUpdate(*nb.LogicalRouter, *nb.LogicalRouter) error
	LogicalRouterDelete(*nb.LogicalRouter) error

	LogicalRouterPolicyAdd(*nb.LogicalRouterPolicy) error
	LogicalRouterPolicyUpdate(*nb.LogicalRouterPolicy, *nb.LogicalRouterPolicy) error
	LogicalRouterPolicyDelete(*nb.LogicalRouterPolicy) error

	LogicalRouterPortAdd(*nb.LogicalRouterPort) error
	LogicalRouterPortUpdate(*nb.LogicalRouterPort, *nb.LogicalRouterPort) error
	LogicalRouterPortDelete(*nb.LogicalRouterPort) error

	LogicalRouterStaticRouteAdd(*nb.LogicalRouterStaticRoute) error
	LogicalRouterStaticRouteUpdate(*nb.LogicalRouterStaticRoute, *nb.LogicalRouterStaticRoute) error
	LogicalRouterStaticRouteDelete(*nb.LogicalRouterStaticRoute) error

	LogicalSwitchAdd(*nb.LogicalSwitch) error
	LogicalSwitchUpdate(*nb.LogicalSwitch, *nb.LogicalSwitch) error
	LogicalSwitchDelete(*nb.LogicalSwitch) error

	LogicalSwitchPortAdd(*nb.LogicalSwitchPort) error
	LogicalSwitchPortUpdate(*nb.LogicalSwitchPort, *nb.LogicalSwitchPort) error
	LogicalSwitchPortDelete(*nb.LogicalSwitchPort) error
}

var (
	backends = map[string]Backend{
		"nm": &nm.Backend,
	}
	b Backend
)

func InitBackend(name string, north client.Client, args []string) error {
	back, found := backends[name]
	if !found {
		return fmt.Errorf("unknown backend name %q", name)
	}
	err := back.Init(north, args)
	if err != nil {
		return err
	}
	b = back
	return nil
}

func OnAdd(table string, obj model.Model) {
	var err error

	logger.Debugf("add %s %#v", table, obj)

	switch o := obj.(type) {
	case *nb.LogicalSwitch:
		err = b.LogicalSwitchAdd(o)
	case *nb.LogicalSwitchPort:
		err = b.LogicalSwitchPortAdd(o)
	default:
		logger.Debugf("add %s ignored", table)
	}

	if err != nil {
		logger.Errorf("add %s: %s", table, err)
	}
}

func OnUpdate(table string, oldObj model.Model, newObj model.Model) {
	var err error

	logger.Debugf("update %s %#v -> %#v", table, oldObj, newObj)

	switch o := oldObj.(type) {
	case *nb.LogicalSwitch:
		err = b.LogicalSwitchUpdate(o, newObj.(*nb.LogicalSwitch))
	case *nb.LogicalSwitchPort:
		err = b.LogicalSwitchPortUpdate(o, newObj.(*nb.LogicalSwitchPort))
	default:
		logger.Debugf("update %s ignored", table)
	}

	if err != nil {
		logger.Errorf("update %s: %s", table, err)
	}
}

func OnDelete(table string, obj model.Model) {
	var err error

	logger.Debugf("delete %s %v", table, obj)

	switch o := obj.(type) {
	case *nb.LogicalSwitch:
		err = b.LogicalSwitchDelete(o)
	case *nb.LogicalSwitchPort:
		err = b.LogicalSwitchPortDelete(o)
	default:
		logger.Debugf("delete %s ignored", table)
	}

	if err != nil {
		logger.Errorf("delete %s: %s", table, err)
	}
}

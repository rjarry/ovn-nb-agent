// SPDX-License-Identifier: MIT
// Copyright (c) 2023 Robin Jarry
package nm

import "github.com/rjarry/ovn-nb-agent/schema"

func (n *NmBackend) LogicalSwitchAdd(
	obj *schema.LogicalSwitch,
) error {
	return nil
}

func (n *NmBackend) LogicalSwitchUpdate(
	oldObj *schema.LogicalSwitch,
	newObj *schema.LogicalSwitch,
) error {
	return nil
}

func (n *NmBackend) LogicalSwitchDelete(
	obj *schema.LogicalSwitch,
) error {
	return nil
}

func (n *NmBackend) LogicalSwitchPortAdd(
	obj *schema.LogicalSwitchPort,
) error {
	return nil
}

func (n *NmBackend) LogicalSwitchPortUpdate(
	oldObj *schema.LogicalSwitchPort,
	newObj *schema.LogicalSwitchPort,
) error {
	return nil
}

func (n *NmBackend) LogicalSwitchPortDelete(
	obj *schema.LogicalSwitchPort,
) error {
	return nil
}

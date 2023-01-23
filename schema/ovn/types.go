// SPDX-License-Identifier: MIT
// Copyright (c) 2023 Robin Jarry

package ovn

type DB interface {
	GetSwitchPort(uuid string) (*LogicalSwitchPort, error)
	GetSwitch(uuid string) (*LogicalSwitch, error)
	GetDHCPOptions(uuid string) (*DHCPOptions, error)
}

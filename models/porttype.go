// SPDX-License-Identifier: MIT
// Copyright (c) 2023 Robin Jarry

package models

import "fmt"

type PortType int

const (
	VIF PortType = iota
	ROUTER
	LOCALNET
	LOCALPORT
	L2GATEWAY
	VTEP
	EXTERNAL
	VIRTUAL
	REMOTE
)

func ParsePortType(s string) (PortType, error) {
	var t PortType
	var err error
	switch s {
	case "":
		t = VIF
	case "router":
		t = ROUTER
	case "localnet":
		t = LOCALNET
	case "localport":
		t = LOCALPORT
	case "l2gateway":
		t = L2GATEWAY
	case "vtep":
		t = VTEP
	case "external":
		t = EXTERNAL
	case "virtual":
		t = VIRTUAL
	case "remote":
		t = REMOTE
	default:
		err = fmt.Errorf("unknown port type: %s", s)
	}
	return t, err
}

func (t PortType) String() string {
	switch t {
	case VIF:
		return "vif"
	case ROUTER:
		return "router"
	case LOCALNET:
		return "localnet"
	case LOCALPORT:
		return "localport"
	case L2GATEWAY:
		return "l2gateway"
	case VTEP:
		return "vtep"
	case EXTERNAL:
		return "external"
	case VIRTUAL:
		return "virtual"
	case REMOTE:
		return "remote"
	}
	return ""
}

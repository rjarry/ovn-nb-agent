// SPDX-License-Identifier: MIT
// Copyright (c) 2023 Robin Jarry

package models

import (
	"net"
	"net/netip"
	"strings"
)

type PortAddress struct {
	Mac net.HardwareAddr
	Ip  netip.Addr
}

func ParsePortAddress(blob string) (*PortAddress, error) {
	tokens := strings.SplitN(blob, " ", 2)
	if len(tokens) != 2 {
		return nil, &net.AddrError{
			Err: "invalid port address format", Addr: blob,
		}
	}
	mac, err := net.ParseMAC(tokens[0])
	if err != nil {
		return nil, err
	}
	ip, err := netip.ParseAddr(tokens[1])
	if err != nil {
		return nil, err
	}
	return &PortAddress{Mac: mac, Ip: ip}, nil
}

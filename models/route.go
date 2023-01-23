// SPDX-License-Identifier: MIT
// Copyright (c) 2023 Robin Jarry

package models

import (
	"errors"
	"fmt"
	"net/netip"
	"strings"
)

type Route struct {
	Prefix  netip.Prefix
	NextHop netip.Addr
}

func ParseRoute(blob string) (Route, error) {
	var route Route
	var err error

	tokens := strings.SplitN(blob, ",", 2)
	if len(tokens) != 2 {
		err = errors.New("invalid route blob format")
		goto end
	}
	route.Prefix, err = netip.ParsePrefix(tokens[0])
	if err != nil {
		goto end
	}
	route.NextHop, err = netip.ParseAddr(tokens[1])
	if err != nil {
		goto end
	}
end:
	return route, err
}

func (r *Route) String() string {
	return fmt.Sprintf("%s via %s", r.Prefix.String(), r.NextHop.String())
}

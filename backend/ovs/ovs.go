// SPDX-License-Identifier: MIT
// Copyright (c) 2023 Robin Jarry
package ovs

import (
	"context"
	"flag"

	"github.com/ovn-org/libovsdb/client"
	"github.com/rjarry/ovn-nb-agent/logger"
)

type OvsBackend struct {
	ovs   client.Client
	north client.Client
}

var Backend OvsBackend

const sockEndpoint string = "unix:/var/run/openvswitch/db.sock"

func (n *OvsBackend) Init(north client.Client, args []string) error {
	var endpoint string

	fs := flag.NewFlagSet("ovs", flag.ContinueOnError)
	fs.StringVar(&endpoint, "o", sockEndpoint, "ovs vswitchd socket endpoint")
	err := fs.Parse(args)
	if err != nil {
		return err
	}

	// init connection
	db, err = schema.FullDatabaseModel()
	if err != nil {
		return err
	}

	ovs, err = client.NewOVSDBClient(db, client.WithEndpoint(endpoint))
	if err != nil {
		return err
	}
	err = ovs.Connect(context.Background())
	if err != nil {
		return err
	}
	logger.Infof("connected to ovs vswitchd db: %s", endpoint)

	n.nm = nm
	n.north = north
	return nil
}

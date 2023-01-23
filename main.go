// SPDX-License-Identifier: MIT
// Copyright (c) 2023 Robin Jarry
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path"
	"syscall"

	"github.com/ovn-org/libovsdb/cache"
	"github.com/ovn-org/libovsdb/client"
	"github.com/ovn-org/libovsdb/model"

	"github.com/rjarry/ovn-nb-agent/backend"
	"github.com/rjarry/ovn-nb-agent/logger"
	"github.com/rjarry/ovn-nb-agent/schema"
)

var version string // set at build time
const sockEndpoint string = "unix:/var/run/ovn/ovnsb_db.sock"

func main() {
	var err error
	var north client.Client
	var monitor string
	var endpoint string
	var backendName string
	var showVersion bool
	var logLevel string
	var db *model.DBModel

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	fs.StringVar(&logLevel, "l", "info", "log level")
	fs.BoolVar(&showVersion, "V", false, "show version and exit")
	fs.StringVar(&endpoint, "s", sockEndpoint, "ovn NB socket endpoint")
	fs.StringVar(&backendName, "b", "nm", "configuration backend")
	_ = fs.Parse(os.Args[1:])

	if showVersion {
		fmt.Printf("%s version %s\n", path.Base(os.Args[0]), version)
		os.Exit(0)
	}

	err = logger.Init(os.Stdout, logLevel)
	if err != nil {
		goto end
	}

	// init connection
	db, err = schema.FullDatabaseModel()
	if err != nil {
		goto end
	}
	north, err = client.NewOVSDBClient(db, client.WithEndpoint(endpoint))
	if err != nil {
		goto end
	}
	err = north.Connect(context.Background())
	if err != nil {
		goto end
	}
	logger.Infof("connected to ovn northbound db: %s", endpoint)

	err = backend.InitBackend(backendName, north, fs.Args())
	if err != nil {
		goto end
	}

	// init monitoring
	monitor, err = north.MonitorAll()
	if err != nil {
		goto end
	}
	north.Cache().AddEventHandler(&cache.EventHandlerFuncs{
		AddFunc:    backend.OnAdd,
		UpdateFunc: backend.OnUpdate,
		DeleteFunc: backend.OnDelete,
	})
	logger.Infof("monitoring")

	<-sigs
end:
	if north != nil && north.Connected() {
		north.MonitorCancel(monitor)
		north.Disconnect()
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}

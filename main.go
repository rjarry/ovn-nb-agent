// SPDX-License-Identifier: MIT
// Copyright (c) 2023 Robin Jarry

package main

import (
	"fmt"
	"net/netip"
	"os"
	"os/signal"
	"path"
	"syscall"

	"github.com/akamensky/argparse"
	"github.com/vishvananda/netlink/nl"

	"github.com/rjarry/ovn-nb-agent/backend"
	"github.com/rjarry/ovn-nb-agent/logger"
	"github.com/rjarry/ovn-nb-agent/ovn"
	"github.com/rjarry/ovn-nb-agent/parse"
)

var version string // set at build time

func main() {
	var err error
	var back backend.Backend
	var north *ovn.NorthboundDB
	var bridgeMappings map[string]string
	var encapIP netip.Addr

	// enable netlink extended ack for verbose error messages
	nl.EnableErrorMessageReporting = true

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	parser := argparse.NewParser("", "OVN northbound agent")
	parser.Flag(
		"V", "version",
		&argparse.Options{
			Help: "Show version and exit",
			Validate: func(args []string) error {
				fmt.Printf("%s version %s\n",
					path.Base(args[0]), version)
				os.Exit(0)
				return nil
			},
		},
	)
	logLevel := parser.Selector(
		"l", "log-level",
		logger.LevelNames(),
		&argparse.Options{
			Default: func() string {
				v := os.Getenv("OVN_LOG_LEVEL")
				if v == "" {
					v = "info"
				}
				return v
			}(),
			Help: "Log level (env: OVN_LOG_LEVEL)",
		},
	)
	endpoint := parser.String(
		"n", "ovn-nb-socket",
		&argparse.Options{
			Default: func() string {
				var sock string = os.Getenv("OVN_NB_DB")
				var runDir string = os.Getenv("OVN_RUNDIR")
				if runDir == "" {
					runDir = "/run/ovn"
				}
				if sock == "" {
					sock = "unix:" + path.Join(
						runDir, "ovnnb_db.sock")
				}
				return sock
			}(),
			Help: "OVN Northbound database socket endpoint (env: OVN_NB_DB)",
		},
	)
	mappings := parser.String(
		"m", "bridge-mappings",
		&argparse.Options{
			Default: func() string {
				return os.Getenv("OVN_BRIDGE_MAPPINGS")
			}(),
			Help: "Logical switches & interface mappings (env: OVN_BRIDGE_MAPPINGS). The required format is NET:IFACE[,NET:IFACE...]",
		},
	)
	encap := parser.String(
		"e", "encap-address",
		&argparse.Options{
			Default: func() string {
				return os.Getenv("OVN_ENCAP_ADDRESS")
			}(),
			Help: "Tunneling encapsulation address (env: OVN_ENCAP_ADDRESS)",
		},
	)

	var names []string
	backends := make(map[string]backend.Backend)
	for _, b := range backend.Backends {
		names = append(names, b.Name())
		b.Arguments(parser)
		backends[b.Name()] = b
	}

	backendName := parser.Selector(
		"b", "backend",
		names,
		&argparse.Options{
			Required: true,
			Help:     "Backend name (required).",
		},
	)

	parser.ExitOnHelp(true)

	err = parser.Parse(os.Args)
	if err == nil && *mappings != "" {
		bridgeMappings, err = parse.ParseMappings(*mappings)
		if err != nil {
			err = fmt.Errorf("[-m|--bridge-mappings]: %w", err)
		}
	}
	if encap != nil && *encap != "" && err == nil {
		encapIP, err = parse.ParseIP(*encap)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}

	err = logger.Init(os.Stdout, *logLevel)
	if err != nil {
		goto end
	}

	back = backends[*backendName]
	logger.Debugf("initializing %s backend", back.Name())
	err = back.Init(bridgeMappings, encapIP)
	if err != nil {
		goto end
	}
	logger.Infof("%s backend initialized", back.Name())

	north, err = ovn.Connect(*endpoint, back)
	if err != nil {
		goto end
	}
	logger.Infof("connected to ovn northbound db: %s", *endpoint)

	logger.Infof("received %s shutting down", <-sigs)
end:
	if back != nil {
		back.Shutdown()
	}
	if north != nil {
		north.Disconnect()
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}

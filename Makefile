# SPDX-Identifier: MIT
# Copyright (c) 2023 Robin Jarry

version = $(shell git describe --long --abbrev=12 --tags --dirty 2>/dev/null || echo 0.1)
src = $(shell find * -type f -name '*.go') go.mod go.sum
go_ldflags :=
go_ldflags += -X main.version=$(version)

all: ovn-nb-agent

ovn-nb-agent: $(src)
	go build -trimpath -ldflags='$(go_ldflags)' -o $@

sandbox: ovn-nb-agent
	@tmp=`mktemp -d`; \
	trap "rm -rf $$tmp" EXIT; \
	set -ex; \
	ovsdb-tool create $$tmp/db schema/northbound/schema.json; \
	ovsdb-server --no-chdir --pidfile=$$tmp/pid \
		-vconsole:off -vsyslog:off --log-file=$$tmp/log \
		--remote=db:OVN_Northbound,NB_Global,connections \
		--unixctl=$$tmp/ctl --remote=punix:$$tmp/ovsdb $$tmp/db & \
	sleep 0.2; \
	./ovn-nb-agent -b linux -l trace -n unix:$$tmp/ovsdb \
		-m external:enp1s0 -e 192.168.122.233; \
	pkill --pidfile=$$tmp/pid; \
	wait

# SPDX-Identifier: MIT
# Copyright (c) 2023 Robin Jarry

version = $(shell git describe --long --abbrev=12 --tags --dirty 2>/dev/null || echo 0.1)
src = $(shell find * -type f -name '*.go') go.mod go.sum
go_ldflags :=
go_ldflags += -X main.version=$(version)

all: ovn-nb-agent

ovn-nb-agent: $(src)
	go build -trimpath -ldflags='$(go_ldflags)' -o $@

// SPDX-License-Identifier: MIT
// Copyright (c) 2023 Robin Jarry

package linux

import (
	"errors"
	"os/exec"
	"syscall"

	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"golang.org/x/sys/unix"
)

// Create a netns and return a netlink socket to interacting with it.
func EnsureNetns(name string) (netns.NsHandle, *netlink.Handle, error) {
	var ns netns.NsHandle
	var nl *netlink.Handle
	var err error

	ns, err = netns.GetFromName(name)
	if errors.Is(err, syscall.ENOENT) {
		// To ensure the netns is properly configured, it must be
		// created from the root namespace (PID 1).
		// See https://serverfault.com/a/961592 for more details.
		// The mount namespaces must also be reset but this can only be
		// done from single threaded processes which is not the case
		// for go programs. Use nsenter to switch back to all the
		// namespaces (net+mnt) of PID 1 before creating the netns.
		cmd := exec.Command(
			"nsenter", "-t", "1", "-a", "ip", "netns", "add", name)
		err = cmd.Run()
		if err != nil {
			goto out
		}
		ns, err = netns.GetFromName(name)
	}
	if err != nil {
		goto out
	}
	nl, err = netlink.NewHandleAt(ns, unix.NETLINK_ROUTE)

out:
	return ns, nl, err
}

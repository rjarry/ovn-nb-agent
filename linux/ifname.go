// SPDX-License-Identifier: MIT
// Copyright (c) 2023 Robin Jarry

package linux

import "golang.org/x/sys/unix"

const MAX_LENGTH = unix.IFNAMSIZ - 1

func CompactIfName(name, prefix, suffix string) string {
	switch {
	case len(prefix)+len(name)+len(suffix) < MAX_LENGTH:
		return prefix + name + suffix
	case prefix == "" && suffix == "":
		return name[:MAX_LENGTH]
	case prefix == "":
		return name[:MAX_LENGTH-len(suffix)] + suffix
	case suffix == "":
		return prefix + name[len(name)-MAX_LENGTH+len(prefix):]
	}
	return prefix + name[:MAX_LENGTH-len(prefix)-len(suffix)] + suffix
}

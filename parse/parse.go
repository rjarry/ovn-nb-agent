package parse

import (
	"errors"
	"fmt"
	"net/netip"
	"regexp"
)

var (
	mappingsRe = regexp.MustCompile(`^[^:,\s]+:[^:,\s]+(,[^:,\s]+:[^:,\s]+)*$`)
	mappingRe  = regexp.MustCompile(`([^:,\s]+):([^:,\s]+)`)
)

// Parse a string in the format:
//
//	"foo:bar,baz:buz,foz:rab"
//
// Into a map:
//
//	{"foo": "bar", "baz": "buz", "foz": "rab"}
func ParseMappings(mappings string) (map[string]string, error) {
	if !mappingRe.MatchString(mappings) {
		return nil, fmt.Errorf("invalid mapping value: %q", mappings)
	}
	ret := make(map[string]string)
	for _, match := range mappingRe.FindAllStringSubmatch(mappings, -1) {
		ret[match[1]] = match[2]
	}
	return ret, nil
}

func ParseIP(s string) (netip.Addr, error) {
	ip, err := netip.ParseAddr(s)
	if err != nil {
		return ip, err
	}
	if !ip.IsGlobalUnicast() {
		return ip, errors.New("unicast address required")
	}
	return ip, nil
}

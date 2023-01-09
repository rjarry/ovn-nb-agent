// Code generated by "libovsdb.modelgen"
// DO NOT EDIT.

package schema

// LogicalSwitchPort defines an object in Logical_Switch_Port table
type LogicalSwitchPort struct {
	UUID             string            `ovsdb:"_uuid"`
	Addresses        []string          `ovsdb:"addresses"`
	Dhcpv4Options    []string          `ovsdb:"dhcpv4_options"`
	Dhcpv6Options    []string          `ovsdb:"dhcpv6_options"`
	DynamicAddresses []string          `ovsdb:"dynamic_addresses"`
	Enabled          []bool            `ovsdb:"enabled"`
	ExternalIDs      map[string]string `ovsdb:"external_ids"`
	HaChassisGroup   []string          `ovsdb:"ha_chassis_group"`
	MirrorRules      []string          `ovsdb:"mirror_rules"`
	Name             string            `ovsdb:"name"`
	Options          map[string]string `ovsdb:"options"`
	ParentName       []string          `ovsdb:"parent_name"`
	PortSecurity     []string          `ovsdb:"port_security"`
	Tag              []int             `ovsdb:"tag"`
	TagRequest       []int             `ovsdb:"tag_request"`
	Type             string            `ovsdb:"type"`
	Up               []bool            `ovsdb:"up"`
}
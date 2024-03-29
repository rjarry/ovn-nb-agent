// Code generated by "libovsdb.modelgen"
// DO NOT EDIT.

package ovn

// DHCPOptions defines an object in DHCP_Options table
type DHCPOptions struct {
	UUID        string            `ovsdb:"_uuid"`
	Cidr        string            `ovsdb:"cidr"`
	ExternalIDs map[string]string `ovsdb:"external_ids"`
	Options     map[string]string `ovsdb:"options"`
}

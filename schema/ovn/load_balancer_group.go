// Code generated by "libovsdb.modelgen"
// DO NOT EDIT.

package ovn

// LoadBalancerGroup defines an object in Load_Balancer_Group table
type LoadBalancerGroup struct {
	UUID         string   `ovsdb:"_uuid"`
	LoadBalancer []string `ovsdb:"load_balancer"`
	Name         string   `ovsdb:"name"`
}

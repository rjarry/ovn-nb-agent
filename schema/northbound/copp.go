// Code generated by "libovsdb.modelgen"
// DO NOT EDIT.

package northbound

// Copp defines an object in Copp table
type Copp struct {
	UUID        string            `ovsdb:"_uuid"`
	ExternalIDs map[string]string `ovsdb:"external_ids"`
	Meters      map[string]string `ovsdb:"meters"`
	Name        string            `ovsdb:"name"`
}
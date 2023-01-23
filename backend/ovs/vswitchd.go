package ovs

import (
	"context"
	"fmt"
	"net/netip"
	"time"

	"github.com/google/uuid"
	"github.com/ovn-org/libovsdb/client"
	"github.com/ovn-org/libovsdb/model"
	"github.com/ovn-org/libovsdb/ovsdb"
	"github.com/rjarry/ovn-nb-agent/parse"
	"github.com/rjarry/ovn-nb-agent/schema/ovs"
)

type Vswitchd struct {
	ovs   client.Client
	cache VswitchdCache

	rootUUID       string
	datapath       string
	encapAddress   netip.Addr
	stagingBridge  string
	bridgeMappings map[string]string
}

const (
	DATAPATH_SYSTEM string = "system"
	DATAPATH_NETDEV        = "netdev"
)

func (v *Vswitchd) Connect(endpoint string) error {
	schema, err := ovs.FullDatabaseModel()
	if err != nil {
		return err
	}
	c, err := client.NewOVSDBClient(schema, client.WithEndpoint(endpoint))
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	err = c.Connect(ctx)
	if err != nil {
		return err
	}

	err = v.cache.Start(c)
	if err != nil {
		return err
	}
	v.ovs = c

	var o ovs.OpenvSwitch
	err = c.Get(&o)
	if err != nil {
		return err
	}

	v.rootUUID = o.UUID
	s, ok := o.ExternalIDs["ovn-encap-ip"]
	if ok && s != "" {
		v.encapAddress, err = parse.ParseIP(s)
		if err != nil {
			return err
		}
	}
	s, ok = o.ExternalIDs["ovn-bridge-mappings"]
	if ok && s != "" {
		v.bridgeMappings, err = parse.ParseMappings(s)
		if err != nil {
			return err
		}
	}
	v.stagingBridge = "br-int" // default value
	s, ok = o.ExternalIDs["ovn-bridge"]
	if ok && s != "" {
		v.stagingBridge = s
	}
	s, ok = o.ExternalIDs["ovn-bridge-datapath-type"]
	if ok && s != "" {
		switch s {
		case DATAPATH_SYSTEM, DATAPATH_NETDEV:
			v.datapath = s
		default:
			return fmt.Errorf(
				"invalid ovn-bridge-datapath-type value %q", s)
		}
	}

	br := ovs.Bridge{Name: v.stagingBridge}
	err = c.Get(&br)
	if err == nil {
		if v.datapath != "" && v.datapath != br.DatapathType {
			br.DatapathType = v.datapath
			err = v.UpdateBridge(&br)
		}
	} else {
		_, err = v.AddBridge(v.stagingBridge)
	}
	if err != nil {
		return err
	}

	return nil
}

func (v *Vswitchd) Disconnect() {
	if v.ovs != nil && v.ovs.Connected() {
		v.cache.Stop(v.ovs)
		v.ovs.Disconnect()
	}
}

func (v *Vswitchd) Cache() *VswitchdCache {
	return &v.cache
}

func (v *Vswitchd) AddBridge(name string) (*ovs.Bridge, error) {
	br := ovs.Bridge{
		UUID:         uuid.NewString(),
		Name:         name,
		DatapathType: v.datapath,
	}
	operations, err := v.ovs.Create(&br)
	if err != nil {
		return nil, err
	}
	root := ovs.OpenvSwitch{UUID: v.rootUUID}
	mut, err := v.ovs.Where(&root).Mutate(&root, model.Mutation{
		Field:   &root.Bridges,
		Mutator: "insert",
		Value:   []string{br.UUID},
	})
	if err != nil {
		return nil, err
	}
	operations = append(operations, mut...)

	reply, err := v.ovs.Transact(operations...)
	if err != nil {
		return nil, err
	}
	_, err = ovsdb.CheckOperationResults(reply, operations)
	return &br, err
}

func (v *Vswitchd) UpdateBridge(br *ovs.Bridge) error {
	operations, err := v.ovs.Where(br).Update(br)
	if err != nil {
		return err
	}
	reply, err := v.ovs.Transact(operations...)
	if err != nil {
		return err
	}
	_, err = ovsdb.CheckOperationResults(reply, operations)
	return err
}

func (v *Vswitchd) AddPort(name string, br *ovs.Bridge) (*ovs.Interface, error) {
	return nil, nil
}

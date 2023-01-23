package ovs

import (
	"sync"

	"github.com/ovn-org/libovsdb/cache"
	"github.com/ovn-org/libovsdb/client"
	"github.com/ovn-org/libovsdb/model"
	"github.com/rjarry/ovn-nb-agent/schema/ovs"
)

type VswitchdCache struct {
	sync.Mutex
	monitor  string
	ifaces   map[string]*ovs.Interface
	ifacesCb map[string]func(*ovs.Interface)
	bridges  map[string]*ovs.Bridge
}

func (c *VswitchdCache) Start(ovs client.Client) error {
	var err error
	ovs.Cache().AddEventHandler(&cache.EventHandlerFuncs{
		AddFunc:    c.onAdd,
		UpdateFunc: c.onUpdate,
		DeleteFunc: c.onDelete,
	})
	c.monitor, err = ovs.MonitorAll()
	if err != nil {
		return err
	}
	return nil
}

func (c *VswitchdCache) Stop(ovs client.Client) error {
	return ovs.MonitorCancel(c.monitor)
}

func (c *VswitchdCache) onAdd(table string, obj model.Model) {
	switch obj := obj.(type) {
	case *ovs.Bridge:
		c.Lock()
		c.Unlock()
	case *ovs.Interface:
		if ifaceId, ok := obj.ExternalIDs["iface-id"]; ok {
			c.Lock()
			defer c.Unlock()
			c.ifaces[ifaceId] = obj
			if cb, ok := c.ifacesCb[ifaceId]; ok {
				go cb(obj)
				delete(c.ifacesCb, ifaceId)
			}
		}
	}
}

func (c *VswitchdCache) onUpdate(table string, _, obj model.Model) {
	switch obj := obj.(type) {
	case *ovs.Bridge:
		c.Lock()
		c.Unlock()
	case *ovs.Interface:
		if ifaceId, ok := obj.ExternalIDs["iface-id"]; ok {
			c.Lock()
			c.ifaces[ifaceId] = obj
			if cb, ok := c.ifacesCb[ifaceId]; ok {
				go cb(obj)
				delete(c.ifacesCb, ifaceId)
			}
			c.Unlock()
		}
	}
}

func (c *VswitchdCache) onDelete(table string, obj model.Model) {
	switch obj := obj.(type) {
	case *ovs.Bridge:
		c.Lock()
		c.Unlock()
	case *ovs.Interface:
		if ifaceId, ok := obj.ExternalIDs["iface-id"]; ok {
			c.Lock()
			delete(c.ifaces, ifaceId)
			c.Unlock()
		}
	}
}

func (c *VswitchdCache) IfaceFromId(ifaceId string) *ovs.Interface {
	c.Lock()
	defer c.Unlock()
	return c.ifaces[ifaceId]
}

func (c *VswitchdCache) IfaceIdCallback(ifaceId string, cb func(*ovs.Interface)) {
	c.Lock()
	defer c.Unlock()
	if iface, ok := c.ifaces[ifaceId]; ok {
		go cb(iface)
	} else {
		c.ifacesCb[ifaceId] = cb
	}
}

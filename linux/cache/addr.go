package cache

import (
	"net/netip"
	"sync"

	"github.com/rjarry/ovn-nb-agent/logger"
	"github.com/vishvananda/netlink"
)

type AddrCache struct {
	sync.Mutex

	done chan struct{}

	cache map[netip.Addr]int
}

func (c *AddrCache) Start() error {
	c.done = make(chan struct{})
	c.cache = make(map[netip.Addr]int)
	updates := make(chan netlink.AddrUpdate)

	go c.monitor(updates)

	return netlink.AddrSubscribeWithOptions(
		updates, c.done, netlink.AddrSubscribeOptions{
			ListExisting: true,
		},
	)
}

func (c *AddrCache) Stop() {
	if c.done != nil {
		close(c.done)
		c.done = nil
	}
}

func (c *AddrCache) GetLinkIndex(ip netip.Addr) (int, bool) {
	c.Lock()
	defer c.Unlock()
	index, found := c.cache[ip]
	return index, found
}

func (c *AddrCache) monitor(updates chan netlink.AddrUpdate) {
	for update := range updates {
		addr, _ := netip.AddrFromSlice([]byte(update.LinkAddress.IP))
		index := update.LinkIndex
		c.Lock()
		if update.NewAddr {
			logger.Tracef("newaddr: %s link %d", addr, index)
			c.cache[addr] = index
		} else {
			logger.Tracef("deladdr: %s link %d", addr, index)
			delete(c.cache, addr)
		}
		c.Unlock()
	}
}

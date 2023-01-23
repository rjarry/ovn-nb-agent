package cache

import (
	"context"
	"sync"

	"github.com/rjarry/ovn-nb-agent/logger"
	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
)

type LinkCache struct {
	sync.Mutex

	done chan struct{}

	// caches
	byIndex map[int32]netlink.Link
	byName  map[string]netlink.Link
	byAlias map[string]netlink.Link

	nameCallbacks  map[string]func(netlink.Link)
	aliasCallbacks map[string]func(netlink.Link)
}

func (c *LinkCache) Start() error {
	c.done = make(chan struct{})
	c.nameCallbacks = make(map[string]func(netlink.Link))
	c.aliasCallbacks = make(map[string]func(netlink.Link))
	c.byIndex = make(map[int32]netlink.Link)
	c.byName = make(map[string]netlink.Link)
	c.byAlias = make(map[string]netlink.Link)
	updates := make(chan netlink.LinkUpdate)

	go c.monitor(updates)

	return netlink.LinkSubscribeWithOptions(
		updates, c.done, netlink.LinkSubscribeOptions{
			ListExisting: true,
		},
	)
}

func (c *LinkCache) Stop() {
	if c.done != nil {
		close(c.done)
		c.done = nil
	}
}

func (c *LinkCache) monitor(updates chan netlink.LinkUpdate) {
	for update := range updates {
		index := update.Index
		name := update.Link.Attrs().Name
		alias := update.Link.Attrs().Alias
		kind := update.Link.Type()
		c.Lock()
		if link, ok := c.byIndex[index]; ok {
			// handle renames
			delete(c.byName, link.Attrs().Name)
			delete(c.byAlias, link.Attrs().Alias)
		}
		switch update.Header.Type {
		case unix.RTM_NEWLINK, unix.RTM_SETLINK:
			logger.Tracef("newlink: %s alias %q (%s)", name, alias, kind)
			c.byIndex[index] = update.Link
			c.byName[name] = update.Link
			if callback, ok := c.nameCallbacks[name]; ok {
				go callback(update.Link)
				delete(c.nameCallbacks, name)
			}
			if alias != "" {
				c.byAlias[alias] = update.Link
				if callback, ok := c.aliasCallbacks[alias]; ok {
					go callback(update.Link)
					delete(c.aliasCallbacks, alias)
				}
			}
		case unix.RTM_DELLINK:
			logger.Tracef("dellink: %s alias %q (%s)", name, alias, kind)
			delete(c.byIndex, index)
			delete(c.byName, name)
			delete(c.byAlias, alias)
		}
		c.Unlock()
	}
}

func (c *LinkCache) FromName(name string) netlink.Link {
	c.Lock()
	defer c.Unlock()
	return c.byName[name]
}

func (c *LinkCache) FromAlias(alias string) netlink.Link {
	c.Lock()
	defer c.Unlock()
	return c.byAlias[alias]
}

func (c *LinkCache) AliasCallback(alias string, cb func(netlink.Link)) {
	c.Lock()
	defer c.Unlock()
	if link, ok := c.byAlias[alias]; ok {
		go cb(link)
	} else {
		c.aliasCallbacks[alias] = cb
	}
}

func (c *LinkCache) NameCallback(name string, cb func(netlink.Link)) {
	c.Lock()
	defer c.Unlock()
	if link, ok := c.byName[name]; ok {
		go cb(link)
	} else {
		c.nameCallbacks[name] = cb
	}
}

func (c *LinkCache) WaitName(ctx context.Context, name string) (netlink.Link, error) {
	var link netlink.Link
	var err error

	c.Lock()
	if link, ok := c.byName[name]; ok {
		c.Unlock()
		return link, nil
	}

	done := make(chan struct{})

	c.nameCallbacks[name] = func(l netlink.Link) {
		link = l
		close(done)
	}

	c.Unlock()

	select {
	case <-done:
		break
	case <-ctx.Done():
		c.Lock()
		err = ctx.Err()
		c.Unlock()
	}

	return link, err
}

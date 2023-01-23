package linux

import (
	"context"
	"fmt"
	"hash/fnv"
	"net"
	"net/netip"
	"time"

	"github.com/rjarry/ovn-nb-agent/linux"
	"github.com/rjarry/ovn-nb-agent/logger"
	"github.com/rjarry/ovn-nb-agent/models"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netlink/nl"
	"golang.org/x/sys/unix"
)

func dhcpIdentifier(bridge *netlink.Bridge, cidr *netip.Prefix) string {
	h := fnv.New32a()
	i := bridge.Index
	h.Write([]byte{byte(i >> 24), byte(i >> 16), byte(i >> 8), byte(i)})
	h.Write(bridge.HardwareAddr)
	h.Write([]byte(cidr.String()))
	return fmt.Sprintf("dhcp-%08x", h.Sum32())
}

func (b *LinuxBackend) configureDHCP(bridge *netlink.Bridge, zone *models.DHCPZone) error {
	name := dhcpIdentifier(bridge, zone.Opts.CIDR())
	veth, err := linux.PrepareDhcpNetns(zone, name)
	if err != nil {
		return err
	}
	logger.Debugf("set veth %s master", name)
	if err = netlink.LinkSetMaster(veth, bridge); err != nil {
		return err
	}
	logger.Debugf("set veth %s isolated", name)
	if err = netlink.LinkSetIsolated(veth, true); err != nil {
		return err
	}
	logger.Debugf("disable veth %s mac learning", name)
	return netlink.LinkSetLearning(veth, false)
}

var (
	broadcastMac = net.HardwareAddr([]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
	udp          = nl.IPPROTO_UDP
)

func (b *LinuxBackend) redirectDhcpRequests(
	port netlink.Link, bridge *netlink.Bridge, opts *models.DHCPOptions,
) error {
	name := dhcpIdentifier(bridge, opts.CIDR())

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	veth, err := b.linkCache.WaitName(ctx, name)
	if err != nil {
		return fmt.Errorf("no such link %q: %s", name, err)
	}

	qdisc := netlink.MakeHandle(0xffff, 0)

	ingress := new(netlink.Ingress)
	ingress.Handle = qdisc
	ingress.Parent = netlink.HANDLE_INGRESS
	ingress.LinkIndex = port.Attrs().Index
	err = netlink.QdiscReplace(ingress)
	if err != nil {
		return err
	}
	err = b.addDhcpFilter(qdisc, unix.ETH_P_IP, broadcastMac, port, veth)
	if err != nil {
		return err
	}
	err = b.addDhcpFilter(qdisc, unix.ETH_P_IP, *opts.ServerMac(), port, veth)
	if err != nil {
		return err
	}
	//err = b.addDhcpFilter(qdisc, unix.ETH_P_IPV6, broadcastMac, port, veth)
	//if err != nil {
	//	return err
	//}
	//err = b.addDhcpFilter(qdisc, unix.ETH_P_IPV6, *opts.ServerMac(), port, veth)
	//if err != nil {
	//	return err
	//}
	return nil
}

func (b *LinuxBackend) addDhcpFilter(
	qdisc uint32, protocol uint16, mac net.HardwareAddr, in, out netlink.Link,
) error {
	filter := new(netlink.Flower)
	filter.Parent = qdisc
	filter.Protocol = protocol
	filter.Priority = 1
	filter.LinkIndex = in.Attrs().Index
	filter.DestEth = mac
	filter.DestEthMask = broadcastMac
	filter.IPProto = &udp
	filter.DestPort = linux.DHCP_SERVER_PORT
	filter.Actions = []netlink.Action{
		&netlink.MirredAction{
			ActionAttrs: netlink.ActionAttrs{
				Action: netlink.TC_ACT_STOLEN,
			},
			MirredAction: netlink.TCA_EGRESS_REDIR,
			Ifindex:      out.Attrs().Index,
		},
	}
	logger.Tracef("tc filter add -> %#v", filter)
	return netlink.FilterAdd(filter)
}

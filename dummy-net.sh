#!/bin/sh

set +x

ip() {
	echo + ip "$@" >&2
	if [ -n "$netns" ]; then
		set -- -n "$netns" "$@"
	fi
	command ip -color=auto "$@"
}
bridge() {
	echo + bridge "$@" >&2
	if [ -n "$netns" ]; then
		set -- -n "$netns" "$@"
	fi
	command bridge -color=auto "$@"
}
case "$TMUX" in
*tmate*)
	alias tmux=tmate
	;;
esac

if [ -n "$TMUX" ]; then
	for win in compute{1,2} guest{1,2,3,4,5,6}; do
		tmux kill-window -t $win
	done
fi

netns=

ip link del br-phy
ip link del p0
ip link del p1

ip -all netns del

if [ "$1" = "-c" ]; then
	exit
fi

ip link add br-phy type bridge vlan_filtering 1
bridge vlan del dev br-phy vid 1 self
ip link set br-phy up
ip link add p0 master br-phy type veth peer name phy0
ip link add p1 master br-phy type veth peer name phy1
ip link set p0 up
ip link set p1 up
bridge vlan del dev p0 vid 1 master
bridge vlan del dev p1 vid 1 master
for vlan in 400 401 402 403 404 405 406 407 408 409; do
	bridge vlan add dev p0 vid $vlan master
	bridge vlan add dev p1 vid $vlan master
done
for v in 1 2 3 4 5 6; do
	ip link add vm$v type veth peer name guest$v
done
ip netns add compute1
ip link set phy0 netns compute1
for v in 1 2 6; do
	ip netns add guest$v
	ip link set vm$v netns compute1
	ip link set guest$v netns guest$v
done
ip netns add compute2
ip link set phy1 netns compute2
for v in 3 4 5; do
	ip netns add guest$v
	ip link set vm$v netns compute2
	ip link set guest$v netns guest$v
done

netns=compute1
echo "netns $netns ========================================"

ip link set lo up
ip link set phy0 up
ip link add link phy0 name tenants type vlan id 404
ip link set tenants up
ip addr add 172.16.13.1/24 dev tenants
ip link add br-internal type bridge vlan_filtering 1
bridge vlan del dev br-internal vid 1 self
ip link set br-internal up
ip link set vm1 up master br-internal
ip link set vm2 up master br-internal
ip link add vx-internal type vxlan id 1337 dstport 4789 group 239.0.13.37 dev tenants ttl 5
ip link set vx-internal up
ip link set vx-internal master br-internal
ip link add br-external type bridge vlan_filtering 1
bridge vlan del dev br-external vid 1 self
ip link set br-external up
ip link add link phy0 external master br-external type vlan id 407
ip link set external up
ip link set vm6 up master br-external

netns=compute2
echo "netns $netns ========================================"

ip link set lo up
ip link set phy1 up
ip link add link phy1 name tenants type vlan id 404
ip link set tenants up
ip addr add 172.16.13.2/24 dev tenants
ip link add br-internal type bridge vlan_filtering 1
bridge vlan del dev br-internal vid 1 self
ip link set br-internal up
ip link set vm3 up master br-internal
ip link set vm4 up master br-internal
ip link add vx-internal type vxlan id 1337 dstport 4789 group 239.0.13.37 dev tenants ttl 5
ip link set vx-internal master br-internal
ip link set vx-internal up
ip link add br-external type bridge vlan_filtering 1
bridge vlan del dev br-external vid 1 self
ip link set br-external up
ip link add link phy1 external master br-external type vlan id 407
ip link set external up
ip link set vm5 up master br-external

for v in 1 2 3 4 5 6; do
	netns=guest$v
	ip link set lo up
	ip link set guest$v up
	if [ "$v" -lt 5 ]; then
		net=10.16.0
	else
		net=10.99.0
	fi
	ip addr add $net.$v/24 dev guest$v
done

if [ -n "$TMUX" ]; then
	for ns in compute{1,2} guest{1,2,3,4,5,6}; do
		tmux new-window -d -n $ns ip netns exec $ns bash -li
	done
fi

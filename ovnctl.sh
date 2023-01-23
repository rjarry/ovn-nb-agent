#!/bin/bash

export OVN_NB_DB="tcp:172.16.254.254:6642"


die() {
	printf "error: %s\n" "$1" >&2
	exit 1
}

mac() {
	local host_or_ip="$1"
	echo "$host_or_ip" | md5sum | \
		sed 's/^\(..\)\(..\)\(..\)\(..\)\(..\).*$/02:\1:\2:\3:\4:\5/'
}

dhcp_uuid() {
	local switch="$1"
	ovn-nbctl -f table -d bare --no-headings list DHCP_Options | \
		sed -rne "s/^([a-f0-9-]+)[[:space:]]+.*switch=$switch\>.*$/\\1/p"
}

dhcp_cidr() {
	local switch="$1"
	ovn-nbctl -f table -d bare --no-headings list DHCP_Options | \
		sed -rne "s/[a-f0-9-]+[[:space:]]+([0-9\./]+)\>.*[[:space:]]+switch=$switch\>.*$/\\1/p"
}

switch() {
	local name="$1"
	local cidr="$2"
	local provider="$3"
	local mtu=1442
	if [ -z "$name" ]; then
		die "<NAME> is required
usage: $0 switch <NAME> <SUBNET.0/PREFIX> [<NETWORK=VLAN>]"
	fi
	if ! [[ "$cidr" =~ ^[0-9]+\.[0-9]+\.[0-9]+\.0/[0-9]+$ ]]; then
		die "
usage: $0 switch <NAME> <SUBNET.0/PREFIX> [<NETWORK=VLAN>]"
	fi
	if [ -n "$provider" ] && ! [[ "$provider" =~ ^[[:alnum:]]+=[0-9]+$ ]]; then
		die "invalid <NETWORK=VLAN> argument
usage: $0 switch <NAME> <SUBNET/PREFIX> [<NETWORK=VLAN>]"
	fi
	local subnet=${cidr%/*}
	local prefix=${cidr#*/}
	local gateway=${subnet%.0}.1
	local meta=${subnet%.0}.2
	local gw_mac=$(mac $gateway)
	local meta_mac=$(mac $meta)

	ovn-nbctl dhcp-options-create "$cidr" "switch=$name"
	local dhcp_uuid=$(dhcp_uuid "$name")
	ovn-nbctl dhcp-options-set-options "$dhcp_uuid" \
		classless_static_route="{169.254.169.254/32,$meta, 0.0.0.0/0,$gateway}" \
		dns_server="{8.8.8.8, 4.4.4.4}" \
		lease_time=43200 \
		mtu=$mtu \
		router=$gateway \
		server_id=$gateway \
		server_mac=$gw_mac

	ovn-nbctl ls-add "$name" -- \
		set Logical_Switch "$name" external-ids:neutron=mtu\="$mtu" -- \
		set Logical_Switch "$name" external-ids:neutron=network_name\="$name" -- \
		set Logical_Switch "$name" other-config:mcast_flood_unregistered\=false -- \
		set Logical_Switch "$name" other-config:mcast_snoop\=false -- \
		set Logical_Switch "$name" other-config:vlan-passthru\=false

	if [ -n "$provider" ]; then
		local network="${provider%=*}"
		local vlan="${provider#*=}"
		port_name=provnet-$(uuidgen)
		ovn-nbctl lsp-add "$name" "$port_name" "" "$vlan" -- \
			lsp-set-type "$port_name" localnet -- \
			lsp-set-enabled "$port_name" enabled -- \
			lsp-set-options "$port_name" network_name="$network" \
				mcast_flood=false mcast_flood_reports=true
	else
		local port_name=$(uuidgen)
		ovn-nbctl lsp-add "$name" "$port_name" -- \
			lsp-set-type "$port_name" localport -- \
			lsp-set-addresses "$port_name" "$meta_mac $meta" -- \
			lsp-set-enabled "$port_name" enabled -- \
			lsp-set-options "$port_name" requested-chassis=""
	fi
}

port() {
	local switch="$1"
	local port="$2"

	if [ -z "$switch" ] || [ -z "$port" ]; then
		die "<SWITCH> and <PORT> are required
usage: $0 port <SWITCH> <PORT>"
	fi

	local dhcp_uuid=$(dhcp_uuid "$switch")
	local dhcp_cidr=$(dhcp_cidr "$switch")
	local subnet=${dhcp_cidr%/*}
	local ip="${subnet%.0}.$(((RANDOM % 252) + 3))"
	local mac=$(mac $ip)

	ovn-nbctl lsp-add "$switch" "$port" -- \
		lsp-set-addresses "$port" "$mac $ip" -- \
		lsp-set-enabled "$port" enabled -- \
		lsp-set-options "$port" mcast_flood_reports=true -- \
		lsp-set-dhcpv4-options "$port" "$dhcp_uuid"
}

start() {
	if [ $# -lt 3 ]; then
		die "invalid arguments
usage: $0 start <COMPUTE_ID> <VM_ID> <PORT_UUID> [...<PORT_UUID>]"
	fi
	local compute="compute$1"
	shift
	local vm="vm$1"
	shift

	# ensure compute namespace exists
	if ! ip netns pids "$compute" >/dev/null; then
		die "COMPUTE_ID ${compute#compute} does not exist"
	fi
	# ensure vm namespace does not exist
	if ip netns pids "$vm" >/dev/null 2>&1; then
		die "VM_ID ${vm#vm} already started"
	fi

	ip netns add "$vm"
	ip -n "$vm" link set lo up

	i=0
	for port in "$@"; do
		local address=$(ovn-nbctl lsp-get-addresses "$port")
		local mac=${address% *}
		ip link add "port$i" type veth peer name "$vm.$i"
		ip link set "port$i" netns "$vm"
		ip link set "$vm.$i" netns "$compute"
		ip -n "$vm" link set "port$i" address "$mac"
		ip -n "$vm" link set "port$i" up
		ip netns exec "$vm" dhclient -nw --no-pid -cf /dev/null \
			-lf "/tmp/dhcp-leases.$vm.port$i"
		ip -n "$compute" link set "$vm.$i" up
		ip -n "$compute" link set "$vm.$i" alias "$port"
		i=$((i+1))
	done

	if [ -n "$TMUX" ]; then
		tmux new-window -d -n "$vm" ip netns exec "$vm" bash -li || :
	fi
}

stop() {
	if [ $# -ne 1 ]; then
		die "invalid arguments
usage: $0 stop <VM_ID>"
	fi
	local vm="vm$1"
	if [ -n "$TMUX" ]; then
		tmux kill-window -t "$vm" || :
	fi
	for pid in $(ip netns pids "$vm"); do
		kill "$pid"
	done
	ip netns del "vm${1?vm_id}"
}

cmd=$1
shift

set -e

case "$cmd" in
switch)
	switch "$@"
	;;
port)
	port "$@"
	;;
vm-start)
	start "$@"
	;;
vm-stop)
	stop "$@"
	;;
*|"")
	die "unknown command: $cmd (valid commands: switch port vm-start vm-stop)"
esac

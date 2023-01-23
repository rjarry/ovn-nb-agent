#!/bin/bash

set -e

num_computes=${1?num_computes}
tmp=$(mktemp -d)

cleanup() {
	pkill dhclient || : 2>/dev/null
	if [ -n "$TMUX" ]; then
		for w in $(tmux list-windows -F '#W'); do
			case "$w" in
			compute[0-9]*|vm[0-9]*)
				tmux kill-window -t "$w" || :
				;;
			esac
		done
	fi
	for c in $(seq $num_computes); do
		ip link del compute$c 2>/dev/null || :
	done
	ip link del br-computes 2>/dev/null || :
	for ns in /run/netns/*; do
		if [ -d "$ns" ]; then
			pids=$(ip netns pids $(basename "$ns") 2>/dev/null) || :
			for pid in $pids; do
				kill "$pid" 2>/dev/null || :
			done
		fi
	done
	ip -all netns del 2>/dev/null || :
	if [ -z "$1" ]; then
		jobs -p | xargs kill 2>/dev/null || :
		rm -rf -- "$tmp" 2>/dev/null || :
	fi
}

cleanup -n
ip link add br-computes type bridge
ip link set br-computes up
ip addr add 172.16.254.254/16 dev br-computes
trap cleanup EXIT

# start norhbound db server
ovsdb-tool create "$tmp/db" schema/ovn/schema.json
ovsdb-server --no-chdir --pidfile="$tmp/db.pid" \
	-vconsole:err -vsyslog:off -vfile:off \
	--remote=db:OVN_Northbound,NB_Global,connections \
	--unixctl="$tmp/ctl" --remote="ptcp:6642:172.16.254.254" "$tmp/db" &

sleep 2

for c in $(seq $num_computes); do
	ip link add compute$c master br-computes type veth peer name phy
	ip link set compute$c up
	ip netns add compute$c
	ip link set phy netns compute$c
	ip -n compute$c link set lo up
	ip -n compute$c link set phy up
	ip="172.16.$(((c >> 8) & 0xff)).$((c & 0xff))"
	ip -n compute$c addr add $ip/16 dev phy
	ip netns exec compute$c ./ovn-nb-agent -b linux -l trace \
		-n "tcp:172.16.254.254:6642" -m external:phy -e $ip 2>&1 \
		| sed "s/^/[compute$c] /" &
	if [ -n "$TMUX" ]; then
		tmux new-window -d -n "compute$c" ip netns exec "compute$c" bash -li || :
	fi
done

while netgraph -4 2>/dev/null | dot -Tsvg > net.svg; do sleep 2; done &

wait

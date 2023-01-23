# OVN Northbound Agent

Daemon program that connects to an OVN Northbound database (OvSdb) and applies
the configuration with alternate backends.

## Build

You need a golang compiler (`1.18` or later).

```
make
```


## Demo (as root)

Should not break your machine but there are no guarantees.

```
pip3 install linux-tools

tmux
tmux rename-window control
tmux new-window -dn sandbox ./sandbox.sh 4

./ovnctl.sh switch tenant 192.168.46.0/24
./ovnctl.sh port tenant tenant.p1
./ovnctl.sh port tenant tenant.p2

./ovnctl.sh vm-start 2 1 tenant.p1
./ovnctl.sh vm-start 4 2 tenant.p2

sleep 5

ip -n vm1 -br addr show
ip -n vm2 -br addr show
```

Ping between vm1 and vm2 should work using the DHCP assigned addresses going
through a VXLAN tunnel.

Open `net.svg` in firefox to see the network configuration.

## License

MIT

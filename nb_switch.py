LogicalSwitch = {
    "UUID": "c7e7725b-0e79-411d-844d-f87b0c4104f2",  # derive to get ifname
    "ExternalIDs": {
        "neutron:mtu": "1442",
        "neutron:network_name": "tenantA",  # ifalias for bridge interface
    },
    "Ports": [
        # vxlan/geneve endpoint (osef?)
        {
            "Addresses": ["fa:16:3e:29:a0:c3 192.168.201.2"],
            "Type": "localport",
        },
        # vm port (tap)
        {
            "Addresses": ["fa:16:3e:b0:a1:ee 192.168.201.170"],
            "Dhcpv4Options": [
                {
                    "Cidr": "192.168.201.0/24",
                    "Options": {
                        "classless_static_route": "{169.254.169.254/32,192.168.201.2, 0.0.0.0/0,192.168.201.1}",
                        "dns_server": "{10.38.5.26, 10.11.5.19}",
                        "lease_time": "43200",
                        "mtu": "1442",
                        "router": "192.168.201.1",
                        "server_id": "192.168.201.1",
                        "server_mac": "fa:16:3e:ee:5f:b3",
                    },
                },
            ],
            "ExternalIDs": { "neutron:cidrs": "192.168.201.170/24" },
            "Options": {
                "mcast_flood_reports": "true",
                "requested-chassis": "compute-0.localdomain",  # XXX: for us
            },
            "Type": "",
        },
    ],
}
LogicalSwitch = {
    "UUID": "4d5880be-af6b-4f80-8fc1-129c4afc2cf1",  # derive to get ifname
    "ExternalIDs": {
        "neutron:mtu": "1500",
        "neutron:network_name": "external",  # ifalias for bridge interface
    },
    "Ports": [
        # vxlan/geneve endpoint (osef?)
        {
            "Addresses": [ "fa:16:3e:f7:51:e9 192.168.122.2" ],
            "Type": "localport",
        },
        # router TODO: ??
        {
            "Addresses": ["router"],
            "ExternalIDs": { "neutron:cidrs": "192.168.122.234/24" },
            "Options": {
                "exclude-lb-vips-from-garp": "true",
                "nat-addresses": "router",
                "router-port": "lrp-2f4c2d16-2e34-42e1-a8f1-d2f7fa12f88f",
            },
            "Type": "router",
        },
        # provider network
        {
            "Options": {
                "mcast_flood": "false",
                "mcast_flood_reports": "true",
                "network_name": "external",  # bridge mappings for phy connection
            },
            "Tag": [406],
            "Type": "localnet",
        },
        # vm port (tap)
        {
            "Addresses": [
                "fa:16:3e:e8:d1:0f 192.168.122.62",
                "unknown",
            ],
            "Dhcpv4Options": [
                {
                    "Cidr": "192.168.122.0/24",
                    "Options": {
                        "classless_static_route": "{169.254.169.254/32,192.168.122.2, 0.0.0.0/0,192.168.122.1}",
                        "dns_server": "{10.38.5.26, 10.11.5.19}",
                        "lease_time": "43200",
                        "mtu": "1500",
                        "router": "192.168.122.1",
                        "server_id": "192.168.122.1",
                        "server_mac": "fa:16:3e:00:d4:eb",
                    },
                },
            ],
            "ExternalIDs": { "neutron:cidrs": "192.168.122.62/24" },
            "Options": {
                "mcast_flood_reports": "true",
                "requested-chassis": "compute-0.localdomain", # XXX: for us
            },
            "Type": "",
        },
    ],
}
LogicalSwitch = {
    "UUID": "d36f3b9f-e727-4df0-b8f0-0bb198422226",
    "ExternalIDs": {
        "neutron:mtu": "1442",
        "neutron:network_name": "internal",
    },
    "Ports": [
        # router TODO: ??
        {
            "Addresses": ["router"],
            "ExternalIDs": { "neutron:cidrs": "192.168.200.1/24" },
            "Options": {
                "router-port": "lrp-fb361220-fd01-4c8b-a227-5d1945ed8945",
            },
            "Type": "router",
        },
        # vm port (tap)
        {
            "Addresses": [ "fa:16:3e:eb:31:93 192.168.200.90" ],
            "Dhcpv4Options": [
                {
                    "Cidr": "192.168.200.0/24",
                    "Options": {
                        "classless_static_route": "{169.254.169.254/32,192.168.200.2, 0.0.0.0/0,192.168.200.1}",
                        "dns_server": "{10.38.5.26}",
                        "lease_time": "43200",
                        "mtu": "1442",
                        "router": "192.168.200.1",
                        "server_id": "192.168.200.1",
                        "server_mac": "fa:16:3e:67:4e:25",
                    },
                },
            ],
            "ExternalIDs": { "neutron:cidrs": "192.168.200.90/24" },
            "Options": {
                "mcast_flood_reports": "true",
                "requested-chassis": "compute-0.localdomain", # XXX: for us
            },
            "Type": "",
        },
        # vxlan/geneve endpoint (osef?)
        {
            "Addresses": [ "fa:16:3e:15:ef:1f 192.168.200.2" ],
            "ExternalIDs": { "neutron:cidrs": "192.168.200.2/24" },
        },
    ],
}
LogicalSwitch = {
    "ExternalIDs": {
        "neutron:mtu": "1500",
        "neutron:network_name": "tenantB",
    },
    "Ports": [
        # vm port (vhost)
        {
            "Addresses": [ "fa:16:3e:2d:83:84 192.168.202.240" ],
            "Dhcpv4Options": [
                "85b44aa7-59a2-4306-bf88-2f4028ecc5e7",
            ],
            "ExternalIDs": {
                "neutron:cidrs": "192.168.202.240/24",
                "neutron:device_owner": "",
                "neutron:port_name": "external-407-vhu-2",
            },
            "Options": {
                "mcast_flood_reports": "true",
                "requested-chassis": "",
            },
            "Type": "",
        },
        # vm port (tap)
        {
            "Addresses": [ "fa:16:3e:fd:8c:db 192.168.202.47" ],
            "Dhcpv4Options": [
                "85b44aa7-59a2-4306-bf88-2f4028ecc5e7",
            ],
            "ExternalIDs": { "neutron:cidrs": "192.168.202.47/24"},
            "Options": {
                "mcast_flood_reports": "true",
                "requested-chassis": "compute-0.localdomain",
            },
            "Type": "",
        },
        # vm port (vhost)
        {
            "Addresses": [ "fa:16:3e:40:ac:a6 192.168.202.218" ],
            "Dhcpv4Options": [
                "85b44aa7-59a2-4306-bf88-2f4028ecc5e7",
            ],
            "ExternalIDs": {
                "neutron:cidrs": "192.168.202.218/24",
                "neutron:device_owner": "",
                "neutron:port_name": "external-407-vhu-1",
            },
            "Options": {
                "mcast_flood_reports": "true",
                "requested-chassis": "",
            },
            "Type": "",
        },
        # provider network
        {
            "Options": {
                "mcast_flood": "false",
                "mcast_flood_reports": "true",
                "network_name": "external",
            },
            "Tag": [407],
            "Type": "localnet",
        },
        # vxlan/geneve endpoint (osef?)
        {
            "Addresses": [ "fa:16:3e:e5:4e:70 192.168.202.2" ],
            "ExternalIDs": { "neutron:cidrs": "192.168.202.2/24" },
            "Type": "localport",
        },
    ],
}


package geoip

import (
	"net"
	"strings"
)

var blockedIPNetworks = mustParseCIDRs(
	// RFC1918 Private IPv4
	"10.0.0.0/8",
	"172.16.0.0/12",
	"192.168.0.0/16",

	// IPv6 Unique Local Address
	"fc00::/7",

	// Carrier Grade NAT RFC6598
	"100.64.0.0/10",

	// Link Local
	"169.254.0.0/16",
	"fe80::/10",
)



func isPublicIP(ipAddress string) bool {

	ip := parseHostIP(ipAddress)

	if ip == nil {
		return false
	}


	return !isBlockedIP(ip)
}



func isBlockedIP(ip net.IP) bool {


	if ip.IsLoopback() {
		return true
	}


	if ip.IsUnspecified() {
		return true
	}


	if ip.IsMulticast() {
		return true
	}


	if ip.IsPrivate() {
		return true
	}


	return containsBlockedNetwork(ip)
}

func containsBlockedNetwork(ip net.IP) bool {

	for _, network := range blockedIPNetworks {

		if network.Contains(ip) {
			return true
		}
	}


	return false
}

func parseHostIP(value string) net.IP {


	value =
		strings.TrimSpace(value)


	if index :=
		strings.IndexByte(value,'%');
	index >= 0 {

		value =
			value[:index]
	}


	return net.ParseIP(value)
}



func mustParseCIDRs(
	cidrs ...string,
) []*net.IPNet {


	networks :=
		make([]*net.IPNet,0,len(cidrs))


	for _,cidr := range cidrs {


		_,network,err :=
			net.ParseCIDR(cidr)


		if err != nil {

			panic(
				"geoip: invalid CIDR "+cidr,
			)
		}


		networks =
			append(
				networks,
				network,
			)
	}


	return networks
}
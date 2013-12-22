package ipv6support

import "net"
import "projectconst"

func IsIPv6Supported() bool {
	addrs, err := net.LookupIP(projectconst.SERVER_HOST_IPV6)
	if len(addrs) < 1 || err != nil {
		return false
	}
	return true
}

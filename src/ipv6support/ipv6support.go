package ipv6support

//import "net"
import "net/http"
import "projectconst"

func IsIPv6Supported() bool {
	_, err := http.Get(projectconst.SERVER_ADDR_IPV6)

	if err != nil {
		return false
	}

	return true
}

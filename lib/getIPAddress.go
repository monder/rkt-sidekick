package lib

import (
	"fmt"
	"net"
)

func GetIPAddress(cidr string) (string, error) {
	_, requirednet, err := net.ParseCIDR(cidr)
	if err != nil {
		return "", err
	}
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if requirednet.Contains(ipnet.IP) {
				return ipnet.IP.String(), nil
			}
		}

	}
	return "", fmt.Errorf("No IP address matching %s found.", cidr)
}

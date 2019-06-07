package common

import "net"

var localIPBlocks []*net.IPNet
var privateIPBlocks []*net.IPNet

func init() {
	for _, cidr := range []string{
		"127.0.0.0/8", // IPv4 loopback
		"::1/128",     // IPv6 loopback
	} {
		_, block, _ := net.ParseCIDR(cidr)
		localIPBlocks = append(localIPBlocks, block)
		privateIPBlocks = append(privateIPBlocks, block)
	}
	for _, cidr := range []string{
		"10.0.0.0/8",     // RFC1918
		"172.16.0.0/12",  // RFC1918
		"192.168.0.0/16", // RFC1918
		"fe80::/10",      // IPv6 link-local
		"fc00::/7",       // IPv6 unique local addr
	} {
		_, block, _ := net.ParseCIDR(cidr)
		privateIPBlocks = append(privateIPBlocks, block)
	}
}

// HostIP returns the primary IP address of the machine.
func HostIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	return conn.LocalAddr().(*net.UDPAddr).IP.String()
}

// IsPrivateHost returns whether the given host resolves within a IANA private IP range.
func IsPrivateHost(host string) bool {
	if host == "" {
		return true
	}
	ip := HostToIP(host)
	return IsPrivateIP(ip)
}

// IsPrivateIP returns whether the given IP is within a IANA private IP range.
func IsPrivateIP(ip net.IP) bool {
	return iPInBlocks(ip, privateIPBlocks)
}

// IsLocalHost returns whether the given host resolves within a local IP range.
func IsLocalHost(host string) bool {
	if host == "" {
		return true
	}

	ip := HostToIP(host)
	return IsLocalIP(ip)
}

// IsLocalIP returns whether the given IP is within a local IP range.
func IsLocalIP(ip net.IP) bool {
	return iPInBlocks(ip, localIPBlocks)
}

// HostToIP resolves a hostname/IP to an IP (if possible).
func HostToIP(host string) net.IP {
	if host == "" {
		return nil
	}

	ip := net.ParseIP(host)
	if len(ip) == 0 {
		addr, err := net.ResolveIPAddr("ip", host)
		if err != nil {
			return nil
		}
		ip = addr.IP
	}

	return ip
}

func iPInBlocks(ip net.IP, blocks []*net.IPNet) bool {
	if len(ip) == 0 {
		return false
	}

	for _, block := range blocks {
		if block.Contains(ip) {
			return true
		}
	}
	return false
}

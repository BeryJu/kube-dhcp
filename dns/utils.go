package dns

import (
	"fmt"
	"strings"

	dhcpv1 "beryju.org/kube-dhcp/api/v1"
)

func getHostname(lease dhcpv1.Lease, domain string) string {
	hostname := strings.ReplaceAll(lease.Spec.Hostname, domain, "")
	hostname = strings.Split(hostname, ".")[0]
	return fmt.Sprintf("%s.%s", hostname, domain)
}

func reverseDNSRecord(ip string) string {
	addressSlice := strings.Split(ip, ".")
	reverseSlice := []string{}

	for i := range addressSlice {
		octet := addressSlice[len(addressSlice)-1-i]
		reverseSlice = append(reverseSlice, octet)
	}
	return fmt.Sprintf("%s.%s.%s.%s.in-addr.arpa", reverseSlice[0], reverseSlice[1], reverseSlice[2], reverseSlice[3])
}

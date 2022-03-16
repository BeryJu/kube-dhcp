package dns

import (
	"fmt"
	"strings"
)

func reverseDNSRecord(ip string) string {
	addressSlice := strings.Split(ip, ".")
	reverseSlice := []string{}

	for i := range addressSlice {
		octet := addressSlice[len(addressSlice)-1-i]
		reverseSlice = append(reverseSlice, octet)
	}
	return fmt.Sprintf("%s.%s.%s.%s.in-addr.arpa", reverseSlice[0], reverseSlice[1], reverseSlice[2], reverseSlice[3])
}

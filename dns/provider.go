package dns

import (
	"strings"

	dhcpv1 "beryju.org/kube-dhcp/api/v1"
	"beryju.org/kube-dhcp/dns/base"
	"beryju.org/kube-dhcp/dns/route53"
)

func GetDNSProviderForScope(scope dhcpv1.Scope) base.DNSProvider {
	switch strings.ToLower(scope.Spec.DNS.Provider) {
	case "route53":
		return route53.New(scope.Spec.DNS.Config)
	case "":
		fallthrough
	default:
		return base.New()
	}
}

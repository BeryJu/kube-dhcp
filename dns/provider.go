package dns

import (
	"strings"

	dhcpv1 "beryju.org/kube-dhcp/api/v1"
)

type DNSProvider interface {
	CreateRecord(lease *dhcpv1.Lease) error
	UpdateRecord(lease *dhcpv1.Lease) error
	DeleteRecord(lease *dhcpv1.Lease) error
}

type BaseDNSProvider struct{}

func NewBaseProvider() *BaseDNSProvider {
	return &BaseDNSProvider{}
}
func (b *BaseDNSProvider) CreateRecord(lease *dhcpv1.Lease) error { return nil }
func (b *BaseDNSProvider) UpdateRecord(lease *dhcpv1.Lease) error { return nil }
func (b *BaseDNSProvider) DeleteRecord(lease *dhcpv1.Lease) error { return nil }

func GetDNSProviderForScope(scope dhcpv1.Scope) (DNSProvider, error) {
	switch strings.ToLower(scope.Spec.DNS.Provider) {
	case "route53":
		return NewRoute53Provider(scope.Spec.DNS.Config)
	case "":
		fallthrough
	default:
		return NewBaseProvider(), nil
	}
}

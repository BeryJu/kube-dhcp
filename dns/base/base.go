package base

import (
	dhcpv1 "beryju.org/kube-dhcp/api/v1"
)

type DNSProvider interface {
	CreateRecord(lease *dhcpv1.Lease) error
	UpdateRecord(lease *dhcpv1.Lease) error
	DeleteRecord(lease *dhcpv1.Lease) error
}

type BaseDNSProvider struct{}

func New() *BaseDNSProvider {
	return &BaseDNSProvider{}
}
func (b *BaseDNSProvider) CreateRecord(lease *dhcpv1.Lease) error { return nil }
func (b *BaseDNSProvider) UpdateRecord(lease *dhcpv1.Lease) error { return nil }
func (b *BaseDNSProvider) DeleteRecord(lease *dhcpv1.Lease) error { return nil }

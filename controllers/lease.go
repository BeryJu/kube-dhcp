package controllers

import (
	"context"
	"net"
	"time"

	corev1 "k8s.io/api/core/v1"

	dhcpv1 "beryju.org/kube-dhcp/api/v1"
	"github.com/insomniacslk/dhcp/dhcpv4"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (r *ScopeReconciler) createLeaseFor(scope *dhcpv1.Scope, conn net.PacketConn, peer net.Addr, m *dhcpv4.DHCPv4) *dhcpv1.Lease {
	meta := metav1.ObjectMeta{
		// TODO: Customisable name
		Name:      m.HostName(),
		Namespace: scope.Namespace,
	}
	spec := &dhcpv1.LeaseSpec{
		LeaseCommonSpec: *scope.Spec.DeepCopy().LeaseTemplate,
		Scope: corev1.LocalObjectReference{
			Name: scope.Name,
		},
	}
	spec.Identifier = m.ClientHWAddr.String()
	spec.Address = r.nextFreeAddress(*scope).String()
	// spec.OptionSet = scope.Spec.OptionSet
	status := dhcpv1.LeaseStatus{
		LastRequest: time.Now().Format(time.RFC3339),
	}
	return &dhcpv1.Lease{
		ObjectMeta: meta,
		Spec:       *spec,
		Status:     status,
	}
}

func (r *ScopeReconciler) replyWithLease(lease *dhcpv1.Lease, conn net.PacketConn, peer net.Addr, m *dhcpv4.DHCPv4) {
	// We need the scope to get the subnet bits
	scope := dhcpv1.Scope{}
	err := r.Get(context.Background(), types.NamespacedName{
		Namespace: lease.Namespace,
		Name:      lease.Spec.Scope.Name,
	}, &scope)
	if err != nil {
		r.l.Error(err, "failed to get scope for lease reply")
		return
	}

	// We need the option set to get the options
	options := dhcpv1.OptionSet{}
	err = r.Get(context.Background(), types.NamespacedName{
		Namespace: lease.Namespace,
		Name:      lease.Spec.OptionSet.Name,
	}, &options)
	if err != nil {
		r.l.Error(err, "failed to get options set for lease reply")
		return
	}

	rep, err := dhcpv4.NewReplyFromRequest(m)
	if err != nil {
		r.l.Error(err, "failed to create reply")
		return
	}
	rep.UpdateOption(dhcpv4.OptMessageType(dhcpv4.MessageTypeAck))

	ipLeaseDuration, err := time.ParseDuration(lease.Spec.AddressLeaseTime)
	if err != nil {
		r.l.Error(err, "failed to parse address lease duration, defaulting", "default", "24h")
		ipLeaseDuration = time.Hour * 24
	}
	rep.UpdateOption(dhcpv4.OptIPAddressLeaseTime(ipLeaseDuration))

	// set subnet
	_, cidr, err := net.ParseCIDR(scope.Spec.SubnetCIDR)
	if err != nil {
		r.l.Error(err, "failed to parse scope cidr, defaulting", "default", "255.255.255.0")
		cidr = &net.IPNet{
			Mask: net.CIDRMask(24, 8),
		}
	}
	rep.UpdateOption(dhcpv4.OptSubnetMask(cidr.Mask))

	rep.UpdateOption(dhcpv4.OptRequestedIPAddress(net.IP(lease.Spec.Address)))

	for _, opt := range options.Spec.Options {
		r.l.V(1).Info("applying options from optionset", "option", opt.Tag)
		rep.UpdateOption(dhcpv4.OptGeneric(dhcpv4.GenericOptionCode(opt.Tag), []byte(opt.Value)))
	}

	if _, err := conn.WriteTo(rep.ToBytes(), peer); err != nil {
		r.l.Error(err, "failed to write reply")
	}
}

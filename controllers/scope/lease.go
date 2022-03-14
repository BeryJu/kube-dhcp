package scope

import (
	"context"
	"encoding/base64"
	"net"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"

	dhcpv1 "beryju.org/kube-dhcp/api/v1"
	"github.com/insomniacslk/dhcp/dhcpv4"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (r *ScopeReconciler) createLeaseFor(scope *dhcpv1.Scope, conn net.PacketConn, peer net.Addr, m *dhcpv4.DHCPv4) *dhcpv1.Lease {
	meta := metav1.ObjectMeta{
		Name:      strings.ToLower(r.templateLeaseName(scope, conn, peer, m)),
		Namespace: scope.Namespace,
	}
	spec := &dhcpv1.LeaseSpec{
		LeaseCommonSpec: *scope.Spec.DeepCopy().LeaseTemplate,
		Scope: corev1.LocalObjectReference{
			Name: scope.Name,
		},
		Hostname: m.HostName(),
	}
	spec.Identifier = m.ClientHWAddr.String()
	spec.Address = r.nextFreeAddress(*scope).String()
	status := dhcpv1.LeaseStatus{
		LastRequest: time.Now().Format(time.RFC3339),
	}
	return &dhcpv1.Lease{
		ObjectMeta: meta,
		Spec:       *spec,
		Status:     status,
	}
}

func (r *ScopeReconciler) findLease(m *dhcpv4.DHCPv4) *dhcpv1.Lease {
	// check all leases to see if we already have this identifier somewhere
	leases := &dhcpv1.LeaseList{}
	err := r.List(context.Background(), leases)
	if err != nil {
		r.l.Error(err, "failed to list leases")
		return nil
	}
	r.l.V(1).Info("cheking for existing lease")
	var match *dhcpv1.Lease
	for _, lease := range leases.Items {
		if lease.Spec.Identifier == m.ClientHWAddr.String() {
			r.l.V(1).Info("found matching lease", "lease", lease)
			match = &lease
			break
		}
	}
	return match
}

func (r *ScopeReconciler) replyWithLease(lease *dhcpv1.Lease, conn net.PacketConn, peer net.Addr, m *dhcpv4.DHCPv4, modifyResponse func(*dhcpv4.DHCPv4) *dhcpv4.DHCPv4) {
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

	rep, err := dhcpv4.NewReplyFromRequest(m)
	if err != nil {
		r.l.Error(err, "failed to create reply")
		return
	}
	rep = modifyResponse(rep)

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

	rep.YourIPAddr = net.ParseIP(lease.Spec.Address)
	rep.UpdateOption(dhcpv4.OptHostName(lease.Spec.Hostname))

	if lease.Spec.OptionSet.Name != "" {
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
		for _, opt := range options.Spec.Options {
			finalVal := make([]byte, 0)
			r.l.V(1).Info("applying options from optionset", "option", opt.Tag)
			if opt.Tag == nil {
				continue
			}

			// Values which are directly converted from string to byte
			if opt.Value != nil {
				i := net.ParseIP(*opt.Value)
				if i == nil {
					finalVal = []byte(*opt.Value)
				} else {
					finalVal = dhcpv4.IPs([]net.IP{i}).ToBytes()
				}
			}

			// For non-stringable values, get b64 decoded values
			if len(opt.Values64) > 0 {
				values64 := make([]byte, 0)
				for _, v := range opt.Values64 {
					va, err := base64.StdEncoding.DecodeString(v)
					if err != nil {
						r.l.Error(err, "failed to convert base64 value to byte")
					} else {
						values64 = append(values64, va...)
					}
				}
				finalVal = values64
			}
			dopt := dhcpv4.OptGeneric(dhcpv4.GenericOptionCode(*opt.Tag), finalVal)
			rep.UpdateOption(dopt)
		}
	}

	r.l.V(1).Info(rep.Summary(), "peer", peer.String())
	if _, err := conn.WriteTo(rep.ToBytes(), peer); err != nil {
		r.l.Error(err, "failed to write reply")
	}
}

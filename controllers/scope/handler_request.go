package scope

import (
	"context"
	"net"
	"time"

	dhcpv1 "beryju.org/kube-dhcp/api/v1"
	"beryju.org/kube-dhcp/dns"
	"github.com/insomniacslk/dhcp/dhcpv4"
)

func (r *ScopeReconciler) handleDHCPRequest(conn net.PacketConn, peer net.Addr, m *dhcpv4.DHCPv4) {
	match := r.findLease(m)

	if match == nil {
		r.l.V(1).Info("no lease found, creating new")
		scope := r.findScopeForRequest(conn, peer, m)
		if scope == nil {
			return
		}
		r.l.V(1).Info("found scope for new lease")
		match = r.createLeaseFor(scope, conn, peer, m)
		r.l.V(1).Info("creating new lease")
		r.createLease(match, scope)
	}

	match.Status.LastRequest = time.Now().Format(time.RFC3339)
	err := r.Update(context.Background(), match)
	if err != nil {
		r.l.Error(err, "failed to update lease")
	}
	r.replyWithLease(match, conn, peer, m, func(d *dhcpv4.DHCPv4) *dhcpv4.DHCPv4 {
		d.UpdateOption(dhcpv4.OptMessageType(dhcpv4.MessageTypeAck))
		return d
	})
}

func (r *ScopeReconciler) createLease(lease *dhcpv1.Lease, scope *dhcpv1.Scope) {
	err := r.Create(context.Background(), lease)
	if err != nil {
		r.l.Error(err, "failed to create lease")
	}

	dns := dns.GetDNSProviderForScope(*scope)
	err = dns.CreateRecord(lease)
	if err != nil {
		r.l.Error(err, "failed to delete DNS record")
	}
}

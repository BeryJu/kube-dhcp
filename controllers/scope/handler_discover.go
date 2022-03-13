package scope

import (
	"context"
	"net"
	"time"

	"github.com/insomniacslk/dhcp/dhcpv4"
)

func (r *ScopeReconciler) handleDHCPDiscover(conn net.PacketConn, peer net.Addr, m *dhcpv4.DHCPv4) {
	match := r.findLease(m)
	found := match != nil

	if !found {
		r.l.V(1).Info("no lease found, creating new")
		scope := r.findScopeForRequest(conn, peer, m)
		if scope == nil {
			return
		}
		r.l.V(1).Info("found scope for new lease")
		match = r.createLeaseFor(scope, conn, peer, m)
		// Don't save the new lease yet as the client hasn't committed to it yet
	}
	// If we already have a saved lease, update its status
	if found {
		match.Status.LastRequest = time.Now().Format(time.RFC3339)
		err := r.Update(context.Background(), match)
		if err != nil {
			r.l.Error(err, "failed to update lease")
		}
	}
	r.replyWithLease(match, conn, peer, m, func(d *dhcpv4.DHCPv4) *dhcpv4.DHCPv4 {
		d.UpdateOption(dhcpv4.OptMessageType(dhcpv4.MessageTypeOffer))
		return d
	})
}

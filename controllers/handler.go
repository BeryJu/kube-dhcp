package controllers

import (
	"context"
	"net"
	"time"

	dhcpv1 "beryju.org/kube-dhcp/api/v1"
	"github.com/insomniacslk/dhcp/dhcpv4"
)

func (r *ScopeReconciler) handler(conn net.PacketConn, peer net.Addr, m *dhcpv4.DHCPv4) {
	r.l.V(1).Info(m.Summary())
	switch mt := m.MessageType(); mt {
	case dhcpv4.MessageTypeDiscover:
		r.handleDHCPDiscover(conn, peer, m)
	case dhcpv4.MessageTypeRequest:
		r.handleDHCPRequest(conn, peer, m)
	default:
		r.l.Info("unsupported message type", "type", mt)
		return
	}
}

func (r *ScopeReconciler) handleDHCPDiscover(conn net.PacketConn, peer net.Addr, m *dhcpv4.DHCPv4) {
	rep, err := dhcpv4.NewReplyFromRequest(m)
	if err != nil {
		r.l.Error(err, "failed to create reply")
		return
	}
	rep.UpdateOption(dhcpv4.OptMessageType(dhcpv4.MessageTypeOffer))

	if _, err := conn.WriteTo(rep.ToBytes(), peer); err != nil {
		r.l.Error(err, "failed to write reply")
	} else {
		r.l.V(1).Info("sent discovery offer")
	}
}

func (r *ScopeReconciler) handleDHCPRequest(conn net.PacketConn, peer net.Addr, m *dhcpv4.DHCPv4) {
	// check all leases to see if we already have this identifier somewhere
	leases := &dhcpv1.LeaseList{}
	err := r.List(context.Background(), leases)
	if err != nil {
		r.l.Error(err, "failed to list leases")
		return
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

	if match == nil {
		r.l.V(1).Info("no lease found, creating new")
		scope := r.findScopeForRequest(conn, peer, m)
		if scope == nil {
			return
		}
		r.l.V(1).Info("found scope for new lease")
		match = r.createLeaseFor(scope, conn, peer, m)
		r.l.V(1).Info("creating new lease")
		err = r.Create(context.Background(), match)
		if err != nil {
			r.l.Error(err, "failed to create lease")
		}
	}
	match.Status.LastRequest = time.Now().Format(time.RFC3339)

	err = r.Update(context.Background(), match)
	if err != nil {
		r.l.Error(err, "failed to update lease")
	}
	r.replyWithLease(match, conn, peer, m)
}

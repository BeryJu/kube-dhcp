package controllers

import (
	"net"

	"github.com/insomniacslk/dhcp/dhcpv4"
)

func (r *ScopeReconciler) handler(conn net.PacketConn, peer net.Addr, m *dhcpv4.DHCPv4) {
	r.l.V(1).Info(m.Summary())
	switch mt := m.MessageType(); mt {
	case dhcpv4.MessageTypeDiscover:
		r.handleDHCPDiscover(conn, peer, m)
	case dhcpv4.MessageTypeRequest:
		r.handleDHCPRequest(conn, peer, m)
	case dhcpv4.MessageTypeRelease:
		// TODO release
		r.l.Info("release handler")
	default:
		r.l.Info("unsupported message type", "type", mt)
		return
	}
}

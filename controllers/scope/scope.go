package scope

import (
	"context"
	"net"
	"net/netip"

	dhcpv1 "beryju.org/kube-dhcp/api/v1"
	"github.com/insomniacslk/dhcp/dhcpv4"
)

func (r *ScopeReconciler) findScopeForRequest(conn net.PacketConn, peer net.Addr, m *dhcpv4.DHCPv4) *dhcpv1.Scope {
	scopeList := &dhcpv1.ScopeList{}
	err := r.List(context.Background(), scopeList)
	if err != nil {
		r.l.Error(err, "failed to list scopes")
		return nil
	}
	r.l.V(1).Info("got all scopes", "scopeCount", len(scopeList.Items))
	var match *dhcpv1.Scope
	for _, scope := range scopeList.Items {
		// TODO: priority and order
		if r.matchScope(scope, conn, peer, m) {
			r.l.V(1).Info("Selected scope based on match", "scope", scope.ObjectMeta.Name)
			match = &scope
		}
		if match == nil && scope.Spec.Default {
			r.l.V(1).Info("Selected scope based on default state", "scope", scope.ObjectMeta.Name)
			match = &scope
		}
	}
	if match != nil {
		r.l.V(1).Info("found scope for request", "scope", match.ObjectMeta.Name)
	}
	return match
}

func (r *ScopeReconciler) matchScope(scope dhcpv1.Scope, conn net.PacketConn, peer net.Addr, m *dhcpv4.DHCPv4) bool {
	// Check cidrs
	_, cidr, err := net.ParseCIDR(scope.Spec.SubnetCIDR)
	if err != nil {
		r.l.Error(err, "failed to parse cidr", "scope", scope.ObjectMeta.Name)
		scope.Status.State = err.Error()
		err = r.Update(context.Background(), &scope)
		if err != nil {
			r.l.Error(err, "failed to write status into scope")
		}
	}
	if cidr.Contains(m.ClientIPAddr) {
		r.l.V(1).Info("Scope CIDR matches client addr", "scope", scope.ObjectMeta.Name, "ip", m.ClientIPAddr.String())
		return true
	}
	return false
}

func (r *ScopeReconciler) nextFreeAddress(scope dhcpv1.Scope) *netip.Addr {
	// Check cidrs
	cidr, err := netip.ParsePrefix(scope.Spec.SubnetCIDR)
	if err != nil {
		r.l.Error(err, "failed to parse cidr", "scope", scope.ObjectMeta.Name)
		scope.Status.State = err.Error()
		err = r.Update(context.Background(), &scope)
		if err != nil {
			r.l.Error(err, "failed to write status into scope")
		}
		return nil
	}
	// get all leases to check
	leases := &dhcpv1.LeaseList{}
	err = r.List(context.Background(), leases)
	if err != nil {
		r.l.Error(err, "failed to list leases")
		return nil
	}

	start, err := netip.ParseAddr(scope.Spec.Range.Start)
	if err != nil {
		r.l.Error(err, "failed to parse start")
		return nil
	}
	end, err := netip.ParseAddr(scope.Spec.Range.End)
	if err != nil {
		r.l.Error(err, "failed to parse end")
		return nil
	}

	initialIp := cidr.Addr()
	for {
		initialIp = initialIp.Next()
		r.l.V(1).Info("checking for free ip", "ip", initialIp.String())
		// Check if IP is in the correct subnet
		if !cidr.Contains(initialIp) {
			return nil
		}
		foundExisting := false
		// Ip is less than the start of the range
		if start.Compare(initialIp) == -1 {
			foundExisting = true
		}
		// Ip is more than the end of the range
		if end.Compare(initialIp) == 1 {
			foundExisting = true
		}
		// check for existing leases
		for _, l := range leases.Items {
			if l.Spec.Address == initialIp.String() {
				foundExisting = true
				break
			}
		}
		if !foundExisting {
			return &initialIp
		}
	}
}

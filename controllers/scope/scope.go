package scope

import (
	"context"
	"math/big"
	"net"

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

func (r *ScopeReconciler) nextFreeAddress(scope dhcpv1.Scope) *net.IP {
	// Check cidrs
	initialIp, cidr, err := net.ParseCIDR(scope.Spec.SubnetCIDR)
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

	for {
		// Get next IP
		ipb := big.NewInt(0).SetBytes([]byte(initialIp))
		ipb.Add(ipb, big.NewInt(1))
		b := ipb.Bytes()
		b = append(make([]byte, len(initialIp)-len(b)), b...)
		initialIp = net.IP(b)
		r.l.V(1).Info("checking for free ip", "ip", initialIp.String())
		// Check if IP is in the correct subnet
		if !cidr.Contains(initialIp) {
			return nil
		}
		foundExisting := false
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

package scope

import (
	"bytes"
	"net"
	"strings"
	"text/template"

	dhcpv1 "beryju.org/kube-dhcp/api/v1"
	"github.com/insomniacslk/dhcp/dhcpv4"
)

type LeaseNameTemplateContext struct {
	dhcp *dhcpv4.DHCPv4
}

func (r *ScopeReconciler) templateLeaseName(scope *dhcpv1.Scope, conn net.PacketConn, peer net.Addr, m *dhcpv4.DHCPv4) string {
	def := m.HostName()
	ctx := LeaseNameTemplateContext{
		dhcp: m,
	}
	tmpl, err := template.New("").Parse(scope.Spec.LeaseNameTemplate)
	if err != nil {
		r.l.Error(err, "failed to render template", "tmpl", scope.Spec.LeaseNameTemplate)
		return def
	}
	var tpl bytes.Buffer
	if err := tmpl.Execute(&tpl, ctx); err != nil {
		r.l.Error(err, "failed to template lease name")
		return def
	}
	return strings.TrimSpace(tpl.String())
}

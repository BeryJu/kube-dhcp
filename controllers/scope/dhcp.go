package scope

import (
	"net"

	"github.com/insomniacslk/dhcp/dhcpv4/server4"
)

func (r *ScopeReconciler) ensureRunning() {
	if r.dsRunning {
		return
	}
	r.l.Info("DHCP Server isn't running, starting")
	go func() {
		r.dsRunning = true
		err := r.startServer()
		if err != nil {
			r.l.Error(err, "failed to start dhcp server")
			panic(err)
		}
	}()
}

func (r *ScopeReconciler) startServer() error {
	laddr := net.UDPAddr{
		IP:   net.ParseIP("0.0.0.0"),
		Port: 1067,
	}
	server, err := server4.NewServer("", &laddr, r.handler)
	if err != nil {
		return err
	}
	r.ds = server
	return r.ds.Serve()
}

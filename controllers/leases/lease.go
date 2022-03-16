package leases

import (
	"context"
	"sync"
	"time"

	dhcpv1 "beryju.org/kube-dhcp/api/v1"
	"beryju.org/kube-dhcp/dns"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// LeaseReconciler reconciles a Lease object
type LeaseReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	l logr.Logger

	queue      map[types.UID]bool
	queueMutex sync.Mutex
}

//+kubebuilder:rbac:groups=dhcp.beryju.org,resources=leases,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=dhcp.beryju.org,resources=leases/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=dhcp.beryju.org,resources=leases/finalizers,verbs=update
func (l *LeaseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l.l = ctrl.Log
	l.l.V(1).Info("lease reconcile run")

	leases := &dhcpv1.LeaseList{}
	err := l.List(ctx, leases)
	if err != nil {
		l.l.Error(err, "failed to list leases")
		return ctrl.Result{}, err
	}

	// this approach probably leaks goroutines all over the place,
	// since waiting ones are never cancelled/removed
	for _, lease := range leases.Items {
		_, qs := l.queue[lease.UID]
		if !qs {
			go l.checkExpired(lease)
		}
	}

	return ctrl.Result{}, nil
}

func (l *LeaseReconciler) checkExpired(lease dhcpv1.Lease) {
	l.queueMutex.Lock()
	l.queue[lease.UID] = true
	l.queueMutex.Unlock()
	created := lease.CreationTimestamp.Time
	dur, err := time.ParseDuration(lease.Spec.AddressLeaseTime)
	if err != nil {
		l.l.Error(err, "failed to parse duration in lease", "lease", lease.Name)
		return
	}
	delta := time.Until(created.Add(dur))
	if delta < 0 {
		l.deleteLease(lease)
	} else {
		time.Sleep(delta)
		l.checkExpired(lease)
		return
	}
}

func (l *LeaseReconciler) deleteLease(lease dhcpv1.Lease) {
	err := l.Delete(context.Background(), &lease)
	if err != nil {
		l.l.Error(err, "failed to delete lease", "lease", lease.Name)
		return
	}

	// Get scope to get DNS config
	var scope dhcpv1.Scope
	err = l.Get(context.Background(), client.ObjectKey{
		Namespace: lease.Namespace,
		Name:      lease.Spec.Scope.Name,
	}, &scope)
	if err != nil {
		l.l.Error(err, "failed to get scope for DNS config")
		return
	}
	dns, err := dns.GetDNSProviderForScope(scope)
	if err != nil {
		l.l.Error(err, "failed to get DNS provider")
		return
	}
	err = dns.DeleteRecord(&lease)
	if err != nil {
		l.l.Error(err, "failed to delete DNS record")
	}
}

// SetupWithManager sets up the controller with the Manager.
func (l *LeaseReconciler) SetupWithManager(mgr ctrl.Manager) error {
	l.queue = make(map[types.UID]bool)
	return ctrl.NewControllerManagedBy(mgr).
		For(&dhcpv1.Lease{}).
		Complete(l)
}

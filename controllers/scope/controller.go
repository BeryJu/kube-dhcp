package scope

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/insomniacslk/dhcp/dhcpv4/server4"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	dhcpv1 "beryju.org/kube-dhcp/api/v1"
)

// ScopeReconciler reconciles a Scope object
type ScopeReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	l logr.Logger

	ds        *server4.Server
	dsRunning bool

	scopes     []dhcpv1.Scope
	optionSets []dhcpv1.OptionSet
}

//+kubebuilder:rbac:groups=dhcp.beryju.org,resources=scopes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=dhcp.beryju.org,resources=scopes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=dhcp.beryju.org,resources=scopes/finalizers,verbs=update
//+kubebuilder:rbac:groups=dhcp.beryju.org,resources=leases,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=dhcp.beryju.org,resources=leases/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=dhcp.beryju.org,resources=leases/finalizers,verbs=update
//+kubebuilder:rbac:groups=dhcp.beryju.org,resources=optionsets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=dhcp.beryju.org,resources=optionsets/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=dhcp.beryju.org,resources=optionsets/finalizers,verbs=update
func (r *ScopeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.l = ctrl.Log

	r.ensureRunning()

	scopeList := &dhcpv1.ScopeList{}
	err := r.List(ctx, scopeList)
	if err != nil {
		r.l.Error(err, "failed to list scopes")
		return ctrl.Result{}, err
	}
	r.l.V(1).Info("got all scopes", "scopeCount", len(scopeList.Items))
	r.scopes = scopeList.Items

	optionSetList := &dhcpv1.OptionSetList{}
	err = r.List(ctx, optionSetList)
	if err != nil {
		r.l.Error(err, "failed to list optionSets")
		return ctrl.Result{}, err
	}
	r.l.V(1).Info("got all optionSets", "optionSetCount", len(optionSetList.Items))
	r.optionSets = optionSetList.Items

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ScopeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&dhcpv1.Scope{}).
		Complete(r)
}

package option_sets

import (
	"context"

	dhcpv1 "beryju.org/kube-dhcp/api/v1"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// OptionSetReconciler reconciles a Lease object
type OptionSetReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	l logr.Logger
}

//+kubebuilder:rbac:groups=dhcp.beryju.org,resources=optionsets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=dhcp.beryju.org,resources=optionsets/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=dhcp.beryju.org,resources=optionsets/finalizers,verbs=update
func (os *OptionSetReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	os.l = ctrl.Log

	sets := &dhcpv1.OptionSetList{}
	err := os.List(ctx, sets)
	if err != nil {
		os.l.Error(err, "failed to list optionSets")
		return ctrl.Result{}, err
	}

	for _, set := range sets.Items {
		setDirty := false
		for _, opt := range set.Spec.Options {
			// Check if we need to update Tag
			if opt.Tag == nil && opt.TagName != nil {
				tag, o := TagMap[*opt.TagName]
				if !o {
					os.l.Info("failed to map tag name to tag", "optionSet", set.Name, "tag", opt.TagName)
				} else {
					opt.Tag = &tag
					setDirty = true
				}
			}
		}
		if setDirty {
			err := os.Update(ctx, &set)
			if err != nil {
				os.l.Error(err, "failed to update set")
			}
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (os *OptionSetReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&dhcpv1.OptionSet{}).
		Complete(os)
}

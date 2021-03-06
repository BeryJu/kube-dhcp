package cmd

import (
	"flag"
	"os"
	"time"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.

	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	dhcpv1 "beryju.org/kube-dhcp/api/v1"
	"beryju.org/kube-dhcp/controllers/leases"
	"beryju.org/kube-dhcp/controllers/option_sets"
	"beryju.org/kube-dhcp/controllers/scope"

	//+kubebuilder:scaffold:imports
	"fmt"

	"github.com/getsentry/sentry-go"
	"github.com/spf13/cobra"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)
var metricsAddr string
var enableLeaderElection bool
var probeAddr string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use: "kube-dhcp",
	Run: func(cmd *cobra.Command, args []string) {
		opts := zap.Options{
			Development: true,
		}
		opts.BindFlags(flag.CommandLine)

		ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

		err := sentry.Init(sentry.ClientOptions{
			Dsn:         "https://110345c457714e22ba1d493f0a5dc639@sentry.beryju.org/16",
			Environment: "development",
			Release:     "kube-dhcp@0.0.0",
		})
		if err != nil {
			ctrl.Log.Info("sentry.Init: %s", err)
		}
		// Flush buffered events before the program terminates.
		// Set the timeout to the maximum duration the program can afford to wait.
		defer sentry.Flush(2 * time.Second)

		mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
			Scheme:                 scheme,
			MetricsBindAddress:     metricsAddr,
			Port:                   9443,
			HealthProbeBindAddress: probeAddr,
			LeaderElection:         enableLeaderElection,
			LeaderElectionID:       "2d6faaa0.beryju.org",
		})
		if err != nil {
			setupLog.Error(err, "unable to start manager")
			os.Exit(1)
		}

		if err = (&scope.ScopeReconciler{
			Client: mgr.GetClient(),
			Scheme: mgr.GetScheme(),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "Scope")
			os.Exit(1)
		}
		if err = (&leases.LeaseReconciler{
			Client: mgr.GetClient(),
			Scheme: mgr.GetScheme(),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "Lease")
			os.Exit(1)
		}
		if err = (&option_sets.OptionSetReconciler{
			Client: mgr.GetClient(),
			Scheme: mgr.GetScheme(),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "OptionSet")
			os.Exit(1)
		}
		//+kubebuilder:scaffold:builder

		if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
			setupLog.Error(err, "unable to set up health check")
			os.Exit(1)
		}
		if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
			setupLog.Error(err, "unable to set up ready check")
			os.Exit(1)
		}

		setupLog.Info("starting manager")
		if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
			setupLog.Error(err, "problem running manager")
			os.Exit(1)
		}

	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&metricsAddr, "metrics-bind-address", ":8088", "The address the metric endpoint binds to.")
	rootCmd.PersistentFlags().StringVar(&probeAddr, "health-probe-bind-address", ":8089", "The address the probe endpoint binds to.")
	rootCmd.PersistentFlags().BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")

	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(dhcpv1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

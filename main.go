/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"flag"
	"github.com/vmware/cbcontainers-operator/cbcontainers/monitor"
	commonState "github.com/vmware/cbcontainers-operator/cbcontainers/state/common"
	"os"

	clusterProcessors "github.com/vmware/cbcontainers-operator/cbcontainers/processors/cluster"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	clusterState "github.com/vmware/cbcontainers-operator/cbcontainers/state/cluster"
	hardeningState "github.com/vmware/cbcontainers-operator/cbcontainers/state/hardening"
	certificatesUtils "github.com/vmware/cbcontainers-operator/cbcontainers/utils/certificates"
	"github.com/vmware/cbcontainers-operator/controllers"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(cbcontainersv1.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "d27fd235.operator.containers.carbonblack.io",
		Logger:                 ctrl.Log,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	defaultMonitorCreator, err := clusterProcessors.NewDefaultMonitorCreator(monitor.NewDefaultHealthChecker(mgr.GetClient(), commonState.DataPlaneNamespaceName), monitor.NewDefaultFeaturesStatusProvider(mgr.GetClient()))
	if err != nil {
		setupLog.Error(err, "unable to create default monitor creator")
		os.Exit(1)
	}

	cbContainersClusterLogger := ctrl.Log.WithName("controllers").WithName("CBContainersCluster")
	if err = (&controllers.CBContainersClusterReconciler{
		Client:              mgr.GetClient(),
		Log:                 cbContainersClusterLogger,
		Scheme:              mgr.GetScheme(),
		ClusterProcessor:    clusterProcessors.NewCBContainerClusterProcessor(cbContainersClusterLogger, clusterProcessors.NewDefaultGatewayCreator(), defaultMonitorCreator),
		ClusterStateApplier: clusterState.NewClusterStateApplier(cbContainersClusterLogger),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "CBContainersCluster")
		os.Exit(1)
	}

	cbContainersHardeningLogger := ctrl.Log.WithName("controllers").WithName("CBContainersHardening")
	if err = (&controllers.CBContainersHardeningReconciler{
		Client:                mgr.GetClient(),
		Log:                   cbContainersHardeningLogger,
		Scheme:                mgr.GetScheme(),
		HardeningStateApplier: hardeningState.NewHardeningStateApplier(cbContainersHardeningLogger, certificatesUtils.NewCertificateCreator()),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "CBContainersHardening")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("health", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("check", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

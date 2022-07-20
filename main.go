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
	"context"
	"flag"
	"fmt"
	"github.com/vmware/cbcontainers-operator/cbcontainers/state"
	"github.com/vmware/cbcontainers-operator/cbcontainers/state/agent_applyment"
	"github.com/vmware/cbcontainers-operator/cbcontainers/state/applyment"
	"github.com/vmware/cbcontainers-operator/cbcontainers/state/operator"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	coreV1 "k8s.io/api/core/v1"

	"github.com/vmware/cbcontainers-operator/cbcontainers/processors"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	operatorcontainerscarbonblackiov1 "github.com/vmware/cbcontainers-operator/api/v1"
	certificatesUtils "github.com/vmware/cbcontainers-operator/cbcontainers/utils/certificates"
	"github.com/vmware/cbcontainers-operator/controllers"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

const NamespaceIdentifier = "default"

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(operatorcontainerscarbonblackiov1.AddToScheme(scheme))
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

	setupLog.Info(fmt.Sprintf("Getting Cluster Identifier: %v uid", NamespaceIdentifier))
	namespace := &coreV1.Namespace{}
	apiReader := mgr.GetAPIReader()
	if err = apiReader.Get(context.Background(), client.ObjectKey{Namespace: NamespaceIdentifier, Name: NamespaceIdentifier}, namespace); err != nil {
		setupLog.Error(err, fmt.Sprintf("unable to get the %v namespace", NamespaceIdentifier))
		os.Exit(1)
	}
	clusterIdentifier := string(namespace.UID)

	setupLog.Info(fmt.Sprintf("Cluster Identifier: %v", clusterIdentifier))

	setupLog.Info("Getting Nodes list")
	nodesList := &coreV1.NodeList{}
	if err := apiReader.List(context.Background(), nodesList); err != nil || nodesList.Items == nil || len(nodesList.Items) < 1 {
		setupLog.Error(err, "couldn't get nodes list")
		os.Exit(1)
	}
	k8sVersion := nodesList.Items[0].Status.NodeInfo.KubeletVersion
	setupLog.Info(fmt.Sprintf("K8s version is: %v", k8sVersion))

	cbContainersAgentLogger := ctrl.Log.WithName("controllers").WithName("CBContainersAgent")
	if err = (&controllers.CBContainersAgentController{
		Client:           mgr.GetClient(),
		Log:              cbContainersAgentLogger,
		Scheme:           mgr.GetScheme(),
		K8sVersion:       k8sVersion,
		ClusterProcessor: processors.NewAgentProcessor(cbContainersAgentLogger, processors.NewDefaultGatewayCreator(), operator.NewEnvVersionProvider(), clusterIdentifier),
		StateApplier:     state.NewStateApplier(agent_applyment.NewAgentComponent(applyment.NewComponentApplier(mgr.GetClient())), k8sVersion, certificatesUtils.NewCertificateCreator(), cbContainersAgentLogger),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "CBContainersAgent")
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

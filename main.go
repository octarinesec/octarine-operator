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
	"errors"
	"flag"
	"fmt"
	"github.com/vmware/cbcontainers-operator/cbcontainers/communication/gateway"
	"github.com/vmware/cbcontainers-operator/cbcontainers/state"
	"github.com/vmware/cbcontainers-operator/cbcontainers/state/agent_applyment"
	"github.com/vmware/cbcontainers-operator/cbcontainers/state/applyment"
	"github.com/vmware/cbcontainers-operator/cbcontainers/state/common"
	"github.com/vmware/cbcontainers-operator/cbcontainers/state/operator"
	"github.com/vmware/cbcontainers-operator/remote_configuration"
	"go.uber.org/zap/zapcore"
	coreV1 "k8s.io/api/core/v1"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"strings"
	"sync"

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

const (
	NamespaceIdentifier = "default"
	httpProxyEnv        = "HTTP_PROXY"
	httpsProxyEnv       = "HTTPS_PROXY"
	noProxyEnv          = "NO_PROXY"
	namespaceEnv        = "OPERATOR_NAMESPACE"
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(operatorcontainerscarbonblackiov1.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

func setupNoProxyEnv(namespace string) error {
	httpProxyLen := len(strings.TrimSpace(os.Getenv(httpProxyEnv)))
	httpsProxyLen := len(strings.TrimSpace(os.Getenv(httpsProxyEnv)))
	_, foundNoProxyVar := os.LookupEnv(noProxyEnv)

	// Don't set NO_PROXY if we don't have any proxies defined, or we already
	// have a NO_PROXY env var present (even if it's empty)
	if httpProxyLen+httpsProxyLen == 0 || foundNoProxyVar {
		return nil
	}

	noProxy, err := controllers.GetDefaultNoProxyValue(namespace)
	if err != nil {
		return fmt.Errorf("unable to detect default NO_PROXY value: %w", err)
	}

	if err = os.Setenv(noProxyEnv, noProxy); err != nil {
		return fmt.Errorf("could not set NO_PROXY value: %w", err)
	}

	setupLog.Info(fmt.Sprintf("using NO_PROXY value %q", noProxy))
	return nil
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", true,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	opts := zap.Options{
		Development: true,
		TimeEncoder: zapcore.RFC3339TimeEncoder,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	setupLog.Info("Getting the namespace where operator is running and which should host the agent")
	operatorNamespace := os.Getenv(namespaceEnv)
	if operatorNamespace == "" {
		setupLog.Info(fmt.Sprintf("Operator namespace variable was not found. Falling back to default %s", common.DataPlaneNamespaceName))
		operatorNamespace = common.DataPlaneNamespaceName
	}
	setupLog.Info(fmt.Sprintf("Operator and agent namespace: %s", operatorNamespace))

	if err := setupNoProxyEnv(operatorNamespace); err != nil {
		setupLog.Error(err, "unable to setup default proxy settings")
		os.Exit(1)
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "d27fd235.operator.containers.carbonblack.io",
		Logger:                 ctrl.Log,
		Namespace:              operatorNamespace,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	clusterIdentifier, k8sVersion := extractConfigurationVariables(mgr)
	operatorVersionProvider := operator.NewEnvVersionProvider()
	var processorGatewayCreator processors.APIGatewayCreator = func(cbContainersCluster *operatorcontainerscarbonblackiov1.CBContainersAgent, accessToken string) (processors.APIGateway, error) {
		return gateway.NewDefaultGatewayCreator().CreateGateway(cbContainersCluster, accessToken)
	}
	cbContainersAgentLogger := ctrl.Log.WithName("controllers").WithName("CBContainersAgent")

	if err = (&controllers.CBContainersAgentController{
		Client:              mgr.GetClient(),
		Log:                 cbContainersAgentLogger,
		Scheme:              mgr.GetScheme(),
		K8sVersion:          k8sVersion,
		Namespace:           operatorNamespace,
		AccessTokenProvider: operator.NewSecretAccessTokenProvider(mgr.GetClient()),
		ClusterProcessor:    processors.NewAgentProcessor(cbContainersAgentLogger, processorGatewayCreator, operatorVersionProvider, clusterIdentifier),
		StateApplier:        state.NewStateApplier(mgr.GetAPIReader(), agent_applyment.NewAgentComponent(applyment.NewComponentApplier(mgr.GetClient())), k8sVersion, operatorNamespace, certificatesUtils.NewCertificateCreator(), cbContainersAgentLogger),
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

	k8sClient := mgr.GetClient()
	log := ctrl.Log.WithName("configurator")
	operatorVersion, err := operatorVersionProvider.GetOperatorVersion()
	if err != nil && !errors.Is(err, operator.ErrNotSemVer) {
		setupLog.Error(err, "unable to read the running operator's version from environment variable")
		os.Exit(1)
	}
	var configuratorGatewayCreator remote_configuration.ApiCreator = func(cbContainersCluster *operatorcontainerscarbonblackiov1.CBContainersAgent, accessToken string) (remote_configuration.ApiGateway, error) {
		return gateway.NewDefaultGatewayCreator().CreateGateway(cbContainersCluster, accessToken)
	}

	applier := remote_configuration.NewConfigurator(
		k8sClient,
		configuratorGatewayCreator,
		log,
		operator.NewSecretAccessTokenProvider(k8sClient),
		operatorVersion,
		operatorNamespace,
		clusterIdentifier,
	)
	applierController := remote_configuration.NewRemoteConfigurationController(applier, log)

	var wg sync.WaitGroup
	wg.Add(2)

	signalsContext := ctrl.SetupSignalHandler()
	go func() {
		defer wg.Done()

		setupLog.Info("starting manager")
		if err := mgr.Start(signalsContext); err != nil {
			setupLog.Error(err, "problem running manager")
			os.Exit(1)
		}
	}()
	go func() {
		defer wg.Done()

		// TODO: Remove once the feature is ready for go-live
		enableConfigurator := os.Getenv("ENABLE_REMOTE_CONFIGURATOR")
		if enableConfigurator == "true" {
			setupLog.Info("Starting remote configurator")
			applierController.RunLoop(signalsContext)
		}
	}()

	wg.Wait()
}

func extractConfigurationVariables(mgr manager.Manager) (clusterIdentifier string, k8sVersion string) {
	setupLog.Info(fmt.Sprintf("Getting Cluster Identifier: %v uid", NamespaceIdentifier))
	namespace := &coreV1.Namespace{}
	apiReader := mgr.GetAPIReader()
	if err := apiReader.Get(context.Background(), client.ObjectKey{Namespace: NamespaceIdentifier, Name: NamespaceIdentifier}, namespace); err != nil {
		setupLog.Error(err, fmt.Sprintf("unable to get the %v namespace", NamespaceIdentifier))
		os.Exit(1)
	}
	clusterIdentifier = string(namespace.UID)

	setupLog.Info(fmt.Sprintf("Cluster Identifier: %v", clusterIdentifier))

	setupLog.Info("Getting Nodes list")
	nodesList := &coreV1.NodeList{}
	if err := apiReader.List(context.Background(), nodesList); err != nil || nodesList.Items == nil || len(nodesList.Items) < 1 {
		setupLog.Error(err, "couldn't get nodes list")
		os.Exit(1)
	}
	k8sVersion = nodesList.Items[0].Status.NodeInfo.KubeletVersion
	setupLog.Info(fmt.Sprintf("K8s version is: %v", k8sVersion))

	return
}

package octarine

import (
	"fmt"
	"github.com/go-logr/logr"
	"github.com/octarinesec/octarine-operator/pkg/monitor_agent"
	"github.com/octarinesec/octarine-operator/pkg/types"
	"github.com/redhat-cop/operator-utils/pkg/util"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/kubernetes"
)

// Instance of the running monitor agent
var agent *monitor_agent.MonitorAgent

// Reconciles the monitor agent - starts it if it isn't already started.
// If the Octarine resource is terminating - stops the monitor agent.
func (r *ReconcileOctarine) reconcileMonitor(reqLogger logr.Logger, octarine *unstructured.Unstructured, octarineSpec *types.OctarineSpec) error {
	reqLogger.V(1).Info("reconciling monitor")

	// CR is deleted - stop agent
	if util.IsBeingDeleted(octarine) {
		if agent != nil {
			reqLogger.V(1).Info("octarine resource is terminating - stopping monitor agent")
			agent.Stop()
			agent = nil
		}

		return nil
	}

	if agent == nil {
		reqLogger.V(1).Info("starting monitor agent")

		k8sClientset, err := kubernetes.NewForConfig(r.GetRestConfig())
		if err != nil {
			return fmt.Errorf("error creating K8s client for the monitor agent: %v", err)
		}

		agent, err = monitor_agent.NewAgent(octarine.GetNamespace(), octarineSpec, k8sClientset)
		if err != nil {
			return fmt.Errorf("error starting monitor agent: %v", err)
		}
		agent.Start()
	}

	return nil
}

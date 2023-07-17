package config_applier

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	"math/rand"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

// TODO: Use interfaces for dependencies?

var versions = []string{"2.12.1", "2.10.0", "2.12.0", "2.11.0"}

var (
	tr  = true
	fal = false
)

type Applier struct {
	K8sClient client.Client
	Logger    logr.Logger
}

type pendingChange struct {
	ID                    string
	Version               *string
	EnableClusterScanning *bool
	EnableRuntime         *bool
}

func (applier *Applier) RunLoop() {
	// TODO: stop on signal
	for {
		applier.Logger.Info("Running config check iteration")

		// TODO: Fix pointers vs non-pointers
		change, err := applier.getPendingChange()
		if err != nil {
			// TODO
			applier.Logger.Error(err, "failed to get existing CR")
			continue
		}
		if change == nil {
			applier.Logger.Info("No pending remote configuration changes found")
			// TODO: Polling interval
			time.Sleep(10 * time.Second)
			continue
		}

		applier.Logger.Info("Applying change", "change", change)
		err = applier.applyChange(*change)
		if err != nil {
			// TODO
			applier.Logger.Error(err, "failed to apply change")
			// TODO: REport error ?
			continue
		}

		err = applier.acknowledgeChange(*change)
		if err != nil {
			// TODO
			panic(err)
		}

		time.Sleep(40 * time.Second)
	}

	// TODO: Staggering in case of multiple changes?
}

func (applier *Applier) getPendingChange() (*pendingChange, error) {
	rand := randomChange()
	return rand, nil
}

func (applier *Applier) acknowledgeChange(change pendingChange) error {

	return nil
}

func (applier *Applier) applyChange(change pendingChange) error {
	cr, err := applier.getCR()
	if err != nil {
		return err
	}
	if cr == nil {
		// TODO: Log
		return nil
	}

	// TODO: Validation!
	if change.Version != nil {
		cr.Spec.Version = *change.Version
	}
	if change.EnableClusterScanning != nil {
		cr.Spec.Components.ClusterScanning.Enabled = change.EnableClusterScanning
	}
	if change.EnableRuntime != nil {
		cr.Spec.Components.RuntimeProtection.Enabled = change.EnableRuntime
	}

	// Apply
	// TODO:  Handle Conflict response and retry
	err = applier.K8sClient.Update(context.TODO(), cr)
	return err
}

func (applier *Applier) getCR() (*cbcontainersv1.CBContainersAgent, error) {

	// TODO: Copied from the controller
	cbContainersAgentsList := &cbcontainersv1.CBContainersAgentList{}
	if err := applier.K8sClient.List(context.TODO(), cbContainersAgentsList); err != nil {
		return nil, fmt.Errorf("couldn't list CBContainersAgent k8s objects: %v", err)
	}

	if cbContainersAgentsList.Items == nil || len(cbContainersAgentsList.Items) == 0 {
		return nil, nil
	}

	if len(cbContainersAgentsList.Items) > 1 {
		return nil, fmt.Errorf("there is more than 1 CBContainersAgent k8s object, please delete unwanted resources")
	}

	return &cbContainersAgentsList.Items[0], nil
}

func randomChange() *pendingChange {
	csRand, runtimeRand, versionRand := rand.Int(), rand.Int(), rand.Intn(len(versions)+1)
	if versionRand == len(versions) {
		return nil
	}

	changeVersion := &versions[versionRand]

	var changeClusterScanning *bool
	var changeRuntime *bool

	switch csRand % 5 {
	case 1, 3:
		changeClusterScanning = &tr
	case 2, 4:
		changeClusterScanning = &fal
	default:
		changeClusterScanning = nil
	}

	switch runtimeRand % 5 {
	case 1, 3:
		changeRuntime = &tr
	case 2, 4:
		changeRuntime = &fal
	default:
		changeRuntime = nil
	}

	return &pendingChange{
		Version:               changeVersion,
		EnableClusterScanning: changeClusterScanning,
		EnableRuntime:         changeRuntime,
	}
}

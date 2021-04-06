package monitor

import "sync"

const (
	HardeningFeature = "guardrails"
	RuntimeFeature   = "nodeguard"
)

type FeaturesStatus struct {
	hardening *singleFeatureStatus
	runtime   *singleFeatureStatus
}

func NewFeaturesStatus() *FeaturesStatus {
	return &FeaturesStatus{
		hardening: newSingleFeatureStatus(),
		runtime:   newSingleFeatureStatus(),
	}
}

func (featuresStatus *FeaturesStatus) GetEnabledFeatures() map[string]bool {
	return map[string]bool{
		HardeningFeature: featuresStatus.hardening.Enabled(),
		RuntimeFeature:   featuresStatus.runtime.Enabled(),
	}
}

func (featuresStatus *FeaturesStatus) SetHardeningAsEnabled() {
	featuresStatus.hardening.SetAsEnabled()
}

func (featuresStatus *FeaturesStatus) SetHardeningAsDisabled() {
	featuresStatus.hardening.SetAsDisabled()
}

func (featuresStatus *FeaturesStatus) RuntimeEnabled() bool {
	return featuresStatus.runtime.Enabled()
}

func (featuresStatus *FeaturesStatus) SetRuntimeAsEnabled() {
	featuresStatus.runtime.SetAsEnabled()
}

func (featuresStatus *FeaturesStatus) SetRuntimeAsDisabled() {
	featuresStatus.runtime.SetAsDisabled()
}

type singleFeatureStatus struct {
	enabled bool
	mux     sync.RWMutex
}

func newSingleFeatureStatus() *singleFeatureStatus {
	return &singleFeatureStatus{enabled: false}
}

func (singleFeatureStatus *singleFeatureStatus) Enabled() bool {
	singleFeatureStatus.mux.RLock()
	defer singleFeatureStatus.mux.RUnlock()
	return singleFeatureStatus.enabled
}

func (singleFeatureStatus *singleFeatureStatus) SetAsEnabled() {
	singleFeatureStatus.setStatus(true)
}

func (singleFeatureStatus *singleFeatureStatus) SetAsDisabled() {
	singleFeatureStatus.setStatus(false)
}

func (singleFeatureStatus *singleFeatureStatus) setStatus(enabled bool) {
	singleFeatureStatus.mux.Lock()
	defer singleFeatureStatus.mux.Unlock()
	singleFeatureStatus.enabled = enabled
}

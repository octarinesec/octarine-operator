package helm_utils

import (
	"github.com/peterbourgon/mergemap"
	"helm.sh/helm/v3/pkg/chart"
)

// Returns the default values for the helm chart (from values.yaml), including its dependency charts.
// First the dependency charts values are set, then the parent chart values (and they may override them, as defined
// by Helm)
func GetDefaultValues(chart *chart.Chart) map[string]interface{} {
	// Get the default values of all the dependency charts (their values.yaml files) and insert to the map with the
	// chart name as the key (this is the structure of the parent chart values - map between dependency chart name and
	// its values, so this is the structure we need for merging)
	depValues := make(map[string]interface{})
	for _, dep := range chart.Dependencies() {
		depValues[dep.Name()] = dep.Values
	}

	return mergemap.Merge(depValues, chart.Values)
}

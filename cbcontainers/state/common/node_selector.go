package common

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/utils/strings/slices"
)

//map[string]string
type NodeTermsBuilder struct {
	podSpec      *v1.PodSpec
	requirements []v1.NodeSelectorRequirement
}

func NewNodeTermsBuilder(podSpec *v1.PodSpec) *NodeTermsBuilder {
	builder := &NodeTermsBuilder{
		podSpec:      podSpec,
		requirements: make([]v1.NodeSelectorRequirement, 0),
	}

	return builder.withOSRequirement()
}

func (builder *NodeTermsBuilder) withOSRequirement() *NodeTermsBuilder {
	return builder.WithRequirement(v1.NodeSelectorRequirement{
		Key:      v1.LabelOSStable,
		Operator: v1.NodeSelectorOpIn,
		Values:   []string{string(v1.Linux)},
	})
}

func (builder *NodeTermsBuilder) WithArchRequirement() *NodeTermsBuilder {
	return builder.WithRequirement(v1.NodeSelectorRequirement{
		Key:      v1.LabelArchStable,
		Operator: v1.NodeSelectorOpIn,
		Values:   []string{"386", "amd64", "amd64p32"},
	})
}

func (builder *NodeTermsBuilder) WithRequirement(requirement v1.NodeSelectorRequirement) *NodeTermsBuilder {
	builder.requirements = append(builder.requirements, requirement)
	return builder
}

func (builder *NodeTermsBuilder) nodeSelector() *v1.NodeSelector {
	if builder.podSpec.Affinity == nil {
		builder.podSpec.Affinity = &v1.Affinity{}
	}
	affinity := builder.podSpec.Affinity

	if affinity.NodeAffinity == nil {
		affinity.NodeAffinity = &v1.NodeAffinity{}
	}
	nodeAffinity := affinity.NodeAffinity

	if nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution == nil {
		nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution = &v1.NodeSelector{}
	}

	return nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution
}

func (builder *NodeTermsBuilder) validateRequirements(selector *v1.NodeSelector) bool {
	if len(selector.NodeSelectorTerms) != 1 {
		return false
	}

	terms := &selector.NodeSelectorTerms[0]
	for _, desiredRequirement := range builder.requirements {
		foundKey := false
		for _, actualRequirement := range terms.MatchExpressions {
			if desiredRequirement.Key == actualRequirement.Key {
				foundKey = true

				if desiredRequirement.Operator != actualRequirement.Operator {
					return false
				}

				if len(desiredRequirement.Values) != len(actualRequirement.Values) {
					return false
				}

				for _, value := range desiredRequirement.Values {
					if !slices.Contains(actualRequirement.Values, value) {
						return false
					}
				}
			}
		}

		if !foundKey {
			return false
		}
	}

	return true
}

func (builder *NodeTermsBuilder) Build() {
	selector := builder.nodeSelector()
	if !builder.validateRequirements(selector) {
		selector.NodeSelectorTerms = []v1.NodeSelectorTerm{
			{MatchExpressions: builder.requirements},
		}
	}
}

//
//func NewCBContainersNodeSelector() CBContainersNodeSelector {
//	return CBContainersNodeSelector{
//		v1.LabelOSStable: string(v1.Linux),
//	}
//}
//
//func (selector CBContainersNodeSelector) WithArch() CBContainersNodeSelector {
//	selector[v1.LabelArchStable] = ""
//	return selector
//}

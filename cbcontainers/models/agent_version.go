package models

type AgentVersion string

const (
	AgentMinVersionNone                 AgentVersion = "none"
	AgentMaxVersionLatest               AgentVersion = "latest"
	AgentVersionUnknown                 AgentVersion = ""
	AgentVersionResolverHeadlessService AgentVersion = "2.4.2"
)

func (v AgentVersion) IsLargerThan(version string) bool {
	if v == AgentMinVersionNone || v == AgentVersionUnknown {
		return false
	}
	return string(v) > version
}

func (v AgentVersion) IsLessThan(version string) bool {
	if v == AgentMaxVersionLatest || v == AgentVersionUnknown {
		return false
	}
	return string(v) < version
}

func (v AgentVersion) IsResolverHeadlessServiceCompatible() bool {
	if v.IsLargerThan(string(AgentVersionResolverHeadlessService)) {
		return true
	}
	return false
}

package models

import "golang.org/x/mod/semver"

type AgentVersion string

const (
	AgentMinVersionNone   AgentVersion = "none"
	AgentMaxVersionLatest AgentVersion = "latest"
	AgentVersionUnknown   AgentVersion = ""
)

func (v AgentVersion) IsLargerThan(version string) bool {
	if v == AgentMinVersionNone || v == AgentVersionUnknown {
		return false
	}
	return semver.Compare(normalizeToSemVer(string(v)), normalizeToSemVer(version)) > 0
}

func (v AgentVersion) IsLessThan(version string) bool {
	if v == AgentMaxVersionLatest || v == AgentVersionUnknown {
		return false
	}
	return semver.Compare(normalizeToSemVer(string(v)), normalizeToSemVer(version)) < 0
}

func normalizeToSemVer(v string) string {
	if v == string(AgentMaxVersionLatest) || v == string(AgentVersionUnknown) || v == string(AgentMinVersionNone) {
		return v
	}

	// semver requires a leading `v` at the front, so we make sure to have one
	if v[0] != 'v' {
		return "v" + v
	}
	// Note: v is not guaranteed to be a valid SemVer here, but it should be passable to the semver library
	return v
}

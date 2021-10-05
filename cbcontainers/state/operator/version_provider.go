package operator

import (
	"fmt"
	"os"
	"regexp"
)

var (
	operatorVersionEnvVariable = "OPERATOR_VERSION"

	// semVerRegexExpr is a regex expression that checks if a string is a regular semantic version.
	//
	// source: https://semver.org/#is-there-a-suggested-regular-expression-regex-to-check-a-semver-string
	semVerRegexExpr = `^(?P<major>0|[1-9]\d*)\.(?P<minor>0|[1-9]\d*)\.(?P<patch>0|[1-9]\d*)(?:-(?P<prerelease>(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+(?P<buildmetadata>[0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`
	// semVerRegex is the compiled regex that checks if a string is a regular semantic version.
	semVerRegex = regexp.MustCompile(semVerRegexExpr)

	// ErrNotSemVer is returned by GetOperatorVersion when the version is not a semantic version
	ErrNotSemVer = fmt.Errorf("version is not a semantic version")
)

// EnvVersionProvider provides the operator version
// based on the current environment.
type EnvVersionProvider struct{}

func NewEnvVersionProvider() *EnvVersionProvider {
	return &EnvVersionProvider{}
}

// GetOperatorVersion gets the operator version from the environment.
func (p *EnvVersionProvider) GetOperatorVersion() (string, error) {
	v := os.Getenv(operatorVersionEnvVariable)
	if v == "" {
		return "", fmt.Errorf("env variable %s is empty", operatorVersionEnvVariable)
	}
	if match := semVerRegex.Match([]byte(v)); !match {
		return "", fmt.Errorf("%w: %v", ErrNotSemVer, v)
	}
	return v, nil
}

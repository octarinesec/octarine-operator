package operator

import (
	"fmt"
	"os"
)

var (
	operatorVersionEnvVariable = "OPERATOR_VERSION"
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
	return v, nil
}

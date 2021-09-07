package operator

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	operatorVersion = "3.0.0"
	provider        = NewEnvVersionProvider()
)

func TestGetOperatorVersionSuccess(t *testing.T) {
	os.Setenv(operatorVersionEnvVariable, operatorVersion)
	defer os.Setenv(operatorVersion, "")

	v, err := provider.GetOperatorVersion()

	require.NoError(t, err)
	require.Equal(t, operatorVersion, v)
}

func TestGetOperatorVersionFail(t *testing.T) {
	os.Setenv(operatorVersionEnvVariable, "")

	v, err := provider.GetOperatorVersion()

	require.Error(t, err)
	require.Equal(t, "", v)
}

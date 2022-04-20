package utils

import (
	"fmt"

	"github.com/hashicorp/go-version"
)

func IsVersionGreaterThan(firstVersion string, secondVersion string) (bool, error) {
	v1, err1 := version.NewVersion(firstVersion)
	v2, err2 := version.NewVersion(secondVersion)

	if err1 != nil || err2 != nil {
		return false, fmt.Errorf("failed to parse version format")
	}

	if v1.GreaterThan(v2) {
		return true, nil
	}

	return false, nil
}

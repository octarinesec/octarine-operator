package common

import (
	"fmt"

	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	coreV1 "k8s.io/api/core/v1"
)

func MutateImage(container *coreV1.Container, desiredImage cbcontainersv1.CBContainersImageSpec, desiredVersion string) {
	desiredTag := desiredImage.Tag
	if desiredTag == "" {
		desiredTag = desiredVersion
	}
	desiredFullImage := fmt.Sprintf("%s:%s", desiredImage.Repository, desiredTag)

	container.Image = desiredFullImage
	container.ImagePullPolicy = desiredImage.PullPolicy
}

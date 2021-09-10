package common

import coreV1 "k8s.io/api/core/v1"

const (
	DesiredTlsRootCAsVolumeName          = "root-cas"
	DesiredTlsRootCAsVolumeMountReadOnly = true
)

func GetVolumeIndexForName(templatePodSpec *coreV1.PodSpec, name string) int {
	for i, volume := range templatePodSpec.Volumes {
		if volume.Name == name {
			return i
		}
	}

	volume := coreV1.Volume{Name: name}
	templatePodSpec.Volumes = append(templatePodSpec.Volumes, volume)
	return len(templatePodSpec.Volumes) - 1
}

func MutateVolumeConfigMapItem(configMapSource *coreV1.ConfigMapVolumeSource, key, path string) {
	for i, item := range configMapSource.Items {
		if item.Key == key {
			configMapSource.Items[i].Path = path
			return
		}
	}

	item := coreV1.KeyToPath{Key: key, Path: path}
	configMapSource.Items = append(configMapSource.Items, item)
}

func MutateVolumesToIncludeRootCasVolume(templatePodSpec *coreV1.PodSpec) {
	rootCAsVolumeIndex := GetVolumeIndexForName(templatePodSpec, DesiredTlsRootCAsVolumeName)
	if templatePodSpec.Volumes[rootCAsVolumeIndex].ConfigMap == nil {
		templatePodSpec.Volumes[rootCAsVolumeIndex].ConfigMap = &coreV1.ConfigMapVolumeSource{
			LocalObjectReference: coreV1.LocalObjectReference{Name: DataPlaneConfigmapName},
		}
	}
	if templatePodSpec.Volumes[rootCAsVolumeIndex].ConfigMap.Items == nil || len(templatePodSpec.Volumes[rootCAsVolumeIndex].ConfigMap.Items) != 1 {
		templatePodSpec.Volumes[rootCAsVolumeIndex].ConfigMap.Items = make([]coreV1.KeyToPath, 0)
	}
	MutateVolumeConfigMapItem(templatePodSpec.Volumes[rootCAsVolumeIndex].ConfigMap, DataPlaneConfigmapTlsRootCAsFilePath, DataPlaneConfigmapTlsRootCAsFilePath)
}

func GetVolumeMountIndexForName(container *coreV1.Container, name string) int {
	for i, volumeMount := range container.VolumeMounts {
		if volumeMount.Name == name {
			return i
		}
	}

	volumeMount := coreV1.VolumeMount{Name: name}
	container.VolumeMounts = append(container.VolumeMounts, volumeMount)
	return len(container.VolumeMounts) - 1
}

func MutateVolumeMount(container *coreV1.Container, volumeMountIndex int, mountPath string, readOnly bool) {
	container.VolumeMounts[volumeMountIndex].MountPath = mountPath
	container.VolumeMounts[volumeMountIndex].ReadOnly = readOnly
}

func MutateVolumeMountToIncludeRootCasVolumeMount(container *coreV1.Container) {
	tlsRootCasVolumeMountIndex := GetVolumeMountIndexForName(container, DesiredTlsRootCAsVolumeName)
	MutateVolumeMount(container, tlsRootCasVolumeMountIndex, DataPlaneConfigmapTlsRootCAsDirPath, DesiredTlsRootCAsVolumeMountReadOnly)
}

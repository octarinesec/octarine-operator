package v1

type ContainerRuntimeType string

func (containerRuntimeType ContainerRuntimeType) String() string {
	return string(containerRuntimeType)
}

const (
	ContainerdContainerRuntime ContainerRuntimeType = "containerd"
	DockerContainerRuntime     ContainerRuntimeType = "docker-daemon"
	CRIOContainerRuntime       ContainerRuntimeType = "cri-o"
)

type CBContainersContainerRuntimeSpec struct {
	Endpoint             string               `json:"endpoint,omitempty"`
	ContainerRuntimeType ContainerRuntimeType `json:"containerRuntimeType,omitempty"`
}

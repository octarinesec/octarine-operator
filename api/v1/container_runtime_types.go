package v1

type ContainerEngineType string

func (engineType ContainerEngineType) String() string {
	return string(engineType)
}

const (
	ContainerdContainerRuntime ContainerEngineType = "containerd"
	DockerContainerRuntime     ContainerEngineType = "docker-daemon"
	CRIOContainerRuntime       ContainerEngineType = "cri-o"
)

type K8sContainerEngineSpec struct {
	Endpoint   string              `json:"endpoint,omitempty"`
	EngineType ContainerEngineType `json:"engineType,omitempty"`
}

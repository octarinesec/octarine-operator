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

var SupportedK8sEngineTypes = []ContainerEngineType{ContainerdContainerRuntime, DockerContainerRuntime}

type K8sContainerEngineSpec struct {
	Endpoint string `json:"endpoint,omitempty"`
	// +kubebuilder:validation:Enum:=containerd;docker-daemon
	EngineType ContainerEngineType `json:"engineType,omitempty"`
}

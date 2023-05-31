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

var SupportedK8sEngineTypes = []ContainerEngineType{ContainerdContainerRuntime, DockerContainerRuntime, CRIOContainerRuntime}

type K8sContainerEngineSpec struct {
	Endpoint string `json:"endpoint,omitempty"`
	// +kubebuilder:validation:Enum:=containerd;docker-daemon;cri-o
	EngineType ContainerEngineType `json:"engineType,omitempty"`

	// +kubebuilder:default:=<>
	// CRIO holds configuration values specific to the CRI-O container engine
	CRIO CRIOSpec `json:"CRIO,omitempty"`
}

type CRIOSpec struct {
	// StoragePath can be used to set the path used by CRI-O to store images on each node.
	// This path will be mounted on the cluster scanner to provide access to the node's images.
	// If the path does not match what CRI-O uses on the nodes, then images will not be found and scanned as expected.
	// If not specified, the default location of CRI-O is used.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:="/var/lib/containers/storage"
	StoragePath string `json:"storagePath,omitempty"`

	// ConfigPath can be used to set the path to CRI-O's configuration file
	// If not specified, the default location for CRI-O is used.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:="/etc/containers/storage.conf"
	ConfigPath string `json:"configPath,omitempty"`
}

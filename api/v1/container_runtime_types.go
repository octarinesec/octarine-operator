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
	// If not specified, the default location of CRI-O is used (/var/lib/containers/storage).
	// +kubebuilder:validation:Optional
	StoragePath string `json:"storagePath,omitempty"`

	// StorageConfigPath can be used to set the path to the storage configuration file used by CRI-O (if any).
	// If not specified, the default location for storage is used (/etc/containers/storage.conf).
	// The files does not need to exist.
	// See https://github.com/containers/storage/blob/main/docs/containers-storage.conf.5.md for more information
	// +kubebuilder:validation:Optional
	StorageConfigPath string `json:"storageConfigPath,omitempty"`

	// ConfigPath can be used to set the path to CRI-O's configuration file.
	// If not specified, the default location is used (/etc/crio/crio.conf).
	// See https://github.com/cri-o/cri-o/blob/main/docs/crio.conf.5.md for more information.
	// +kubebuilder:validation:Optional
	ConfigPath string `json:"configPath,omitempty"`
}

package processors

type ClusterRegistrar interface {
	RegisterCluster() error
}

type CBContainerClusterProcessor struct {
	clusterRegistrar ClusterRegistrar
}

func NewCBContainerClusterProcessor(clusterRegistrar ClusterRegistrar) *CBContainerClusterProcessor {
	return &CBContainerClusterProcessor{
		clusterRegistrar: clusterRegistrar,
	}
}

func (processor *CBContainerClusterProcessor) Process() error {
	return processor.clusterRegistrar.RegisterCluster()
}

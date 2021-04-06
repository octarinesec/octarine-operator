package reporters

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	grpcCommunication "github.com/vmware/cbcontainers-operator/cbcontainers/communication/grpc"
	monitorModels "github.com/vmware/cbcontainers-operator/cbcontainers/monitor/models"
	pb "github.com/vmware/cbcontainers-operator/cbcontainers/monitor/protobuf"
	"google.golang.org/grpc"
	"time"
)

type GrpcMonitorReporter struct {
	connection *grpc.ClientConn
}

func NewGrpcMonitorReporter(host string, port uint32, certPool *x509.CertPool, cert *tls.Certificate) (*GrpcMonitorReporter, error) {
	connection, err := grpcCommunication.GetGrpcConnection(host, port, certPool, cert)
	if err != nil {
		return nil, err
	}

	return &GrpcMonitorReporter{
		connection: connection,
	}, nil
}

func (reporter *GrpcMonitorReporter) Close() error {
	return reporter.connection.Close()
}

func (reporter *GrpcMonitorReporter) SendMonitorMessage(message monitorModels.HealthReportMessage) error {
	grpcMessage, err := reporter.convertMessageToGrpcMessage(message)
	if err != nil {
		return err
	}

	client := pb.NewMonitorClient(reporter.connection)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err = client.HandleHealthReport(ctx, grpcMessage, grpc.WaitForReady(true))
	return err
}

func (reporter *GrpcMonitorReporter) convertMessageToGrpcMessage(message monitorModels.HealthReportMessage) (*pb.HealthReport, error) {
	servicesSpec, err := reporter.convertWorkloadsReportsToGrpcMessage(message.Workloads)
	if err != nil {
		return nil, err
	}

	webhooksSpec, err := reporter.convertWebhooksReportsToGrpcMessage(message.Webhooks)
	if err != nil {
		return nil, err
	}

	return &pb.HealthReport{
		Account:           message.Account,
		Domain:            message.Cluster,
		Version:           message.Version,
		EnabledComponents: message.EnabledComponents,
		Services:          servicesSpec,
		Webhooks:          webhooksSpec,
	}, nil
}

func (reporter *GrpcMonitorReporter) convertWorkloadsReportsToGrpcMessage(workloadsReports map[string]monitorModels.WorkloadHealthReport) (map[string]*pb.ServiceHealthReport, error) {
	reports := make(map[string]*pb.ServiceHealthReport)
	for workloadName, workloadReport := range workloadsReports {
		grpcKind, err := reporter.convertWorkloadKindToGrpcKind(workloadReport.Kind)
		if err != nil {
			return nil, err
		}

		reports[workloadName] = &pb.ServiceHealthReport{
			Kind:   grpcKind,
			Status: workloadReport.Status,
			Labels: workloadReport.Labels,
			Spec: &pb.ServiceSpec{
				Replicas:   workloadReport.Spec.Replicas,
				Containers: reporter.convertContainersReportsToGrpcMessage(workloadReport.Spec.Containers),
			},
			Replicas: reporter.convertWorkloadReplicasReportsToGrpcMessage(workloadReport.ReplicasReports),
		}
	}

	return reports, nil
}

func (reporter *GrpcMonitorReporter) convertContainersReportsToGrpcMessage(containersReports map[string]monitorModels.ContainerHealthReport) map[string]*pb.ContainerSpec {
	grpcContainersMessage := make(map[string]*pb.ContainerSpec)
	for containerName, containerReport := range containersReports {
		grpcContainersMessage[containerName] = &pb.ContainerSpec{Image: containerReport.Image}
	}

	return grpcContainersMessage
}

func (reporter *GrpcMonitorReporter) convertWorkloadKindToGrpcKind(kind monitorModels.WorkloadKind) (pb.ServiceHealthReport_Kind, error) {
	switch kind {
	case monitorModels.WorkloadKindDeployment:
		return pb.ServiceHealthReport_DEPLOYMENT, nil
	case monitorModels.WorkloadKindDaemonSet:
		return pb.ServiceHealthReport_DAEMONSET, nil
	}

	return 0, fmt.Errorf("failed to convert workload kind %v", kind)
}

func (reporter *GrpcMonitorReporter) convertWebhooksReportsToGrpcMessage(webhooksReports map[string]monitorModels.WebhookHealthReport) (map[string]*pb.WebhookHealthReport, error) {
	webhooksGrpcMessage := make(map[string]*pb.WebhookHealthReport)
	for webhookName, webhookReport := range webhooksReports {
		grpcType, err := reporter.convertWebhookTypeToGrpcType(webhookReport.Type)
		if err != nil {
			return nil, err
		}

		webhooksGrpcMessage[webhookName] = &pb.WebhookHealthReport{
			WebhookType: grpcType,
			Uid:         webhookReport.Uid,
		}
	}

	return webhooksGrpcMessage, nil
}

func (reporter *GrpcMonitorReporter) convertWebhookTypeToGrpcType(webhookType monitorModels.WebhookType) (pb.WebhookHealthReport_WebhookType, error) {
	switch webhookType {
	case monitorModels.WebhookTypeValidating:
		return pb.WebhookHealthReport_VALIDATING, nil
	case monitorModels.WebhookTypeMutating:
		return pb.WebhookHealthReport_MUTATING, nil
	}

	return 0, fmt.Errorf("failed to convert webhook type %v", webhookType)
}

func (reporter *GrpcMonitorReporter) convertWorkloadReplicasReportsToGrpcMessage(workloadReplicasReports map[string]monitorModels.WorkloadReplicaHealthReport) map[string]*pb.ReplicaHealth {
	grpcReplicasMessage := make(map[string]*pb.ReplicaHealth)
	for replicaName, replicaReport := range workloadReplicasReports {
		grpcReplicasMessage[replicaName] = &pb.ReplicaHealth{
			Node:   replicaReport.Node,
			Status: replicaReport.Status,
			Spec: &pb.ReplicaSpec{
				Containers: reporter.convertContainersReportsToGrpcMessage(replicaReport.Spec.Containers),
			},
		}
	}

	return grpcReplicasMessage
}

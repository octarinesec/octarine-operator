package grpc_client

import (
	"context"
	"fmt"
	pb "github.com/octarinesec/octarine-operator/pkg/monitor_agent/protobuf"
	"github.com/octarinesec/octarine-operator/pkg/octarine_api"
	"github.com/octarinesec/octarine-operator/pkg/types"
	"google.golang.org/grpc"
	"time"
)

type GRPCClient struct {
	Connection *grpc.ClientConn
}

func NewGRPCClient(apiClient *octarine_api.OctarineApiClient, octarineSpec *types.OctarineSpec) (*GRPCClient, error) {
	conn, err := getGrpcConnection(apiClient, octarineSpec)
	if err != nil {
		return nil, fmt.Errorf("couldn't create GRPC connection: %v", err)
	}

	return &GRPCClient{conn}, nil
}

func (gc *GRPCClient) Close() {
	gc.Connection.Close()
}

func (gc *GRPCClient) SendMonitorMessage(message *pb.HealthReport) error {
	client := pb.NewMonitorClient(gc.Connection)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := client.HandleHealthReport(ctx, message, grpc.WaitForReady(true))
	return err
}

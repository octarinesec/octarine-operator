package grpc_client

import (
	"crypto/tls"
	"fmt"
	"github.com/octarinesec/octarine-operator/pkg/octarine_api"
	"github.com/octarinesec/octarine-operator/pkg/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/encoding/gzip"
)

func getGrpcConnection(apiClient *octarine_api.OctarineApiClient, octarineSpec *types.OctarineSpec) (*grpc.ClientConn, error) {
	certPool, cert, err := apiClient.GetOctarineCertificates(octarineSpec.Global.Octarine.Account, octarineSpec.Global.Octarine.Domain, "monitor-agent")
	if err != nil {
		return nil, fmt.Errorf("failed getting Octarine certificates: %v", err)
	}

	host := fmt.Sprintf("%s:%d", octarineSpec.Global.Octarine.Messageproxy.Host, octarineSpec.Global.Octarine.Messageproxy.Port)

	creds := credentials.NewTLS(&tls.Config{
		ServerName:         "messageproxy",
		Certificates:       []tls.Certificate{*cert},
		RootCAs:            certPool,
		InsecureSkipVerify: true,
	})

	conn, err := grpc.Dial(host, grpc.WithTransportCredentials(creds), grpc.WithDefaultCallOptions(grpc.UseCompressor(gzip.Name)))
	if err != nil {
		return nil, err
	}

	return conn, nil
}

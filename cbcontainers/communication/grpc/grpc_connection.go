package grpc

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/encoding/gzip"
)

const (
	MessageProxyServerName = "messageproxy"
)

func GetGrpcConnection(host string, port int, certPool *x509.CertPool, cert *tls.Certificate) (*grpc.ClientConn, error) {
	address := fmt.Sprintf("%s:%d", host, port)
	tlsCredentials := credentials.NewTLS(&tls.Config{
		ServerName:         MessageProxyServerName,
		Certificates:       []tls.Certificate{*cert},
		RootCAs:            certPool,
		InsecureSkipVerify: true,
	})

	return grpc.Dial(address,
		grpc.WithTransportCredentials(tlsCredentials),
		grpc.WithDefaultCallOptions(grpc.UseCompressor(gzip.Name)),
	)
}

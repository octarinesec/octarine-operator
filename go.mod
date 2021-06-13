module github.com/vmware/cbcontainers-operator

go 1.15

require (
	github.com/cloudflare/cfssl v1.4.1
	github.com/go-logr/logr v0.4.0
	github.com/go-resty/resty/v2 v2.5.0
	github.com/golang/mock v1.6.0
	github.com/golang/protobuf v1.5.2
	github.com/stretchr/testify v1.7.0
	google.golang.org/grpc v1.36.1
	google.golang.org/protobuf v1.26.0
	k8s.io/api v0.20.5
	k8s.io/apimachinery v0.20.5
	k8s.io/client-go v0.20.5
	sigs.k8s.io/controller-runtime v0.8.3
)

replace github.com/vmware/cbcontainers-operator => ./

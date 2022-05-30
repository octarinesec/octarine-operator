module github.com/vmware/cbcontainers-operator

go 1.15

require (
	github.com/cloudflare/cfssl v1.4.1
	github.com/go-logr/logr v1.2.0
	github.com/go-resty/resty/v2 v2.5.0
	github.com/golang/mock v1.6.0
	github.com/stretchr/testify v1.7.0
	k8s.io/api v0.23.5
	k8s.io/apimachinery v0.23.5
	k8s.io/client-go v0.23.5
	sigs.k8s.io/controller-runtime v0.11.2
)

replace github.com/vmware/cbcontainers-operator => ./

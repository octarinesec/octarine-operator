module github.com/vmware/cbcontainers-operator

go 1.15

require (
	github.com/cloudflare/cfssl v1.4.1
	github.com/go-logr/logr v0.3.0
	github.com/go-resty/resty/v2 v2.5.0
	github.com/onsi/ginkgo v1.14.1
	github.com/onsi/gomega v1.10.2
	k8s.io/api v0.19.2
	k8s.io/apimachinery v0.19.2
	k8s.io/client-go v0.19.2
	sigs.k8s.io/controller-runtime v0.7.0
)

replace github.com/vmware/cbcontainers-operator => /Users/benrub/Development/octarinesec/octarine-operator

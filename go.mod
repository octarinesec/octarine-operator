module github.com/octarinesec/octarine-operator

go 1.13

require (
	github.com/cloudflare/cfssl v1.4.1
	github.com/go-logr/logr v0.1.0
	github.com/go-resty/resty/v2 v2.3.0
	github.com/golang/protobuf v1.3.2
	github.com/mitchellh/mapstructure v1.1.2
	github.com/operator-framework/operator-sdk v0.18.0
	github.com/peterbourgon/mergemap v0.0.0-20130613134717-e21c03b7a721
	github.com/redhat-cop/operator-utils v0.2.4
	github.com/spf13/pflag v1.0.5
	google.golang.org/grpc v1.27.0
	helm.sh/helm/v3 v3.2.0
	k8s.io/api v0.18.2
	k8s.io/apimachinery v0.18.2
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/kubernetes v1.13.0
	sigs.k8s.io/controller-runtime v0.6.0
)

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.2+incompatible // Required by OLM
	k8s.io/client-go => k8s.io/client-go v0.18.2 // Required by prometheus-operator
)

# Current Operator version
VERSION ?= 4.0.0
IMAGE_TAG_BASE ?= cbartifactory/octarine-operator
# Default bundle image tag
BUNDLE_IMG ?= $(IMAGE_TAG_BASE)-bundle:v$(VERSION)
# Options for 'bundle-build'
ifneq ($(origin CHANNELS), undefined)
BUNDLE_CHANNELS := --channels=$(CHANNELS)
endif
ifneq ($(origin DEFAULT_CHANNEL), undefined)
BUNDLE_DEFAULT_CHANNEL := --default-channel=$(DEFAULT_CHANNEL)
endif
BUNDLE_METADATA_OPTS ?= $(BUNDLE_CHANNELS) $(BUNDLE_DEFAULT_CHANNEL)

# Image URL to use all building/pushing image targets
IMG ?= controller:latest
# Image URL to use all building/pushing image targets
OPERATOR_REPLICAS ?= 1

CRD_OPTIONS ?= "crd:crdVersions=v1"
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion) - supported in v1beta1 only
CRD_OPTIONS_V1BETA1 ?= "crd:trivialVersions=true,crdVersions=v1beta1"

PATH_TO_RELEASE := config/default
PATH_TO_RELEASE := $(if $(findstring v1beta1, $(CRD_VERSION)), $(PATH_TO_RELEASE)_v1beta1, $(PATH_TO_RELEASE))

PATH_TO_CRDS := config/crd
PATH_TO_CRDS := $(if $(findstring v1beta1, $(CRD_VERSION)), $(PATH_TO_CRDS)_v1beta1, $(PATH_TO_CRDS))

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Setting SHELL to bash allows bash commands to be executed by recipes.
# This is a requirement for 'setup-envtest.sh' in the test target.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

all: build

OS = $(shell go env GOOS)
ARCH = $(shell go env GOARCH)


# Run tests
# Set default shell as bash
SHELL := /bin/bash
ENVTEST_ASSETS_DIR=$(shell pwd)/testbin
test: generate fmt vet manifests
	mkdir -p ${ENVTEST_ASSETS_DIR}
	test -f ${ENVTEST_ASSETS_DIR}/setup-envtest.sh || curl -sSLo ${ENVTEST_ASSETS_DIR}/setup-envtest.sh https://raw.githubusercontent.com/kubernetes-sigs/controller-runtime/v0.8.3/hack/setup-envtest.sh
	source ${ENVTEST_ASSETS_DIR}/setup-envtest.sh; fetch_envtest_tools $(ENVTEST_ASSETS_DIR); setup_envtest_env $(ENVTEST_ASSETS_DIR); go test ./... -coverprofile cover.out

# Build manager binary
manager: generate fmt vet
	go build -o bin/manager main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet manifests
	OPERATOR_VERSION=${VERSION} go run ./main.go

# Run with Delve for development purposes against the configured Kubernetes cluster in ~/.kube/config
# Delve is a debugger for the Go programming language. More info: https://github.com/go-delve/delve
# Note: use kill -SIGINT $pid to stop delve if it hangs
run-delve: generate fmt vet manifests
	go build -gcflags "all=-trimpath=$(shell go env GOPATH) -N -l" -o bin/manager main.go
	OPERATOR_VERSION=${VERSION} dlv --listen=:2345 --headless=true --api-version=2 --accept-multiclient exec ./bin/manager

install: manifests kustomize ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build $(PATH_TO_CRDS) | kubectl apply -f -

uninstall: manifests kustomize ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build $(PATH_TO_CRDS) | kubectl delete -f -

# Generate and bundle all operator components in a single YAML file
create_operator_spec: manifests kustomize
	rm -f operator.yaml
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG} && $(KUSTOMIZE) edit set replicas operator=${OPERATOR_REPLICAS}
	- $(KUSTOMIZE) build $(PATH_TO_RELEASE) >> operator.yaml
	git restore config/manager/kustomization.yaml

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: create_operator_spec
	- kubectl apply -f operator.yaml
	rm -f operator.yaml

# UnDeploy controller from the configured Kubernetes cluster in ~/.kube/config
undeploy: create_operator_spec
	- kubectl delete -f operator.yaml
	rm -f operator.yaml

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases
# This is needed since controller-gen does not support empty object/maps like {} but they are helpful to propagate default values up from nested objects
	for filename in $$(ls config/crd/bases) ; do \
		XT=config/crd/bases/$$filename ; \
		sed 's/default: <>/default: {}/g' $$XT >> $$XT.temp ; \
		rm -f $$XT ; \
		mv $$XT.temp $$XT ; \
	done
# The above modification is not needed for v1beta1 as defaults are not supported there at all
	$(CONTROLLER_GEN) $(CRD_OPTIONS_V1BETA1) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd_v1beta1/bases

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

# Generate code
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

# Build the docker image
docker-build:
	docker build -t ${IMG} --build-arg OPERATOR_VERSION=${VERSION} .

# Push the docker image
docker-push:
	docker push ${IMG}

# Download controller-gen locally if necessary
CONTROLLER_GEN = $(shell pwd)/bin/controller-gen
controller-gen:
	$(call go-get-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen@v0.4.1)

# Download kustomize locally if necessary
KUSTOMIZE = $(shell pwd)/bin/kustomize
kustomize:
	$(call go-get-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v3@v3.8.7)

# go-get-tool will 'go get' any package $2 and install it to $1.
PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
define go-get-tool
@[ -f $(1) ] || { \
set -e ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
go mod init tmp ;\
echo "Downloading $(2)" ;\
GOBIN=$(PROJECT_DIR)/bin go get $(2) ;\
rm -rf $$TMP_DIR ;\
}
endef

# Generate bundle manifests and metadata, then validate generated files.
.PHONY: bundle
bundle: manifests kustomize
	operator-sdk generate kustomize manifests -q
	cd config/manager && $(KUSTOMIZE) edit set image controller=$(IMG)
	$(KUSTOMIZE) build config/manifests | operator-sdk generate bundle -q --overwrite --version $(VERSION) $(BUNDLE_METADATA_OPTS)
	operator-sdk bundle validate ./bundle

# Build the bundle image.
.PHONY: bundle-build
bundle-build:
	docker build -f bundle.Dockerfile -t $(BUNDLE_IMG) .

.PHONY: opm
OPM = ./bin/opm
opm:
ifeq (,$(wildcard $(OPM)))
ifeq (,$(shell which opm 2>/dev/null))
	@{ \
	set -e ;\
	mkdir -p $(dir $(OPM)) ;\
	curl -sSLo $(OPM) https://github.com/operator-framework/operator-registry/releases/download/v1.15.1/$(OS)-$(ARCH)-opm ;\
	chmod +x $(OPM) ;\
	}
else
OPM = $(shell which opm)
endif
endif
BUNDLE_IMGS ?= $(BUNDLE_IMG)
CATALOG_IMG ?= $(IMAGE_TAG_BASE)-catalog:v$(VERSION) ifneq ($(origin CATALOG_BASE_IMG), undefined) FROM_INDEX_OPT := --from-index $(CATALOG_BASE_IMG) endif
.PHONY: catalog-build
catalog-build: opm
	$(OPM) index add --container-tool docker --mode semver --tag $(CATALOG_IMG) --bundles $(BUNDLE_IMGS) $(FROM_INDEX_OPT)

.PHONY: catalog-push
catalog-push: ## Push the catalog image.
	$(MAKE) docker-push IMG=$(CATALOG_IMG)
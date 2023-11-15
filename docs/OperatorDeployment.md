## Operator Deployment

### Prerequisites
Kubernetes 1.18+ is supported.

### From script:
```
export OPERATOR_VERSION=v6.1.0
export OPERATOR_SCRIPT_URL=https://setup.containers.carbonblack.io/$OPERATOR_VERSION/operator-apply.sh
curl -s $OPERATOR_SCRIPT_URL | bash
```

{OPERATOR_VERSION} is of the format "v{VERSION}"

Versions list: [Releases](https://github.com/octarinesec/octarine-operator/releases)

### From Source Code
Clone the git project and deploy the operator from the source code

By default, the operator utilizes CustomResourceDefinitions v1, which requires Kubernetes 1.16+.
Deploying an operator with CustomResourceDefinitions v1beta1 (deprecated in Kubernetes 1.16, removed in Kubernetes 1.22) can be done - see the relevant section below.

#### Create the operator image
```
make docker-build docker-push IMG={IMAGE_NAME}
```

#### Deploy the operator resources
```
make deploy IMG={IMAGE_NAME}
```

* View [Developer Guide](docs/developers.md) to see how deploy the operator without using an image

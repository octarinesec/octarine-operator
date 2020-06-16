# Octarine Operator development
This section contains relevant information for developing the Octarine operator.

## Prerequisites
Please make sure to install the following in order to develop & run the operator:
- operator-sdk (https://sdk.operatorframework.io/docs/install-operator-sdk/)
- Docker
- golang v1.14

## Run locally
```shell script
operator-sdk run local --watch-namespace "octarine-dataplane" --operator-flags='--zap-level=3'
```

## Build & Release
1. Update the version in `version/version.go`
2. Build the operator:
```shell script
operator-sdk build octarinesec/octarine-operator:<version>
```
3. Push the image to the registry:
```shell script
docker push octarinesec/octarine-operator:<version>
```
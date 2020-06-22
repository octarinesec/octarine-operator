# Octarine Operator development
This section contains relevant information for developing the Octarine operator.

## Prerequisites
Please make sure to install the following in order to develop & run the operator:
- operator-sdk (https://sdk.operatorframework.io/docs/install-operator-sdk/)
- Docker
- golang v1.14

## Run locally
> Before running the operator locally on your cluster for the first time, you'll need to deploy the operator resources:
> ```shell script
> kubectl apply -Rf deploy/
> ```
```shell script
OPERATOR_NAME=octarine-operator SERVICE_ACCOUNT_NAME=octarine-operator IMAGE_PULL_SECRET_NAME=octarine-operator-registry-secret operator-sdk run local --watch-namespace "octarine-dataplane" --operator-flags='--zap-level=3'
```

## Build & Release
1. Update the version in `version/version.go` and in `helm-charts/octarine-operator/Chart.yaml`
2. Build the operator:
```shell script
operator-sdk build octarinesec/octarine-operator:<version>
```
3. Push the image to the registry:
```shell script
docker push octarinesec/octarine-operator:<version>
```
4. Create helm chart package:
```shell script
helm package helm-charts/octarine-operator -d helm-repo
```
5. Update helm repo index:
```shell script
helm repo index --url https://octarinesec.github.io/octarine-operator .
```
6. Push the chart package and the updated `index.yaml` to GitHub (`master` branch)
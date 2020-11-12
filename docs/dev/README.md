# Octarine Operator development
This section contains relevant information for developing the Octarine operator.

## Prerequisites
Please make sure to install the following in order to develop & run the operator:
- operator-sdk (https://sdk.operatorframework.io/docs/install-operator-sdk/)
- Docker
- golang v1.14
- Helm 2 or 3

## Run locally
> Before running the operator locally on your cluster for the first time, you'll need to deploy the operator resources:
> ```shell script
> kubectl create ns octarine-dataplane
> kubectl apply -n octarine-dataplane -f deploy/crds/operator.octarinesec.com_octarines_crd.yaml
> kubectl apply -n octarine-dataplane -f deploy/cluster_role.yaml
> kubectl apply -n octarine-dataplane -f deploy/cluster_role_binding.yaml
> kubectl apply -n octarine-dataplane -f deploy/role.yaml
> kubectl apply -n octarine-dataplane -f deploy/role_binding.yaml
> kubectl apply -n octarine-dataplane -f deploy/service_account.yaml
> ```
```shell script
OPERATOR_NAME=octarine-operator SERVICE_ACCOUNT_NAME=octarine-operator IMAGE_PULL_SECRET_NAME=octarine-operator-registry-secret operator-sdk run local --watch-namespace "octarine-dataplane" --operator-flags='--zap-level=3'
```

*After running the operator, refer to the [Custom Resource documentation](docs/octarine_cr.md) in order to deploy Octarine CR.*

## Modify the backend (CP) address
By default, the DP will work with `main.octarinesec.com`.  
In order to work with your backend, add the `api` and `messageproxy` params to the Octarine CR:
```yaml
cat <<EOF | kubectl apply --namespace octarine-dataplane -f -
apiVersion: operator.octarinesec.com/v1alpha1
kind: Octarine
metadata:
  name: octarine
spec:
  global:
    octarine:
      version: <octarine version>
      account: <account (CB org key)>
      domain: <domain, group:member>
      accessTokenSecret: octarine-access-token
      api:
        host: <api address>
        port: 443
        adapterName: <api-adapter name, eg. octarine-<release name>>
      messageproxy:
        host: <messageproxy address>
        port: 50051
EOF
```

## Build & Release
1. Update the version in `version/version.go`, `helm-charts/octarine-operator/Chart.yaml` & `helm-charts/octarine-operator/values.yaml` (`image.tag`) & `deploy/operator.yaml` (`spec.spec.containers.image`) 
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
6. Push the chart package and the updated resources to GitHub (`master` branch)
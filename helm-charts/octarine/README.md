# Octarine Helm Charts
This repo contains helm chart for installing Octarine dataplane.

## Prerequisites
- Kubernetes 1.3+

- Helm installed, Tiller pod is running

## Deployment
1. Label the `octarine-dataplane` namespace if it already exists (if it doesn't exist, helm will automatically create & label it): 
```sh
kubectl label namespace octarine-dataplane name=octarine-dataplane
```
>Labeling the namespace is required due to usage of a validating webhook.

2. Install the chart with the release name `octarine` in the `octarine-dataplane` namespace:
```sh
helm install --name octarine --namespace octarine-dataplane ./octarine --set imageCredentials.username=<docker username> --set imageCredentials.password=<docker password>  --set global.octarine.account=<your account name> --set global.octarine.domain=<the domain name> --set global.octarine.accessToken=<your access token>
```

## Rolling upgrade
Upgrade the `octarine` release to the desired version:
```sh
helm upgrade octarine ./octarine --reuse-values --set global.octarine.version=<version>
```

## Uninstalling the chart
Uninstall the `octarine` release:
```sh
helm delete octarine
```

## Customize the configuration
The following table lists the configurable parameters of the octarine chart and their default values.

### Required
Parameter | Description | Default
--------- | ----------- | -------
`imageCredentials.username` | The Docker registry username | `nil`
`imageCredentials.password` | The Docker registry password | `nil`
>The Docker username & password will be provided to you by Octarine

### Global
Parameter | Description | Default
--------- | ----------- | -------
`global.imagePullSecret` | The name of an **existing** imagePullSecret to use.<br>If provided, imageCredentials will be ignored and an imagePullSecret won't be created. | `nil`
`guardrails.enabled` | Install and enable Guardrails | `true`
`nodeguard.enabled` | Install and enable Nodeguard | `true`

## Container registry credentials
The Octarine componenets are available in our private repository.  
By default, the chart creates an imagePullSecret based on the imageCredentials in the `values.yaml`.

You can create & provide your own imagePullSecret:
1. Create & label the octarine-dataplane namespace:
```sh
kubectl create namespace octarine-dataplane
kubectl label namespace octarine-dataplane name=octarine-dataplane
```
>Labeling the namespace is required due to usage of a validating webhook.

2. Create image pull secret for Octarine's private DockerHub registry:
```sh
kubectl create secret docker-registry octarine-registry-secret -n octarine-dataplane --docker-server=https://index.docker.io/v1/ --docker-username=<your username> --docker-password=<your password> --docker-email=<your email>
```

3. Provide the created secret name when installing the chart in the `global.imagePullSecret` param of the `values.yaml`

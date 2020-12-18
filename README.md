# VMware Carbon Black Cloud Container Operator 
## Cloud Container Operator Overview 

The the Carbon Black Cloud Container Operator runs within a Kubernetes cluster.The Container Operator is a set of controllers which deploy and manage the VMware Carbon Black Cloud Container components. 
 
 Capabilities
 * Deploy and manage the Container Essentials product bundle (including the configuration and the image scanning for Kubernetes security)! 
 * Deploy and manage the Container Advanced product bundle (including the runtime for Kubernetes security) 
 * Automatically fetch and deploy the Carbon Black Cloud Container private image registry secret
 * Automatically register the Carbon Black Cloud Container cluster
 * Manage the Container Essentials validationng webhook - dynamically manage the admission control webhook to avoid possible downtime
 * Monitor and report agent availability to the Carbon Black console

The Carbon Black Cloud Container Operator utilizes the operator-framework to create a hybrid operator, which combines helm and go operators. 
The helm controller within the operator is responsible for managing the Cloud Container components deployment, and the go controller monitors and manages them. 

## Deployment and Upgrade 
### Prerequisites
Kubernetes 1.13+ 

### Deploy using Helm 
The operator Helm chart supports Helm 3 and Helm 2 (if you're using Helm 2, make sure that Tiller pod is running).
Install the chart with the release name octarine in the octarine-dataplane namespace: 

```sh
helm repo add octarine-operator https://octarinesec.github.io/octarine-operatorhelm repo updatehelm upgrade --install --namespace octarine-dataplane octarine-operator octarine-operator/octarine-operator 
```
### Deploy K8s resources 
Create the octarine-dataplane namespace and deploy the resources from the deploy dir:
```
kubectl create namespace octarine-dataplane
kubectl label namespace octarine-dataplane name=octarine-dataplane
kubectl apply -n octarine-dataplane -Rf deploy
```
The namespace label is required due to the validating webhook configured by Container Essentials, and in order to ensure the service availability. 

*After deploying the Operator,please refer to the Octarine [Custom Resource documentation](docs/octarine_cr.md) in order to deploy Octarine dataplane components.*

### Rolling upgrade 
The deployment command can be used for upgrading as well.  
Note: Make sure to update the helm repo before. 
```sh
helm repo updatehelm upgrade --install --namespace octarine-dataplane octarine-operator octarine-operator/octarine-operator 
```
### Uninstalling the Carbon Black Cloud Container Operator 
You can uninstall the Carbon Black Cloud Container Operator in three ways: 

1. If you created an Octarine Carbon Black Cloud Container resource to install the Carbon Black components, please delete it before uninstalling the operator: 

```sh
kubectl delete octarines.operator.octarinesec.com octarine 
```

2. If you deployed the operator using Helm Uninstall the octarine release: 
```sh 
helm delete octarine-operator 
```

Delete the Octarine CRD which was created by helm: 
```sh
kubectl delete crd octarines.operator.octarinesec.com 
```

3. If you deployed the operator using its plain K8s resources, uninstall it by running: 

```sh
kubectl delete -n octarine-dataplane -Rf deploy 
```

## Customize the configuration 

The following table lists the configurable parameters of the operator chart and their default values. 
| Parameter      | Description                                | Default |
| -------------- | ------------------------------------------ | ------- |
| `replicaCount` | The number of the operator replicas to run | `1`     |
## Logs 
You can enable verbose logging and set the verbosity level using the --zap-level flag of the operator executable. 
See the args value within the values.yaml.

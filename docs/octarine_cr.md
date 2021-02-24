# Carbon Black Cloud Container Custom Resource 

The Carbon Black Cloud Container operator watches a [Kubernetes Custom Resource (CR)](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/) of kind Octarine, group `operator.octarinesec.com`, version `v1alpha1`. 
The configurations defined by the Carbon Black Custom Resource will be used by the operator to deploy the agent component

## Installing Cloud Container components
1. Create a Carbon Black Cloud access token. The token is in the form `<token>/<token ID>`.

2. Create a secret with the access token in the namespace in which the operator is running:

> Replace `<CB Token>` with the accessjwt you copied in the previous step

```shell script
kubectl create secret generic octarine-access-token --namespace octarine-dataplane --from-literal=accessToken=<CB Token>
```

3. Create the Cloud Container CR in the namespace in which the operator is running:
> Replace the `agent version`, `org-id` and `cluster` with your values.  
> If you created the access token secret with a different name than the one in the previous step, change the `accessTokenSecret` value accordingly.

```shell script
cat <<EOF | kubectl apply --namespace octarine-dataplane -f -
apiVersion: operator.octarinesec.com/v1alpha1
kind: Octarine
metadata:
  name: octarine
spec:
  global:
    octarine:
      version: <agent version>
      account: <org-id>
      domain: <cluster, group:member>
      accessTokenSecret: octarine-access-token
EOF
```

## Cloud Container CR spec
The `spec` of the Cloud Container CR overrides the default values of the Octarine dataplane helm chart (`values.yaml`). You can override any default value by adding it to the CR spec.

### Required parameters
| Parameter                           | Description                                                     | Default |
| ----------------------------------- | --------------------------------------------------------------- | ------- |
| `global.octarine.version`           | Version of the Carbon Black Cloud components                    |
| `global.octarine.account`           | Carbon Black Cloud account name                                 |
| `global.octarine.domain`            | Carbon Black Cloud cluster name                                 |
| `global.octarine.accessTokenSecret` | Name of a secret containing the Carbon Black Cloud access token |

### Optional parameters
> This is a partial list of the optional parameters. For the full parameters list, please refer to the Cloud Container chart in `helm-charts/octarine`.

| Parameter                                   | Description                                                                                                                                                               | Default |
| ------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------- |
| `guardrails.enabled`                        | Whether Container Essentials should be deployed or not                                                                                                                    | `true`  |
| `guardrails.admissionController.autoManage` | If `true`, the operator will deploy the Container Essential validating webhook config when Guardrails service is available and delete it when the service is unavailable. | `true`  |
| `guardrails.enforcer.replicaCount`          | The number of the enforcer replicas to run                                                                                                                                | `1`     |
| `guardrails.enforcer.env`                   | Configure additional environment variables for the enforcer. example: HTTP_PROXY, HTTPS_PROXY, NO_PROXY                                                                    | `1`     |
| `guardrails.stateReporter.env`              | Configure additional environment variables for the state-reporter. example: HTTP_PROXY, HTTPS_PROXY, NO_PROXY                                                              | `1`     |

## Updating Cloud Container components configuration
You can update any of the configurable parameters in an existing Cloud Container CR spec. The operator will perform the corresponding changes in the Cloud Container components.

## Uninstalling Carbon Black Cloud Container components
You can uninstall Carbon Black Cloud Container components by deleting the Cloud Container CR:
```shell script
kubectl delete octarines.operator.octarinesec.com octarine
```

# Octarine Custom Resource
The Octarine operator watches a CR of kind `Octarine`, group `operator.octarinesec.com`, version `v1alpha1`.  
The Octarine custom resource defines the parameters for the Octarine components. The operator will deploy the Octarine components based on this custom resource.

## Installing Octarine components
1. Create a Carbon Black Cloud access token. The token is in the form `<token>/<token ID>`.

2. Create a secret with the access token in the namespace in which the operator is running:

> Replace `<CB Token>` with the accessjwt you copied in the previous step

```shell script
kubectl create secret generic octarine-access-token --namespace octarine-dataplane --from-literal=accessToken=<CB Token>
```

3. Create the Octarine CR in the namespace in which the operator is running:
> Replace the `octarine version`, `account` and `domain` with your values.  
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
      version: <octarine version>
      account: <account>
      domain: <domain, group:member>
      accessTokenSecret: octarine-access-token
EOF
```

## Octarine CR spec
The `spec` of the Octarine CR overrides the default values of the Octarine dataplane helm chart (`values.yaml`). You can override any default value by adding it to the CR spec.

### Required parameters
Parameter | Description | Default
--------- | ----------- | -------
`global.octarine.version` | Version of the Octarine components | 
`global.octarine.account` | Octarine account name | 
`global.octarine.domain` | Octarine cluster name | 
`global.octarine.accessTokenSecret` | Name of a secret containing the Octarine access token |

### Optional parameters
> This is a partial list of the optional parameters. For the full parameters list, please refer to the Octarine chart in `helm-charts/octarine`.

Parameter | Description | Default
--------- | ----------- | -------
`guardrails.enabled` | Whether Guardrails should be deployed or not | `true`
`guardrails.admissionController.autoManage` | If `true`, the operator will deploy the Guardrails validating webhook config when Guardrails service is available and delete it when the service is unavailable. | `true`
`guardrails.enforcer.replicaCount` | The number of the enforcer replicas to run | `1`
`nodeguard.enabled` | Whether Nodeguard should be deployed or not | `true`
`nodeguard.worker.interfacePrefixes.container` | Prefix of the container network interface | `NO_PREFIX`
`nodeguard.worker.interfacePrefixes.external` | Prefix of the external network interface | `NO_PREFIX`

## Updating Octarine components configuration
You can update any of the configurable parameters in an existing Octarine CR spec. The operator will perform the corresponding changes in the Octarine components.

## Uninstalling Octarine components
You can uninstall Octarine components by deleting the Octarine CR:
```shell script
kubectl delete octarines.operator.octarinesec.com octarine
```

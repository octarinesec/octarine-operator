# Octarine Custom Resource
The Octarine operator watches a CR of kind `Octarine`, group `operator.octarinesec.com`, version `v1alpha1`.  
The Octarine custom resource defines the parameters for the Octarine components. The operator will deploy the Octarine components based on this custom resource.

## Installing Octarine components
In order to deploy Octarine components to your cluster, replace the `version`, `account` and `domain` in the following command and execute it on the same namespace in which the operator is deployed:
```shell script
cat <<EOF | kubectl apply -f -
apiVersion: operator.octarinesec.com/v1alpha1
kind: Octarine
metadata:
  name: octarine
spec:
  global:
    octarine:
      version: <octarine dataplane version>
      account: <account name>
      domain: <octarine domain, group:member>
EOF
```

## Octarine CR spec
The `spec` of the Octarine CR overrides the default values of the Octarine dataplane helm chart (`values.yaml`). You can override any default value by adding it to the CR spec.

These are some main configurable parameters:

Parameter | Description | Default
--------- | ----------- | -------
`global.octarine.version` | Version of the Octarine components | 
`global.octarine.account` | Octarine account name | 
`global.octarine.domain` | Octarine cluster name | 
`guardrails.enabled` | Whether Guardrails should be deployed or not | `true`
`guardrails.admissionController.autoManage` | If `true`, the operator will deploy the Guardrails validating webhook config when Guardrails service is available and delete it when the service is unavailable. | `true`
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
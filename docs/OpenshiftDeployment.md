## Deploying on Openshift

The operator and its agent require elevated permissions to operate properly. However, this violates the default SecurityContextConstraints on most Openshift clusters, hence the components fail to start.
This can be fixed by applying the following custom security constraint configurations on the cluster (cluster admin priveleges required).

```yaml
kind: SecurityContextConstraints
apiVersion: security.openshift.io/v1
metadata:
  name: scc-anyuid
runAsUser:
  type: MustRunAsNonRoot
allowHostPID: false
allowHostPorts: false
allowHostNetwork: false
allowHostDirVolumePlugin: false
allowHostIPC: false
allowPrivilegedContainer: false
readOnlyRootFilesystem: true
seLinuxContext:
  type: RunAsAny
fsGroup:
  type: RunAsAny
supplementalGroups:
  type: RunAsAny
users:
- system:serviceaccount:cbcontainers-dataplane:cbcontainers-operator
- system:serviceaccount:cbcontainers-dataplane:cbcontainers-enforcer
- system:serviceaccount:cbcontainers-dataplane:cbcontainers-state-reporter
- system:serviceaccount:cbcontainers-dataplane:cbcontainers-monitor
- system:serviceaccount:cbcontainers-dataplane:cbcontainers-runtime-resolver
---
kind: SecurityContextConstraints
apiVersion: security.openshift.io/v1
metadata:
  name: scc-image-scanning # This probably needs to be fixed in the actual deployment
runAsUser:
  type: RunAsAny
allowHostPID: false
allowHostPorts: false
allowHostNetwork: false
allowHostDirVolumePlugin: false
allowHostIPC: false
allowPrivilegedContainer: false
readOnlyRootFilesystem: false
seLinuxContext:
  type: RunAsAny
fsGroup:
  type: RunAsAny
supplementalGroups:
  type: RunAsAny
allowedCapabilities:
- 'NET_BIND_SERVICE'
users:
- system:serviceaccount:cbcontainers-dataplane:cbcontainers-image-scanning
---
kind: SecurityContextConstraints
apiVersion: security.openshift.io/v1
metadata:
  name: scc-node-agent
runAsUser:
  type: RunAsAny
allowHostPID: true
allowHostPorts: false
allowHostNetwork: true
allowHostDirVolumePlugin: true
allowHostIPC: false
allowPrivilegedContainer: true
readOnlyRootFilesystem: false
seLinuxContext:
  type: RunAsAny
fsGroup:
  type: RunAsAny
supplementalGroups:
  type: RunAsAny
volumes:
- configMap
- downwardAPI
- emptyDir
- hostPath
- persistentVolumeClaim
- projected
- secret
users:
- system:serviceaccount:cbcontainers-dataplane:cbcontainers-agent-node
```
### Uninstalling on Openshift

Add this SecurityContextConstraints
before running the operator uninstall command

```yaml
kind: SecurityContextConstraints
apiVersion: security.openshift.io/v1
metadata:
  name: scc-edr-cleaner
runAsUser:
  type: RunAsAny
allowHostPID: true
allowHostPorts: false
allowHostNetwork: true
allowHostDirVolumePlugin: true
allowHostIPC: false
allowPrivilegedContainer: true
readOnlyRootFilesystem: false
seLinuxContext:
  type: RunAsAny
fsGroup:
  type: RunAsAny
supplementalGroups:
  type: RunAsAny
volumes:
- configMap
- downwardAPI
- emptyDir
- hostPath
- persistentVolumeClaim
- projected
- secret
users:
- system:serviceaccount:cbcontainers-edr-sensor-cleaners:cbcontainers-edr-sensor-cleaner
```
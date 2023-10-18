## Reading Metrics With Prometheus

The operator metrics are protected by kube-auth-proxy.

You will need to grant permissions to your Prometheus server to allow it to scrape the protected metrics.

You can create a ClusterRole and bind it with ClusterRoleBinding to the service account that your Prometheus server uses.

If you don't have such cluster role & cluster role binding configured, you can use the following:

Cluster Role:
```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
name: cbcontainers-metrics-reader
rules:
- nonResourceURLs:
    - /metrics
      verbs:
      - get
```

Cluster Role binding creation:
```sh
kubectl create clusterrolebinding metrics --clusterrole=cbcontainers-metrics-reader --serviceaccount=<prometheus-namespace>:<prometheus-service-account-name>
```

### When using Prometheus Operator

Use the following ServiceMonitor to start scraping metrics from the CBContainers operator:
* Make sure that your Prometheus custom resource service monitor selectors match it.
```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  labels:
    control-plane: operator
  name: cbcontainers-operator-metrics-monitor
  namespace: cbcontainers-dataplane
spec:
  endpoints:
  - bearerTokenFile: /var/run/secrets/kubernetes.io/serviceaccount/token
    path: /metrics
    port: https
    scheme: https
    tlsConfig:
      insecureSkipVerify: true
  selector:
    matchLabels:
      control-plane: operator
```
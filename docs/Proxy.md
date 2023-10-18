## Using HTTP proxy

Configuring the Carbon Black Cloud Container services to use HTTP proxy can be done by enabling the centralized proxy settings or by setting HTTP_PROXY, HTTPS_PROXY and NO_PROXY environment variables manually.
The centralized proxy settings apply an HTTP proxy configuration for all components, while the manual setting of environment variables allows this to be done on a per component basis.
If both HTTP proxy environment variables and centralized proxy settings are provided, the environment variables would take precedence.
The operator does not make use of the centralized proxy settings, so you have to use the environment variables for it instead.

### Configure centralized proxy settings

In order to configure the proxy environment variables in the Operator, use the following command to patch the Operator deployment:
```sh
kubectl set env -n cbcontainers-dataplane deployment cbcontainers-operator HTTP_PROXY="<proxy-url>" HTTPS_PROXY="<proxy-url>" NO_PROXY="<kubernetes-api-server-ip>/<range>"
```

Update the `CBContainersAgent` CR with the centralized proxy settings (`kubectl edit cbcontainersagents.operator.containers.carbonblack.io cbcontainers-agent`):

```yaml
spec:
  components:
    settings:
      proxy:
        enabled: true
        httpProxy: "<proxy-url>"
        httpsProxy: "<proxy-url>"
        noProxy: "<exclusion1>,<exclusion2>"
```

You can disable the centralized proxy settings without having to delete them, by setting the `enabled` key above to `false`.

By default, the centralized proxy settings take care of determining the API server IP address(es) and the necessary proxy exclusions for the cbcontainers-dataplane namespace.
These determined values are automatically appended to the `noProxy` values from above or the specified `NO_PROXY` environment variable for a particular component.
However, if you wish to change those pre-determined values, you can specify the `noProxySuffix` key at the same level as the `noProxy` key.
It has the same format as the `noProxy` key and its values are treated in the same way as if they were pre-determined.
One can also force nothing to be appended to `noProxy` or `NO_PROXY` by setting `noProxySuffix` to an empty string.

### Configure HTTP proxy environment variables (per component proxy settings)

In order to configure those environment variables for the basic, Runtime and Image Scanning  components,
update the `CBContainersAgent` CR using the proxy environment variables (`kubectl edit cbcontainersagents.operator.containers.carbonblack.io cbcontainers-agent`):

```yaml
spec:
  components:
    basic:
      enforcer:
        env:
          HTTP_PROXY: "<proxy-url>"
          HTTPS_PROXY: "<proxy-url>"
          NO_PROXY: "<kubernetes-api-server-ip>/<range>"
      stateReporter:
        env:
          HTTP_PROXY: "<proxy-url>"
          HTTPS_PROXY: "<proxy-url>"
          NO_PROXY: "<kubernetes-api-server-ip>/<range>"
    runtimeProtection:
      resolver:
        env:
          HTTP_PROXY: "<proxy-url>"
          HTTPS_PROXY: "<proxy-url>"
          NO_PROXY: "<kubernetes-api-server-ip>/<range>"
      sensor:
        env:
          HTTP_PROXY: "<proxy-url>"
          HTTPS_PROXY: "<proxy-url>"
          NO_PROXY: "<kubernetes-api-server-ip>/<range>,cbcontainers-runtime-resolver.cbcontainers-dataplane.svc.cluster.local"
    clusterScanning:
      clusterScanner:
        env:
          HTTP_PROXY: "<proxy-url>"
          HTTPS_PROXY: "<proxy-url>"
          NO_PROXY: "<kubernetes-api-server-ip>/<range>,cbcontainers-image-scanning-reporter.cbcontainers-dataplane.svc.cluster.local"
      imageScanningReporter:
        env:
          HTTP_PROXY: "<proxy-url>"
          HTTPS_PROXY: "<proxy-url>"
          NO_PROXY: "<kubernetes-api-server-ip>/<range>"
```

It is very important to configure the NO_PROXY environment variable with the value of the Kubernetes API server IP.

Finding the API-server IP:
```sh
kubectl -n default get service kubernetes -o=jsonpath='{..clusterIP}'
```

### Other proxy considerations

When using non-transparent HTTPS proxy you will need to configure the agent to use the proxy certificate authority:
```yaml
spec:
  gateways:
    gatewayTLS:
      rootCAsBundle: <Base64 encoded proxy CA>
```
Another option will be to allow the agent communicate without verifying the certificate. this option is not recommended and exposes the agent to MITM attack.
```yaml
spec:
  gateways:
    gatewayTLS:
      insecureSkipVerify: true
```
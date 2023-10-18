## Changing the source of the images

By default, all the images for the operator and agent deployment are going to be pulled from Docker Hub.

We understand that some companies might not want to pull images from Docker Hub and would prefer to mirror them into their internal repositories.

For that reason, we allow specifying the image yourself.
To do that modify the `CBContainersAgent` resource you're applying to your cluster.

Modify the following properties to specify the image for each service:

- monitor - `spec.components.basic.monitor.image`
- enforcer - `spec.components.basic.enforcer.image`
- state-reporter - `spec.components.basic.stateReporter.image`
- runtime-resolver - `spec.components.runtimeProtection.resolver.image`
- runtime-sensor - `spec.components.runtimeProtection.sensor.image`
- image-scanning-reporter - `spec.components.clusterScanning.imageScanningReporter.image`
- cluster-scanner - `spec.components.clusterScanning.clusterScanner.image`

The `image` object consists of 4 properties:

- `repository` - the repository of the image, e.g. `docker.io/my-org/monitor`
- `tag` - the version tag of the image, e.g. `1.0.0`, `latest`, etc.
- `pullPolicy` - the pull policy for that image, e.g. `IfNotPresent`, `Always`, or `Never`.
  See [docs](https://kubernetes.io/docs/concepts/containers/images/#image-pull-policy).
- `pullSecrets` - the image pull secrets that are going to be used to pull the container images.
  The secrets must already exist in the cluster.
  See [docs](https://kubernetes.io/docs/tasks/configure-pod-container/pull-image-private-registry/).

A sample configuration can look like this:

```yaml
spec:
  monitor:
    image:
      repository: docker.io/my-org/monitor
      tag: 1.0.0
      pullPolicy: Always
      pullSecrets:
        - my-pull-secret
```

This means that the operator will try to run the monitor service from the `docker.io/my-org/monitor:1.0.0` container image and the kubelet will be instruted to **always** pull the image, using the `my-pull-secret` secret.

### Using a shared secret for all images

If you want to use just one pull secret to pull all the custom images, you don't need to add it every single image configuration.
Instead you can specify it(them) under `spec.settings.imagePullSecrets`.

The secrets you put on that list will be added to the `imagePullSecrets` list of ALL agent workloads.

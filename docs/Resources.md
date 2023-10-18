## Changing components resources:
```yaml
spec:
  components:
    basic:
      monitor:
        resources:
          limits:
            cpu: 200m
            memory: 256Mi
          requests:
            cpu: 30m
            memory: 64Mi
      enforcer:
        resources:
          #### DESIRED RESOURCES SPEC - for hardening enforcer container
      stateReporter:
        resources:
          #### DESIRED RESOURCES SPEC - for hardening state reporter container
    runtimeProtection:
      resolver:
        resources:
          #### DESIRED RESOURCES SPEC - for runtime resolver container
      sensor:
        resources:
          #### DESIRED RESOURCES SPEC - for node-agent runtime container
    clusterScanning:
      imageScanningReporter:
        resources:
          #### DESIRED RESOURCES SPEC - for image scanning reporter pod
      clusterScanner:
        resources:
          #### DESIRED RESOURCES SPEC - for node-agent cluster-scanner container
```
#### Cluster Scanner Component Memory
The `clusterScanning.clusterScanner` component, tries by default to scan images with size up to 1GB.
To do so, its recommended resources are:
```yaml
resources:
  requests:
    cpu: 100m
    memory: 1Gi
  limits:
    cpu: 2000m
    memory: 6Gi
```

If your images are larger than 1GB, and you want to scan them, you'll need to allocate higher memory resources in the
component's `requests.memory` & `limits.memory`, and add an environment variable `MAX_COMPRESSED_IMAGE_SIZE_MB`, to override
the max images size in MB, the scanner tries to scan.

For example, setting the cluster scanner to be able to scan images up to 1.5 GB configuration will be:
```yaml
spec:
  components:
    clusterScanning:
      clusterScanner:
        env:
          MAX_COMPRESSED_IMAGE_SIZE_MB: "1536" // 1536 MB == 1.5 GB
        resources:
          requests:
            cpu: 100m
            memory: 2Gi
          limits:
            cpu: 2000m
            memory: 5Gi
```

If your nodes have low memory, and you want the cluster scanner to consume less memory, you need to reduce the
component's `requests.memory` & `limits.memory` , and override the `MAX_COMPRESSED_IMAGE_SIZE_MB`, to be less than 1GB (1024MB).

For example, assigning lower memory resources, and set the cluster-scanner to try and scan images up to 250MB:
```yaml
spec:
  components:
    clusterScanning:
      clusterScanner:
        env:
          MAX_COMPRESSED_IMAGE_SIZE_MB: "250" // 250 MB
        resources:
          requests:
            cpu: 100m
            memory: 250Mi
          limits:
            cpu: 2000m
            memory: 1Gi
```

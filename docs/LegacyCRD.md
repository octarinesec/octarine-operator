## Utilizing v1beta1 CustomResourceDefinition versions
The operator supports Kubernetes clusters from v1.13+.
The CustomResourceDefinition APIs were in beta stage in those cluster and were later promoted to GA in v1.16. They are no longer served as of v1.22 of Kubernetes.

To maintain compatibility, this operator offers 2 sets of CustomResourceDefinitions - one under the `apiextensions/v1beta1` API and one under `apiextensons/v1`.

By default, all operations in the repository like `deploy` or `install` work with the v1 version of the `apiextensions` API. Utilizing `v1beta1` is supported by passing the `CRD_VERSION=v1beta1` option when running make.
Note that both `apiextensions/v1` and `apiextensions/v1beta1` versions of the CRDs are generated and maintained by `make` - only commands that use the final output work with 1 version at a time.

For example, this command will deploy the operator resources on the current cluster but utilizing the `apiextensions/v1beta1` API version for them.

```
make deploy CRD_VERSION=v1beta1
```
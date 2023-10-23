# VMware Carbon Black Cloud Container Operator
## Overview

The Carbon Black Cloud Container Operator runs within a Kubernetes cluster. The Container Operator is a set of controllers which deploy and manage the VMware Carbon Black Cloud Container components.

Capabilities
* Deploy and manage the Container Essentials product bundle (including the configuration and the image scanning for Kubernetes security)!
* Automatically fetch and deploy the Carbon Black Cloud Container private image registry secret
* Automatically register the Carbon Black Cloud Container cluster
* Manage the Container Essentials validating webhook - dynamically manage the admission control webhook to avoid possible downtime
* Monitor and report agent availability to the Carbon Black console

The Carbon Black Cloud Container Operator utilizes the operator-framework to create a GO operator, which is responsible for managing and monitoring the Cloud Container components deployment.

## Compatibility Matrix

| Operator version | Kubernetes Sensor Component Version | Minimum Kubernetes Version |
|------------------|-------------------------------------|----------------------------|
| v6.0.x           | 2.10.0, 2.11.0, 2.12.0, 3.0.0       | 1.18                       |
| v5.6.x           | 2.10.0, 2.11.0, 2.12.0              | 1.16                       |
| v5.5.x           | 2.10.0, 2.11.0                      | 1.16                       |

## Install

First, you need to install the CBC operator on the cluster:

[Operator Deployment](OperatorDeployment.md)

Then you need to deploy the CBC Agent on top of the operator:

[Agent Deployment](AgentDeployment.md)



For OpenShift clusters, follow the OpenShift Deployment instructions:

[OpenShift Deployment and Uninstall](OpenshiftDeployment.md)


## Full Uninstall

### Uninstalling the Carbon Black Cloud Container Operator

```sh
export OPERATOR_VERSION=v6.0.2
export OPERATOR_SCRIPT_URL=https://setup.containers.carbonblack.io/$OPERATOR_VERSION/operator-apply.sh
curl -s $OPERATOR_SCRIPT_URL | bash -s -- -u 
```

* Notice that the above command will delete the Carbon Black Container custom resources definitions and instances.

## Documentation
1. [Setting up Prometheus access](Prometheus.md)
2. [CRD Configuration](crds.md)
3. [Resource spec Configuration](Resources.md)
4. [Using HTTP proxy](Proxy.md)
5. [Configuring image sources](ImageSources.md)
6. [RBAC Configuration](rbac.md)

## Developers Guide
A developers guide for building and configuring the operator:

[Developers Guide](developers.md)

## Helm Charts Documentation
[VMware Carbon Black Cloud Container Helm Charts Documentation](../charts/README.md)


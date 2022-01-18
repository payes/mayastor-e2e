# Mayastor E2E agent test pod

## Introduction
Derived from `ubuntu`
This agent runs as a daemonset in the cluster to help run commands remotely.

## Building
Run `./build.sh`
This builds the image `mayadata/e2e-agent`

# Deploying
The e2e-agent yaml includes a configmap and the e2e-agent daemonset definition.
`E2E_HOST_ADDR` is needed to be set before deploying the yaml.
```
Kubectl apply -f e2e-agent.yaml
```

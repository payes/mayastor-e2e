# Mayastor E2E agent test pod

## Introduction
Derived from `ubuntu`
This agent runs as a daemonset in the cluster to help perform the following e2e related tasks:
1. Node reboot
2. Create rules to drop connections from other k8s nodes
3. Create rules to accept connections from other k8s nodes
4. Create rules to accept connections from the e2e host machine

## Building
Run `./build.sh`
This builds the image `mayadata/e2e-agent`

# Deploying
The e2e-agent yaml includes a configmap and the e2e-agent daemonset definition.
`E2E_HOST_ADDR` is needed to be set before deploying the yaml.
```
Kubectl apply -f e2e-agent.yaml
```

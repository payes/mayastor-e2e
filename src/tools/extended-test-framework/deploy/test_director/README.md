# Sealed secrets

For security reasons we are using sealed secrets approach from [bitnami](https://github.com/bitnami-labs/sealed-secrets) to store encrypted secrets inside GIT. Master YAML file with certificates is stored inside Jenkins credentials.

## Setup for deployment

1. Install a bitnami from their [sites](https://github.com/bitnami-labs/sealed-secrets/releases) and download the ***controller.yaml*** with same version
2. Once the sealed secret is installed start your k8s cluster and run `kubectl apply -f <path-to-downloaded-controller-step-1>`.
3. Now we are able to save key pair for our sealed secrets. We only need to run `kubectl get secret -n kube-system -l sealedsecrets.bitnami.com/sealed-secrets-key -o yaml > master.yaml`
    1. If you are using Jenkins to build a cluster you need to store this file inside Jenkins credentials (SEALED_SECRET_YAML). **!!!Only do this in consultation with others as it will affect all other Jenkins builds!!!**
    2. Otherwise, store it at safe place
4. Prepare standard secret config yaml with base64 encoded values [k8s guide](https://kubernetes.io/docs/concepts/configuration/secret/)

For example:
   ```yaml
   apiVersion: v1
   kind: Secret
   metadata:
   name: test-director-secret
   namespace: mayastor-e2e #create or change NS
   type: Opaque
   data:
     xray_client_id: RVENDYxzOURFMkVY=
     xray_client_secret: MGT3YmIyNmY2OWU4NQ==
     slack_web_hook: aHRy9o29rcy5zbGckpy=
     jira_api_token: WlRjBiM0pBYldGNVlXUmhkR0VTFNGRjRTMW80Yl
     #key: base64-value (all these values should be visible in our password manager or ask your colleagues)
   ```
5. Further to generate encrypted sealed secret yaml file run command `kubeseal --format=yaml < <path-to-k8s-secret-file-step-4> > sealed_secrets.yaml`
6. We have everything to build a cluster with secrets

## Deployment
- `kubectl apply -f <sealed-secret-master-file-from-step-3>`
- `kubectl apply -f <controller-file-step-2>`
- `kubectl apply -f <sealed-secret-file-step-5>`

## Delete

- `kubectl delete secret <secret-name> -n <namespace>`
- `kubectl delete sealedsecret <secret-name> -n <namespace>`
- `kubectl delete secret -n kube-system -l sealedsecrets.bitnami.com/sealed-secrets-key`
- `kubectl delete -f <controller-file-step-2>`

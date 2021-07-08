# K8s to secrets manager

Takes a Kubernetes secret and imports it into AWS Secrets Manager for easy migration to [kubernetes-external-secrets](https://github.com/external-secrets/kubernetes-external-secrets)

## Usage

```console
Usage of k8s-to-secretsmanager:
      --binary                should the resulting secret in secretsmanager be of type binary
  -c, --cluster string        EKS cluster the secret is going to used in
  -d, --description string    (Optional) description to use for new secret in Secrets manager
      --dry-run               Print create secret input instead of running the actual create secret command(s) THIS WILL PRINT SECRETS TO STDOUT BE CAREFUL
  -h, --help                  Print help text
  -k, --kmskey string         KMS key to use for encryption of secret in secretsmanager
  -n, --namespace string      namespace of the secret in kubernetes
  -p, --profile string        AWS profile to use for API calls (default "default")
  -r, --region string         region to use for AWS API calls (default "us-east-1")
  -s, --secret string         name of secret to retrieve from kubernetes and use in new secretsmanager secret
      --tags stringToString   list of key=value pairs to use for tags for secretsmanager secret (separated by comma) (default [uploaded:by=k8s-to-secretsmanager])
  -v, --version               Print version info
```

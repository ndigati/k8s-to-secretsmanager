# K8s to secrets manager

Takes a Kubernetes secret and imports it into AWS Secrets Manager for easy migration to [kubernetes-external-secrets](https://github.com/external-secrets/kubernetes-external-secrets)

## Why?

I mostly created this to play with client-go and the AWS v2 go sdk. It also makes it a little easier to migrate secrets from the built in kubernetes secrets to secretsmanager without having to write a long-ish bash pipeline or storing `-o yaml` files on my laptop.

## Usage

### Prerequistes 

- A working kubeconfig for access to a kubernetes cluster
- AWS credentials (either through a named profile or static keys)
- A KMS key arn to encrypt the uploaded secrets with

---

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

## Building

Build requires a working Go toolchain (Only tested with go 1.16+). There's a [Makefile](./Makefile) to run common build and testing commands.

To get a working binary for your platform/OS you can just run `make` and it will place the binary in the repo's `build/` directory.
module github.com/ndigati/k8s-to-secretsmanager

go 1.16

require (
	github.com/aws/aws-sdk-go-v2 v1.7.0
	github.com/aws/aws-sdk-go-v2/config v1.4.0
	github.com/aws/aws-sdk-go-v2/service/kms v1.4.0
	github.com/aws/aws-sdk-go-v2/service/secretsmanager v1.4.0
	github.com/spf13/pflag v1.0.5
	k8s.io/api v0.15.12
	k8s.io/apimachinery v0.15.12
	k8s.io/client-go v0.15.12
	k8s.io/klog/v2 v2.9.0
)

package main

import (
	"fmt"
	"os"
	"runtime"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	flag "github.com/spf13/pflag"
)

var gitCommit string

// getK8sSecret fetches the secret from kubernetes and returns the full secret object
// the data in the returned object in *not* base64 encoded so you need to encode it
// if you expect base64 data
// https://github.com/kubernetes/client-go/issues/198
func getK8sSecret(name string, namespace string) (*v1.Secret, error) {
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	overrides := &clientcmd.ConfigOverrides{}

	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, overrides)
	config, err := kubeConfig.ClientConfig()
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset.CoreV1().Secrets(namespace).Get(name, metav1.GetOptions{})
}

func main() {
	var help bool
	var version bool
	var dryRun bool

	var binary bool
	var secretName string
	var secretNamespace string
	var clusterName string
	var description string
	var kmsKey string
	var tags map[string]string
	var defaultTags = map[string]string{
		"uploaded:by": "k8s-to-secretsmanager",
	}
	var region string
	var profile string

	flag.BoolVar(&binary, "binary", false, "should the resulting secret in secretsmanager be of type binary")
	flag.BoolVar(&dryRun, "dry-run", false, "Print create secret input instead of running the actual create secret command(s) THIS WILL PRINT SECRETS TO STDOUT BE CAREFUL")
	flag.StringVarP(&secretName, "secret", "s", "", "name of secret to retrieve from kubernetes and use in new secretsmanager secret")
	flag.StringVarP(&secretNamespace, "namespace", "n", "", "namespace of the secret in kubernetes")
	flag.StringVarP(&clusterName, "cluster", "c", "", "EKS cluster the secret is going to used in")
	flag.StringVarP(&description, "description", "d", "", "(Optional) description to use for new secret in Secrets manager")
	flag.StringVarP(&kmsKey, "kmskey", "k", "", "KMS key to use for encryption of secret in secretsmanager")
	flag.StringToStringVar(&tags, "tags", defaultTags, "list of key=value pairs to use for tags for secretsmanager secret (separated by comma)")

	flag.StringVarP(&region, "region", "r", "us-east-1", "region to use for AWS API calls")
	flag.StringVarP(&profile, "profile", "p", "default", "AWS profile to use for API calls")

	flag.BoolVarP(&help, "help", "h", false, "Print help text")
	flag.BoolVarP(&version, "version", "v", false, "Print version info")

	flag.Parse()

	klog.InitFlags(nil)

	if help {
		flag.Usage()
		os.Exit(0)
	}

	if version {
		fmt.Printf(`%s
commit: %s
go version: %s
`, "k8s-to-secretsmanager", gitCommit, runtime.Version())
		os.Exit(0)
	}

	if secretName == "" {
		klog.Error("secret name variable is required")
		flag.Usage()
		os.Exit(1)
	}
	if secretNamespace == "" {
		klog.Error("secret namespace variable is required")
		flag.Usage()
		os.Exit(1)
	}
	if clusterName == "" {
		klog.Error("eks cluster name variable is required")
		flag.Usage()
		os.Exit(1)
	}
	if kmsKey == "" {
		klog.Error("KMS key flag is required")
		flag.Usage()
		os.Exit(1)
	}

	secret, err := getK8sSecret(secretName, secretNamespace)
	if err != nil {
		klog.Fatalf("Failed to get secret %s from kubernetes: %s", secretName, err)
	}

	eksSecretName := generateEKSSecretName(clusterName, secretNamespace, secretName)
	awsClient := awsInit(region, profile, dryRun)

	input, err := awsClient.createSecretInput(eksSecretName, description, kmsKey, binary, secret, tags)
	if err != nil {
		klog.Fatalf("Failed to create secret input: %v", err)
	}

	err = awsClient.createAWSSecret(input, secretName)
	if err != nil {
		klog.Fatalf("Failed to create secretsmanager secret: %v", err)
	}
}

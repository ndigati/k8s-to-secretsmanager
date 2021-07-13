package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
)

// EKSSecretName name of the new secret in secrets manager
// Format is: "eks/<cluster>/<namespace>/<name>"
const EKSSecretName = "eks/%s/%s/%s"

type KmsListKeysAPI interface {
	ListKeys(ctx context.Context, params *kms.ListKeysInput, optFns ...func(*kms.Options)) (*kms.ListKeysOutput, error)
}

type SecretsManagerCreateSecretAPI interface {
	CreateSecret(ctx context.Context, params *secretsmanager.CreateSecretInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.CreateSecretOutput, error)
}

type secretCreator struct {
	region  string
	profile string
	dryRun  bool

	kmsClient KmsListKeysAPI
	smClient  SecretsManagerCreateSecretAPI
}

func generateEKSSecretName(cluster, namespace, name string) string {
	return fmt.Sprintf(EKSSecretName, cluster, namespace, name)
}

func awsInit(region, profile string, dryRun bool) secretCreator {
	cfg, err := awsconfig.LoadDefaultConfig(
		context.TODO(),
		awsconfig.WithRegion(region),
		awsconfig.WithSharedConfigProfile(profile))
	if err != nil {
		klog.Fatalf("unable to load aws SDK config, %v", err)
	}

	creator := secretCreator{}
	creator.region = region
	creator.profile = profile
	creator.dryRun = dryRun

	creator.kmsClient = kms.NewFromConfig(cfg)
	creator.smClient = secretsmanager.NewFromConfig(cfg)
	return creator
}

func (c secretCreator) createSecretInput(name, description, kmskey string, isBinary bool, secret *v1.Secret, tags map[string]string) (*secretsmanager.CreateSecretInput, error) {
	input := &secretsmanager.CreateSecretInput{}
	input.Name = &name
	for k, v := range tags {
		tag := types.Tag{
			Key:   aws.String(k),
			Value: aws.String(v),
		}
		input.Tags = append(input.Tags, tag)
	}
	input.Description = aws.String(description)
	klog.Infof("Using kms key: %s for encryption", kmskey)
	input.KmsKeyId = aws.String(kmskey)

	if len(secret.Data) == 0 {
		return nil, errors.New("no data in requested secret")
	}

	if isBinary {
		if len(secret.Data) > 1 {
			return nil, errors.New("too many binary values in secret")
		}
		for _, value := range secret.Data {
			input.SecretBinary = value
		}
	} else {
		var decodedData map[string]string = make(map[string]string)
		for key, value := range secret.Data {
			decodedData[key] = string(value)
		}
		j, err := json.Marshal(decodedData)
		s := string(j)
		if err != nil {
			return nil, errors.New("failed to marshal json")
		}
		input.SecretString = aws.String(s)
	}
	return input, nil
}

func (c secretCreator) createAWSSecret(input *secretsmanager.CreateSecretInput) error {
	if c.dryRun {
		output, err := json.Marshal(&input)
		if err != nil {
			klog.Exit("Failed to marshal JSON input")
		}
		klog.Info("Would have run CreateSecret with the following input")
		fmt.Println(string(output))
		return nil
	}

	output, err := c.smClient.CreateSecret(context.TODO(), input)
	if err != nil {
		return err
	}

	klog.Info("Successfully created secret!\n")
	outjson, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		klog.Error("Failed to marshal output json")
	}
	fmt.Println(string(outjson))
	return nil
}

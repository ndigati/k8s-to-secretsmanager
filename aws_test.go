package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	v1 "k8s.io/api/core/v1"
)

type MockListKeys struct{}
type MockCreateSecret struct{}

func (m *MockListKeys) ListKeys(ctx context.Context, params *kms.ListKeysInput, optFns ...func(*kms.Options)) (*kms.ListKeysOutput, error) {
	return &kms.ListKeysOutput{
		Keys: []types.KeyListEntry{
			{KeyArn: aws.String("fake-arn"), KeyId: aws.String("fake-id")},
		},
	}, nil
}

func (m *MockCreateSecret) CreateSecret(ctx context.Context, params *secretsmanager.CreateSecretInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.CreateSecretOutput, error) {
	return &secretsmanager.CreateSecretOutput{
		ARN:  aws.String("fake-arn"),
		Name: params.Name,
	}, nil
}

func TestGenerateEKSSecretName(t *testing.T) {
	tables := []struct {
		cluster   string
		namespace string
		name      string
		expected  string
	}{
		{"test", "default", "hello-secret", "eks/test/default/hello-secret"},
		{"prod", "myapp", "postgres-connection", "eks/prod/myapp/postgres-connection"},
	}

	for _, table := range tables {
		result := generateEKSSecretName(table.cluster, table.namespace, table.name)
		if result != table.expected {
			t.Errorf("Resulting secret manager does not match format, got: %s, expected %s", result, table.expected)
		}
	}
}

func TestCreateSecretInput(t *testing.T) {
	secret1json := `{
"apiVersion": "v1",
"kind": "Secret",
"metadata": {
  "name": "my-secret"
},
"data": {
  "key1": "dmFsdWUx"
}
}
`
	secret2json := `{
"apiVersion": "v1",
"kind": "Secret",
"metadata": {
  "name": "my-secret"
},
"data": {
  "key1": "dmFsdWUx",
  "key2": "d29ybGQK"
}
}
`
	var secret1 v1.Secret
	err := json.Unmarshal([]byte(secret1json), &secret1)
	if err != nil {
		t.Error("Failed to decode test1 json string")
	}

	var secret2 v1.Secret
	err = json.Unmarshal([]byte(secret2json), &secret2)
	if err != nil {
		t.Error("Failed to decode test2 json string")
	}

	expected1 := &secretsmanager.CreateSecretInput{}
	expected1.Name = aws.String("eks/test/default/my-secret")
	expected1.Description = aws.String("Test secret default")
	expected1.KmsKeyId = aws.String("fake-arn")
	expected1.SecretString = aws.String(`{"key1":"value1"}`)

	expected2 := &secretsmanager.CreateSecretInput{}
	expected2.Name = aws.String("eks/test/default/my-secret")
	expected2.Description = aws.String("Test secret binary")
	expected2.KmsKeyId = aws.String("fake-arn")
	expected2.SecretBinary = []byte("value1")

	expected3 := &secretsmanager.CreateSecretInput{}

	tables := []struct {
		name          string
		description   string
		kmsKey        string
		isBinary      bool
		secret        *v1.Secret
		tags          map[string]string
		expected      *secretsmanager.CreateSecretInput
		expectedError error
	}{
		{"eks/test/default/my-secret", "Test secret default", "fake-arn", false, &secret1, nil, expected1, nil},
		{"eks/test/default/my-secret", "Test secret binary", "fake-arn", true, &secret1, nil, expected2, nil},
		{"eks/test/default/my-secret", "Test secret too many binary", "fake-arn", true, &secret2, nil, expected3, errors.New("too many binary values in secret")},
	}

	c := secretCreator{
		region:  "us-east-1",
		profile: "doesnotexist",

		kmsClient: &MockListKeys{},
		smClient:  &MockCreateSecret{},
	}

	for _, table := range tables {
		result, err := c.createSecretInput(table.name, table.description, table.kmsKey, table.isBinary, table.secret, table.tags)
		if err == nil && table.expectedError == nil {
			if !compareCreateSecretInputs(table.expected, result) {
				t.Errorf("Resulting secret input does not match:\nexpected: %#v\n \ngot: %#v", table.expected, result)
			}
		}
		// Should probably have a custom error type but this works for now
		// TODO: define custom error type and compare that instead of the error strings
		if err != nil && table.expectedError != nil {
			if err.Error() != table.expectedError.Error() {
				t.Errorf("Test case didn't match expected error, got: %s, expected: %s", err.Error(), table.expectedError.Error())
			}
		}

	}
}

func compareCreateSecretInputs(a, b *secretsmanager.CreateSecretInput) bool {
	if (a.Name != nil) && (b.Name != nil) && (*a.Name != *b.Name) {
		return false
	}
	if (a.Description != nil) && (b.Description != nil) && (*a.Description != *b.Description) {
		return false
	}
	if (a.KmsKeyId != nil) && (b.KmsKeyId != nil) && *a.KmsKeyId != *b.KmsKeyId {
		return false
	}
	if !bytes.Equal(a.SecretBinary, b.SecretBinary) {
		return false
	}
	if (a.SecretString != nil) && (b.SecretString != nil) && (*a.SecretString != *b.SecretString) {
		return false
	}

	if len(a.Tags) != len(b.Tags) {
		return false
	}
	for i, tag := range a.Tags {
		if *tag.Key != *b.Tags[i].Key {
			return false
		}
		if *tag.Value != *b.Tags[i].Value {
			return false
		}
	}

	return true
}

package main

import (
	"reflect"
	"testing"
)

func TestMergeTags(t *testing.T) {
	tables := []struct {
		one      map[string]string
		two      map[string]string
		expected map[string]string
	}{
		// Normal tag pairs
		{
			map[string]string{
				"uploaded:by": "k8s-to-secretsmanager",
			},
			map[string]string{
				"tag1": "value",
			},
			map[string]string{
				"uploaded:by": "k8s-to-secretsmanager",
				"tag1":        "value",
			},
		},
		// One set of tags is empty
		{
			map[string]string{
				"uploaded:by": "k8s-to-secretsmanager",
			},
			map[string]string{},
			map[string]string{
				"uploaded:by": "k8s-to-secretsmanager",
			},
		},
	}

	for _, table := range tables {
		result := mergeTags(table.one, table.two)
		if !reflect.DeepEqual(result, table.expected) {
			t.Errorf("Resulting merged tags are not equal, got: %#v, expected: %#v", result, table.expected)
		}
	}
}

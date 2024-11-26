package generator

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	v1 "kusionstack.io/kusion-api-go/api.kusion.io/v1"
)

func TestNamespaceResource(t *testing.T) {
	tests := []struct {
		name      string
		namespace string
		resWanted *v1.Resource
		errWanted bool
	}{
		{
			name:      "empty namespace",
			namespace: "",
			resWanted: nil,
			errWanted: true,
		},
		{
			name:      "valid namespace",
			namespace: "testNS",
			resWanted: &v1.Resource{
				ID:   "v1:Namespace:testNS",
				Type: "Kubernetes",
				Attributes: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "Namespace",
					"metadata": map[string]interface{}{
						"creationTimestamp": nil,
						"name":              "testNS",
					},
					"spec":   make(map[string]interface{}),
					"status": make(map[string]interface{}),
				},
				DependsOn: nil,
				Extensions: map[string]interface{}{
					"GVK": "/v1, Kind=Namespace",
				},
			},
			errWanted: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NamespaceResource(context.Background(), tt.namespace)
			if (err != nil) != tt.errWanted {
				t.Errorf("NamespaceResource() error = %v, errWanted = %v", err, tt.errWanted)
			}
			require.Equal(t, tt.resWanted, got)
		})
	}
}

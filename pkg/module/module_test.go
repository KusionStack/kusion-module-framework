package module

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
	v1 "kusionstack.io/kusion-api-go/api.kusion.io/v1"
	"kusionstack.io/kusion-module-framework/pkg/module/proto"
)

type mockFrameworkModule struct{}

var wl = []byte(`_type: service.Service
replicas: 1
labels: {}
annotations: {}
dirs: {}
schedule: '* * * * *'`)

var k8sWorkload = []byte(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
spec:
  replicas: 3
  selector:
    matchLabels:
      app: test-app
  template:
    metadata:
      labels:
        app: test-app
    spec:
      containers:
      - name: test-container
        image: test-image
`)

func (m *mockFrameworkModule) Generate(ctx context.Context, req *GeneratorRequest) (*GeneratorResponse, error) {
	var workload map[string]interface{}
	_ = yaml.Unmarshal(k8sWorkload, &workload)
	response := &GeneratorResponse{
		Resources: []v1.Resource{
			{
				ID:         "mock-resource",
				Type:       v1.Kubernetes,
				Attributes: workload,
			},
		},
		Patcher: &v1.Patcher{
			Labels: map[string]string{
				"new-label": "label-value",
			},
			JSONPatchers: map[string]v1.JSONPatcher{
				"mock-resource": {
					Type:    v1.JSONPatch,
					Payload: []byte(`[{"op":"replace","path":"/spec/replicas","value":4}]`),
				},
			},
		},
	}
	return response, nil
}

func TestGenerateWithValidRequest(t *testing.T) {
	ctx := context.Background()
	fmw := &FrameworkModuleWrapper{
		Module: &mockFrameworkModule{},
	}
	req := &proto.GeneratorRequest{
		Project:        "testProject",
		Stack:          "testStack",
		App:            "testApp",
		Workload:       wl,
		DevConfig:      []byte(`{"key":"value"}`),
		PlatformConfig: []byte(`{"key":"value"}`),
		Context:        []byte(`{"key":"value"}`),
		SecretStore:    []byte(`{"key":"value"}`),
	}

	resp, err := fmw.Generate(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestGenerateWithNilWorkload(t *testing.T) {
	ctx := context.Background()
	fmw := &FrameworkModuleWrapper{
		Module: &mockFrameworkModule{},
	}
	req := &proto.GeneratorRequest{
		Project: "testProject",
		Stack:   "testStack",
		App:     "testApp",
	}

	_, err := fmw.Generate(ctx, req)
	assert.NoError(t, err)
}

func TestGenerateWithInvalidWorkload(t *testing.T) {
	ctx := context.Background()
	fmw := &FrameworkModuleWrapper{
		Module: &mockFrameworkModule{},
	}
	req := &proto.GeneratorRequest{
		Project:  "testProject",
		Stack:    "testStack",
		App:      "testApp",
		Workload: k8sWorkload,
	}

	_, err := fmw.Generate(ctx, req)
	assert.NoError(t, err)
}

func TestGenerateWithEmptyRequest(t *testing.T) {
	ctx := context.Background()
	fmw := &FrameworkModuleWrapper{
		Module: &mockFrameworkModule{},
	}
	var req *proto.GeneratorRequest

	_, err := fmw.Generate(ctx, req)
	assert.Error(t, err)
}

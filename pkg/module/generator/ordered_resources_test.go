package generator

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "kusionstack.io/kusion-api-go/api.kusion.io/v1"
)

var (
	fakeDeployment = map[string]interface{}{
		"apiVersion": "apps/v1",
		"kind":       "Deployment",
		"metadata": map[string]interface{}{
			"namespace": "foo",
			"name":      "bar",
		},
		"spec": map[string]interface{}{
			"replica": 1,
			"template": map[string]interface{}{
				"spec": map[string]interface{}{
					"containers": []map[string]interface{}{
						{
							"image": "foo.bar.com:v1",
							"name":  "bar",
						},
					},
				},
			},
		},
	}
	fakeService = map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Service",
		"metadata": map[string]interface{}{
			"namespace": "foo",
			"name":      "bar",
		},
	}
	fakeNamespace = map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Namespace",
		"metadata": map[string]interface{}{
			"name": "foo",
		},
	}
	genOldResources = func() *v1.Resources {
		return &v1.Resources{
			{
				ID:         "apps/v1:Deployment:foo:bar",
				Type:       v1.Kubernetes,
				Attributes: fakeDeployment,
			},
			{
				ID:         "v1:Service:foo:bar",
				Type:       v1.Kubernetes,
				Attributes: fakeService,
			},
			{
				ID:         "v1:Namespace:foo",
				Type:       v1.Kubernetes,
				Attributes: fakeNamespace,
			},
		}
	}
)

func TestDependsKindsGraphResources(t *testing.T) {
	tests := []struct {
		name              string
		resources         v1.Resources
		dependsKindsGraph map[string][]string
		resExpected       v1.Resources
		errExpected       bool
	}{
		{
			name:              "empty resources",
			resources:         v1.Resources{},
			dependsKindsGraph: nil,
			resExpected:       nil,
			errExpected:       true,
		},
		{
			name:              "resources with default graph",
			resources:         *genOldResources(),
			dependsKindsGraph: nil,
			resExpected: v1.Resources{
				{
					ID:         "apps/v1:Deployment:foo:bar",
					Type:       v1.Kubernetes,
					Attributes: fakeDeployment,
					DependsOn: []string{
						"v1:Namespace:foo",
						"v1:Service:foo:bar",
					},
				},
				{
					ID:         "v1:Service:foo:bar",
					Type:       v1.Kubernetes,
					Attributes: fakeService,
					DependsOn: []string{
						"v1:Namespace:foo",
					},
				},
				{
					ID:         "v1:Namespace:foo",
					Type:       v1.Kubernetes,
					Attributes: fakeNamespace,
				},
			},
			errExpected: false,
		},
		{
			name:      "resources with specified graph",
			resources: *genOldResources(),
			dependsKindsGraph: map[string][]string{
				"Service":   {"Deployment"},
				"Namespace": {"Service", "Deployment"},
			},
			resExpected: v1.Resources{
				{
					ID:         "apps/v1:Deployment:foo:bar",
					Type:       v1.Kubernetes,
					Attributes: fakeDeployment,
				},
				{
					ID:         "v1:Service:foo:bar",
					Type:       v1.Kubernetes,
					Attributes: fakeService,
					DependsOn: []string{
						"apps/v1:Deployment:foo:bar",
					},
				},
				{
					ID:         "v1:Namespace:foo",
					Type:       v1.Kubernetes,
					Attributes: fakeNamespace,
					DependsOn: []string{
						"v1:Service:foo:bar",
						"apps/v1:Deployment:foo:bar",
					},
				},
			},
			errExpected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := OrderedResources(context.Background(), tt.resources, tt.dependsKindsGraph)
			if (err != nil) != tt.errExpected {
				t.Errorf("OrderedResources() error = %v, errWanted = %v", err, tt.errExpected)
			}
			require.Equal(t, tt.resExpected, got)
		})
	}
}

func TestResourceKind(t *testing.T) {
	r := &resource{
		Type: v1.Kubernetes,
		Attributes: map[string]interface{}{
			"kind": "Deployment",
		},
	}

	expected := "Deployment"
	actual := r.kubernetesKind()

	assert.Equal(t, expected, actual)
}

func TestInjectAllDependsOn(t *testing.T) {
	resources := genOldResources()
	dependsKindsGraph := map[string][]string{
		"Namespace": {},
	}

	expected := []string{"v1:Namespace:foo"}
	actual := resource((*resources)[0])
	actual.injectDependsOn(dependsKindsGraph, *resources)

	assert.Equal(t, expected, actual.DependsOn)
}

func TestFindDependKinds(t *testing.T) {
	r := &resource{
		Type: v1.Kubernetes,
		Attributes: map[string]interface{}{
			"kind": "Deployment",
		},
	}

	expected := []string{
		"Namespace",
		"ResourceQuota",
		"PersistentVolumeClaim",
		"ServiceAccount",
		"PodSecurityPolicy",
		"ConfigMap",
		"Secret",
		"Service",
		"LimitRange",
	}
	actual := r.findDependKinds(DefaultDependsKindsGraph)

	assert.Equal(t, expected, actual)
}

func TestFindDependResources(t *testing.T) {
	dependKind := "Namespace"
	resources := genOldResources()

	expected := []*v1.Resource{
		{
			ID:         "v1:Namespace:foo",
			Type:       v1.Kubernetes,
			Attributes: fakeNamespace,
		},
	}
	actual := findDependResources(dependKind, *resources)

	assert.Equal(t, expected, actual)
}

func TestDefaultDependsGraphValidation(t *testing.T) {
	for resourceKind, dependResourceKinds := range DefaultDependsKindsGraph {
		for _, dependsResourceKind := range dependResourceKinds {
			if _, exists := DefaultDependsKindsGraph[dependsResourceKind]; !exists {
				t.Errorf("resource type %q depends on a non-existent resource type %q", resourceKind, dependsResourceKind)
			}
		}
	}
}

func TestNoCyclesInDefaultDependencyGraph(t *testing.T) {
	if HasCycleInGraph(DefaultDependsKindsGraph) {
		t.Errorf("Cycle detected in the dependency graph!")
	}
}

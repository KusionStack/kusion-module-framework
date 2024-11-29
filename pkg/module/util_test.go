package module

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "kusionstack.io/kusion-api-go/api.kusion.io/v1"
)

func TestForeachOrdered(t *testing.T) {
	m := map[string]int{
		"b": 2,
		"a": 1,
		"c": 3,
	}

	result := ""
	err := ForeachOrdered(m, func(key string, value int) error {
		result += key
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, "abc", result)
}

func TestGenericPtr(t *testing.T) {
	value := 42
	ptr := GenericPtr(value)
	assert.Equal(t, &value, ptr)
}

func TestMergeMaps(t *testing.T) {
	map1 := map[string]string{
		"a": "1",
		"b": "2",
	}

	map2 := map[string]string{
		"c": "3",
		"d": "4",
	}

	merged := MergeMaps(map1, nil, map2)

	expected := map[string]string{
		"a": "1",
		"b": "2",
		"c": "3",
		"d": "4",
	}

	assert.Equal(t, expected, merged)
}

func TestKubernetesResourceID(t *testing.T) {
	typeMeta := metav1.TypeMeta{
		APIVersion: "apps/v1",
		Kind:       "Deployment",
	}

	objectMeta := metav1.ObjectMeta{
		Namespace: "example",
		Name:      "my-deployment",
	}

	id := KubernetesResourceID(typeMeta, objectMeta)
	assert.Equal(t, "apps/v1:Deployment:example:my-deployment", id)
}

func TestUniqueAppName(t *testing.T) {
	projectName := "my-project"
	stackName := "my-stack"
	appName := "my-app"

	expected := "my-project-my-stack-my-app"
	result := UniqueAppName(projectName, stackName, appName)

	assert.Equal(t, expected, result)
}

func TestUniqueAppLabels(t *testing.T) {
	projectName := "my-project"
	appName := "my-app"

	expected := map[string]string{
		"app.kubernetes.io/part-of": projectName,
		"app.kubernetes.io/name":    appName,
	}

	result := UniqueAppLabels(projectName, appName)

	assert.Equal(t, expected, result)
}

func TestPatchResource(t *testing.T) {
	resources := map[string][]*v1.Resource{
		"/v1, Kind=Namespace": {
			{
				ID:   "v1:Namespace:default",
				Type: "Kubernetes",
				Attributes: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "Namespace",
					"metadata": map[string]interface{}{
						"name": "default",
					},
				},
				Extensions: map[string]interface{}{
					"GVK": "/v1, Kind=Namespace",
				},
			},
		},
	}
	assert.NoError(
		t,
		PatchResource(resources, "/v1, Kind=Namespace", func(ns *corev1.Namespace) error {
			ns.Labels = map[string]string{
				"foo": "bar",
			}
			return nil
		}),
	)
	assert.Equal(
		t,
		map[string]interface{}{
			"foo": "bar",
		},
		resources["/v1, Kind=Namespace"][0].Attributes["metadata"].(map[string]interface{})["labels"].(map[string]interface{}),
	)
}

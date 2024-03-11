package module

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	v1 "kusionstack.io/kusion/pkg/apis/core/v1"
)

func WrapK8sResourceToKusionResource(id string, resource any) (*v1.Resource, error) {
	gvk := resource.(runtime.Object).GetObjectKind().GroupVersionKind().String()

	// fixme: this function converts int to int64 by default
	unstructured, err := runtime.DefaultUnstructuredConverter.ToUnstructured(resource)
	if err != nil {
		return nil, err
	}
	return &v1.Resource{
		ID:         id,
		Type:       v1.Kubernetes,
		Attributes: unstructured,
		DependsOn:  nil,
		Extensions: map[string]any{
			v1.ResourceExtensionGVK: gvk,
		},
	}, nil
}

// KubernetesResourceID returns the unique ID of a Kubernetes resource
// based on its type and metadata.
func KubernetesResourceID(typeMeta metav1.TypeMeta, objectMeta metav1.ObjectMeta) string {
	// resource id example: apps/v1:Deployment:nginx:nginx-deployment
	id := typeMeta.APIVersion + ":" + typeMeta.Kind + ":"
	if objectMeta.Namespace != "" {
		id += objectMeta.Namespace + ":"
	}
	id += objectMeta.Name
	return id
}

// UniqueAppName returns a unique name for a workload based on its project and app name.
func UniqueAppName(projectName, stackName, appName string) string {
	return projectName + "-" + stackName + "-" + appName
}

// UniqueAppLabels returns a map of labels that identify an app based on its project and name.
func UniqueAppLabels(projectName, appName string) map[string]string {
	return map[string]string{
		"app.kubernetes.io/part-of": projectName,
		"app.kubernetes.io/name":    appName,
	}
}

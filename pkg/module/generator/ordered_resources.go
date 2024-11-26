package generator

import (
	"context"
	"errors"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	v1 "kusionstack.io/kusion-api-go/api.kusion.io/v1"
)

type resource v1.Resource

// defaultOrderedKinds provides the default order of Kubernetes resource kinds.
var defaultOrderedKinds = []string{
	"Namespace",
	"ResourceQuota",
	"StorageClass",
	"CustomResourceDefinition",
	"ServiceAccount",
	"PodSecurityPolicy",
	"Role",
	"ClusterRole",
	"RoleBinding",
	"ClusterRoleBinding",
	"ConfigMap",
	"Secret",
	"Endpoints",
	"Service",
	"LimitRange",
	"PriorityClass",
	"PersistentVolume",
	"PersistentVolumeClaim",
	"Deployment",
	"StatefulSet",
	"CronJob",
	"PodDisruptionBudget",
	"MutatingWebhookConfiguration",
	"ValidatingWebhookConfiguration",
}

// OrderedResources returns a list of Kusion Resources with the injected `dependsOn`
// in a specified order.
func OrderedResources(ctx context.Context, resources v1.Resources, orderedKinds []string) (v1.Resources, error) {
	if len(orderedKinds) == 0 {
		orderedKinds = defaultOrderedKinds
	}

	if len(resources) == 0 {
		return nil, errors.New("empty resources")
	}

	for i := 0; i < len(resources); i++ {
		// Continue if the resource is not a Kubernetes resource.
		if resources[i].Type != v1.Kubernetes {
			continue
		}

		// Inject dependsOn of the resource.
		r := (*resource)(&resources[i])
		r.injectDependsOn(orderedKinds, resources)
		resources[i] = v1.Resource(*r)
	}

	return resources, nil
}

// kubernetesKind returns the kubernetes kind of the given resource.
func (r resource) kubernetesKind() string {
	u := &unstructured.Unstructured{}
	u.SetUnstructuredContent(r.Attributes)
	return u.GetKind()
}

// injectDependsOn injects all dependsOn relationships for the given resource and dependent kinds.
func (r *resource) injectDependsOn(orderedKinds []string, rs []v1.Resource) {
	kinds := r.findDependKinds(orderedKinds)
	for _, kind := range kinds {
		drs := findDependResources(kind, rs)
		r.appendDependsOn(drs)
	}
}

// appendDependsOn injects dependsOn relationships for the given resource and dependent resources.
func (r *resource) appendDependsOn(dependResources []*v1.Resource) {
	for _, dr := range dependResources {
		r.DependsOn = append(r.DependsOn, dr.ID)
	}
}

// findDependKinds returns the dependent resource kinds for the specified kind.
func (r *resource) findDependKinds(orderedKinds []string) []string {
	curKind := r.kubernetesKind()
	dependKinds := make([]string, 0)
	for _, previousKind := range orderedKinds {
		if curKind == previousKind {
			break
		}
		dependKinds = append(dependKinds, previousKind)
	}
	return dependKinds
}

// findDependResources returns the dependent resources of the specified kind.
func findDependResources(dependKind string, rs []v1.Resource) []*v1.Resource {
	var dependResources []*v1.Resource
	for i := 0; i < len(rs); i++ {
		if resource(rs[i]).kubernetesKind() == dependKind {
			dependResources = append(dependResources, &rs[i])
		}
	}
	return dependResources
}

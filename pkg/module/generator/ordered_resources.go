package generator

import (
	"context"
	"errors"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	v1 "kusionstack.io/kusion-api-go/api.kusion.io/v1"
)

type resource v1.Resource

// DefaultDependsKindsGraph defines the default dependency relationships between
// Kubernetes resource kinds. This graph maps each resource kind to the list of
// resource kinds it potentially depends on (not strictly required, but commonly
// associated in practice).
//
// Structure:
//   - Key: The resource kind (e.g., "Deployment")
//   - Value: Slice of resource kinds this resource may depend on
//
// Example:
//
//	"Deployment": {"Namespace", "ServiceAccount", ...}
var DefaultDependsKindsGraph = map[string][]string{
	"Namespace":                      {},
	"ResourceQuota":                  {"Namespace"},
	"StorageClass":                   {},
	"CustomResourceDefinition":       {},
	"ServiceAccount":                 {"Namespace"},
	"PodSecurityPolicy":              {},
	"Role":                           {"Namespace"},
	"ClusterRole":                    {},
	"RoleBinding":                    {"Namespace", "ServiceAccount", "Role"},
	"ClusterRoleBinding":             {"ServiceAccount", "ClusterRole"},
	"ConfigMap":                      {"Namespace"},
	"Secret":                         {"Namespace"},
	"Endpoints":                      {"Namespace"},
	"Service":                        {"Namespace", "Endpoints"},
	"LimitRange":                     {"Namespace", "StorageClass"},
	"PriorityClass":                  {},
	"PersistentVolume":               {"StorageClass"},
	"PersistentVolumeClaim":          {"Namespace", "ResourceQuota", "StorageClass", "PersistentVolume"},
	"Deployment":                     {"Namespace", "ResourceQuota", "PersistentVolumeClaim", "ServiceAccount", "PodSecurityPolicy", "ConfigMap", "Secret", "Service", "LimitRange"},
	"StatefulSet":                    {"Namespace", "ResourceQuota", "PersistentVolumeClaim", "ServiceAccount", "PodSecurityPolicy", "ConfigMap", "Secret", "Service", "LimitRange"},
	"CronJob":                        {"Namespace", "ResourceQuota", "PersistentVolumeClaim", "ServiceAccount", "PodSecurityPolicy", "ConfigMap", "Secret", "Service", "LimitRange"},
	"PodDisruptionBudget":            {"Namespace", "Deployment", "StatefulSet", "CronJob"},
	"MutatingWebhookConfiguration":   {"Namespace", "ServiceAccount", "RoleBinding", "ClusterRoleBinding", "ConfigMap", "Secret", "Service"},
	"ValidatingWebhookConfiguration": {"Namespace", "ServiceAccount", "RoleBinding", "ClusterRoleBinding", "ConfigMap", "Secret", "Service"},
}

// OrderedResources returns a list of Kusion Resources with the injected `dependsOn`
// in a specified order.
func OrderedResources(ctx context.Context, resources v1.Resources, dependsKindsGraph map[string][]string) (v1.Resources, error) {
	if dependsKindsGraph == nil {
		dependsKindsGraph = DefaultDependsKindsGraph
	}
	if HasCycleInGraph(dependsKindsGraph) {
		return nil, errors.New("find cycles in giving depends kinds grach")
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
		r.injectDependsOn(dependsKindsGraph, resources)
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
func (r *resource) injectDependsOn(dependsKindsGraph map[string][]string, rs []v1.Resource) {
	kinds := r.findDependKinds(dependsKindsGraph)
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
func (r *resource) findDependKinds(dependsKindsGraph map[string][]string) []string {
	curKind := r.kubernetesKind()
	if _, exists := dependsKindsGraph[curKind]; !exists {
		depends := []string{}
		for resourceKinds, resourceKindsDepends := range dependsKindsGraph {
			depends = append(depends, resourceKinds)
			for _, resourceKindDepend := range resourceKindsDepends {
				// if this curKind is depends by other kinds, and not in this graph,
				// return empty depends.
				if curKind == resourceKindDepend {
					return []string{}
				}
			}
		}
		// if this curKind is not depends by any other kinds, and not in this graph,
		// curkind will depends on all kinds in this graph.
		return depends
	}
	return dependsKindsGraph[curKind]
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

// HasCycleInGraph checks if there's a cycle in the dependency graph.
// Returns true if a cycle is detected, false otherwise.
func HasCycleInGraph(graph map[string][]string) bool {
	// Track visited nodes and recursion stack for cycle detection
	visited := make(map[string]bool)
	recursionStack := make(map[string]bool)

	// Check each node in the graph
	for node := range graph {
		if !visited[node] {
			if hasCycle(node, visited, recursionStack, graph) {
				return true // Cycle detected
			}
		}
	}

	return false // No cycle found
}

// hasCycle performs DFS to detect cycles recursively
func hasCycle(node string, visited, recursionStack map[string]bool, graph map[string][]string) bool {
	if recursionStack[node] {
		return true // Cycle detected
	}

	if visited[node] {
		return false // Already visited and no cycle found
	}

	// Mark as visited and add to recursion stack
	visited[node] = true
	recursionStack[node] = true

	// Recursively check dependencies
	for _, dep := range graph[node] {
		if hasCycle(dep, visited, recursionStack, graph) {
			return true
		}
	}

	// Remove from recursion stack after processing
	recursionStack[node] = false
	return false
}

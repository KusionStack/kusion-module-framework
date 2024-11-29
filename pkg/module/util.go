package module

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	v1 "kusionstack.io/kusion-api-go/api.kusion.io/v1"
	"kusionstack.io/kusion-module-framework/pkg/log"
)

const (
	ImportIDKey = "kusionstack.io/import-id"
)

var ErrEmptyTFProviderVersion = errors.New("empty terraform provider version")

var defaultTFHost = "registry.terraform.io"

func WrapK8sResourceToKusionResource(id string, resource runtime.Object) (*v1.Resource, error) {
	gvk := resource.GetObjectKind().GroupVersionKind().String()

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

// KubernetesResourceID returns the ID of a Kubernetes resource based on its type and metadata.
// Resource ID usually should be unique in one resource list.
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

// WrapTFResourceToKusionResource wraps the Terraform resource into the format of the Kusion resource.
func WrapTFResourceToKusionResource(
	providerCfg ProviderConfig,
	resType string,
	resourceID string,
	attributes map[string]interface{},
	dependsOn []string,
) (*v1.Resource, error) {
	extensions, err := TerraformProviderExtensions(providerCfg, resType)
	if err != nil {
		return nil, err
	}

	return &v1.Resource{
		ID:         resourceID,
		Type:       v1.Terraform,
		Attributes: attributes,
		DependsOn:  dependsOn,
		Extensions: extensions,
	}, nil
}

// ProviderConfig contains the full configurations of a specified provider. It is the combination
// of the specified provider's config in blocks "terraform.required_providers" and "providers" in
// the terraform hcl file, where the former is described by fields Source and Version, and the latter
// is described by ProviderMeta.
type ProviderConfig struct {
	// Source of the provider.
	Source string `yaml:"source" json:"source"`
	// Version of the provider.
	Version string `yaml:"version" json:"version"`
	// ProviderMeta is used to describe configs in the terraform hcl "provider" block.
	ProviderMeta v1.GenericConfig `yaml:"providerMeta" json:"providerMeta"`
}

// TerraformResourceID returns the Kusion resource ID of the Terraform resource.
// Resource ID usually should be unique in one resource list.
func TerraformResourceID(providerCfg ProviderConfig, resType, resName string) (string, error) {
	if providerCfg.Version == "" {
		return "", ErrEmptyTFProviderVersion
	}

	var providerNamespace, providerName string
	srcAttrs := strings.Split(providerCfg.Source, "/")
	if len(srcAttrs) == 3 {
		providerNamespace = srcAttrs[1]
		providerName = srcAttrs[2]
	} else if len(srcAttrs) == 2 {
		providerNamespace = srcAttrs[0]
		providerName = srcAttrs[1]
	} else {
		return "", fmt.Errorf("invalid terraform provider source: %s", providerCfg.Source)
	}

	return strings.Join([]string{
		providerNamespace,
		providerName,
		resType,
		resName,
	}, ":"), nil
}

// TerraformProviderExtensions returns the Kusion resource extension of the Terraform provider.
func TerraformProviderExtensions(providerCfg ProviderConfig, resType string) (map[string]any, error) {
	if providerCfg.Version == "" {
		return nil, ErrEmptyTFProviderVersion
	}
	if providerCfg.Source == "" {
		return nil, fmt.Errorf("empty terraform provider source")
	}
	if resType == "" {
		return nil, fmt.Errorf("empty resource type")
	}

	// Conduct whether to use the default Terraform provider registry host
	// according to the source of the provider config.
	// For example, "hashicorp/aws" means using the default TF provider registry,
	// while "registry.customized.io/hashicorp/aws" implies to use a customized registry host.
	var providerURL string
	srcAttrs := strings.Split(providerCfg.Source, "/")
	if len(srcAttrs) == 3 {
		providerURL = strings.Join([]string{providerCfg.Source, providerCfg.Version}, "/")
	} else if len(srcAttrs) == 2 {
		providerURL = strings.Join([]string{defaultTFHost, providerCfg.Source, providerCfg.Version}, "/")
	} else {
		return nil, fmt.Errorf("invalid terraform provider source: %s", providerCfg.Source)
	}

	return map[string]any{
		"provider":     providerURL,
		"providerMeta": providerCfg.ProviderMeta,
		"resourceType": resType,
	}, nil
}

// TerraformProviderRegion returns the resource region from the Terraform provider configs.
func TerraformProviderRegion(providerCfg ProviderConfig) string {
	region, ok := providerCfg.ProviderMeta["region"]
	if !ok {
		return ""
	}

	return region.(string)
}

// PatchHealthPolicyToExtension patch the health policy to the `extensions` field of the Kusion resource.
// Support Kubernetes resource only.
func PatchHealthPolicyToExtension(resource *v1.Resource, healthPolicy string) error {
	if resource == nil {
		return fmt.Errorf("resource is nil")
	}

	healthPolicyMap := make(map[string]any)

	if resource.Extensions == nil {
		resource.Extensions = make(map[string]interface{})
	}

	if resource.Type == v1.Kubernetes {
		healthPolicyMap[v1.FieldKCLHealthCheckKCL] = healthPolicy
		resource.Extensions[v1.FieldHealthPolicy] = healthPolicyMap
		return nil
	}

	log.Warnf("patch health policy to extension skipped for resource %s, resource type %s is not supported", resource.ID, resource.Type)

	return nil
}

// PatchImportResourcesToExtension patch the imported resource to the `extensions` field of the Kusion resource.
// Support TF resource only.
func PatchImportResourcesToExtension(resource *v1.Resource, importedResource string) error {
	if resource == nil {
		return fmt.Errorf("resource is nil")
	}

	if resource.Extensions == nil {
		resource.Extensions = make(map[string]interface{})
	}

	if resource.Type == v1.Terraform {
		resource.Extensions[ImportIDKey] = importedResource
		// remove the resource attribute to avoid update conflict when using terraform import
		resource.Attributes = make(map[string]interface{})
		return nil
	}

	log.Warnf("patch import resource to extension skipped for resource %s, resource type %s is not supported", resource.ID, resource.Type)

	return nil
}

// PatchKubeConfigPathToExtension patch the kubeConfig path to the `extensions` field of the Kusion resource.
// 1. If $KUBECONFIG environment variable is set, then it is used.
// 2. If not, and the `kubeConfig` in resource extensions is set, then it is used.
// 3. Otherwise, ${HOME}/.kube/config is used.
func PatchKubeConfigPathToExtension(resource *v1.Resource, kubeConfigPath string) error {
	if resource == nil {
		return fmt.Errorf("resource is nil")
	}

	if resource.Extensions == nil {
		resource.Extensions = make(map[string]interface{})
	}
	resource.Extensions[v1.ResourceExtensionKubeConfig] = kubeConfigPath

	return nil
}

// IgnoreModules todo@dayuan delete this condition after workload is changed into a module
var IgnoreModules = map[string]bool{
	"service": true,
	"job":     true,
}

// ForeachOrdered executes the given function on each
// item in the map in order of their keys.
func ForeachOrdered[T any](m map[string]T, f func(key string, value T) error) error {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, k := range keys {
		v := m[k]
		if err := f(k, v); err != nil {
			return err
		}
	}

	return nil
}

// GenericPtr returns a pointer to the provided value.
func GenericPtr[T any](i T) *T {
	return &i
}

// MergeMaps merges multiple map[string]string into one
// map[string]string.
// If a map is nil, it skips it and moves on to the next one. For each
// non-nil map, it iterates over its key-value pairs and adds them to
// the merged map. Finally, it returns the merged map.
func MergeMaps(maps ...map[string]string) map[string]string {
	merged := make(map[string]string)

	for _, m := range maps {
		if len(m) == 0 {
			continue
		}
		for k, v := range m {
			merged[k] = v
		}
	}

	if len(merged) == 0 {
		return nil
	}
	return merged
}

// KusionPathDependency returns the implicit resource dependency path based on
// the resource id and name with the "$kusion_path" prefix.
func KusionPathDependency(id, name string) string {
	return "$kusion_path." + id + "." + name
}

// PatchResource patches the resource with the given patch.
func PatchResource[T any](resources map[string][]*v1.Resource, gvk string, patchFunc func(*T) error) error {
	var obj T
	for _, r := range resources[gvk] {
		// convert unstructured to typed object
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(r.Attributes, &obj); err != nil {
			return err
		}

		if err := patchFunc(&obj); err != nil {
			return err
		}

		// convert typed object to unstructured
		updated, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&obj)
		if err != nil {
			return err
		}
		r.Attributes = updated
	}
	return nil
}

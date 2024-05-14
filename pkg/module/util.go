package module

import (
	"errors"
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
)

var ErrEmptyTFProviderVersion = errors.New("empty terraform provider version")

var defaultTFHost = "registry.terraform.io"

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

// WrapTFResourceToKusionResource wraps the Terraform resource into the format of
// the Kusion resource.
func WrapTFResourceToKusionResource(
	id string,
	attributes, extensions map[string]any,
	denpendsOn []string,
) (*v1.Resource, error) {
	return &v1.Resource{
		ID:         id,
		Type:       v1.Terraform,
		Attributes: attributes,
		DependsOn:  denpendsOn,
		Extensions: extensions,
	}, nil
}

// ProviderConfig contains the full configurations of a specified provider. It is the combination
// of the specified provider's config in blocks "terraform/required_providers" and "providers" in
// terraform hcl file, where the former is described by fields Source and Version, and the latter
// is described by GenericConfig cause different provider has different config.
type ProviderConfig struct {
	// Source of the provider.
	Source string `yaml:"source" json:"source"`

	// Version of the provider.
	Version string `yaml:"version" json:"version"`

	// GenericConfig is used to describe the config of a specified terraform provider.
	v1.GenericConfig `yaml:",inline,omitempty" json:",inline,omitempty"`
}

// TerraformResourceID returns the Kusion resource ID of the Terraform resource.
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
func TerraformProviderExtensions(
	providerCfg ProviderConfig,
	providerMeta map[string]any, resType string,
) (map[string]any, error) {
	if providerCfg.Version == "" {
		return nil, ErrEmptyTFProviderVersion
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
		"providerMeta": providerMeta,
		"resourceType": resType,
	}, nil
}

// TerraformProviderRegion returns the resource region from the Terraform provider configs.
func TerraformProviderRegion(providerCfg ProviderConfig) string {
	region, ok := providerCfg.GenericConfig["region"]
	if !ok {
		return ""
	}

	return region.(string)
}

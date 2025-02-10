// Copyright 2024 KusionStack Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package terraform

import (
	"errors"
	"fmt"
	"strings"

	svchost "github.com/hashicorp/terraform-svchost"
	v1 "kusionstack.io/kusion-api-go/api.kusion.io/v1"

	"kusionstack.io/kusion-module-framework/pkg/resources"
)

// DefaultProviderRegistryHost is the hostname used for provider addresses that do
// not have an explicit hostname.
const DefaultProviderRegistryHost = "registry.terraform.io"

var (
	// errInvalidSource means provider's source is invalid, which must be in [<HOSTNAME>/]<NAMESPACE>/<TYPE> format
	errInvalidSource = errors.New(`invalid provider source string, must be in the format "[hostname/]namespace/name"`)
	// errInvalidVersion means provider's version constraint is invalid, which must not be empty.
	errInvalidVersion = errors.New("invalid provider version constraint")
	// errInvalidResourceTypeOrName resourceType or resourceName is invalid, which must not be empty.
	errInvalidResourceTypeOrName = errors.New("resourceType or resourceName is empty")
)

// NewProvider constructs a provider instance from given source, version and configuration arguments.
func NewProvider(providerConfigs map[string]any, source, version string) (Provider, error) {
	var ret Provider

	if version == "" {
		return ret, errInvalidVersion
	}

	_, err := parseProviderSourceString(source)
	if err != nil {
		return ret, err
	}

	ret.Source = source
	ret.Version = version
	ret.ProviderConfigs = providerConfigs
	return ret, nil
}

// ToKusionResourceID takes provider, resource info and returns string representing Kusion qualified resource ID.
func ToKusionResourceID(p Provider, resourceType, resourceName string) (string, error) {
	if p.Version == "" {
		return "", errInvalidVersion
	}
	if resourceType == "" || resourceName == "" {
		return "", errInvalidResourceTypeOrName
	}
	tfProvider, err := parseProviderSourceString(p.Source)
	if err != nil {
		return "", err
	}

	tfProviderIDStr := tfProvider.IDString()
	return strings.Join([]string{tfProviderIDStr, resourceType, resourceName}, resources.SegmentSeparator), nil
}

// NewKusionResource creates a Kusion Resource object with the given resourceType, resourceID, attributes.
func NewKusionResource(p Provider, resourceType, resourceID string,
	attrs map[string]interface{}, dependsOn []string,
) (*v1.Resource, error) {
	if resourceType == "" {
		return nil, errInvalidResourceTypeOrName
	}

	// put provider info into extensions
	extensions := make(map[string]interface{}, 3)
	extensions["provider"] = p.String()
	extensions["providerMeta"] = p.ProviderConfigs
	extensions["resourceType"] = resourceType

	return &v1.Resource{
		ID:         resourceID,
		Type:       v1.Terraform,
		Attributes: attrs,
		DependsOn:  dependsOn,
		Extensions: extensions,
	}, nil
}

// parseProviderSourceString parses the source attribute and returns a terraform provider.
//
// The following are valid source string formats:
//
//	name
//	namespace/name
//	hostname/namespace/name
func parseProviderSourceString(str string) (TFProvider, error) {
	var ret TFProvider

	// split the source string into individual components
	parts := strings.Split(str, "/")
	if len(parts) == 0 || len(parts) > 3 {
		return ret, errInvalidSource
	}

	// check for an invalid empty string in any part
	for i := range parts {
		if parts[i] == "" {
			return ret, errInvalidSource
		}
	}

	// check the 'name' portion, which is always the last part
	givenName := parts[len(parts)-1]
	ret.Type = givenName
	ret.Hostname = DefaultProviderRegistryHost

	if len(parts) == 1 {
		ret.Namespace = "hashicorp"
		return ret, nil
	}

	if len(parts) >= 2 {
		// the namespace is always the second-to-last part
		givenNamespace := parts[len(parts)-2]
		ret.Namespace = givenNamespace
	}

	// Final Case: 3 parts
	if len(parts) == 3 {
		// the namespace is always the first part in a three-part source string
		hn, err := svchost.ForComparison(parts[0])
		if err != nil {
			return ret, fmt.Errorf("invalid provider source hostname %q in source %q: %s", hn, str, err)
		}
		ret.Hostname = hn
	}

	return ret, nil
}

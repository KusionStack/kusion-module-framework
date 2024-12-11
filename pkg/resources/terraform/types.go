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
	svchost "github.com/hashicorp/terraform-svchost"

	"kusionstack.io/kusion-module-framework/pkg/resources"
)

// Provider contains all the information of a specified Kusion provider, which not only includes the
// source address, version constraint of required provider, but also various configuration arguments used
// to configure required provider before Terraform can use them.
type Provider struct {
	// Source address of the provider.
	Source string `yaml:"source" json:"source"`
	// Version constraint of the provider.
	Version string `yaml:"version" json:"version"`
	// Configuration arguments of the provider.
	ProviderConfigs map[string]any `yaml:"providerConfigs" json:"providerConfigs"`
}

// String returns a qualified string, intended for use in resource extension.
func (p Provider) String() string {
	return p.Source + "/" + p.Version
}

// TFProvider encapsulates a single terraform provider type.
type TFProvider struct {
	Type      string
	Namespace string
	Hostname  svchost.Hostname
}

// String returns an FQN string, intended for use in machine-readable output.
func (tp TFProvider) String() string {
	return tp.Hostname.ForDisplay() + resources.SegmentSeparator + tp.Namespace + resources.SegmentSeparator + tp.Type
}

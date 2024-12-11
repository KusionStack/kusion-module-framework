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

package kubernetes

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	v1 "kusionstack.io/kusion-api-go/api.kusion.io/v1"

	"kusionstack.io/kusion-module-framework/pkg/resources"
)

// ToKusionResourceID returns the Kusion resource ID for the given Kubernetes object specified by
// its GroupVersionKind and ObjectMeta.
func ToKusionResourceID(gvk schema.GroupVersionKind, objectMeta metav1.ObjectMeta) string {
	// resource id example: apps/v1:Deployment:nginx:nginx-deployment
	if gvk.Group == "" {
		gvk.Group = "core"
	}

	id := gvk.Group + resources.SegmentSeparator + gvk.Version + resources.SegmentSeparator + gvk.Kind
	if objectMeta.Namespace != "" {
		id += resources.SegmentSeparator + objectMeta.Namespace
	}
	id += resources.SegmentSeparator + objectMeta.Name
	return id
}

// NewKusionResource creates a Kusion Resource object with the given obj and objectMeta.
func NewKusionResource(obj runtime.Object, objectMeta metav1.ObjectMeta) (*v1.Resource, error) {
	// TODO: this function converts int to int64 by default
	unstructured, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return nil, err
	}

	gvk := obj.GetObjectKind().GroupVersionKind()
	return &v1.Resource{
		ID:         ToKusionResourceID(gvk, objectMeta),
		Type:       v1.Kubernetes,
		Attributes: unstructured,
		DependsOn:  nil,
		Extensions: map[string]any{
			v1.ResourceExtensionGVK: gvk.String(),
		},
	}, nil
}

package generator

import (
	"context"
	"errors"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "kusionstack.io/kusion-api-go/api.kusion.io/v1"
	"kusionstack.io/kusion-module-framework/pkg/module"
)

// NamespaceResource returns a Kubernetes Namespace resource wrapped into the form
// of Kusion Resource.
func NamespaceResource(ctx context.Context, namespace string) (*v1.Resource, error) {
	if namespace == "" {
		return nil, errors.New("empty namespace")
	}

	ns := &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Namespace",
			APIVersion: corev1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}

	id := module.KubernetesResourceID(ns.TypeMeta, ns.ObjectMeta)

	return module.WrapK8sResourceToKusionResource(id, ns)
}

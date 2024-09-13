package image

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	imagev1alpha1 "github.com/rkgcloud/image-sync-controller/api/image/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/ptr"
	"reconciler.io/runtime/reconcilers"
)

func SinglePodReconcile() reconcilers.SubReconciler[*imagev1alpha1.PodSync] {
	return &reconcilers.ChildReconciler[*imagev1alpha1.PodSync, *corev1.Pod, *corev1.PodList]{
		Name: "SinglePodReconcile",
		DesiredChild: func(ctx context.Context, resource *imagev1alpha1.PodSync) (*corev1.Pod, error) {
			return &corev1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Pod",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%s", resource.Name, "-sample-pod"),
					Namespace: resource.Namespace,
					Labels: map[string]string{
						"image-pod": resource.Name,
						"app":       "nginx"},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Image: "nginx:latest",
						},
					},
					ActiveDeadlineSeconds: ptr.Int64(100),
				},
			}, nil
		},
		MergeBeforeUpdate: func(current, desired *corev1.Pod) {
			desired.Labels = current.Labels
			desired.Annotations = current.Annotations
		},
		ReflectChildStatusOnParent: func(ctx context.Context, parent *imagev1alpha1.ImageSync, child *corev1.Pod, err error) {
			log := logr.FromContextOrDiscard(ctx)
			if err != nil {
				log.Error(err, "failed to merge child pod status")
			}
		},
	}
}

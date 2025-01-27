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
	return &reconcilers.Advice[*imagev1alpha1.PodSync]{
		Reconciler: &reconcilers.ChildReconciler[*imagev1alpha1.PodSync, *corev1.Pod, *corev1.PodList]{
			Name: "SinglePodReconcile",
			DesiredChild: func(ctx context.Context, resource *imagev1alpha1.PodSync) (*corev1.Pod, error) {
				podToDeploy := &corev1.Pod{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Pod",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      fmt.Sprintf("%s-%s", resource.Name, "-sample-pod"),
						Namespace: resource.Namespace,
						Labels: map[string]string{
							"image-pod": resource.Name,
							"app":       resource.Name},
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Image: "nginx:latest",
								Name:  "nginx",
							},
						},
						ActiveDeadlineSeconds: ptr.Int64(100),
					},
				}

				return podToDeploy, nil
			},
			MergeBeforeUpdate: func(current, desired *corev1.Pod) {
				desired.Labels = current.Labels
				desired.Annotations = current.Annotations
			},
			ReflectChildStatusOnParent: func(ctx context.Context, parent *imagev1alpha1.PodSync, child *corev1.Pod, err error) {
				log := logr.FromContextOrDiscard(ctx)
				if err != nil {
					log.Error(err, "failed to merge child pod status")
				}

				if child == nil {
					return
				}

				parent.Status.PodName = child.Name
			},
		},
		After: func(ctx context.Context, resource *imagev1alpha1.PodSync, result reconcilers.Result, err error) (reconcilers.Result, error) {
			if err != nil {
				resource.Status.PodName = err.Error()
				// set failed status here
				return result, err
			}
			return result, err
		},
	}
}

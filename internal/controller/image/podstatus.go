package image

import (
	"context"
	"fmt"
	"strings"

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
				var containers []corev1.Container
				if resource.Spec.Containers != nil && len(resource.Spec.Containers) > 0 {
					for _, container := range resource.Spec.Containers {
						containers = append(containers, corev1.Container{
							Image:           container.ImageURL,
							Name:            fmt.Sprintf("%s-container", strings.ToLower(resource.Name)),
							Args:            container.Args,
							ImagePullPolicy: corev1.PullIfNotPresent,
						})
					}
				}
				if containers == nil || len(containers) == 0 {
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
								"app":       resource.Name},
						},
						Spec: corev1.PodSpec{
							Containers:            containers,
							ActiveDeadlineSeconds: ptr.Int64(100),
						},
					}, nil
				}

				return nil, nil
			},
			ChildObjectManager: &reconcilers.UpdatingObjectManager[*corev1.Pod]{
				MergeBeforeUpdate: func(current, desired *corev1.Pod) {
					desired.Labels = current.Labels
					desired.Annotations = current.Annotations
				},
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
				parent.Status.MarkPodSyncConditionPodCreated(ctx)

				if child.Status.Phase == corev1.PodRunning {
					parent.Status.MarkPodSyncConditionPodRunning(ctx)
				}

				if child.Status.Phase == corev1.PodSucceeded {
					parent.Status.MarkPodSyncConditionPodCompleted(ctx)
				}
			},
		},
		After: func(ctx context.Context, resource *imagev1alpha1.PodSync, result reconcilers.Result, err error) (reconcilers.Result, error) {
			if err != nil {
				resource.Status.MarkPodSyncConditionPodCreateFailed(ctx, err.Error())
				return result, err
			}
			return result, err
		},
	}
}

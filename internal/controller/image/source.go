package image

import (
	"context"
	"errors"

	imagev1alpha1 "github.com/rkgcloud/image-sync-controller/api/image/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"reconciler.io/runtime/reconcilers"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
)

func SourceReconcile() reconcilers.SubReconciler[*imagev1alpha1.ImageSync] {
	return &reconcilers.SyncReconciler[*imagev1alpha1.ImageSync]{
		Name: "SourceReconcile",
		Setup: func(ctx context.Context, mgr ctrl.Manager, bldr *builder.Builder) error {
			bldr.Watches(&corev1.Secret{}, reconcilers.EnqueueTracked(ctx))
			return nil
		},
		SyncWithResult: func(ctx context.Context, resource *imagev1alpha1.ImageSync) (reconcilers.Result, error) {
			if resource.Spec.SourceImage.Image == "" {
				return reconcilers.Result{
					Requeue: true,
				}, errors.New("source image cannot be empty")
			}
			return reconcilers.Result{}, nil
		},
	}
}

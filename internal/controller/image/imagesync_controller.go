/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package image

import (
	imagev1alpha1 "github.com/rkgcloud/image-sync-controller/api/image/v1alpha1"
	"reconciler.io/runtime/reconcilers"
)

const imageSyncFinalizer = "image.apps.rkgcloud.com/finalizer"

// +kubebuilder:rbac:groups=image.apps.rkgcloud.com,resources=imagesyncs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=image.apps.rkgcloud.com,resources=imagesyncs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=image.apps.rkgcloud.com,resources=imagesyncs/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;create;patch;update;delete;watch
// +kubebuilder:rbac:groups=core,resources=serviceaccounts,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;update;patch;delete

func ImageSyncReconcile(c reconcilers.Config) *reconcilers.ResourceReconciler[*imagev1alpha1.ImageSync] {
	return &reconcilers.ResourceReconciler[*imagev1alpha1.ImageSync]{
		Name: "ImageSync",
		Reconciler: &reconcilers.WithFinalizer[*imagev1alpha1.ImageSync]{
			Finalizer: imageSyncFinalizer,
			Reconciler: reconcilers.Sequence[*imagev1alpha1.ImageSync]{
				SourceReconcile(),
			},
		},
		Config: c,
	}
}

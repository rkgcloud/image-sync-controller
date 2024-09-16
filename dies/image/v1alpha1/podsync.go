package v1alpha1

import (
	imagev1alpha1 "github.com/rkgcloud/image-sync-controller/api/image/v1alpha1"
)

// +die:object=true
type _ = imagev1alpha1.PodSync

// +diey
type _ = imagev1alpha1.PodSyncStatus

// +die
type _ = imagev1alpha1.ContainerInfo

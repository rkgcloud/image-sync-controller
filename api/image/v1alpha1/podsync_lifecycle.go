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

package v1alpha1

import (
	"context"

	"reconciler.io/runtime/apis"
)

var (
	PodSyncLabelKey = GroupVersion.Group + "/pod-sync"
)

const (
	PodSyncConditionSucceeded    = apis.ConditionReady
	PodSyncConditionPodCreated   = "PodCreated"
	PodSyncConditionPodRunning   = "PodRunning"
	PodSyncConditionPodCompleted = "PodCompleted"
)

var PodSyncConditionSet = apis.NewLivingConditionSet(
	PodSyncConditionPodCreated,
	PodSyncConditionPodRunning,
	PodSyncConditionPodCompleted,
)

func (s *PodSyncStatus) InitializeConditions(ctx context.Context) {
	conditionManager := ImageSyncConditionSet.ManageWithContext(ctx, s)
	conditionManager.InitializeConditions()
}

func (s *PodSyncStatus) GetConditionsAccessor() apis.ConditionsAccessor {
	return &s.Status
}

func (s *PodSyncStatus) GetConditionSet() apis.ConditionSet {
	return PodSyncConditionSet
}

func (s *PodSyncStatus) MarkPodSyncConditionPodCreated(ctx context.Context) {
	PodSyncConditionSet.ManageWithContext(ctx, s).MarkTrue(PodSyncConditionPodCreated,
		"PodCreated", "The Pod has been created")
}

func (s *PodSyncStatus) MarkPodSyncConditionPodCreateFailed(ctx context.Context, message string) {
	PodSyncConditionSet.ManageWithContext(ctx, s).MarkFalse(PodSyncConditionPodCreated,
		"PodCreateFailed", message)
}

func (s *PodSyncStatus) MarkPodSyncConditionPodRunning(ctx context.Context) {
	PodSyncConditionSet.ManageWithContext(ctx, s).MarkTrue(PodSyncConditionPodRunning,
		"PodRunning", "The Pod is running")
}

func (s *PodSyncStatus) MarkPodSyncConditionPodRunFailed(ctx context.Context, message string) {
	PodSyncConditionSet.ManageWithContext(ctx, s).MarkFalse(PodSyncConditionPodRunning,
		"PodRunFailed", message)
}

func (s *PodSyncStatus) MarkPodSyncConditionPodCompleted(ctx context.Context) {
	PodSyncConditionSet.ManageWithContext(ctx, s).MarkTrue(PodSyncConditionPodCompleted,
		"PodCompleted", "The Pod is completed")
}

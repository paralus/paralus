package step

import (
	"testing"

	clusterv2 "github.com/paralus/paralus/proto/types/controller"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	exv1beta1 "k8s.io/api/extensions/v1beta1"
)

func TestObjectState_ExtensionsV1Beta1Deployment(t *testing.T) {
	tests := []struct {
		name       string
		deployment *exv1beta1.Deployment
		wantState  clusterv2.StepObjectState
		wantReason string
	}{
		{
			name: "complete when available with no unavailable replicas",
			deployment: &exv1beta1.Deployment{
				Status: exv1beta1.DeploymentStatus{
					UnavailableReplicas: 0,
					Conditions: []exv1beta1.DeploymentCondition{
						{Type: exv1beta1.DeploymentAvailable, Status: corev1.ConditionTrue, Message: "all good"},
					},
				},
			},
			wantState:  clusterv2.StepObjectComplete,
			wantReason: "all good",
		},
		{
			name: "failed when not progressing",
			deployment: &exv1beta1.Deployment{
				Status: exv1beta1.DeploymentStatus{
					Conditions: []exv1beta1.DeploymentCondition{
						{Type: exv1beta1.DeploymentProgressing, Status: corev1.ConditionFalse, Message: "stuck"},
					},
				},
			},
			wantState:  clusterv2.StepObjectFailed,
			wantReason: "stuck",
		},
		{
			name: "failed on replica failure even if progressing",
			deployment: &exv1beta1.Deployment{
				Status: exv1beta1.DeploymentStatus{
					Conditions: []exv1beta1.DeploymentCondition{
						{Type: exv1beta1.DeploymentProgressing, Status: corev1.ConditionTrue},
						{Type: exv1beta1.DeploymentReplicaFailure, Status: corev1.ConditionTrue, Message: "replica error"},
					},
				},
			},
			wantState:  clusterv2.StepObjectFailed,
			wantReason: "replica error",
		},
		{
			name: "created when progressing but not yet available",
			deployment: &exv1beta1.Deployment{
				Status: exv1beta1.DeploymentStatus{
					Conditions: []exv1beta1.DeploymentCondition{
						{Type: exv1beta1.DeploymentProgressing, Status: corev1.ConditionTrue, Message: "rolling out"},
					},
				},
			},
			wantState:  clusterv2.StepObjectCreated,
			wantReason: "rolling out",
		},
		{
			name: "available but with unavailable replicas is not complete",
			deployment: &exv1beta1.Deployment{
				Status: exv1beta1.DeploymentStatus{
					UnavailableReplicas: 1,
					Conditions: []exv1beta1.DeploymentCondition{
						{Type: exv1beta1.DeploymentAvailable, Status: corev1.ConditionTrue},
						{Type: exv1beta1.DeploymentProgressing, Status: corev1.ConditionTrue, Message: "still rolling"},
					},
				},
			},
			wantState:  clusterv2.StepObjectCreated,
			wantReason: "still rolling",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state, reason := ObjectState(tt.deployment)
			assert.Equal(t, tt.wantState, state)
			assert.Equal(t, tt.wantReason, reason)
		})
	}
}

func TestObjectState_AppsV1Deployment(t *testing.T) {
	tests := []struct {
		name       string
		deployment *appsv1.Deployment
		wantState  clusterv2.StepObjectState
		wantReason string
	}{
		{
			name: "complete when available with no unavailable replicas",
			deployment: &appsv1.Deployment{
				Status: appsv1.DeploymentStatus{
					UnavailableReplicas: 0,
					Conditions: []appsv1.DeploymentCondition{
						{Type: appsv1.DeploymentAvailable, Status: corev1.ConditionTrue, Message: "all good"},
					},
				},
			},
			wantState:  clusterv2.StepObjectComplete,
			wantReason: "all good",
		},
		{
			name: "failed when not progressing",
			deployment: &appsv1.Deployment{
				Status: appsv1.DeploymentStatus{
					Conditions: []appsv1.DeploymentCondition{
						{Type: appsv1.DeploymentProgressing, Status: corev1.ConditionFalse, Message: "stuck"},
					},
				},
			},
			wantState:  clusterv2.StepObjectFailed,
			wantReason: "stuck",
		},
		{
			name: "failed on replica failure even if progressing",
			deployment: &appsv1.Deployment{
				Status: appsv1.DeploymentStatus{
					Conditions: []appsv1.DeploymentCondition{
						{Type: appsv1.DeploymentProgressing, Status: corev1.ConditionTrue},
						{Type: appsv1.DeploymentReplicaFailure, Status: corev1.ConditionTrue, Message: "replica error"},
					},
				},
			},
			wantState:  clusterv2.StepObjectFailed,
			wantReason: "replica error",
		},
		{
			name: "created when progressing but not yet available",
			deployment: &appsv1.Deployment{
				Status: appsv1.DeploymentStatus{
					Conditions: []appsv1.DeploymentCondition{
						{Type: appsv1.DeploymentProgressing, Status: corev1.ConditionTrue, Message: "rolling out"},
					},
				},
			},
			wantState:  clusterv2.StepObjectCreated,
			wantReason: "rolling out",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state, reason := ObjectState(tt.deployment)
			assert.Equal(t, tt.wantState, state)
			assert.Equal(t, tt.wantReason, reason)
		})
	}
}

func int32Ptr(i int32) *int32 { return &i }

func TestObjectState_StatefulSet(t *testing.T) {
	t.Run("complete when ready replicas match desired", func(t *testing.T) {
		s := &appsv1.StatefulSet{
			Spec:   appsv1.StatefulSetSpec{Replicas: int32Ptr(3)},
			Status: appsv1.StatefulSetStatus{ReadyReplicas: 3},
		}
		state, reason := ObjectState(s)
		assert.Equal(t, clusterv2.StepObjectComplete, state)
		assert.Equal(t, "all required replicas ready", reason)
	})

	t.Run("created when ready replicas are fewer than desired", func(t *testing.T) {
		s := &appsv1.StatefulSet{
			Spec:   appsv1.StatefulSetSpec{Replicas: int32Ptr(3)},
			Status: appsv1.StatefulSetStatus{ReadyReplicas: 1},
		}
		state, reason := ObjectState(s)
		assert.Equal(t, clusterv2.StepObjectCreated, state)
		assert.Equal(t, "in progress", reason)
	})
}

func TestObjectState_DaemonSet(t *testing.T) {
	t.Run("complete when ready matches desired", func(t *testing.T) {
		ds := &appsv1.DaemonSet{
			Status: appsv1.DaemonSetStatus{DesiredNumberScheduled: 3, NumberReady: 3},
		}
		state, reason := ObjectState(ds)
		assert.Equal(t, clusterv2.StepObjectComplete, state)
		assert.Equal(t, "all required replicas ready", reason)
	})

	t.Run("created when ready is fewer than desired", func(t *testing.T) {
		ds := &appsv1.DaemonSet{
			Status: appsv1.DaemonSetStatus{DesiredNumberScheduled: 3, NumberReady: 1},
		}
		state, reason := ObjectState(ds)
		assert.Equal(t, clusterv2.StepObjectCreated, state)
		assert.Equal(t, "in progress", reason)
	})
}

func TestObjectState_Job(t *testing.T) {
	t.Run("complete on JobComplete=True", func(t *testing.T) {
		j := &batchv1.Job{Status: batchv1.JobStatus{
			Conditions: []batchv1.JobCondition{{Type: batchv1.JobComplete, Status: corev1.ConditionTrue, Message: "done"}},
		}}
		state, reason := ObjectState(j)
		assert.Equal(t, clusterv2.StepObjectComplete, state)
		assert.Equal(t, "done", reason)
	})

	t.Run("failed on JobFailed=True", func(t *testing.T) {
		j := &batchv1.Job{Status: batchv1.JobStatus{
			Conditions: []batchv1.JobCondition{{Type: batchv1.JobFailed, Status: corev1.ConditionTrue, Message: "oops"}},
		}}
		state, reason := ObjectState(j)
		assert.Equal(t, clusterv2.StepObjectFailed, state)
		assert.Equal(t, "oops", reason)
	})

	t.Run("created when no terminal condition is set", func(t *testing.T) {
		j := &batchv1.Job{}
		state, reason := ObjectState(j)
		assert.Equal(t, clusterv2.StepObjectCreated, state)
		assert.Equal(t, "not completed", reason)
	})
}

func TestObjectState_UnhandledType(t *testing.T) {
	// A ConfigMap has no dedicated handler, so ObjectState falls through to
	// the default branch and reports it as complete.
	state, reason := ObjectState(&corev1.ConfigMap{})
	assert.Equal(t, clusterv2.StepObjectComplete, state)
	assert.Contains(t, reason, "not handled")
}

func TestJobState(t *testing.T) {
	t.Run("complete on JobComplete=True", func(t *testing.T) {
		j := &batchv1.Job{Status: batchv1.JobStatus{
			Conditions: []batchv1.JobCondition{{Type: batchv1.JobComplete, Status: corev1.ConditionTrue, Message: "done"}},
		}}
		state, reason := JobState(j)
		assert.Equal(t, clusterv2.StepJobComplete, state)
		assert.Equal(t, "done", reason)
	})

	t.Run("failed on JobFailed=True", func(t *testing.T) {
		j := &batchv1.Job{Status: batchv1.JobStatus{
			Conditions: []batchv1.JobCondition{{Type: batchv1.JobFailed, Status: corev1.ConditionTrue, Message: "oops"}},
		}}
		state, reason := JobState(j)
		assert.Equal(t, clusterv2.StepJobFailed, state)
		assert.Equal(t, "oops", reason)
	})

	t.Run("created when no terminal condition is set", func(t *testing.T) {
		j := &batchv1.Job{}
		state, reason := JobState(j)
		assert.Equal(t, clusterv2.StepJobCreated, state)
		assert.Equal(t, "not completed", reason)
	})

	t.Run("panics on a non-Job object, since JobState assumes the type", func(t *testing.T) {
		assert.Panics(t, func() {
			JobState(&corev1.ConfigMap{})
		})
	})
}

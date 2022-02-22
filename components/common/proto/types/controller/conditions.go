package controller

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:generate=false

// NewNamespaceConditionFunc is the function signature for creating new namespace condition
type NewNamespaceConditionFunc func(status ConditionStatus, reason string) *NamespaceCondition

// +kubebuilder:object:generate=false

// NewTaskletConditionFunc is the function signature for creating new tasklet condition
type NewTaskletConditionFunc func(status ConditionStatus, reason string) *TaskletCondition

// +kubebuilder:object:generate=false

// NewTaskConditionFunc is the function signature for creating new task condition
type NewTaskConditionFunc func(status ConditionStatus, reason string) *TaskCondition

// +kubebuilder:object:generate=false

// NamespaceConditionFunc checks if condition type is complete
type NamespaceConditionFunc func(n *Namespace) bool

// +kubebuilder:object:generate=false

// TaskletConditionFunc checks if condition type is complete
type TaskletConditionFunc func(t *Tasklet) bool

// +kubebuilder:object:generate=false

// TaskConditionFunc checks if condition type is complete
type TaskConditionFunc func(t *Task) bool

// +kubebuilder:object:generate=false

// GetTaskConditionReasonFunc returns the reason of the task condition
type GetTaskConditionReasonFunc func(t *Task) string

// +kubebuilder:object:generate=false

// GetTaskletConditionReasonFunc returns the reason of the tasklet condition
type GetTaskletConditionReasonFunc func(t *Tasklet) string

// +kubebuilder:object:generate=false

// GetNamespaceConditionReasonFunc returns the reason for namespace condition
type GetNamespaceConditionReasonFunc func(n *Namespace) string

// utility methods for creating/checking conditions
var (
	NewNamespaceUpsert     NewNamespaceConditionFunc = newNamespaceCondition(NamespaceUpsert)
	NewNamespaceInit       NewNamespaceConditionFunc = newNamespaceCondition(NamespaceInit)
	NewNamespaceCreate     NewNamespaceConditionFunc = newNamespaceCondition(NamespaceCreate)
	NewNamespacePostCreate NewNamespaceConditionFunc = newNamespaceCondition(NamespacePostCreate)
	NewNamespacePreDelete  NewNamespaceConditionFunc = newNamespaceCondition(NamespacePreDelete)
	NewNamespaceReady      NewNamespaceConditionFunc = newNamespaceCondition(NamespaceReady)

	NewTaskletUpsert      NewTaskletConditionFunc = newTaskletCondition(TaskletUpsert)
	NewTaskletInit        NewTaskletConditionFunc = newTaskletCondition(TaskletInit)
	NewTaskletInstall     NewTaskletConditionFunc = newTaskletCondition(TaskletInstall)
	NewTaskletPostInstall NewTaskletConditionFunc = newTaskletCondition(TaskletPostInstall)
	NewTaskletPreDelete   NewTaskletConditionFunc = newTaskletCondition(TaskletPreDelete)
	NewTaskletReady       NewTaskletConditionFunc = newTaskletCondition(TaskletReady)

	NewTaskUpsert    NewTaskConditionFunc = newTaskCondition(TaskUpsert)
	NewTaskInit      NewTaskConditionFunc = newTaskCondition(TaskInit)
	NewTaskletCreate NewTaskConditionFunc = newTaskCondition(TaskletCreate)
	NewTaskPreDelete NewTaskConditionFunc = newTaskCondition(TaskPreDelete)
	NewTaskReady     NewTaskConditionFunc = newTaskCondition(TaskReady)

	TaskReadyFailedReason    GetTaskConditionReasonFunc    = getTaskConditionReason(Failed, TaskReady)
	TaskConvergeFailedReason GetTaskConditionReasonFunc    = getTaskConditionReason(Failed, TaskInit, TaskletCreate)
	TaskletReadyFailedReason GetTaskletConditionReasonFunc = getTaskletConditionReason(Failed, TaskletReady)
	TaskletFailedReason      GetTaskletConditionReasonFunc = getTaskletConditionReason(Failed, TaskletInit, TaskletInstall, TaskletPostInstall, TaskletReady)

	NamespaceConvergeFailedReason GetNamespaceConditionReasonFunc = getNamespaceConditionReason(Failed, NamespaceInit, NamespaceCreate, NamespacePostCreate)
	NamespaceReadyReason          GetNamespaceConditionReasonFunc = getNamespaceConditionReason(Failed, NamespaceReady)
)

func newNamespaceCondition(conditionType NamespaceConditionType) func(status ConditionStatus, reason string) *NamespaceCondition {
	return func(status ConditionStatus, reason string) *NamespaceCondition {
		return &NamespaceCondition{
			Type:           string(conditionType),
			Status:         string(status),
			Reason:         reason,
			LastUpdateTime: &metav1.Time{time.Now()},
		}
	}
}

func newTaskletCondition(conditionType TaskletConditionType) func(status ConditionStatus, reason string) *TaskletCondition {
	return func(status ConditionStatus, reason string) *TaskletCondition {
		return &TaskletCondition{
			Type:           string(conditionType),
			Status:         string(status),
			Reason:         reason,
			LastUpdateTime: &metav1.Time{time.Now()},
		}
	}
}

func newTaskCondition(conditionType TaskConditionType) func(status ConditionStatus, reason string) *TaskCondition {
	return func(status ConditionStatus, reason string) *TaskCondition {
		return &TaskCondition{
			Type:           string(conditionType),
			Status:         string(status),
			Reason:         reason,
			LastUpdateTime: &metav1.Time{time.Now()},
		}
	}
}

func getTaskletConditionReason(conditionStatus ConditionStatus, conditionTypes ...TaskletConditionType) func(t *Tasklet) string {
	return func(t *Tasklet) string {
		for _, condition := range t.Status.Conditions {
			for _, conditionType := range conditionTypes {
				if condition.Type == string(conditionType) {
					if condition.Status == string(conditionStatus) {
						return condition.Reason
					}
				}

			}
		}
		return ""
	}
}

func getTaskConditionReason(conditionStatus ConditionStatus, conditionTypes ...TaskConditionType) func(t *Task) string {
	return func(t *Task) string {
		for _, conditionType := range conditionTypes {
			for _, condition := range t.Status.Conditions {
				if condition.Type == string(conditionType) {
					if condition.Status == string(conditionStatus) {
						return condition.Reason
					}

				}
			}
		}

		return ""
	}
}

func getNamespaceConditionReason(conditionStatus ConditionStatus, conditionTypes ...NamespaceConditionType) func(n *Namespace) string {
	return func(n *Namespace) string {
		for _, conditionType := range conditionTypes {
			for _, condition := range n.Status.Conditions {
				if condition.Type == string(conditionType) {
					if condition.Status == string(conditionStatus) {
						return condition.Reason
					}

				}
			}
		}

		return ""
	}
}

// SetNamespaceCondition sets namespace condition
func SetNamespaceCondition(n *Namespace, condition NamespaceCondition) {
	found := false
	for i := range n.Status.Conditions {
		if n.Status.Conditions[i].Type == condition.Type {
			found = true
			n.Status.Conditions[i] = &condition
		}
	}
	if !found {
		n.Status.Conditions = append(n.Status.Conditions, &condition)
	}
}

// SetTaskletCondition sets tasklet condition
func SetTaskletCondition(t *Tasklet, condition TaskletCondition) {
	found := false
	for i := range t.Status.Conditions {
		if t.Status.Conditions[i].Type == condition.Type {
			found = true
			t.Status.Conditions[i] = &condition
		}
	}
	if !found {
		t.Status.Conditions = append(t.Status.Conditions, &condition)
	}
}

// SetTaskCondition sets task condition
func SetTaskCondition(t *Task, condition TaskCondition) {
	found := false
	for i := range t.Status.Conditions {
		if t.Status.Conditions[i].Type == condition.Type {
			found = true
			t.Status.Conditions[i] = &condition
		}
	}
	if !found {
		t.Status.Conditions = append(t.Status.Conditions, &condition)
	}
}

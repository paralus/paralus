package controller

import (
	"testing"
)

func TestCheckNamespaceConditions(t *testing.T) {
	namespaceConditionChecker := checkNamespaceConditions(
		withNamespaceConditionStatus(Complete),
		withNamespaceConditionType(NamespaceInit),
		withNamespaceConditionType(NamespaceCreate),
	)

	n := new(Namespace)
	n.Status.Conditions = append(n.Status.Conditions, NewNamespaceCreate(Complete, "test"), NewNamespaceInit(Failed, "test"))

	if namespaceConditionChecker(n) {
		t.Error("expeted false")
	}

	namespaceConditionChecker = checkNamespaceConditions(
		withNamespaceConditionStatus(Complete),
		withNamespaceConditionType(NamespaceInit),
		withNamespaceConditionType(NamespaceCreate),
		withNamespaceConditionShortCircuit(),
	)

	if !namespaceConditionChecker(n) {
		t.Error("expeted true")
	}

	n = new(Namespace)
	n.Status.Conditions = append(n.Status.Conditions, NewNamespaceCreate(Complete, "test"), NewNamespaceInit(Complete, "test"))

	if !namespaceConditionChecker(n) {
		t.Error("expeted true")
	}

}

func TestCheckTaskConditions(t *testing.T) {
	taskConditionChecker := checkTaskConditions(
		withTaskConditionStatus(Complete),
		withTaskConditionType(TaskInit),
		withTaskConditionType(TaskletCreate),
	)

	task := new(Task)
	task.Status.Conditions = append(task.Status.Conditions, NewTaskInit(Complete, "test"), NewTaskletCreate(Failed, "test"))

	if taskConditionChecker(task) {
		t.Error("expeted false")
	}

	taskConditionChecker = checkTaskConditions(
		withTaskConditionStatus(Complete),
		withTaskConditionType(TaskInit),
		withTaskConditionType(TaskletCreate),
		withTaskConditionShortCircuit(),
	)

	if !taskConditionChecker(task) {
		t.Error("expeted true")
	}

	task = new(Task)
	task.Status.Conditions = append(task.Status.Conditions, NewTaskInit(Complete, "test"), NewTaskletCreate(Complete, "test"))

	if !taskConditionChecker(task) {
		t.Error("expeted true")
	}

}

func TestCheckTaskletConditions(t *testing.T) {
	taskletConditionChecker := checkTaskletConditions(
		withTaskletConditionStatus(Complete),
		withTaskletConditionType(TaskletInit),
		withTaskletConditionType(TaskletInstall),
	)

	tasklet := new(Tasklet)
	tasklet.Status.Conditions = append(tasklet.Status.Conditions, NewTaskletInit(Complete, "test"), NewTaskletInstall(Failed, "test"))

	if taskletConditionChecker(tasklet) {
		t.Error("expeted false")
	}

	taskletConditionChecker = checkTaskletConditions(
		withTaskletConditionStatus(Complete),
		withTaskletConditionType(TaskletInit),
		withTaskletConditionType(TaskletInstall),
		withTaskletConditionShortCircuit(),
	)

	if !taskletConditionChecker(tasklet) {
		t.Error("expeted true")
	}

	tasklet = new(Tasklet)
	tasklet.Status.Conditions = append(tasklet.Status.Conditions, NewTaskletInit(Complete, "test"), NewTaskletInit(Complete, "test"))

	if !taskletConditionChecker(tasklet) {
		t.Error("expeted true")
	}

}

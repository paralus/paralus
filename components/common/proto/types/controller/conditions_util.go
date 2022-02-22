package controller

type namespaceConditionCheckOptions struct {
	conditionStatus ConditionStatus
	conditionTypes  []NamespaceConditionType
	shortCircuit    bool
}

type taskConditionCheckOptions struct {
	conditionStatus ConditionStatus
	conditionTypes  []TaskConditionType
	shortCircuit    bool
}

type taskletConditionCheckOptions struct {
	conditionStatus ConditionStatus
	conditionTypes  []TaskletConditionType
	shortCircuit    bool
}

type namespaceConditionCheckOption func(*namespaceConditionCheckOptions)
type taskConditionCheckOption func(*taskConditionCheckOptions)
type taskletConditionCheckOption func(*taskletConditionCheckOptions)

func withNamespaceConditionStatus(status ConditionStatus) namespaceConditionCheckOption {
	return func(opts *namespaceConditionCheckOptions) {
		opts.conditionStatus = status
	}
}

func withNamespaceConditionType(conditionType NamespaceConditionType) namespaceConditionCheckOption {
	return func(opts *namespaceConditionCheckOptions) {
		opts.conditionTypes = append(opts.conditionTypes, conditionType)
	}
}

func withNamespaceConditionShortCircuit() namespaceConditionCheckOption {
	return func(opts *namespaceConditionCheckOptions) {
		opts.shortCircuit = true
	}
}

func withTaskConditionStatus(status ConditionStatus) taskConditionCheckOption {
	return func(opts *taskConditionCheckOptions) {
		opts.conditionStatus = status
	}
}

func withTaskConditionType(conditionType TaskConditionType) taskConditionCheckOption {
	return func(opts *taskConditionCheckOptions) {
		opts.conditionTypes = append(opts.conditionTypes, conditionType)
	}
}

func withTaskConditionShortCircuit() taskConditionCheckOption {
	return func(opts *taskConditionCheckOptions) {
		opts.shortCircuit = true
	}
}

func withTaskletConditionStatus(status ConditionStatus) taskletConditionCheckOption {
	return func(opts *taskletConditionCheckOptions) {
		opts.conditionStatus = status
	}
}

func withTaskletConditionType(conditionType TaskletConditionType) taskletConditionCheckOption {
	return func(opts *taskletConditionCheckOptions) {
		opts.conditionTypes = append(opts.conditionTypes, conditionType)
	}
}

func withTaskletConditionShortCircuit() taskletConditionCheckOption {
	return func(opts *taskletConditionCheckOptions) {
		opts.shortCircuit = true
	}
}

func checkNamespaceConditions(opts ...namespaceConditionCheckOption) NamespaceConditionFunc {

	checkOpts := new(namespaceConditionCheckOptions)
	for _, opt := range opts {
		opt(checkOpts)
	}

	return func(n *Namespace) bool {
		if !checkOpts.shortCircuit {
			allStatisfied := true
			found := false
			for _, condition := range n.Status.Conditions {
				for _, conditionType := range checkOpts.conditionTypes {
					if condition.Type == string(conditionType) {
						found = true
						if condition.Status != string(checkOpts.conditionStatus) {
							allStatisfied = false
						}
					}
				}
			}
			return found && allStatisfied
		}

		for _, condition := range n.Status.Conditions {
			for _, conditionType := range checkOpts.conditionTypes {
				if condition.Type == string(conditionType) {
					if condition.Status == string(checkOpts.conditionStatus) {
						return true
					}
				}
			}
		}
		return false

	}
}

func checkTaskConditions(opts ...taskConditionCheckOption) TaskConditionFunc {

	checkOpts := new(taskConditionCheckOptions)
	for _, opt := range opts {
		opt(checkOpts)
	}

	return func(n *Task) bool {
		if !checkOpts.shortCircuit {
			allStatisfied := true
			found := false
			for _, condition := range n.Status.Conditions {
				for _, conditionType := range checkOpts.conditionTypes {
					if condition.Type == string(conditionType) {
						found = true
						if condition.Status != string(checkOpts.conditionStatus) {
							allStatisfied = false
						}
					}
				}
			}
			return found && allStatisfied
		}

		for _, condition := range n.Status.Conditions {
			for _, conditionType := range checkOpts.conditionTypes {
				if condition.Type == string(conditionType) {
					if condition.Status == string(checkOpts.conditionStatus) {
						return true
					}
				}
			}
		}
		return false

	}
}

func checkTaskletConditions(opts ...taskletConditionCheckOption) TaskletConditionFunc {

	checkOpts := new(taskletConditionCheckOptions)
	for _, opt := range opts {
		opt(checkOpts)
	}

	return func(n *Tasklet) bool {
		if !checkOpts.shortCircuit {
			allStatisfied := true
			found := false
			for _, condition := range n.Status.Conditions {
				for _, conditionType := range checkOpts.conditionTypes {
					if condition.Type == string(conditionType) {
						found = true
						if condition.Status != string(checkOpts.conditionStatus) {
							allStatisfied = false
						}
					}
				}
			}
			return found && allStatisfied
		}

		for _, condition := range n.Status.Conditions {
			for _, conditionType := range checkOpts.conditionTypes {
				if condition.Type == string(conditionType) {
					if condition.Status == string(checkOpts.conditionStatus) {
						return true
					}
				}
			}
		}
		return false

	}
}

// exported namespace condition utils
var (
	IsNamespaceUpserted NamespaceConditionFunc = checkNamespaceConditions(
		withNamespaceConditionStatus(Complete),
		withNamespaceConditionType(NamespaceUpsert))

	IsNamespaceInited NamespaceConditionFunc = checkNamespaceConditions(
		withNamespaceConditionStatus(Complete),
		withNamespaceConditionType(NamespaceInit))

	IsNamespaceCreated NamespaceConditionFunc = checkNamespaceConditions(
		withNamespaceConditionStatus(Complete),
		withNamespaceConditionType(NamespaceCreate))

	IsNamespacePostCreated NamespaceConditionFunc = checkNamespaceConditions(
		withNamespaceConditionStatus(Complete),
		withNamespaceConditionType(NamespacePostCreate))

	IsNamespacePreDeleted NamespaceConditionFunc = checkNamespaceConditions(
		withNamespaceConditionStatus(Complete),
		withNamespaceConditionType(NamespacePreDelete))

	IsNamespaceReady NamespaceConditionFunc = checkNamespaceConditions(
		withNamespaceConditionStatus(Complete),
		withNamespaceConditionType(NamespaceReady))

	IsNamespaceReadyFailed NamespaceConditionFunc = checkNamespaceConditions(
		withNamespaceConditionStatus(Failed),
		withNamespaceConditionType(NamespaceReady))

	IsNamespaceConverged NamespaceConditionFunc = checkNamespaceConditions(
		withNamespaceConditionStatus(Complete),
		withNamespaceConditionType(NamespaceInit),
		withNamespaceConditionType(NamespaceCreate),
		withNamespaceConditionType(NamespacePostCreate),
	)

	IsNamespaceConvergeFailed NamespaceConditionFunc = checkNamespaceConditions(
		withNamespaceConditionStatus(Failed),
		withNamespaceConditionType(NamespaceInit),
		withNamespaceConditionType(NamespaceCreate),
		withNamespaceConditionType(NamespacePostCreate),
		withNamespaceConditionShortCircuit(),
	)
)

// exported task condition utils
var (
	IsTaskUpserted TaskConditionFunc = checkTaskConditions(
		withTaskConditionStatus(Complete),
		withTaskConditionType(TaskUpsert))

	IsTaskInited TaskConditionFunc = checkTaskConditions(
		withTaskConditionStatus(Complete),
		withTaskConditionType(TaskInit))

	IsTaskletCreated TaskConditionFunc = checkTaskConditions(
		withTaskConditionStatus(Complete),
		withTaskConditionType(TaskletCreate))

	IsTaskPreDeleted TaskConditionFunc = checkTaskConditions(
		withTaskConditionStatus(Complete),
		withTaskConditionType(TaskPreDelete))

	IsTaskReady TaskConditionFunc = checkTaskConditions(
		withTaskConditionStatus(Complete),
		withTaskConditionType(TaskReady))

	IsTaskFailed TaskConditionFunc = checkTaskConditions(
		withTaskConditionStatus(Failed),
		withTaskConditionType(TaskInit),
		withTaskConditionType(TaskletCreate),
		withTaskConditionType(TaskReady),
		withTaskConditionShortCircuit(),
	)

	IsTaskConverged TaskConditionFunc = checkTaskConditions(
		withTaskConditionStatus(Complete),
		withTaskConditionType(TaskInit),
		withTaskConditionType(TaskletCreate),
	)

	IsTaskConvergeFailed TaskConditionFunc = checkTaskConditions(
		withTaskConditionStatus(Failed),
		withTaskConditionType(TaskInit),
		withTaskConditionType(TaskletCreate),
		withTaskConditionShortCircuit(),
	)
)

// exported tasklet utils
var (
	IsTaskletUpserted TaskletConditionFunc = checkTaskletConditions(
		withTaskletConditionStatus(Complete),
		withTaskletConditionType(TaskletUpsert))

	IsTaskletInited TaskletConditionFunc = checkTaskletConditions(
		withTaskletConditionStatus(Complete),
		withTaskletConditionType(TaskletInit))

	IsTaskletInstalled TaskletConditionFunc = checkTaskletConditions(
		withTaskletConditionStatus(Complete),
		withTaskletConditionType(TaskletInstall))

	IsTaskletPostInstalled TaskletConditionFunc = checkTaskletConditions(
		withTaskletConditionStatus(Complete),
		withTaskletConditionType(TaskletPostInstall))

	IsTaskletPreDeleted TaskletConditionFunc = checkTaskletConditions(
		withTaskletConditionStatus(Complete),
		withTaskletConditionType(TaskletPreDelete))

	IsTaskletReady TaskletConditionFunc = checkTaskletConditions(
		withTaskletConditionStatus(Complete),
		withTaskletConditionType(TaskletReady))

	IsTaskletFailed TaskletConditionFunc = checkTaskletConditions(
		withTaskletConditionStatus(Failed),
		withTaskletConditionType(TaskletInit),
		withTaskletConditionType(TaskletInstall),
		withTaskletConditionType(TaskletPostInstall),
		withTaskletConditionType(TaskletReady),
		withTaskletConditionShortCircuit(),
	)
)

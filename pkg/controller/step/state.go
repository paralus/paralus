package step

import (
	"fmt"

	clusterv2 "github.com/paralus/paralus/proto/types/controller"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	exv1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
)

func getExV1Beta1DeploymentState(d *exv1beta1.Deployment) (state clusterv2.StepObjectState, reason string) {
	var available bool
	var progressing bool = true
	var replicasFailure bool
	var progressMessage, availableMessage, replicaFailureMessage string

	for _, condition := range d.Status.Conditions {
		if condition.Type == exv1beta1.DeploymentAvailable {
			if condition.Status == corev1.ConditionTrue {
				available = true
			}
			availableMessage = condition.Message
		}
		if condition.Type == exv1beta1.DeploymentProgressing {
			if condition.Status != corev1.ConditionTrue {
				progressing = false
			}
			progressMessage = condition.Message
		}
		if condition.Type == exv1beta1.DeploymentReplicaFailure {
			if condition.Status == corev1.ConditionTrue {
				replicasFailure = true
				replicaFailureMessage = condition.Message
			}

		}
	}

	switch {
	case available && d.Status.UnavailableReplicas == 0:
		state = clusterv2.StepObjectComplete
		reason = availableMessage
		return
	case !progressing || replicasFailure:
		state = clusterv2.StepObjectFailed
		if !progressing {
			reason = progressMessage
		} else {
			reason = replicaFailureMessage
		}

		return
	default:
		state = clusterv2.StepObjectCreated
		reason = progressMessage
		return
	}

}

func getAppsV1DeploymentState(d *appsv1.Deployment) (state clusterv2.StepObjectState, reason string) {

	var available bool
	var progressing bool = true
	var replicasFailure bool
	var progressMessage, availableMessage, replicaFailureMessage string

	for _, condition := range d.Status.Conditions {
		if condition.Type == appsv1.DeploymentAvailable {
			if condition.Status == corev1.ConditionTrue {
				available = true
			}
			availableMessage = condition.Message

		}
		if condition.Type == appsv1.DeploymentProgressing {
			if condition.Status != corev1.ConditionTrue {
				progressing = false
			}
			progressMessage = condition.Message

		}
		if condition.Type == appsv1.DeploymentReplicaFailure {
			if condition.Status == corev1.ConditionTrue {
				replicasFailure = true
				replicaFailureMessage = condition.Message
			}

		}
	}

	switch {
	case available && d.Status.UnavailableReplicas == 0:
		state = clusterv2.StepObjectComplete
		reason = availableMessage
		return
	case !progressing || replicasFailure:
		state = clusterv2.StepObjectFailed
		if !progressing {
			reason = progressMessage
		} else {
			reason = replicaFailureMessage
		}

		return
	default:
		state = clusterv2.StepObjectCreated
		reason = progressMessage
		return

	}

}

func getStatefulSetState(s *appsv1.StatefulSet) (state clusterv2.StepObjectState, reason string) {

	if s.Status.ReadyReplicas == *s.Spec.Replicas {
		state = clusterv2.StepObjectComplete
		reason = "all required replicas ready"
		return
	}
	// for _, condition := range s.Status.Conditions {
	// 	if condition.Type
	// }

	state = clusterv2.StepObjectCreated
	reason = "in progress"
	return
}

func getDaemonSetState(ds *appsv1.DaemonSet) (state clusterv2.StepObjectState, reason string) {
	if ds.Status.DesiredNumberScheduled == ds.Status.NumberReady {
		state = clusterv2.StepObjectComplete
		reason = "all required replicas ready"
		return
	}

	state = clusterv2.StepObjectCreated
	reason = "in progress"
	return
}

func getJobState(j *batchv1.Job) (state clusterv2.StepObjectState, reason string) {
	for _, condition := range j.Status.Conditions {
		if condition.Type == batchv1.JobComplete {
			if condition.Status == corev1.ConditionTrue {
				state = clusterv2.StepObjectComplete
				reason = condition.Message
				return
			}
		}
		if condition.Type == batchv1.JobFailed {
			if condition.Status == corev1.ConditionTrue {
				state = clusterv2.StepObjectFailed
				reason = condition.Message
				return
			}
		}
	}
	state = clusterv2.StepObjectCreated
	reason = "not completed"
	return
}

func getPersistentVolumeClaimState(pvc *corev1.PersistentVolumeClaim) (state clusterv2.StepObjectState, reason string) {
	if pvc.Status.Phase == corev1.ClaimBound {
		state = clusterv2.StepObjectComplete
		reason = "claim bound"
		return
	}

	state = clusterv2.StepObjectCreated
	reason = "in progress"
	return
}

// ObjectState returns the object state of runtime object
func ObjectState(o runtime.Object) (state clusterv2.StepObjectState, reason string) {
	switch o.(type) {
	case *exv1beta1.Deployment:
		d := o.(*exv1beta1.Deployment)
		state, reason = getExV1Beta1DeploymentState(d)
	case *appsv1.Deployment:
		d := o.(*appsv1.Deployment)
		state, reason = getAppsV1DeploymentState(d)
	case *appsv1.StatefulSet:
		s := o.(*appsv1.StatefulSet)
		state, reason = getStatefulSetState(s)
	case *appsv1.DaemonSet:
		ds := o.(*appsv1.DaemonSet)
		state, reason = getDaemonSetState(ds)
	case *batchv1.Job:
		j := o.(*batchv1.Job)
		state, reason = getJobState(j)
	// case *corev1.PersistentVolumeClaim:
	// 	pvc := o.(*corev1.PersistentVolumeClaim)
	// 	state, reason = getPersistentVolumeClaimState(pvc)
	default:
		state = clusterv2.StepObjectComplete
		reason = fmt.Sprintf("object type %T not handled", o)
	}

	return
}

// JobState returns the job state
func JobState(o runtime.Object) (state clusterv2.StepJobState, reason string) {
	j := o.(*batchv1.Job)
	for _, condition := range j.Status.Conditions {
		if condition.Type == batchv1.JobComplete {
			if condition.Status == corev1.ConditionTrue {
				state = clusterv2.StepJobComplete
				reason = condition.Message
				return
			}
		}
		if condition.Type == batchv1.JobFailed {
			if condition.Status == corev1.ConditionTrue {
				state = clusterv2.StepJobFailed
				reason = condition.Message
				return
			}
		}
	}
	state = clusterv2.StepJobCreated
	reason = "not completed"
	return
}

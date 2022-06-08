package step

import (
	"context"

	"github.com/paralus/paralus/pkg/controller/apply"
	cr "github.com/paralus/paralus/pkg/controller/runtime"
	"github.com/paralus/paralus/pkg/controller/scheme"
	"github.com/paralus/paralus/pkg/controller/util"
	hash "github.com/paralus/paralus/pkg/hasher"
	clusterv2 "github.com/paralus/paralus/proto/types/controller"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	ref "k8s.io/client-go/tools/reference"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var (
	stepLog = logf.Log.WithName("cluster-v2-step")
	ma      = meta.NewAccessor()
)

// Handler is the interface for working with steps
type Handler interface {
	// Handle executes the step
	Handle(ctx context.Context, owner metav1.Object, step clusterv2.StepTemplate) (status clusterv2.StepStatus)
	// Status returns new status for current status
	HandleStatus(ctx context.Context, owner metav1.Object, original clusterv2.StepStatus) (current clusterv2.StepStatus)
	// Delete deletes a given step
	HandleDelete(ctx context.Context, owner metav1.Object, stage clusterv2.StageSpec) error
}

type handler struct {
	apply.Applier
}

// NewHandler returns object handler that handles step object
func NewHandler(a apply.Applier) Handler {
	return &handler{a}
}

var _ Handler = (*handler)(nil)

func (h *handler) getObject(ctx context.Context, ref *corev1.ObjectReference) (o client.Object, err error) {
	gvk := schema.FromAPIVersionAndKind(ref.APIVersion, ref.Kind)
	nn := types.NamespacedName{
		Namespace: ref.Namespace,
		Name:      ref.Name,
	}
	if o, err = util.NewObject(gvk); err != nil {
		return
	}

	err = h.Get(ctx, nn, o)
	return
}

func (h *handler) getObjectState(ctx context.Context, original clusterv2.StepStatus) (state clusterv2.StepObjectState, reason string, err error) {
	if original.ObjectRef == nil {
		state = clusterv2.StepObjectState(original.ObjectState)
		reason = original.ObjectReason
		return
	}

	var o runtime.Object
	o, err = h.getObject(ctx, original.ObjectRef)
	if err != nil {
		return
	}

	state, reason = ObjectState(o)

	return
}

func (h *handler) getJobState(ctx context.Context, ref *corev1.ObjectReference) (state clusterv2.StepJobState, reason string, err error) {
	if ref == nil {
		state = clusterv2.StepJobComplete
		reason = "not configured"
		return
	}

	var o runtime.Object
	o, err = h.getObject(ctx, ref)
	if err != nil {
		return
	}

	state, reason = JobState(o)

	return
}

func (h *handler) HandleStatus(ctx context.Context, owner metav1.Object, original clusterv2.StepStatus) (current clusterv2.StepStatus) {
	log := stepLog.WithValues("owner", types.NamespacedName{
		Namespace: owner.GetNamespace(),
		Name:      owner.GetName(),
	}, "step", original.Name)

	jobRef := original.JobRef

	var objectState clusterv2.StepObjectState
	var objectReason string
	var jobState clusterv2.StepJobState
	var jobReason string
	var err error

	current = *original.DeepCopy()

	if objectState, objectReason, err = h.getObjectState(ctx, original); err != nil {
		log.Info("unable to get object state", "error", err)
		return original
	}

	if jobState, jobReason, err = h.getJobState(ctx, jobRef); err != nil {
		log.Info("unable to get job state", "error", err)
		return original
	}

	log.Info("step", "objectState", objectState, "objectReason", objectReason, "jobState", jobState, "jobReason", jobReason)

	current.ObjectState = string(objectState)
	current.ObjectReason = objectReason
	current.JobState = string(jobState)
	current.JobReason = jobReason

	if objectState == clusterv2.StepObjectComplete && jobState == clusterv2.StepJobComplete {
		current.State = string(clusterv2.StepComplete)
		current.Reason = "complete"
		return
	}

	if objectState == clusterv2.StepObjectFailed {
		current.State = string(clusterv2.StepFailed)
		current.Reason = objectReason
		return
	}

	if jobState == clusterv2.StepJobFailed {
		current.State = string(clusterv2.StepFailed)
		current.Reason = jobReason
		return
	}

	return
}

func (h *handler) deleteStep(ctx context.Context, owner metav1.Object, step clusterv2.StepTemplate) error {
	log := stepLog.WithValues("owner", types.NamespacedName{
		Namespace: owner.GetNamespace(),
		Name:      owner.GetName(),
	}, "step", step.Name)

	log.Info("deleting step", "name", step.Name)

	if step.Object != nil {
		log.Info("step object configured for step", "gvk", step.Object.TypeMeta.GroupVersionKind())
		o, _, err := cr.ToUnstructuredObject(step.Object)
		if err != nil {
			return err
		}
		var objectKey client.ObjectKey
		var existing client.Object

		objectKey = client.ObjectKey{
			Name:      o.GetName(),
			Namespace: o.GetNamespace(),
		}

		existing, err = util.NewObject(o.GetObjectKind().GroupVersionKind())

		if err != nil {
			return err
		}

		err = h.Get(ctx, objectKey, existing)
		if err != nil {
			if apierrs.IsNotFound(err) {
				return nil
			}
			return err
		}

		if eo, ok := existing.(metav1.Object); ok {
			if !util.OwnsObject(owner, eo) {
				return nil
			}
			log.Info("deleting object", "name", step.Name)
			err = h.Delete(ctx, existing, client.PropagationPolicy(metav1.DeletePropagationBackground))
			if err != nil {
				return err
			}

		}

	}

	if step.JobTemplate != nil {
		log.Info("step job configured for step")
	}

	return nil
}

func (h *handler) HandleDelete(ctx context.Context, owner metav1.Object, stage clusterv2.StageSpec) error {
	log := stepLog.WithValues("owner", types.NamespacedName{
		Namespace: owner.GetNamespace(),
		Name:      owner.GetName(),
	})

	for _, step := range stage {
		log.Info("deleting step", "name", step.Name)
		err := h.deleteStep(ctx, owner, step)
		if err != nil {
			log.Info("unable to delete step", "name", step.Name, "error", err)
			return err
		}

	}

	return nil
}

func (h *handler) Handle(ctx context.Context, owner metav1.Object, step clusterv2.StepTemplate) (status clusterv2.StepStatus) {
	log := stepLog.WithValues("owner", types.NamespacedName{
		Namespace: owner.GetNamespace(),
		Name:      owner.GetName(),
	}, "step", step.Name)

	status = clusterv2.StepStatus{
		Name: step.Name,
	}
	// handle step object
	if step.Object != nil {
		accessor, err := step.Object.Accessor()
		if err != nil {
			log.Error(err, "unable to get step object accessor")
			status.ObjectState = string(clusterv2.StepObjectFailed)
			status.ObjectReason = err.Error()
			status.State = string(clusterv2.StepFailed)
			status.Reason = err.Error()
			return
		}
		gvk, err := accessor.GroupVersionKind()
		if err != nil {
			log.Error(err, "unable to get step object gvk")
			status.ObjectState = string(clusterv2.StepObjectFailed)
			status.ObjectReason = err.Error()
			status.State = string(clusterv2.StepFailed)
			status.Reason = err.Error()
			return
		}
		log.Info("step object configured for step", "gvk", gvk)
		o, _, err := cr.ToUnstructuredObject(step.Object)
		if err != nil {
			log.Error(err, "unable to convert step object")
			status.ObjectState = string(clusterv2.StepObjectFailed)
			status.ObjectReason = err.Error()
			status.State = string(clusterv2.StepFailed)
			status.Reason = err.Error()
			return
		}

		err = util.SetOwnerReference(owner, o)
		if err != nil {
			log.Error(err, "unable to set controller reference")
			status.ObjectState = string(clusterv2.StepObjectFailed)
			status.ObjectReason = err.Error()
			status.State = string(clusterv2.StepFailed)
			status.Reason = err.Error()
			return
		}

		err = h.Apply(ctx, o)
		if err != nil {
			log.Error(err, "unable to patch step object", "name", step.Name)
			status.ObjectState = string(clusterv2.StepObjectFailed)
			status.ObjectReason = err.Error()
			status.State = string(clusterv2.StepFailed)
			status.Reason = err.Error()
			return
		}

		objRef, _ := ref.GetReference(scheme.Scheme, o)
		status.ObjectState = string(clusterv2.StepObjectCreated)
		status.ObjectReason = "created"
		status.ObjectRef = objRef
		hash, _ := hash.GetHash(o)
		status.ObjectSpecHash = hash
	}

	// handle step job
	if step.JobTemplate != nil {
		log.Info("step job configured for step")
	}

	status.State = string(clusterv2.StepExecuted)

	return

}

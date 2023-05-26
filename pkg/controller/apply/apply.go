package apply

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/paralus/paralus/pkg/controller/client"
	scheme "github.com/paralus/paralus/pkg/controller/scheme"
	"github.com/paralus/paralus/pkg/controller/util"
	clusterv2 "github.com/paralus/paralus/proto/types/controller"
	apixv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/retry"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var (
	applyLog = logf.Log.WithName("cluster-v2-apply")
)

var (
	crdv1beta1GVK = schema.GroupVersionKind{
		Group:   apixv1beta1.SchemeGroupVersion.Group,
		Version: apixv1beta1.SchemeGroupVersion.Version,
		Kind:    "CustomResourceDefinition",
	}
)

var knownApplyUpdateGroups = func() map[string]struct{} {
	return map[string]struct{}{
		clusterv2.GroupVersion.Group: {},
	}
}()

// isApplyUpdate checks if object should be updated for apply operation
func isApplyUpdate(o runtime.Object) bool {
	group := o.GetObjectKind().GroupVersionKind().Group
	if _, ok := knownApplyUpdateGroups[group]; ok {
		return true
	}
	return false
}

// Options are the options for apply operation
type Options struct {
	// if UseUpdate is set, then update is used instead of patch
	UseUpdate bool
	// if DontCreate is set, then object is not created if it is not present;
	// it is only updated/patched when present
	DontCreate bool
}

// Option is the functional apply options
type Option func(*Options)

// WithUseUpdate sets if update should be used instead of patch for apply
// operation
func WithUseUpdate(o runtime.Object) Option {
	return func(opts *Options) {
		opts.UseUpdate = isApplyUpdate(o)
	}
}

// WithForceUseUpdate sets if update should be used instead of patch for apply
// operation, irrespective of whether the object is of paralus domain or not
func WithForceUseUpdate() Option {
	return func(opts *Options) {
		opts.UseUpdate = true
	}
}

// WithDontCreate sets DontCreate flag
func WithDontCreate() Option {
	return func(opts *Options) {
		opts.DontCreate = true
	}
}

// Applier is the interface for applying patch to runtime objects
type Applier interface {
	Apply(ctx context.Context, obj ctrlclient.Object, opts ...Option) error
	ApplyStatus(ctx context.Context, obj ctrlclient.Object, statusObj interface{}) error
	ctrlclient.Client
}

type applier struct {
	dynamic bool
	ctrlclient.Client
}

// NewApplier returns new applier
func NewApplier(client ctrlclient.Client) Applier {
	return &applier{false, client}
}

// NewDynamicApplier returns a new applier whose client is dynamically refreshed
// when new CRDs are installed
func NewDynamicApplier() (Applier, error) {
	c, err := client.New()
	if err != nil {
		return nil, err
	}

	return &applier{true, c}, nil
}

func isCRD(gvk schema.GroupVersionKind) bool {
	//applyLog.Info("is crd", "gvk", gvk)
	switch gvk {
	case crdv1beta1GVK:
		return true
	}
	return false
}

func (a *applier) Apply(ctx context.Context, obj ctrlclient.Object, opts ...Option) error {
	var applyOpts = new(Options)
	for _, f := range opts {
		f(applyOpts)
	}

	// added to preserve backward compatibility with other code
	if !applyOpts.UseUpdate {
		applyOpts.UseUpdate = isApplyUpdate(obj)
	}

	gvk := obj.GetObjectKind().GroupVersionKind()
	log := applyLog.WithValues("gvk", gvk)

	var objectKey ctrlclient.ObjectKey
	var current ctrlclient.Object
	var err error

	if mo, ok := obj.(metav1.Object); ok {

		objectKey = ctrlclient.ObjectKey{
			Name:      mo.GetName(),
			Namespace: mo.GetNamespace(),
		}

	}

	gvk, err = GetGVK(obj)
	if err != nil {
		return err
	}

	current, err = util.NewObject(gvk)

	if err != nil {
		err = fmt.Errorf("unable to create new object %s", err.Error())
		return err
	}

	//refresh client before applying a unknow object
	if !util.KnownObject(gvk) && a.dynamic {
		c, err := client.New()
		if err != nil {
			log.Info("error in creating the refreshed client ", "err", err)
			err = fmt.Errorf("unable to create new client for dynamic applier %s", err.Error())
			return err
		}
		a.Client = c
	}

	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		err := a.Get(ctx, objectKey, current)
		if err != nil {
			if apierrs.IsNotFound(err) {
				// if don't create flag is set and object is not found
				if applyOpts.DontCreate {
					return err
				}

				err = a.Create(ctx, obj)
				if err != nil {
					err = fmt.Errorf("unable to create step object %s", err.Error())
					return err
				}

				if a.dynamic && isCRD(gvk) {

					// wait until the crds are sync
					// TODO : what happens when you get an error ???
					err = a.pollCRDUntilEstablished(ctx, 180*time.Second, obj, objectKey)
					if err != nil {
						log.Info("error in polling ", "err", err)
						return nil
					}

					log.Info("crd created, refreshing client")
					a.Client, err = client.New()
					if err != nil {
						log.Info("error in creating the refreshed client ", "err", err)
						return err
					}
					log.Info("crd created, refreshed client")
				}
				return nil
			}
			err = fmt.Errorf("unable to get step object %s", err.Error())
			return err
		}

		current.GetObjectKind().SetGroupVersionKind(gvk)

		if applyOpts.UseUpdate {
			err = updateObject(current, obj)
			if err != nil {
				return err
			}
			err = a.Update(ctx, current)
			if err != nil {
				return err
			}
		} else {
			err = a.Patch(ctx, obj, NewPatch(current))
			if err != nil {
				err = fmt.Errorf("unable to patch step object %s", err.Error())
				return err

			}
		}

		obj.GetObjectKind().SetGroupVersionKind(gvk)

		return nil
	})
}

func (a *applier) pollCRDUntilEstablished(ctx context.Context, timeout time.Duration, obj ctrlclient.Object, objectKey types.NamespacedName) error {
	return wait.PollImmediate(time.Second, timeout, func() (bool, error) {

		crd := &apixv1beta1.CustomResourceDefinition{}
		err := scheme.Scheme.Convert(obj, crd, nil)
		if err != nil {
			return false, fmt.Errorf("unable to convert to CRD type: %v", err)
		}

		err = a.Get(ctx, objectKey, obj)
		if err != nil {
			applyLog.Info("error in getting the object", "err", err)
		}

		applyLog.Info("checking crd status ", "name", crd.Spec.Names, "crd status", crd.Status)
		for _, cond := range crd.Status.Conditions {
			switch cond.Type {
			case apixv1beta1.Established:
				if cond.Status == apixv1beta1.ConditionTrue {
					return true, nil
				}
			case apixv1beta1.NamesAccepted:
				if cond.Status == apixv1beta1.ConditionFalse {
					return false, fmt.Errorf("naming conflict detected for CRD %s", crd.GetName())
				}
			}
		}

		return false, nil
	})
}

func getGVKIfNotFound(obj runtime.Object) (schema.GroupVersionKind, error) {
	currentGVK := obj.GetObjectKind().GroupVersionKind()
	formedGVK := schema.GroupVersionKind{}

	kind := currentGVK.Kind
	if len(kind) == 0 {
		gvks, _, err := scheme.Scheme.ObjectKinds(obj)
		if err != nil {
			return formedGVK, err
		}
		kind = gvks[0].Kind
	}

	var listMeta metav1.Common
	objectMeta, err := meta.Accessor(obj)
	if err != nil {
		listMeta, err = meta.CommonAccessor(obj)
		if err != nil {
			return formedGVK, err
		}
	} else {
		listMeta = objectMeta
	}

	version := currentGVK.GroupVersion().String()
	if len(version) == 0 {
		selfLink := listMeta.GetSelfLink()
		if len(selfLink) == 0 {
			return formedGVK, ErrNoSelfLink
		}
		selfLinkURL, err := url.Parse(selfLink)
		if err != nil {
			return formedGVK, err
		}
		// example paths: /<prefix>/<version>/*
		parts := strings.Split(selfLinkURL.Path, "/")
		if len(parts) < 3 {
			return formedGVK, fmt.Errorf("unexpected self link format: '%v'; got version '%v'", selfLink, version)
		}
		version = parts[2]
	}

	formedGVK.Kind = kind
	formedGVK.Version = version

	return formedGVK, nil
}

func (a *applier) ApplyStatus(ctx context.Context, obj ctrlclient.Object, statusObj interface{}) error {
	var objectKey ctrlclient.ObjectKey
	var original ctrlclient.Object
	var err error
	if mo, ok := obj.(metav1.Object); ok {
		objectKey = ctrlclient.ObjectKey{
			Name:      mo.GetName(),
			Namespace: mo.GetNamespace(),
		}
	}

	gvk, err := GetGVK(obj)
	if err != nil {
		return err
	}

	original, err = util.NewObject(gvk)

	if err != nil {
		return err
	}

	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		err := a.Get(ctx, objectKey, original)
		if err != nil {
			return err
		}
		return a.Status().Patch(ctx, obj, NewStatus(original, statusObj))
	})
}

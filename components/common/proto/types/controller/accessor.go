package controller

import (
	"errors"
	"fmt"
	"strings"

	"github.com/valyala/fastjson"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	accessorLog = ctrl.Log.WithName("accessor")
)

const (
	apiVersion        = "apiVersion"
	kind              = "kind"
	metadata          = "metadata"
	name              = "name"
	namespace         = "namespace"
	labels            = "labels"
	annotations       = "annotations"
	creationTimestamp = "creationTimestamp"
	generateName      = "generateName"
	ownerReferences   = "ownerReferences"
	resourceVersion   = "resourceVersion"
	selfLink          = "selfLink"
	uid               = "uid"
	status            = "status"
	emptyObject       = "{}"
	emptyArray        = "[]"
)

var (
	// ErrNoMetadata is returned when metadata field is not present in the object
	ErrNoMetadata = errors.New("metadata filed not present")
)

// +kubebuilder:object:generate=false

// Accessor is the interface for accessing k8s fields from step object
type Accessor interface {
	// Kind returns the kind of the step object
	Kind() (string, error)
	// SetKind sets kind for step object
	SetKind(string) error
	// APIVersion returns api version for the step object
	APIVersion() (string, error)
	// SetAPIVersion sets api version for step object
	SetAPIVersion(string) error

	// GroupVersionKind returns group version kind of the step object
	GroupVersionKind() (schema.GroupVersionKind, error)

	// Namespace returns namespace of the step object
	Namespace() (string, error)
	// SetNamespace sets namespace for step object
	SetNamespace(string) error
	// Name returns name of the step object
	Name() (string, error)
	// SetName sets name for the step object
	SetName(string) error
	// Labels returns labels of the step object
	Labels() (map[string]string, error)
	// SetLabels sets labels for the step object
	SetLabels(map[string]string) error
	// Annotations returns annotations of the step object
	Annotations() (map[string]string, error)
	// SetAnnotations sets annotations for the step object
	SetAnnotations(map[string]string) error
	// SetRaw sets json serialized bytes value at keys
	SetRaw(value []byte, keys ...string) error
	// GetRaw gets json serialized bytes value at keys
	GetRaw(keys ...string) ([]byte, error)
	// ResetAutoFields resets fields auto populated by the server in the step object
	ResetAutoFields() error
	// Bytes returns the mutated step object
	Bytes() []byte
}

// +kubebuilder:object:generate=false

type accessor struct {
	*fastjson.Value
}

// +kubebuilder:object:generate=false

func newAccessor(b []byte) (Accessor, error) {
	v, err := fastjson.ParseBytes(b)
	if err != nil {
		return nil, err
	}
	return &accessor{v}, nil
}

var _ Accessor = (*accessor)(nil)

func ensureKeys(obj *fastjson.Object, keys ...string) error {
	var err error
	if len(keys) == 0 {
		return nil
	}

	if obj.Get(keys[0]) == nil || obj.Get(keys[0]).Type() == fastjson.TypeNull {
		obj.Set(keys[0], fastjson.MustParse(emptyObject))
	}

	obj, err = obj.Get(keys[0]).Object()
	if err != nil {
		return err
	}

	return ensureKeys(obj, keys[1:]...)
}

func ensureKeysWithDefaultValue(obj *fastjson.Object, value *fastjson.Value, keys ...string) error {
	var err error
	if obj.Get(keys[0]) == nil || obj.Get(keys[0]).Type() == fastjson.TypeNull {
		//If this is the last leg of the keys, set the default value (which would be either Array or Object
		if len(keys) == 1 {
			obj.Set(keys[0], value)
		} else {
			//For any node above the leaf node, we must create an empty object only.
			//Currently we don't support setting raw on a path which has an array in the middle of the path
			obj.Set(keys[0], fastjson.MustParse(emptyObject))
		}
	}
	if len(keys) == 1 {
		return nil
	}
	obj, err = obj.Get(keys[0]).Object()
	if err != nil {
		return err
	}
	return ensureKeysWithDefaultValue(obj, value, keys[1:]...)
}

func (a *accessor) ensureKeys(keys ...string) error {
	return a.ensureKeysWithDefaultValue(fastjson.MustParse(emptyObject), keys...)
}

func (a *accessor) ensureKeysWithDefaultValue(value *fastjson.Value, keys ...string) error {
	o, err := a.Object()
	if err != nil {
		return err
	}
	err = ensureKeysWithDefaultValue(o, value, keys...)
	return err
}

func (a *accessor) Kind() (string, error) {
	if ok := a.Exists(kind); ok {
		v, err := a.Get(kind).StringBytes()
		if err != nil {
			return "", err
		}
		return string(v), nil
	}
	return "", fmt.Errorf("%s not found", kind)
}

func (a *accessor) SetKind(k string) error {
	sb := new(strings.Builder)
	sb.WriteString(`"`)
	sb.WriteString(k)
	sb.WriteString(`"`)
	v, err := fastjson.Parse(sb.String())
	if err != nil {
		return err
	}

	a.Set(kind, v)
	return nil
}

func (a *accessor) APIVersion() (string, error) {
	if ok := a.Exists(apiVersion); ok {
		v, err := a.Get(apiVersion).StringBytes()
		if err != nil {
			return "", err
		}
		return string(v), nil
	}
	return "", fmt.Errorf("%s not found", apiVersion)
}

func (a *accessor) SetAPIVersion(k string) error {
	sb := new(strings.Builder)
	sb.WriteString(`"`)
	sb.WriteString(k)
	sb.WriteString(`"`)
	v, err := fastjson.Parse(sb.String())
	if err != nil {
		return err
	}

	a.Set(apiVersion, v)
	return nil
}

func (a *accessor) GroupVersionKind() (gvk schema.GroupVersionKind, err error) {
	var kind, apiVersion string
	var gv schema.GroupVersion
	kind, err = a.Kind()
	if err != nil {
		return
	}
	apiVersion, err = a.APIVersion()
	if err != nil {
		return
	}

	gv, err = schema.ParseGroupVersion(apiVersion)
	if err != nil {
		return
	}

	gvk.Kind = kind
	gvk.Group = gv.Group
	gvk.Version = gv.Version

	return
}

func (a *accessor) Namespace() (string, error) {
	if ok := a.Exists(metadata, namespace); ok {
		v, err := a.Get(metadata, namespace).StringBytes()
		if err != nil {
			return "", err
		}
		return string(v), nil
	}
	return "", nil
}

func (a *accessor) SetNamespace(k string) error {
	err := a.ensureKeys(metadata)
	if err != nil {
		return err
	}
	sb := new(strings.Builder)
	sb.WriteString(`"`)
	sb.WriteString(k)
	sb.WriteString(`"`)
	v, err := fastjson.Parse(sb.String())
	if err != nil {
		return err
	}

	o, err := a.Get(metadata).Object()
	if err != nil {
		return err
	}

	o.Set(namespace, v)
	return nil
}

func (a *accessor) Name() (string, error) {
	if ok := a.Exists(metadata, name); ok {
		v, err := a.Get(metadata, name).StringBytes()
		if err != nil {
			return "", err
		}
		return string(v), nil
	}
	return "", nil
}

func (a *accessor) SetName(k string) error {
	err := a.ensureKeys(metadata)
	if err != nil {
		return err
	}
	sb := new(strings.Builder)
	sb.WriteString(`"`)
	sb.WriteString(k)
	sb.WriteString(`"`)
	v, err := fastjson.Parse(sb.String())
	if err != nil {
		return err
	}

	o, err := a.Get(metadata).Object()
	if err != nil {
		return err
	}

	o.Set(name, v)
	return nil
}

func (a *accessor) Labels() (map[string]string, error) {
	if ok := a.Exists(metadata, labels); ok {

		if a.Get(metadata, labels).Type() == fastjson.TypeNull {
			return map[string]string{}, nil
		}

		o, err := a.Get(metadata, labels).Object()
		if err != nil {
			return nil, err
		}

		lbls := make(map[string]string)

		o.Visit(func(k []byte, v *fastjson.Value) {
			vb, err := v.StringBytes()
			if err != nil {
				return
			}

			lbls[string(k)] = string(vb)
		})
		return lbls, nil
	}
	return map[string]string{}, nil
}

func (a *accessor) SetLabels(lbls map[string]string) error {
	err := a.ensureKeys(metadata, labels)
	if err != nil {
		return err
	}

	elbls, err := a.Get(metadata, labels).Object()
	if err != nil {
		return err
	}

	for k, v := range lbls {
		sb := new(strings.Builder)
		sb.WriteString(`"`)
		sb.WriteString(v)
		sb.WriteString(`"`)
		jv, err := fastjson.Parse(sb.String())
		if err != nil {
			return err
		}
		elbls.Set(k, jv)
	}

	return nil
}

func (a *accessor) Annotations() (map[string]string, error) {
	if ok := a.Exists(metadata, annotations); ok {
		if a.Get(metadata, annotations).Type() == fastjson.TypeNull {
			return map[string]string{}, nil
		}

		o, err := a.Get(metadata, annotations).Object()
		if err != nil {
			return nil, err
		}

		ants := make(map[string]string)

		o.Visit(func(k []byte, v *fastjson.Value) {
			vb, err := v.StringBytes()
			if err != nil {
				return
			}
			ants[string(k)] = string(vb)
		})
		return ants, nil
	}
	return map[string]string{}, nil
}

func (a *accessor) SetAnnotations(ants map[string]string) error {
	err := a.ensureKeys(metadata, annotations)
	if err != nil {
		return err
	}

	eants, err := a.Get(metadata, annotations).Object()
	if err != nil {
		return err
	}

	for k, v := range ants {

		sb := new(strings.Builder)
		sb.WriteString(`"`)
		sb.WriteString(v)
		sb.WriteString(`"`)
		jv, err := fastjson.Parse(sb.String())
		if err != nil {
			return err
		}
		eants.Set(k, jv)
	}

	return nil
}

func (a *accessor) ResetAutoFields() error {
	if a.Get(metadata) == nil {
		return ErrNoMetadata
	}
	o, err := a.Get(metadata).Object()
	if err != nil {
		return err
	}
	o.Del(creationTimestamp)
	o.Del(generateName)
	o.Del(ownerReferences)
	o.Del(resourceVersion)
	o.Del(selfLink)
	o.Del(uid)

	a.Del(status)

	return nil
}

func (a *accessor) SetRaw(value []byte, keys ...string) error {
	v, err := fastjson.ParseBytes(value)
	if err != nil {
		return err
	}
	var defaultValue *fastjson.Value

	switch v.Type() {
	case fastjson.TypeArray:
		defaultValue = fastjson.MustParse(emptyArray)
	default:
		defaultValue = fastjson.MustParse(emptyObject)
	}

	err = a.ensureKeysWithDefaultValue(defaultValue, keys...)
	if err != nil {
		return err
	}

	o, err := a.Get(keys[:len(keys)-1]...).Object()
	if err != nil {
		return err
	}

	// if the value of the last key is an object
	if eo, err := o.Get(keys[len(keys)-1]).Object(); err == nil {
		// and the value to be set is also an object
		if vo, err := v.Object(); err == nil {
			// merge the value k v
			vo.Visit(func(k []byte, v *fastjson.Value) {
				eo.Set(string(k), v)
			})
		} else {
			// otherwise set the value
			eo.Set(keys[len(keys)-1], v)
		}
	} else {
		// otherwise set the last key and value on parent
		o.Set(keys[len(keys)-1], v)
	}

	return nil
}

func (a *accessor) GetRaw(keys ...string) ([]byte, error) {
	v := a.Get(keys...)
	if v != nil {
		return v.MarshalTo(nil), nil
	}
	return nil, fmt.Errorf("not found %s", strings.Join(keys, "."))
}

func (a *accessor) Bytes() []byte {
	return a.MarshalTo(nil)
}

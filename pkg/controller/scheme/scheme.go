package scheme

import (
	"sync"

	apiv2 "github.com/paralus/paralus/proto/types/controller"
	// DO NOT UPDATE
	// API Extensions v1 is not available in k8s v1.14.x
	apixv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	kjson "k8s.io/apimachinery/pkg/runtime/serializer/json"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
)

type addToScheme func(s *runtime.Scheme) error

var (
	// Scheme is the runtime scheme
	Scheme *runtime.Scheme

	// Serializer is the JSON serializer for handling runtime objects
	Serializer runtime.Serializer
)

func init() {

	var once sync.Once

	once.Do(func() {
		Scheme = runtime.NewScheme()

		for _, f := range []addToScheme{
			clientgoscheme.AddToScheme,
			apixv1beta1.AddToScheme,
			apiv2.AddToScheme,
		} {
			err := f(Scheme)
			if err != nil {
				panic(err)
			}
		}

		Serializer = kjson.NewSerializer(kjson.DefaultMetaFactory, Scheme, Scheme, false)
	})

}

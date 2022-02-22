package tail

import (
	"testing"

	"github.com/julienschmidt/httprouter"
)

// CASE 1 /api
// CASE 2 /apis
// CASE 3 /apis/storage.k8s.io/v1beta1
// CASE 4 /api/v1/namespaces
// CASE 5 /api/v1/pods
// CASE 6 /api/v1/namespaces/rafay-system
// CASE 7 /api/v1/namespaces/rafay-system/configmaps/relay-agent-config
// CASE 8 /apis/k3s.cattle.io/v1/namespaces/kube-system/addons

func TestURL(t *testing.T) {
	r := httprouter.New()

	r.Handle("GET", "/api", _dummyHandler)
	r.Handle("GET", "/api/:version", _dummyHandler)
	r.Handle("GET", "/api/:version/:kind0", _dummyHandler)
	r.Handle("GET", "/api/:version/:kind0/:namespace", _dummyHandler)
	r.Handle("GET", "/api/:version/:kind0/:namespace/:kind1", _dummyHandler)
	r.Handle("GET", "/api/:version/:kind0/:namespace/:kind1/:name", _dummyHandler)
	r.Handle("GET", "/apis", _dummyHandler)
	r.Handle("GET", "/apis/:group/:version", _dummyHandler)
	r.Handle("GET", "/apis/:group/:version/:kind0", _dummyHandler)
	r.Handle("GET", "/apis/:group/:version/:kind0/:namespace", _dummyHandler)
	r.Handle("GET", "/apis/:group/:version/:kind0/:namespace/:kind1", _dummyHandler)
	r.Handle("GET", "/apis/:group/:version/:kind0/:namespace/:kind1/:name", _dummyHandler)

	if h, params, _ := r.Lookup("GET", "/api/v1/pods"); h != nil {
		t.Log(params)
	} else {
		t.Error("expected match")
		return
	}

	if h, params, _ := r.Lookup("GET", "/apis/storage.k8s.io/v1beta1"); h != nil {
		t.Log(params)
	} else {
		t.Error("expected match")
		return
	}

	if h, params, _ := r.Lookup("GET", "/api/v1/namespaces/rafay-system/configmaps/relay-agent-config"); h != nil {
		t.Log(params)
	} else {
		t.Error("expected match")
		return
	}

	if h, params, _ := r.Lookup("GET", "/apis/k3s.cattle.io/v1/namespaces/kube-system/addons"); h != nil {
		t.Log(params)
	} else {
		t.Error("expected match")
		return
	}

	if h, params, _ := r.Lookup("GET", "/apis/metrics.k8s.io/v1beta1"); h != nil {
		t.Log(params)
	} else {
		t.Error("expected match")
		return
	}

	if h, params, _ := r.Lookup("GET", "/api/v1/namespaces/test-ns"); h != nil {
		t.Log(params)
	} else {
		t.Error("expected match")
		return
	}

}

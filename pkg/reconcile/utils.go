package reconcile

import (
	"encoding/json"

	"github.com/paralus/paralus/pkg/event"
)

func resourceToKey(r event.Resource) string {
	b, _ := json.Marshal(r)
	return string(b)
}

func keyToResource(k string) event.Resource {
	var r event.Resource
	json.Unmarshal([]byte(k), &r)
	return r
}

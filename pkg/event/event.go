package event

import (
	"github.com/segmentio/encoding/json"
)

// ResourceEventType is type of resource event
type ResourceEventType int

// Resource event types
const (
	EventTypeNotSet ResourceEventType = iota
	ResourceCreate
	ResourceUpdate
	ResourceDelete
	ResourceUpdateStatus
)

// Resource represents an event generated when resource changes
// it should be looked as an parameter to the event handler callback
type Resource struct {
	PartnerID      string            `json:"pa,omitempty"`
	OrganizationID string            `json:"or,omitempty"`
	ProjectID      string            `json:"pr,omitempty"`
	ID             string            `json:"id,omitempty"`
	Name           string            `json:"n,omitempty"`
	EventType      ResourceEventType `json:"t,omitempty"`
	Username       string            `json:"un,omitempty"`
	Account        string            `json:"acc,omitempty"`
}

// Key is the key for this event which can be used as a cache key etc
func (sr Resource) Key() string {
	b, _ := json.Marshal(&sr)
	return string(b)
}

// Handler is the interface for notifying resource changes
type Handler interface {
	OnChange(r Resource)
}

// HandlerFuncs is a utility for creating event handler with functions
type HandlerFuncs struct {
	OnChangeFunc func(r Resource)
	NameFunc     func() string
}

// OnChange is the callback for notifying when resource is changed
func (f HandlerFuncs) OnChange(r Resource) {
	if f.OnChangeFunc != nil {
		f.OnChangeFunc(r)
	}
}

var _ Handler = (*HandlerFuncs)(nil)

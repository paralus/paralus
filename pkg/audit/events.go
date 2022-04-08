package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	logv2 "github.com/RafayLabs/rcloud-base/pkg/log"
	commonv3 "github.com/RafayLabs/rcloud-base/proto/types/commonpb/v3"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
)

var (
	_log = logv2.GetLogger()
)

type (
	// EventVersion is the version of event
	EventVersion string
	// EventOrigin is the origin of the event
	EventOrigin string
	// EventCategory is the category of the event
	EventCategory string
	// EventTopic is the topic to which event has to be published
	EventTopic string
)

// Audit events constants
const (
	RawLogsEventsTopic EventTopic    = "dp-raw-logs-events"
	VersionV1          EventVersion  = "1.0"
	OriginCore         EventOrigin   = "core"
	OriginCluster      EventOrigin   = "cluster"
	AuditCategory      EventCategory = "AUDIT"
)

// EventActorAccount Event's initiator account
type EventActorAccount struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

// EventActor Event's initiator
type EventActor struct {
	Type           string            `json:"type"`
	PartnerID      string            `json:"partner_id"`
	OrganizationID string            `json:"organization_id"`
	Account        EventActorAccount `json:"account"`
	Groups         []string          `json:"groups"`
}

// EventClient Event's client
type EventClient struct {
	Type      string `json:"type"`
	IP        string `json:"ip"`
	UserAgent string `json:"user_agent"`
	Host      string `json:"host"`
}

// EventDetail Event's detail
type EventDetail struct {
	Message string            `json:"message"`
	Meta    map[string]string `json:"meta"`
}

// Event is struct to hold event data
type Event struct {
	Version        EventVersion  `json:"version"`
	Category       EventCategory `json:"category"`
	Origin         EventOrigin   `json:"origin"`
	Portal         string        `json:"portal"`
	Type           string        `json:"type"`
	PartnerID      string        `json:"partner_id"`
	OrganizationID string        `json:"organization_id"`
	ProjectID      string        `json:"project_id"`
	Actor          *EventActor   `json:"actor"`
	Client         *EventClient  `json:"client"`
	Detail         *EventDetail  `json:"detail"`
	Timestamp      string        `json:"timestamp"`
}

type createEventOptions struct {
	version        EventVersion
	origin         EventOrigin
	category       EventCategory
	topic          EventTopic
	partnerID      string
	organizationID string
	projectID      string
	ctx            context.Context
	accountID      string
	username       string
	groups         []string
}

// WithVersion sets version for audit event
func WithVersion(version EventVersion) CreateEventOption {
	return func(opts *createEventOptions) {
		opts.version = version
	}
}

// WithOrigin sets origin for audit event
func WithOrigin(origin EventOrigin) CreateEventOption {
	return func(opts *createEventOptions) {
		opts.origin = origin
	}
}

// WithCategory sets category for audit event
func WithCategory(category EventCategory) CreateEventOption {
	return func(opts *createEventOptions) {
		opts.category = category
	}
}

// WithTopic sets topic for audit event
func WithTopic(topic EventTopic) CreateEventOption {
	return func(opts *createEventOptions) {
		opts.topic = topic
	}
}

// WithPartnerID sets partner id for audit event
func WithPartnerID(partnerID string) CreateEventOption {
	return func(opts *createEventOptions) {
		opts.partnerID = partnerID
	}
}

// WithOrganizationID sets organization id for audit event
func WithOrganizationID(organizationID string) CreateEventOption {
	return func(opts *createEventOptions) {
		opts.organizationID = organizationID
	}
}

// WithProjectID sets project id for audit event
func WithProjectID(projectID string) CreateEventOption {
	return func(opts *createEventOptions) {
		opts.projectID = projectID
	}
}

// WithContext sets context for audit event
func WithContext(ctx context.Context) CreateEventOption {
	return func(opts *createEventOptions) {
		opts.ctx = ctx
	}
}

// WithAccountID sets account id for audit event
func WithAccountID(accountID string) CreateEventOption {
	return func(opts *createEventOptions) {
		opts.accountID = accountID
	}
}

// WithUsername sets username for audit event
func WithUsername(username string) CreateEventOption {
	return func(opts *createEventOptions) {
		opts.username = username
	}
}

// WithGroups sets groups for audit event
func WithGroups(groups []string) CreateEventOption {
	return func(opts *createEventOptions) {
		opts.groups = groups
	}
}

// CreateEventOption is the functional arg for creating audit event
type CreateEventOption func(opts *createEventOptions)

// CreateEvent creates an event
func CreateEvent(event *Event, opts ...CreateEventOption) error {

	cOpts := createEventOptions{}
	for _, opt := range opts {
		opt(&cOpts)
	}

	t := time.Now()
	dateArray := strings.Fields(t.String())
	timestamp := fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02d.%06d%s",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), dateArray[2])
	event.Timestamp = timestamp

	event.Version = cOpts.version
	event.Category = cOpts.category
	event.Origin = cOpts.origin

	event.PartnerID = cOpts.partnerID
	event.OrganizationID = cOpts.organizationID
	event.ProjectID = cOpts.projectID

	if event.Client == nil {
		event.Client = getEventClientFromContext(cOpts.ctx)
	}

	if event.Actor == nil {
		event.Actor = getActor(cOpts)
	}

	payload, err := json.Marshal(event)
	if err != nil {
		_log.Infow("unable to marshal audit event", "error", err)
		return err
	}
	fmt.Println("event:", string(payload)) // TODO: Switch to writing to audit file
	return nil
}

func getEventClientFromContext(ctx context.Context) *EventClient {
	if ctx == nil {
		return nil
	}
	md, _ := metadata.FromIncomingContext(ctx)
	var ua, h, ip, requestType string
	userAgents := md.Get("grpcgateway-user-agent")
	if len(userAgents) > 0 {
		ua = userAgents[0]
		if strings.HasPrefix(ua, "RCTL") {
			requestType = "CLI"
		} else {
			requestType = "BROWSER"
		}
	} else {
		ua = "-"          // Not available
		requestType = "-" // Not available
	}
	hosts := md.Get("x-forwarded-host")
	if len(hosts) > 0 {
		h = hosts[0]
	} else {
		h = "-" // Not available
	}
	ips := md.Get("x-forwarded-for")
	if len(ips) > 0 {
		ip = ips[0]
	} else {
		ip = "-" // Not available
	}

	return &EventClient{
		Type:      requestType,
		UserAgent: ua,
		Host:      h,
		IP:        ip,
	}
}

func getActor(cOpts createEventOptions) *EventActor {
	account := EventActorAccount{
		ID:       cOpts.accountID,
		Username: cOpts.username,
	}
	return &EventActor{
		Type:           "USER",
		PartnerID:      cOpts.partnerID,
		OrganizationID: cOpts.organizationID,
		Account:        account,
		Groups:         cOpts.groups,
	}
}

func GetActorFromSessionData(sd *commonv3.SessionData) *EventActor {
	pid := sd.GetPartner() // TODO: have this pulled from headers
	oid := sd.GetOrganization()
	accountID := sd.GetAccount()
	username := sd.GetUsername()
	account := EventActorAccount{
		ID:       accountID,
		Username: username,
	}
	groups := sd.Groups // TODO: get groups (in interceptor?)

	// Set org id to string "null" for users with PARTNER_ADMIN role
	if oid == "" {
		oid = "null"
	}
	if pid == "" {
		pid = "null"
	}

	return &EventActor{
		Type:           "USER",
		PartnerID:      pid,
		OrganizationID: oid,
		Account:        account,
		Groups:         groups,
	}
}

func GetClientFromRequest(r *http.Request) *EventClient {
	return &EventClient{
		Type:      "BROWSER",
		IP:        r.Header.Get("X-Forwarded-For"),
		UserAgent: r.UserAgent(),
		Host:      r.Host,
	}
}

func GetClientFromSessionData(sd *commonv3.SessionData) *EventClient {
	return &EventClient{
		Type:      "BROWSER",
		IP:        sd.GetClientIp(),
		UserAgent: sd.GetClientUa(),
		Host:      sd.GetClientHost(),
	}
}

func GetEvent(r *http.Request, sd *commonv3.SessionData, detail *EventDetail, eventType string, projectID string) *Event {
	event := &Event{
		Actor:     GetActorFromSessionData(sd),
		Client:    GetClientFromRequest(r),
		Detail:    detail,
		Type:      eventType,
		Portal:    "OPS",
		ProjectID: projectID,
	}

	return event
}

func CreateV1Event(al *zap.Logger, sd *commonv3.SessionData, detail *EventDetail, eventType string, projectID string) error {
	actor := GetActorFromSessionData(sd)
	client := GetClientFromSessionData(sd)

	if projectID == "" {
		projectID = "null"
	}

	event := &Event{
		Version:   VersionV1,
		Category:  AuditCategory,
		Origin:    OriginCore,
		Actor:     actor,
		Client:    client,
		Detail:    detail,
		Type:      eventType,
		Portal:    "OPS", // TODO: What is the portal?
		ProjectID: projectID,
	}

	event.PartnerID = actor.PartnerID
	event.OrganizationID = actor.OrganizationID

	go WriteEvent(event, al)
	return nil
}

func WriteEvent(event *Event, al *zap.Logger) {
	al.Info(
		"audit",
		zap.String("version", string(event.Version)),
		zap.String("category", string(event.Category)),
		zap.String("origin", string(event.Origin)),
		zap.Reflect("actor", event.Actor),
		zap.Reflect("client", event.Client),
		zap.Reflect("detail", event.Detail),
		zap.String("type", event.Type),
		zap.String("portal", event.Portal),
		zap.String("project_id", event.ProjectID),
		zap.String("partner_id", event.PartnerID),
		zap.String("organization_id", event.OrganizationID),
	)
}

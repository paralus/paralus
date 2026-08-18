package audit

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	commonv3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
	"google.golang.org/grpc/metadata"
)

func TestCreateEvent(t *testing.T) {
	// CreateEvent mutates `event` synchronously (Version/Category/Origin/
	// Project/Client/Actor) before handing off to `go WriteEvent(...)`, so
	// these fields can be asserted immediately without waiting on the
	// goroutine.
	t.Run("populates fields from options and defaults Client/Actor", func(t *testing.T) {
		al := zap.NewNop()
		event := &Event{}

		err := CreateEvent(al, event,
			WithVersion(VersionV1),
			WithCategory(AuditCategory),
			WithOrigin(OriginCore),
			WithProject("proj-1"),
			WithUsername("bob"),
			WithGroups([]string{"team-a"}),
			WithTopic(RawLogsEventsTopic),
			WithAccountID("acct-1"),
		)
		require.NoError(t, err)

		assert.Equal(t, VersionV1, event.Version)
		assert.Equal(t, AuditCategory, event.Category)
		assert.Equal(t, OriginCore, event.Origin)
		assert.Equal(t, "proj-1", event.Project)
		assert.NotEmpty(t, event.Timestamp)
		require.NotNil(t, event.Actor)
		assert.Equal(t, "bob", event.Actor.Account.Username)
		assert.Equal(t, []string{"team-a"}, event.Actor.Groups)
	})

	t.Run("does not overwrite an explicitly set Client or Actor", func(t *testing.T) {
		al := zap.NewNop()
		explicitClient := &EventClient{Type: "CUSTOM"}
		explicitActor := &EventActor{Type: "SERVICE"}
		event := &Event{Client: explicitClient, Actor: explicitActor}

		err := CreateEvent(al, event)
		require.NoError(t, err)
		assert.Same(t, explicitClient, event.Client)
		assert.Same(t, explicitActor, event.Actor)
	})

	t.Run("derives Client from grpc incoming metadata when context is set", func(t *testing.T) {
		al := zap.NewNop()
		md := metadata.New(map[string]string{
			"grpcgateway-user-agent": "RCTL/1.0",
			"x-forwarded-host":       "api.example.com",
			"x-forwarded-for":        "1.2.3.4",
		})
		ctx := metadata.NewIncomingContext(context.Background(), md)

		event := &Event{}
		err := CreateEvent(al, event, WithContext(ctx))
		require.NoError(t, err)

		require.NotNil(t, event.Client)
		assert.Equal(t, "CLI", event.Client.Type)
		assert.Equal(t, "RCTL/1.0", event.Client.UserAgent)
		assert.Equal(t, "api.example.com", event.Client.Host)
		assert.Equal(t, "1.2.3.4", event.Client.IP)
	})

	t.Run("BROWSER type when user agent does not start with RCTL", func(t *testing.T) {
		al := zap.NewNop()
		md := metadata.New(map[string]string{"grpcgateway-user-agent": "Mozilla/5.0"})
		ctx := metadata.NewIncomingContext(context.Background(), md)

		event := &Event{}
		err := CreateEvent(al, event, WithContext(ctx))
		require.NoError(t, err)
		assert.Equal(t, "BROWSER", event.Client.Type)
	})

	t.Run("placeholder values when metadata is absent but context is set", func(t *testing.T) {
		al := zap.NewNop()
		ctx := metadata.NewIncomingContext(context.Background(), metadata.New(nil))

		event := &Event{}
		err := CreateEvent(al, event, WithContext(ctx))
		require.NoError(t, err)
		assert.Equal(t, "-", event.Client.Type)
		assert.Equal(t, "-", event.Client.UserAgent)
		assert.Equal(t, "-", event.Client.Host)
		assert.Equal(t, "-", event.Client.IP)
	})

	t.Run("nil Client when no context option is given", func(t *testing.T) {
		al := zap.NewNop()
		event := &Event{}
		err := CreateEvent(al, event)
		require.NoError(t, err)
		assert.Nil(t, event.Client)
	})
}

func TestGetActorFromSessionData(t *testing.T) {
	sd := &commonv3.SessionData{Username: "alice", Groups: []string{"g1", "g2"}}
	actor := GetActorFromSessionData(sd)
	assert.Equal(t, "USER", actor.Type)
	assert.Equal(t, "alice", actor.Account.Username)
	assert.Equal(t, []string{"g1", "g2"}, actor.Groups)
}

func TestGetClientFromRequest(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "http://example.com/path", nil)
	r.Header.Set("X-Forwarded-For", "9.9.9.9")
	r.Header.Set("User-Agent", "test-agent")

	client := GetClientFromRequest(r)
	assert.Equal(t, "BROWSER", client.Type)
	assert.Equal(t, "9.9.9.9", client.IP)
	assert.Equal(t, "test-agent", client.UserAgent)
	assert.Equal(t, "example.com", client.Host)
}

func TestGetClientFromSessionData(t *testing.T) {
	sd := &commonv3.SessionData{ClientIp: "1.1.1.1", ClientUa: "ua-1", ClientHost: "host-1"}
	client := GetClientFromSessionData(sd)
	assert.Equal(t, "BROWSER", client.Type)
	assert.Equal(t, "1.1.1.1", client.IP)
	assert.Equal(t, "ua-1", client.UserAgent)
	assert.Equal(t, "host-1", client.Host)
}

func TestGetEvent(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "http://example.com/path", nil)
	sd := &commonv3.SessionData{Username: "alice"}
	detail := &EventDetail{Message: "did a thing"}

	event := GetEvent(r, sd, detail, "cluster.create", "proj-1")
	assert.Equal(t, "cluster.create", event.Type)
	assert.Equal(t, "proj-1", event.Project)
	assert.Equal(t, "OPS", event.Portal)
	assert.Same(t, detail, event.Detail)
	assert.Equal(t, "alice", event.Actor.Account.Username)
}

func TestCreateV1Event(t *testing.T) {
	al := zap.NewNop()
	sd := &commonv3.SessionData{Username: "alice"}
	detail := &EventDetail{Message: "did a thing"}

	err := CreateV1Event(al, sd, detail, "cluster.create", "proj-1")
	assert.NoError(t, err)
}

func TestWriteEvent(t *testing.T) {
	core, logs := observer.New(zap.InfoLevel)
	al := zap.New(core)

	event := &Event{
		Version:  VersionV1,
		Category: AuditCategory,
		Origin:   OriginCore,
		Type:     "cluster.create",
		Portal:   "OPS",
		Project:  "proj-1",
		Actor:    &EventActor{Type: "USER"},
		Client:   &EventClient{Type: "BROWSER"},
		Detail:   &EventDetail{Message: "hello"},
	}

	WriteEvent(event, al)

	require.Equal(t, 1, logs.Len())
	entry := logs.All()[0]
	assert.Equal(t, "audit", entry.Message)
	fields := entry.ContextMap()
	assert.Equal(t, "1.0", fields["version"])
	assert.Equal(t, "cluster.create", fields["type"])
	assert.Equal(t, "proj-1", fields["project"])
}

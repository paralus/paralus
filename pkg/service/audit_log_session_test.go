package service

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/paralus/paralus/pkg/common"
	commonv3 "github.com/paralus/paralus/proto/types/commonpb/v3"
)

// Issue 344: a normal/IdP user opening audit logs panic'd in
// ValidateUserAuditReadRequest via uuid.MustParse("") and crash-looped the pod.
func TestValidateUserAuditReadRequestEmptyPartnerDoesNotPanic(t *testing.T) {
	sd := &commonv3.SessionData{
		Account:      uuid.NewString(),
		Organization: uuid.NewString(),
		Partner:      "",
		Username:     "idp-user@example.com",
	}
	ctx := context.WithValue(context.Background(), common.SessionDataKey, sd)

	defer func() {
		if rec := recover(); rec != nil {
			t.Fatalf("ValidateUserAuditReadRequest panicked: %v", rec)
		}
	}()

	err := ValidateUserAuditReadRequest(ctx, []string{"demo"}, nil, false)
	if err == nil {
		t.Fatal("expected an error for an empty partner UUID, got nil")
	}
	if !strings.Contains(err.Error(), "partner") {
		t.Fatalf("error should name the partner field, got %v", err)
	}
}

func TestValidateUserAuditReadRequestInvalidAccountDoesNotPanic(t *testing.T) {
	sd := &commonv3.SessionData{
		Account:      "not-a-uuid",
		Organization: uuid.NewString(),
		Partner:      uuid.NewString(),
	}
	ctx := context.WithValue(context.Background(), common.SessionDataKey, sd)
	err := ValidateUserAuditReadRequest(ctx, nil, nil, false)
	if err == nil || !strings.Contains(err.Error(), "account") {
		t.Fatalf("expected account parse error, got %v", err)
	}
}

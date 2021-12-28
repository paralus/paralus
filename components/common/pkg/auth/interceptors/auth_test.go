package interceptors

import (
	"context"
	"testing"

	"google.golang.org/grpc/metadata"
)

func TestGetRSID(t *testing.T) {
	mapWithRsid := make(map[string]string)
	mapWithRsid[cookieHeader] = "csrftoken=FpkPDGFZ2sAfgB1N7qZuQQfv7swEZb7Xr0q4e9tHlZAncVbt5u1a3rCa93Q2fC5s; rsid=v7eylddj5p2fl0l8q5wb6kton1t3kby5; logo_link=; support_email=support@rafay.co"
	md := metadata.New(mapWithRsid)
	ctx := metadata.NewIncomingContext(context.Background(), md)
	rsid := getRsid(ctx)
	if rsid != "v7eylddj5p2fl0l8q5wb6kton1t3kby5" {
		t.Errorf("wrong rsid")
	}
}

package authv3

import "testing"

func TestMetadataString(t *testing.T) {
	cases := []struct {
		name string
		m    map[string]interface{}
		want string
	}{
		{
			name: "canonical keys",
			m:    map[string]interface{}{"Organization": "org-1", "Partner": "part-1"},
			want: "org-1",
		},
		{
			name: "lowercase upgrade SQL keys",
			m:    map[string]interface{}{"organization": "org-lower", "partner": "part-lower"},
			want: "org-lower",
		},
		{
			name: "canonical wins over lowercase",
			m:    map[string]interface{}{"Organization": "org-1", "organization": "org-lower"},
			want: "org-1",
		},
		{
			name: "empty canonical falls through",
			m:    map[string]interface{}{"Organization": "", "organization": "org-lower"},
			want: "org-lower",
		},
		{
			name: "missing",
			m:    map[string]interface{}{"other": "x"},
			want: "",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := metadataString(tc.m, "Organization", "organization")
			if got != tc.want {
				t.Fatalf("got %q want %q", got, tc.want)
			}
		})
	}

	if got := metadataString(map[string]interface{}{"partner": "p-lower"}, "Partner", "partner"); got != "p-lower" {
		t.Fatalf("partner: got %q", got)
	}
}

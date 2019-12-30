package main

import (
	whitelabel "github.com/fastmail/pobox-bulk-go"
	"github.com/google/go-cmp/cmp"
	"testing"
)

func TestFilterToDomain(t *testing.T) {
	tests := []struct {
		desc       string
		domain     string
		have, want whitelabel.Routes
	}{
		{
			desc:   "empty -> empty",
			domain: "n/a",
			have:   whitelabel.Routes{},
			want:   whitelabel.Routes{},
		},
		{
			desc:   "@example.org only",
			domain: "example.org",
			have: whitelabel.Routes{
				"a@example.org": &whitelabel.Route{},
				"b@example.org": &whitelabel.Route{},
			},
			want: whitelabel.Routes{
				"a@example.org": &whitelabel.Route{},
				"b@example.org": &whitelabel.Route{},
			},
		},
		{
			desc:   "multiple domains",
			domain: "example.org",
			have: whitelabel.Routes{
				"a@example.org":    &whitelabel.Route{},
				"b@notexample.org": &whitelabel.Route{},
			},
			want: whitelabel.Routes{
				"a@example.org": &whitelabel.Route{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got := filterToDomain(tt.have, tt.domain)
			if !cmp.Equal(got, tt.want) {
				t.Errorf("mismatch: -want, +got\n%s", cmp.Diff(tt.want, got))
			}
		})
	}
}

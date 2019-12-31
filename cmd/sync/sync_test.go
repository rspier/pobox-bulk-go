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

func TestComputeChanges(t *testing.T) {

	rA := &whitelabel.Route{Fwd: "a@a.com"}
	rB := &whitelabel.Route{Fwd: "b@b.com"}
	rC := &whitelabel.Route{Fwd: "c@c.com"}

	tests := []struct {
		desc     string
		cur, fut whitelabel.Routes
		want     whitelabel.Routes
	}{
		{
			desc: "empty",
			cur:  whitelabel.Routes{},
			fut:  whitelabel.Routes{},
			want: whitelabel.Routes{},
		},
		{
			desc: "delete",
			cur:  whitelabel.Routes{"a": rA},
			fut:  whitelabel.Routes{},
			want: whitelabel.Routes{"a": nil},
		},
		{
			desc: "delete one",
			cur:  whitelabel.Routes{"a": rA, "b": rB},
			fut:  whitelabel.Routes{"b": rB},
			want: whitelabel.Routes{"a": nil},
		},
		{
			desc: "add",
			cur:  whitelabel.Routes{},
			fut:  whitelabel.Routes{"a": rA},
			want: whitelabel.Routes{"a": rA},
		},
		{
			desc: "add one",
			cur:  whitelabel.Routes{"c": rC},
			fut:  whitelabel.Routes{"a": rA, "c": rC},
			want: whitelabel.Routes{"a": rA},
		},
		{
			desc: "change",
			cur:  whitelabel.Routes{"a": rA},
			fut:  whitelabel.Routes{"a": rB},
			want: whitelabel.Routes{"a": rB},
		},
		{
			desc: "change one",
			cur:  whitelabel.Routes{"a": rA, "c": rC},
			fut:  whitelabel.Routes{"a": rB, "c": rC},
			want: whitelabel.Routes{"a": rB},
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got := computeChanges(tt.cur, tt.fut)
			if !cmp.Equal(got, tt.want) {
				t.Errorf("mismatch: -want, +got\n%s", cmp.Diff(tt.want, got))
			}
		})
	}
}

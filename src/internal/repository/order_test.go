package repository

import (
	"errors"
	"slices"
	"strings"
	"testing"
)

func TestCheckID(t *testing.T) {
	for _, tc := range []struct {
		id   string
		want error
	}{
		{"a", nil},
		{strings.Repeat("x", MaxIDLength), nil},
		{strings.Repeat("ä", MaxIDLength), nil}, // counted in characters
		{"", ErrInvalid},
		{strings.Repeat("x", MaxIDLength+1), ErrInvalid},
	} {
		if err := CheckID("story", tc.id); !errors.Is(err, tc.want) || (tc.want == nil) != (err == nil) {
			t.Errorf("CheckID(%d characters) = %v, want %v", len([]rune(tc.id)), err, tc.want)
		}
	}
}

func TestMoveIDs(t *testing.T) {
	order := []string{"a", "b", "c", "d", "e"}
	for _, tc := range []struct {
		ids    []string
		before string
		want   []string
		err    error
	}{
		{ids: []string{"d"}, before: "a", want: []string{"d", "a", "b", "c", "e"}},
		{ids: []string{"a"}, before: "", want: []string{"b", "c", "d", "e", "a"}},
		{ids: []string{"e"}, before: "", want: []string{"a", "b", "c", "d", "e"}},
		{ids: []string{"a"}, before: "b", want: []string{"a", "b", "c", "d", "e"}},
		// Moved IDs keep the order given, not their original order
		{ids: []string{"e", "a"}, before: "c", want: []string{"b", "e", "a", "c", "d"}},
		{ids: []string{"b", "d"}, before: "", want: []string{"a", "c", "e", "b", "d"}},
		{ids: nil, before: "a", err: ErrInvalid},
		{ids: []string{"a", "a"}, before: "c", err: ErrInvalid},
		{ids: []string{"a", "b"}, before: "a", err: ErrInvalid},
		{ids: []string{"x"}, before: "a", err: ErrNotFound},
		{ids: []string{"a"}, before: "x", err: ErrNotFound},
	} {
		in := slices.Clone(order)
		got, err := MoveIDs(in, tc.ids, tc.before)
		if !slices.Equal(in, order) {
			t.Errorf("MoveIDs(%v, %q) changed its input to %v", tc.ids, tc.before, in)
		}
		if tc.err != nil {
			if !errors.Is(err, tc.err) {
				t.Errorf("MoveIDs(%v, %q) err = %v, want %v", tc.ids, tc.before, err, tc.err)
			}
			continue
		}
		if err != nil || !slices.Equal(got, tc.want) {
			t.Errorf("MoveIDs(%v, %q) = %v, %v; want %v", tc.ids, tc.before, got, err, tc.want)
		}
	}
}

func TestSwapIDs(t *testing.T) {
	order := []string{"a", "b", "c"}
	got, err := SwapIDs(order, "c", "a")
	if err != nil || !slices.Equal(got, []string{"c", "b", "a"}) {
		t.Errorf("SwapIDs(c, a) = %v, %v; want [c b a]", got, err)
	}
	if !slices.Equal(order, []string{"a", "b", "c"}) {
		t.Errorf("SwapIDs changed its input to %v", order)
	}
	for _, tc := range []struct {
		a, b string
		err  error
	}{
		{"a", "a", ErrInvalid},
		{"a", "x", ErrNotFound},
		{"x", "a", ErrNotFound},
	} {
		if _, err := SwapIDs(order, tc.a, tc.b); !errors.Is(err, tc.err) {
			t.Errorf("SwapIDs(%s, %s) err = %v, want %v", tc.a, tc.b, err, tc.err)
		}
	}
}

func TestCheckDistinct(t *testing.T) {
	if err := CheckDistinct([]string{"a", "b"}); err != nil {
		t.Errorf("CheckDistinct(a, b) = %v, want nil", err)
	}
	if err := CheckDistinct(nil); err != nil {
		t.Errorf("CheckDistinct(nil) = %v, want nil", err)
	}
	if err := CheckDistinct([]string{"a", "b", "a"}); !errors.Is(err, ErrInvalid) {
		t.Errorf("CheckDistinct(a, b, a) = %v, want %v", err, ErrInvalid)
	}
}

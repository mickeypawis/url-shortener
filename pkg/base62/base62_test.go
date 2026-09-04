package base62_test

import (
	"testing"

	"github.com/mickeypawis/url-shortener/pkg/base62"
)

func TestGenerateLength(t *testing.T) {
	s, err := base62.Generate(7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s) != 7 {
		t.Fatalf("expected length 7, got %d", len(s))
	}
}

func TestGenerateAlphabet(t *testing.T) {
	s, err := base62.Generate(200)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, r := range s {
		alnum := (r >= '0' && r <= '9') || (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z')
		if !alnum {
			t.Fatalf("unexpected character %q in generated code", r)
		}
	}
}

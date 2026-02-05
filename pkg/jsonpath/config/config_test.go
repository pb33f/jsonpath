package config

import "testing"

func TestLazyContextTrackingOption(t *testing.T) {
	cfg := New()
	if cfg.LazyContextTrackingEnabled() {
		t.Fatalf("expected lazy context tracking disabled by default")
	}

	cfg = New(WithLazyContextTracking())
	if !cfg.LazyContextTrackingEnabled() {
		t.Fatalf("expected lazy context tracking enabled with option")
	}
}

package config

import "testing"

type legacyConfigImplementation struct{}

func (legacyConfigImplementation) PropertyNameEnabled() bool        { return false }
func (legacyConfigImplementation) JSONPathPlusEnabled() bool        { return true }
func (legacyConfigImplementation) LazyContextTrackingEnabled() bool { return false }

var _ Config = legacyConfigImplementation{}

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

func TestSpectralCompatibilityOption(t *testing.T) {
	cfg := New(WithSpectralCompatibility())
	if !SpectralCompatibilityEnabled(cfg) {
		t.Fatal("expected Spectral compatibility to be enabled")
	}
	if !cfg.JSONPathPlusEnabled() || !cfg.PropertyNameEnabled() {
		t.Fatal("expected Spectral compatibility to imply JSONPath Plus and property-name extensions")
	}
	if err := Validate(cfg); err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}
}

func TestSpectralAndExplicitPropertyNameExtensionAreCompatible(t *testing.T) {
	cfg := New(WithSpectralCompatibility(), WithPropertyNameExtension())
	if err := Validate(cfg); err != nil {
		t.Fatalf("redundant property-name option should be valid: %v", err)
	}
	if !cfg.PropertyNameEnabled() || !SpectralCompatibilityEnabled(cfg) {
		t.Fatal("expected both implied settings to remain enabled")
	}
}

func TestSpectralAndStrictConflictIsOrderIndependent(t *testing.T) {
	tests := []Config{
		New(WithSpectralCompatibility(), WithStrictRFC9535()),
		New(WithStrictRFC9535(), WithSpectralCompatibility()),
	}
	for _, cfg := range tests {
		if err := Validate(cfg); err == nil {
			t.Fatal("expected incompatible dialects to be rejected")
		}
	}
}

func TestLegacyConfigImplementationsRemainCompatible(t *testing.T) {
	cfg := Config(legacyConfigImplementation{})
	if SpectralCompatibilityEnabled(cfg) {
		t.Fatal("legacy Config implementation unexpectedly enabled Spectral compatibility")
	}
	if err := Validate(cfg); err != nil {
		t.Fatalf("legacy Config implementation failed validation: %v", err)
	}
}

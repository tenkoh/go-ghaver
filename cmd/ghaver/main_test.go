package main

import "testing"

func TestParseCLIWithSHA(t *testing.T) {
	cfg, _, err := parseCLI([]string{"--sha"})
	if err != nil {
		t.Fatalf("parseCLI returned error: %v", err)
	}
	if !cfg.SHA {
		t.Fatalf("expected SHA to be true when --sha is provided")
	}
}

func TestParseCLIDefaults(t *testing.T) {
	cfg, _, err := parseCLI(nil)
	if err != nil {
		t.Fatalf("parseCLI returned error: %v", err)
	}
	if cfg.SHA {
		t.Fatalf("expected SHA to be false by default")
	}
}

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
	if cfg.Version {
		t.Fatalf("expected Version to be false when --sha is provided")
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
	if cfg.Version {
		t.Fatalf("expected Version to be false by default")
	}
}

func TestParseCLIWithVersion(t *testing.T) {
	cfg, _, err := parseCLI([]string{"--version"})
	if err != nil {
		t.Fatalf("parseCLI returned error: %v", err)
	}
	if !cfg.Version {
		t.Fatalf("expected Version to be true when --version is provided")
	}
	if cfg.SHA {
		t.Fatalf("expected SHA to be false when --version is provided")
	}
}

func TestParseCLIWithShortVersion(t *testing.T) {
	cfg, _, err := parseCLI([]string{"-v"})
	if err != nil {
		t.Fatalf("parseCLI returned error: %v", err)
	}
	if !cfg.Version {
		t.Fatalf("expected Version to be true when -v is provided")
	}
	if cfg.SHA {
		t.Fatalf("expected SHA to be false when -v is provided")
	}
}

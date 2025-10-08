package ghaver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGitHubAPIVersionsIncludesReleasesAndTags(t *testing.T) {
	handler := http.NewServeMux()
	handler.HandleFunc("/repos/octo/act/releases", func(w http.ResponseWriter, r *http.Request) {
		releases := []map[string]any{
			{
				"tag_name":         "v3.0.0",
				"name":             "Third",
				"draft":            false,
				"prerelease":       true,
				"published_at":     time.Date(2023, 7, 1, 0, 0, 0, 0, time.UTC),
				"target_commitish": "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef",
			},
			{
				"tag_name":         "v2.0.0",
				"name":             "Second",
				"draft":            false,
				"prerelease":       false,
				"published_at":     time.Date(2023, 6, 1, 0, 0, 0, 0, time.UTC),
				"target_commitish": "branch",
			},
		}
		json.NewEncoder(w).Encode(releases)
	})
	handler.HandleFunc("/repos/octo/act/tags", func(w http.ResponseWriter, r *http.Request) {
		tags := []map[string]any{
			{
				"name": "v2.0.0",
				"commit": map[string]any{
					"sha": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				},
			},
			{
				"name": "v1.0.0",
				"commit": map[string]any{
					"sha": "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
				},
			},
		}
		json.NewEncoder(w).Encode(tags)
	})
	handler.HandleFunc("/repos/octo/act/git/ref/tags/v3.0.0", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"object": map[string]any{"sha": "cccccccccccccccccccccccccccccccccccccccc"},
		})
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	api := NewGitHubAPI(server.Client())
	api.baseURL = server.URL

	versions, err := api.Versions(context.Background(), "octo", "act")
	if err != nil {
		t.Fatalf("Versions returned error: %v", err)
	}

	if len(versions) != 3 {
		t.Fatalf("expected 3 versions, got %d", len(versions))
	}

	if versions[0].Owner != "octo" || versions[0].Repo != "act" {
		t.Errorf("unexpected coordinates for release[0]: %+v", versions[0])
	}
	if versions[0].Tag != "v3.0.0" || versions[0].SHA != "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef" || !versions[0].Prerelease {
		t.Errorf("unexpected release[0]: %+v", versions[0])
	}
	if versions[1].Owner != "octo" || versions[1].Repo != "act" {
		t.Errorf("unexpected coordinates for release[1]: %+v", versions[1])
	}
	if versions[1].Tag != "v2.0.0" || versions[1].SHA != "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" {
		t.Errorf("unexpected release[1]: %+v", versions[1])
	}
	if versions[2].Owner != "octo" || versions[2].Repo != "act" {
		t.Errorf("unexpected coordinates for tag fallback: %+v", versions[2])
	}
	if versions[2].Tag != "v1.0.0" || versions[2].SHA != "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb" {
		t.Errorf("unexpected tag fallback: %+v", versions[2])
	}
}

func TestGitHubAPIVersionsFallsBackToTags(t *testing.T) {
	handler := http.NewServeMux()
	handler.HandleFunc("/repos/octo/empty/releases", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]any{})
	})
	handler.HandleFunc("/repos/octo/empty/tags", func(w http.ResponseWriter, r *http.Request) {
		tags := []map[string]any{
			{
				"name": "v0.1.0",
				"commit": map[string]any{
					"sha": "dddddddddddddddddddddddddddddddddddddddd",
				},
			},
		}
		json.NewEncoder(w).Encode(tags)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	api := NewGitHubAPI(server.Client())
	api.baseURL = server.URL

	versions, err := api.Versions(context.Background(), "octo", "empty")
	if err != nil {
		t.Fatalf("Versions returned error: %v", err)
	}
	if len(versions) != 1 {
		t.Fatalf("expected 1 version, got %d", len(versions))
	}
	if versions[0].Owner != "octo" || versions[0].Repo != "empty" {
		t.Fatalf("unexpected coordinates: %+v", versions[0])
	}
	if versions[0].Tag != "v0.1.0" || versions[0].SHA != "dddddddddddddddddddddddddddddddddddddddd" {
		t.Fatalf("unexpected version: %+v", versions[0])
	}
}

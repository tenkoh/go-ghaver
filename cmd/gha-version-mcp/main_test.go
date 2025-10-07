package main

import (
	"context"
	"errors"
	"strconv"
	"testing"

	"github.com/tenkoh/go-ghaver/internal/ghaver"
)

type fakeGitHubClient struct {
	versions []ghaver.Version
	err      error
}

func (f *fakeGitHubClient) Versions(ctx context.Context, owner, repo string) ([]ghaver.Version, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.versions, nil
}

func TestHandleRequestSuccess(t *testing.T) {
	client := &fakeGitHubClient{
		versions: []ghaver.Version{{Tag: "v2.0.1"}, {Tag: "v2.0.0"}, {Tag: "v1.9.0"}, {Tag: "v2.0.1"}},
	}

	limit := 3
	out, err := handleRequest(context.Background(), client, selectInput{Repository: "owner/repo", Limit: &limit})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{"v2.0.1", "v2.0.0", "v1.9.0"}
	if len(out.Tags) != len(expected) {
		t.Fatalf("expected %d tags, got %d", len(expected), len(out.Tags))
	}
	for i, tag := range expected {
		if out.Tags[i] != tag {
			t.Errorf("tag[%d] = %q, want %q", i, out.Tags[i], tag)
		}
	}
}

func TestHandleRequestInvalidRepository(t *testing.T) {
	_, err := handleRequest(context.Background(), &fakeGitHubClient{}, selectInput{Repository: "invalid"})
	if err == nil {
		t.Fatal("expected error for invalid repository, got nil")
	}
}

func TestHandleRequestInvalidLimit(t *testing.T) {
	zero := 0
	_, err := handleRequest(context.Background(), &fakeGitHubClient{}, selectInput{Repository: "owner/repo", Limit: &zero})
	if err == nil {
		t.Fatal("expected error for zero limit")
	}
}

func TestHandleRequestClampLimit(t *testing.T) {
	over := 120
	client := &fakeGitHubClient{
		versions: make([]ghaver.Version, 0, 130),
	}
	for i := 0; i < 130; i++ {
		client.versions = append(client.versions, ghaver.Version{Tag: makeTag(i)})
	}

	out, err := handleRequest(context.Background(), client, selectInput{Repository: "owner/repo", Limit: &over})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out.Tags) != maxLimit {
		t.Fatalf("expected tags to be clamped to %d, got %d", maxLimit, len(out.Tags))
	}
}

func TestHandleRequestAPIFailure(t *testing.T) {
	client := &fakeGitHubClient{err: errors.New("boom")}
	_, err := handleRequest(context.Background(), client, selectInput{Repository: "owner/repo"})
	if err == nil {
		t.Fatal("expected error from API failure")
	}
}

func makeTag(i int) string {
	return "tag" + strconv.Itoa(i)
}

package ghaver

import (
	"context"
	"errors"
	"testing"
)

type stubSelector struct {
	idxs []int
	call int
	err  error
}

func (s *stubSelector) Select(items []string) (int, error) {
	if s.err != nil {
		return 0, s.err
	}
	if s.call >= len(s.idxs) {
		return 0, errors.New("unexpected select call")
	}
	idx := s.idxs[s.call]
	s.call++
	return idx, nil
}

type stubGitHubClient struct {
	versions []Version
	err      error
}

func (c *stubGitHubClient) Versions(ctx context.Context, owner, repo string) ([]Version, error) {
	if c.err != nil {
		return nil, c.err
	}
	return c.versions, nil
}

func TestRunnerRun(t *testing.T) {
	actions := []Action{{Owner: "actions", Repo: "checkout"}}
	versions := []Version{{Tag: "v3", SHA: "abc"}}

	r := Runner{
		Actions:  actions,
		Selector: &stubSelector{idxs: []int{0, 0}},
		Client:   &stubGitHubClient{versions: versions},
	}

	got, err := r.Run(context.Background(), Options{})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if got.Tag != "v3" {
		t.Fatalf("expected v3 tag, got %s", got.Tag)
	}
}

func TestRunnerRunRequiresSHA(t *testing.T) {
	actions := []Action{{Owner: "actions", Repo: "checkout"}}
	versions := []Version{{Tag: "v3", SHA: ""}}

	r := Runner{
		Actions:  actions,
		Selector: &stubSelector{idxs: []int{0, 0}},
		Client:   &stubGitHubClient{versions: versions},
	}

	_, err := r.Run(context.Background(), Options{WithSHA: true})
	if err == nil {
		t.Fatalf("expected error when SHA missing")
	}
}

func TestFormatVersionItems(t *testing.T) {
	items := formatVersionItems([]Version{
		{Tag: "v1", Name: "Release"},
		{Tag: "v10", Name: "Latest"},
	})

	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0] == items[1] {
		t.Fatalf("formatted items should differ")
	}
	if want := "v10"; items[1][:3] != want {
		t.Fatalf("expected padded tag to start with %s, got %q", want, items[1])
	}
}

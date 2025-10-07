package ghaver

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const githubAPIBase = "https://api.github.com"

// HTTPClient is the subset of http.Client used by the GitHub API client.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Version captures metadata needed for selection and output.
type Version struct {
	Tag        string
	Name       string
	SHA        string
	Source     string
	Published  time.Time
	Prerelease bool
}

// GitHubClient fetches version information for actions.
type GitHubClient interface {
	Versions(ctx context.Context, owner, repo string) ([]Version, error)
}

// GitHubAPI implements GitHubClient using the public REST API.
type GitHubAPI struct {
	client    HTTPClient
	baseURL   string
	token     string
	userAgent string
}

// NewGitHubAPI creates a GitHub API client using the provided HTTP client.
func NewGitHubAPI(httpClient HTTPClient) *GitHubAPI {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 15 * time.Second}
	}

	token := strings.TrimSpace(os.Getenv("GITHUB_TOKEN"))
	return &GitHubAPI{
		client:    httpClient,
		baseURL:   githubAPIBase,
		token:     token,
		userAgent: "go-ghaver",
	}
}

// Versions returns release information ordered by published time, falling back to tags.
func (c *GitHubAPI) Versions(ctx context.Context, owner, repo string) ([]Version, error) {
	releases, err := c.listReleases(ctx, owner, repo)
	if err != nil {
		return nil, err
	}

	tags, err := c.listTags(ctx, owner, repo)
	if err != nil {
		return nil, err
	}
	shaByTag := make(map[string]string, len(tags))
	for _, t := range tags {
		if t.SHA != "" {
			shaByTag[t.Name] = t.SHA
		}
	}

	versions := make([]Version, 0, len(releases))
	for _, r := range releases {
		if r.Draft {
			continue
		}
		sha := shaByTag[r.TagName]
		if sha == "" && looksLikeSHA(r.TargetCommitish) {
			sha = r.TargetCommitish
		}
		if sha == "" {
			if fetched, fetchErr := c.fetchTagSHA(ctx, owner, repo, r.TagName); fetchErr == nil {
				sha = fetched
			}
		}
		versions = append(versions, Version{
			Tag:        r.TagName,
			Name:       firstNonEmpty(r.Name, r.TagName),
			SHA:        sha,
			Source:     "release",
			Published:  r.PublishedAt,
			Prerelease: r.Prerelease,
		})
	}

	if len(versions) == 0 {
		versions = make([]Version, 0, len(tags))
	}

	seen := make(map[string]bool, len(versions)+len(tags))
	for i := range versions {
		seen[versions[i].Tag] = true
	}

	for _, t := range tags {
		if seen[t.Name] {
			continue
		}
		versions = append(versions, Version{
			Tag:    t.Name,
			Name:   t.Name,
			SHA:    t.SHA,
			Source: "tag",
		})
	}

	if len(versions) == 0 {
		return nil, fmt.Errorf("no releases or tags found for %s/%s", owner, repo)
	}

	return versions, nil
}

type release struct {
	TagName         string    `json:"tag_name"`
	Name            string    `json:"name"`
	Draft           bool      `json:"draft"`
	Prerelease      bool      `json:"prerelease"`
	PublishedAt     time.Time `json:"published_at"`
	TargetCommitish string    `json:"target_commitish"`
}

type tag struct {
	Name string `json:"name"`
	SHA  string
}

type tagAPIResponse struct {
	Name   string        `json:"name"`
	Commit tagCommitInfo `json:"commit"`
}

type tagCommitInfo struct {
	SHA string `json:"sha"`
}

func (c *GitHubAPI) listReleases(ctx context.Context, owner, repo string) ([]release, error) {
	req, err := c.newRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/releases", owner, repo), map[string]string{"per_page": "30"})
	if err != nil {
		return nil, err
	}

	var releases []release
	if err := c.doJSON(req, &releases); err != nil {
		return nil, err
	}
	return releases, nil
}

func (c *GitHubAPI) listTags(ctx context.Context, owner, repo string) ([]tag, error) {
	req, err := c.newRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/tags", owner, repo), map[string]string{"per_page": "100"})
	if err != nil {
		return nil, err
	}

	var apiResp []tagAPIResponse
	if err := c.doJSON(req, &apiResp); err != nil {
		return nil, err
	}

	tags := make([]tag, 0, len(apiResp))
	for _, t := range apiResp {
		tags = append(tags, tag{Name: t.Name, SHA: t.Commit.SHA})
	}
	return tags, nil
}

func (c *GitHubAPI) fetchTagSHA(ctx context.Context, owner, repo, tagName string) (string, error) {
	req, err := c.newRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/git/ref/tags/%s", owner, repo, url.PathEscape(tagName)), nil)
	if err != nil {
		return "", err
	}

	var resp struct {
		Object struct {
			SHA string `json:"sha"`
		} `json:"object"`
	}
	if err := c.doJSON(req, &resp); err != nil {
		return "", err
	}
	return resp.Object.SHA, nil
}

func (c *GitHubAPI) newRequest(ctx context.Context, method, path string, query map[string]string) (*http.Request, error) {
	base, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, err
	}
	rel, err := url.Parse(path)
	if err != nil {
		return nil, err
	}
	u := base.ResolveReference(rel)
	if query != nil {
		q := u.Query()
		for k, v := range query {
			q.Set(k, v)
		}
		u.RawQuery = q.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", c.userAgent)
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	return req, nil
}

func (c *GitHubAPI) doJSON(req *http.Request, v any) error {
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<10))
		return fmt.Errorf("github api error: %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	if v == nil {
		io.Copy(io.Discard, resp.Body)
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(v)
}

func looksLikeSHA(s string) bool {
	if len(s) != 40 {
		return false
	}
	for _, r := range s {
		if !(r >= '0' && r <= '9') && !(r >= 'a' && r <= 'f') && !(r >= 'A' && r <= 'F') {
			return false
		}
	}
	return true
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

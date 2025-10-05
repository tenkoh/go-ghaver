package ghaver

import (
	"context"
	"fmt"
	"strings"
)

// Options configures the execution of the runner.
type Options struct {
	WithSHA bool
}

// Runner orchestrates the selection flow and version lookup.
type Runner struct {
	Actions  []Action
	Selector Selector
	Client   GitHubClient
}

// Run performs the full workflow and returns the chosen version.
func (r *Runner) Run(ctx context.Context, opts Options) (Version, error) {
	if len(r.Actions) == 0 {
		return Version{}, fmt.Errorf("no actions available")
	}
	if r.Selector == nil {
		return Version{}, fmt.Errorf("selector is not configured")
	}
	if r.Client == nil {
		return Version{}, fmt.Errorf("GitHub client is not configured")
	}

	actionItems := make([]string, len(r.Actions))
	for i, action := range r.Actions {
		actionItems[i] = action.Display()
	}

	idx, err := r.Selector.Select(actionItems)
	if err != nil {
		return Version{}, err
	}

	selected := r.Actions[idx]

	versions, err := r.Client.Versions(ctx, selected.Owner, selected.Repo)
	if err != nil {
		return Version{}, err
	}

	displayItems := formatVersionItems(versions)
	vidx, err := r.Selector.Select(displayItems)
	if err != nil {
		return Version{}, err
	}

	version := versions[vidx]
	if opts.WithSHA && version.SHA == "" {
		return Version{}, fmt.Errorf("version %s has no associated commit SHA", version.Tag)
	}

	return version, nil
}

func formatVersionItems(versions []Version) []string {
	items := make([]string, len(versions))
	maxTagLen := 0
	for _, v := range versions {
		if l := len(v.Tag); l > maxTagLen {
			maxTagLen = l
		}
	}

	for i, v := range versions {
		labelParts := []string{padRight(v.Tag, maxTagLen)}
		if v.Name != "" && v.Name != v.Tag {
			labelParts = append(labelParts, v.Name)
		}
		if v.Prerelease {
			labelParts = append(labelParts, "(pre-release)")
		}
		items[i] = strings.Join(labelParts, "  ")
	}
	return items
}

func padRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}

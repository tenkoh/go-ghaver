package ghaver

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// Action represents a GitHub Action identified by owner and repository.
type Action struct {
	Owner string
	Repo  string
}

// ActionsFromJSON parses a JSON payload that maps owners to their repositories.
func ActionsFromJSON(data []byte) ([]Action, error) {
	var raw map[string][]string
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse actions.json: %w", err)
	}

	actions := make([]Action, 0, len(raw))
	for owner, repos := range raw {
		for _, repo := range repos {
			repo = strings.TrimSpace(repo)
			if repo == "" {
				continue
			}
			actions = append(actions, Action{Owner: owner, Repo: repo})
		}
	}

	sort.Slice(actions, func(i, j int) bool {
		if actions[i].Owner == actions[j].Owner {
			return actions[i].Repo < actions[j].Repo
		}
		return actions[i].Owner < actions[j].Owner
	})

	return actions, nil
}

// Display returns the owner/repository string for the action.
func (a Action) Display() string {
	return fmt.Sprintf("%s/%s", a.Owner, a.Repo)
}

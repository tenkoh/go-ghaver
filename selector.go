package ghaver

import (
	"errors"
	"fmt"

	fzf "github.com/koki-develop/go-fzf"
)

// Selector chooses an index from the provided list of items.
type Selector interface {
	Select(items []string) (int, error)
}

// ErrSelectionAborted indicates the user aborted the selection.
var ErrSelectionAborted = errors.New("selection aborted")

// FZFSelector integrates go-fzf for interactive selection.
type FZFSelector struct {
	options []fzf.Option
}

// NewFZFSelector constructs a selector with optional configuration.
func NewFZFSelector(opts ...fzf.Option) (*FZFSelector, error) {
	return &FZFSelector{options: opts}, nil
}

// Select returns the index of the item chosen by the user.
func (s *FZFSelector) Select(items []string) (int, error) {
	if len(items) == 0 {
		return 0, fmt.Errorf("no items to select")
	}

	finder, err := fzf.New(s.options...)
	if err != nil {
		return 0, fmt.Errorf("init fzf: %w", err)
	}
	defer finder.Quit()

	idxs, err := finder.Find(items, func(i int) string {
		return items[i]
	})
	if err != nil {
		if errors.Is(err, fzf.ErrAbort) {
			return 0, ErrSelectionAborted
		}
		return 0, err
	}
	if len(idxs) == 0 {
		return 0, ErrSelectionAborted
	}
	return idxs[0], nil
}

// Close releases underlying resources held by the selector.
func (s *FZFSelector) Close() {
	// No persistent resources to release.
}

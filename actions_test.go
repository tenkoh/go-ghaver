package ghaver

import "testing"

func TestActionsFromJSON(t *testing.T) {
	data := []byte(`{
        "aws-actions": ["configure-aws-credentials"],
        "actions": ["checkout", "cache"]
    }`)

	actions, err := ActionsFromJSON(data)
	if err != nil {
		t.Fatalf("ActionsFromJSON returned error: %v", err)
	}

	if len(actions) != 3 {
		t.Fatalf("expected 3 actions, got %d", len(actions))
	}

	got := []string{
		actions[0].Display(),
		actions[1].Display(),
		actions[2].Display(),
	}
	want := []string{
		"actions/cache",
		"actions/checkout",
		"aws-actions/configure-aws-credentials",
	}

	for i := range want {
		if got[i] != want[i] {
			t.Errorf("index %d: got %q, want %q", i, got[i], want[i])
		}
	}
}

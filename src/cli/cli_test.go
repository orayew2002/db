package cli

import "testing"

func TestParseCMDRecognizesUpdate(t *testing.T) {
	if got := parseCMD("UPDATE users SET name='Alice' WHERE id=1"); got != Update {
		t.Fatalf("expected %q action, got %q", Update, got)
	}
}

func TestParseAssignments(t *testing.T) {
	vals, err := parseAssignments(`name='Alice Smith',age=42,active=true,email=NULL`)
	if err != nil {
		t.Fatal(err)
	}

	if vals["name"] != "Alice Smith" {
		t.Fatalf("expected quoted string to be preserved, got %#v", vals["name"])
	}
	if vals["age"] != int64(42) {
		t.Fatalf("expected integer parse, got %#v", vals["age"])
	}
	if vals["active"] != true {
		t.Fatalf("expected bool parse, got %#v", vals["active"])
	}
	if vals["email"] != nil {
		t.Fatalf("expected NULL to decode to nil, got %#v", vals["email"])
	}
}

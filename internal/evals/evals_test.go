package evals

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseTest(t *testing.T) {
	raw := `+++
base = "php-composer"
intent = "Upgrade the dependencies of this project to newer versions"

[[graders]]
type = "tool_used"
tool = "Bash"
input_match = "composer outdated --direct"

[[graders]]
type = "git_dirty"
min = 1
max = 3
+++

Agent reports that it raised dependencies.
`
	cfg, body, err := ParseTest(raw)
	if err != nil {
		t.Fatalf("ParseTest() error: %v", err)
	}

	if cfg.Base != "php-composer" {
		t.Errorf("Base = %q", cfg.Base)
	}
	if cfg.Intent != "Upgrade the dependencies of this project to newer versions" {
		t.Errorf("Intent = %q", cfg.Intent)
	}
	if len(cfg.Graders) != 2 {
		t.Fatalf("expected 2 graders, got %d", len(cfg.Graders))
	}
	if cfg.Graders[0].Type != "tool_used" || cfg.Graders[0].Tool != "Bash" {
		t.Errorf("grader 0 = %+v", cfg.Graders[0])
	}
	if cfg.Graders[1].Type != "git_dirty" || cfg.Graders[1].Min != 1 || cfg.Graders[1].Max != 3 {
		t.Errorf("grader 1 = %+v", cfg.Graders[1])
	}
	if body != "Agent reports that it raised dependencies." {
		t.Errorf("body = %q", body)
	}
}

func TestParseTestNoFrontmatter(t *testing.T) {
	_, _, err := ParseTest("just some text")
	if err == nil {
		t.Error("expected error for no frontmatter")
	}
}

func TestParseTestMissingBase(t *testing.T) {
	raw := `+++
intent = "test"
+++
text
`
	_, _, err := ParseTest(raw)
	if err == nil {
		t.Error("expected error for missing base")
	}
}

func TestGraderFileContent(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "package.json")
	os.WriteFile(f, []byte(`{"version": "2.0.0"}`), 0644)

	ctx := &EvalContext{Dir: dir}

	g := Grader{Type: "file_content", Path: "package.json", Pattern: `"version": "2.0.0"`}
	passed, _ := g.Evaluate(ctx)
	if !passed {
		t.Error("expected file_content to pass")
	}

	g2 := Grader{Type: "file_content", Path: "package.json", Pattern: `"version": "3.0.0"`, Match: "not_contains"}
	passed2, _ := g2.Evaluate(ctx)
	if !passed2 {
		t.Error("expected not_contains to pass")
	}

	g3 := Grader{Type: "file_content", Path: "package.json", Pattern: `"version": "3.0.0"`}
	passed3, _ := g3.Evaluate(ctx)
	if passed3 {
		t.Error("expected missing pattern to fail")
	}
}

func TestGraderFileExists(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "README.md"), []byte("# hi"), 0644)

	ctx := &EvalContext{Dir: dir}

	g := Grader{Type: "file_exists", Path: "README.md"}
	passed, _ := g.Evaluate(ctx)
	if !passed {
		t.Error("expected file_exists to pass")
	}

	g2 := Grader{Type: "file_not_exists", Path: "nope.md"}
	passed2, _ := g2.Evaluate(ctx)
	if !passed2 {
		t.Error("expected file_not_exists to pass")
	}

	g3 := Grader{Type: "file_not_exists", Path: "README.md"}
	passed3, _ := g3.Evaluate(ctx)
	if passed3 {
		t.Error("expected file_not_exists to fail")
	}
}

func TestResultOK(t *testing.T) {
	r := Result{Passed: []string{"a"}, Failed: nil}
	if !r.OK() {
		t.Error("expected OK")
	}
	r2 := Result{Failed: []string{"b"}}
	if r2.OK() {
		t.Error("expected not OK")
	}
}

package evals

import (
	"bytes"
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

func TestGraderExitCode(t *testing.T) {
	dir := t.TempDir()
	ctx := &EvalContext{Dir: dir}

	// Comando que retorna 0
	g := Grader{Type: "exit_code", Command: "echo ok", Code: 0}
	passed, _ := g.Evaluate(ctx)
	if !passed {
		t.Error("expected exit_code 0 to pass")
	}

	// Comando que retorna 0 sem especificar code
	g2 := Grader{Type: "exit_code", Command: "echo ok"}
	passed2, _ := g2.Evaluate(ctx)
	if !passed2 {
		t.Error("expected exit_code without code to pass")
	}

	// Comando que retorna 1
	g3 := Grader{Type: "exit_code", Command: "exit 1", Code: 0}
	passed3, _ := g3.Evaluate(ctx)
	if passed3 {
		t.Error("expected exit_code 1 to fail when expecting 0")
	}

	// Comando que retorna 1 e esperamos 1
	g4 := Grader{Type: "exit_code", Command: "exit 1", Code: 1}
	passed4, _ := g4.Evaluate(ctx)
	if !passed4 {
		t.Error("expected exit_code 1 to pass when expecting 1")
	}

	// Sem comando
	g5 := Grader{Type: "exit_code"}
	passed5, msg := g5.Evaluate(ctx)
	if passed5 {
		t.Error("expected no command to fail")
	}
	if msg != "comando não especificado para exit_code" {
		t.Errorf("unexpected message: %s", msg)
	}
}

func TestGraderOutputContains(t *testing.T) {
	dir := t.TempDir()
	log := bytes.NewBufferString("hello world")
	ctx := &EvalContext{Dir: dir, Log: log}

	// Output contém string
	g := Grader{Type: "output_contains", Output: "hello"}
	passed, _ := g.Evaluate(ctx)
	if !passed {
		t.Error("expected output_contains to pass")
	}

	// Output não contém string
	log2 := bytes.NewBufferString("hello world")
	ctx2 := &EvalContext{Dir: dir, Log: log2}
	g2 := Grader{Type: "output_contains", Output: "notfound"}
	passed2, _ := g2.Evaluate(ctx2)
	if passed2 {
		t.Error("expected output_contains to fail")
	}

	// Output contém padrão regex
	log3 := bytes.NewBufferString("hello world")
	ctx3 := &EvalContext{Dir: dir, Log: log3}
	g3 := Grader{Type: "output_contains", Pattern: "hello\\s+world"}
	passed3, _ := g3.Evaluate(ctx3)
	if !passed3 {
		t.Error("expected regex pattern to pass")
	}

	// not_contains
	log4 := bytes.NewBufferString("hello world")
	ctx4 := &EvalContext{Dir: dir, Log: log4}
	g4 := Grader{Type: "output_contains", Output: "notfound", Match: "not_contains"}
	passed4, _ := g4.Evaluate(ctx4)
	if !passed4 {
		t.Error("expected not_contains to pass")
	}

	log5 := bytes.NewBufferString("hello world")
	ctx5 := &EvalContext{Dir: dir, Log: log5}
	g5 := Grader{Type: "output_contains", Output: "hello", Match: "not_contains"}
	passed5, _ := g5.Evaluate(ctx5)
	if passed5 {
		t.Error("expected not_contains to fail when pattern exists")
	}
}

func TestGraderCommandExists(t *testing.T) {
	dir := t.TempDir()
	ctx := &EvalContext{Dir: dir}

	// Comando que existe
	g := Grader{Type: "command_exists", Command: "go"}
	passed, _ := g.Evaluate(ctx)
	if !passed {
		t.Error("expected 'go' to exist")
	}

	// Comando que não existe
	g2 := Grader{Type: "command_exists", Command: "nonexistent_command_xyz"}
	passed2, _ := g2.Evaluate(ctx)
	if passed2 {
		t.Error("expected nonexistent command to fail")
	}

	// Sem comando
	g3 := Grader{Type: "command_exists"}
	passed3, msg := g3.Evaluate(ctx)
	if passed3 {
		t.Error("expected no command to fail")
	}
	if msg != "comando não especificado para command_exists" {
		t.Errorf("unexpected message: %s", msg)
	}
}

func TestGraderSnapshot(t *testing.T) {
	dir := t.TempDir()
	snapshotDir := filepath.Join(dir, "snapshots")
	os.MkdirAll(snapshotDir, 0755)

	log := bytes.NewBufferString("snapshot content")
	ctx := &EvalContext{Dir: dir, Log: log}

	// Criar snapshot novo
	g := Grader{Type: "snapshot", Snapshot: "snapshots/test.txt"}
	passed, msg := g.Evaluate(ctx)
	if !passed {
		t.Error("expected snapshot creation to pass")
	}
	if msg != "snapshot criado: snapshots/test.txt" {
		t.Errorf("unexpected message: %s", msg)
	}

	// Verificar snapshot existente (deve passar)
	log2 := bytes.NewBufferString("snapshot content")
	ctx2 := &EvalContext{Dir: dir, Log: log2}
	g2 := Grader{Type: "snapshot", Snapshot: "snapshots/test.txt"}
	passed2, _ := g2.Evaluate(ctx2)
	if !passed2 {
		t.Error("expected matching snapshot to pass")
	}

	// Verificar snapshot diferente (deve falhar)
	log3 := bytes.NewBufferString("different content")
	ctx3 := &EvalContext{Dir: dir, Log: log3}
	g3 := Grader{Type: "snapshot", Snapshot: "snapshots/test.txt"}
	passed3, _ := g3.Evaluate(ctx3)
	if passed3 {
		t.Error("expected different snapshot to fail")
	}
}

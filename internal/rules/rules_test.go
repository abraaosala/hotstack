package rules

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseOrdersFiles(t *testing.T) {
	dir := t.TempDir()
	writeRule(t, dir, "02_style.md", "use tabs")
	writeRule(t, dir, "01_general.md", "be clean")
	writeRule(t, dir, "10_project.mdc", "project context")
	writeRule(t, dir, "ignore.txt", "not a rule")

	rules, err := Parse(dir)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	if len(rules) != 3 {
		t.Fatalf("expected 3 rules, got %d", len(rules))
	}

	want := []string{"01_general.md", "02_style.md", "10_project.mdc"}
	for i, r := range rules {
		if r.Filename != want[i] {
			t.Errorf("order[%d] = %s, want %s", i, r.Filename, want[i])
		}
	}
}

func TestParseOrdersFilesWithoutPrefix(t *testing.T) {
	dir := t.TempDir()
	writeRule(t, dir, "foo.md", "content")

	rules, err := Parse(dir)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	if len(rules) != 1 || rules[0].Order != 999 {
		t.Fatalf("expected unnumbered rule with order 999, got %+v", rules)
	}
}

func TestRuleTitle(t *testing.T) {
	rule := Rule{Filename: "02_code_style.md"}
	if got := rule.Title(); got != "code style" {
		t.Errorf("Title() = %q, want %q", got, "code style")
	}
}

func TestLoadConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	content := `[project]
name = "demo"
description = "Demo project"

[agents]
cursor = true
copilot = false
claude = true
opencode = true
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig() error: %v", err)
	}

	if cfg.Project.Name != "demo" {
		t.Errorf("Project.Name = %q, want demo", cfg.Project.Name)
	}

	agents := cfg.EnabledAgents()
	want := []string{"claude", "cursor", "opencode"}
	if len(agents) != 3 {
		t.Fatalf("EnabledAgents() = %v, want %v", agents, want)
	}
	for i, a := range want {
		if agents[i] != a {
			t.Errorf("EnabledAgents()[%d] = %s, want %s", i, agents[i], a)
		}
	}
}

func writeRule(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

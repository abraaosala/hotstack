package skills

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseFrontmatter(t *testing.T) {
	meta, body := parseFrontmatter(`---
name: deps-upgrade
description: "Upgrade each dependency to a newer version"
disable-model-invocation: true
---

# Deps Upgrade

Raise each dependency.
`)
	if meta["name"] != "deps-upgrade" {
		t.Errorf("name = %q", meta["name"])
	}
	if meta["description"] != "Upgrade each dependency to a newer version" {
		t.Errorf("description = %q", meta["description"])
	}
	if meta["disable-model-invocation"] != "true" {
		t.Errorf("disable-model-invocation = %q", meta["disable-model-invocation"])
	}
	if body != "# Deps Upgrade\n\nRaise each dependency.\n" {
		t.Errorf("body = %q", body)
	}
}

func TestParseFrontmatterNoFrontmatter(t *testing.T) {
	meta, body := parseFrontmatter("# Just a skill\n\nNo metadata here.")
	if len(meta) != 0 {
		t.Errorf("expected empty meta, got %v", meta)
	}
	if body == "" {
		t.Error("expected body to be returned")
	}
}

func TestLoadOne(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "example")
	mkdir(t, skillDir)
	mkdir(t, filepath.Join(skillDir, "scripts"))
	write(t, filepath.Join(skillDir, "SKILL.md"), `---
name: example
description: Example skill
---

Instructions here
`)
	write(t, filepath.Join(skillDir, "scripts", "run.sh"), "#!/bin/sh\necho hi\n")

	skill, err := LoadOne(dir, "example")
	if err != nil {
		t.Fatal(err)
	}
	if skill == nil {
		t.Fatal("expected skill to load")
	}
	if skill.Name != "example" {
		t.Errorf("Name = %q", skill.Name)
	}
	if skill.Description != "Example skill" {
		t.Errorf("Description = %q", skill.Description)
	}
	if len(skill.Scripts) != 1 || skill.Scripts[0] != "run.sh" {
		t.Errorf("Scripts = %v", skill.Scripts)
	}
}

func TestLoadOneMissing(t *testing.T) {
	skill, err := LoadOne(t.TempDir(), "nope")
	if err != nil {
		t.Fatal(err)
	}
	if skill != nil {
		t.Errorf("expected nil skill, got %+v", skill)
	}
}

func TestLoadMultiple(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"one", "two"} {
		skillDir := filepath.Join(dir, name)
		mkdir(t, skillDir)
		write(t, filepath.Join(skillDir, "SKILL.md"), "---\nname: "+name+"\n---\n\nBody")
	}

	all, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 2 {
		t.Fatalf("expected 2 skills, got %d", len(all))
	}

	names := map[string]bool{}
	for _, s := range all {
		names[s.Name] = true
	}
	if !names["one"] || !names["two"] {
		t.Errorf("loaded names = %v", names)
	}
}

func mkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0755); err != nil {
		t.Fatal(err)
	}
}

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

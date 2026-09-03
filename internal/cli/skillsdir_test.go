package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func writeSkill(t *testing.T, dir, name string) {
	t.Helper()
	skillDir := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Join(skillDir, "scripts"), 0755); err != nil {
		t.Fatal(err)
	}
	md := "---\nname: " + name + "\ndescription: \"teste\"\n---\n\n# " + name + "\n"
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(md), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "scripts", "run.sh"), []byte("echo hi"), 0755); err != nil {
		t.Fatal(err)
	}
}

func TestListSkillsLocalAndGlobal(t *testing.T) {
	dir := t.TempDir()
	global := filepath.Join(t.TempDir(), "skills")

	writeSkill(t, filepath.Join(dir, ".hot", "skills"), "local-skill")
	writeSkill(t, global, "global-skill")

	t.Setenv("HOTSTACK_HOME", filepath.Dir(global))

	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldWD)
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	all, err := listSkills()
	if err != nil {
		t.Fatal(err)
	}

	names := map[string]bool{}
	for _, s := range all {
		names[s.Name] = true
	}
	if !names["local-skill"] || !names["global-skill"] {
		t.Fatalf("esperava local-skill e global-skill, obtive: %v", names)
	}
}

func TestFindSkillPreferLocalOverGlobal(t *testing.T) {
	dir := t.TempDir()
	global := filepath.Join(t.TempDir(), "skills")

	writeSkill(t, filepath.Join(dir, ".hot", "skills"), "dup")
	writeSkill(t, global, "dup")
	writeSkill(t, global, "only-global")

	t.Setenv("HOTSTACK_HOME", filepath.Dir(global))

	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldWD)
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	// local "dup" tem precedência
	skill, err := findSkill("dup")
	if err != nil {
		t.Fatal(err)
	}
	if skill == nil {
		t.Fatal("esperava encontrar dup")
	}
	if skill.Directory != filepath.Join(dir, ".hot", "skills", "dup") {
		t.Fatalf("esperava skill local, obtive %s", skill.Directory)
	}

	// apenas global
	skill, err = findSkill("only-global")
	if err != nil {
		t.Fatal(err)
	}
	if skill == nil {
		t.Fatal("esperava encontrar only-global")
	}
	if skill.Directory != filepath.Join(global, "only-global") {
		t.Fatalf("esperava skill global, obtive %s", skill.Directory)
	}

	// scripts resolvem para a skill global (directory absoluto)
	if _, err := os.Stat(filepath.Join(skill.Directory, "scripts", "run.sh")); err != nil {
		t.Fatal(err)
	}
}

func TestListSkillsOutsideProject(t *testing.T) {
	global := filepath.Join(t.TempDir(), "skills")
	writeSkill(t, global, "g")

	t.Setenv("HOTSTACK_HOME", filepath.Dir(global))

	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldWD)
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatal(err)
	}

	all, err := listSkills()
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 1 || all[0].Name != "g" {
		t.Fatalf("esperava apenas a skill global g, obtive %+v", all)
	}
}

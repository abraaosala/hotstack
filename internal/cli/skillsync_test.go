package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/abraaosala/hotstack/internal/rules"
)

func TestCopyLocalSkills(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, ".hot", "skills")
	writeSkill(t, src, "demo")

	dest := filepath.Join(dir, ".opencode", "skills")
	n, err := copyLocalSkills(src, dest)
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("esperava 1 skill copiada, obtive %d", n)
	}

	md := filepath.Join(dest, "demo", "SKILL.md")
	data, err := os.ReadFile(md)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Fatal("SKILL.md vazio")
	}

	// não sobrescreve destino existente
	custom := []byte("---\nname: demo\n---\n\nCustom.\n")
	os.WriteFile(md, custom, 0644)
	n2, err := copyLocalSkills(src, dest)
	if err != nil {
		t.Fatal(err)
	}
	if n2 != 0 {
		t.Fatalf("esperava 0 cópias (já existe), obtive %d", n2)
	}
	after, _ := os.ReadFile(md)
	if string(after) != string(custom) {
		t.Fatal("destino existente foi sobrescrevido")
	}
}

func TestSkillTargets(t *testing.T) {
	cfg := rules.Config{}
	cfg.Agents.Claude = true
	cfg.Agents.Copilot = true

	got := skillTargets(cfg)
	// copilot não tem diretório de skills; claude gera .claude e .agents
	want := map[string]bool{
		".claude/skills": true,
		".agents/skills": true,
	}
	if len(got) != len(want) {
		t.Fatalf("esperava %d alvos, obtive %v", len(want), got)
	}
	for _, g := range got {
		if !want[g] {
			t.Fatalf("alvo inesperado: %s", g)
		}
	}
}

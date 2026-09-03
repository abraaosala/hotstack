package cli

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestCopyEmbeddedSkills(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "skills")

	n, err := copyEmbeddedSkills(dest)
	if err != nil {
		t.Fatal(err)
	}
	if n == 0 {
		t.Fatal("esperava pelo menos 1 skill embutida copiada")
	}

	skillMD := filepath.Join(dest, "deps-upgrade", "SKILL.md")
	if _, err := os.Stat(skillMD); err != nil {
		t.Fatalf("SKILL.md não copiado: %v", err)
	}
	data, err := os.ReadFile(skillMD)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Fatal("SKILL.md vazio")
	}

	script := filepath.Join(dest, "deps-upgrade", "scripts", "report.sh")
	if _, err := os.Stat(script); err != nil {
		t.Fatalf("report.sh não copiado: %v", err)
	}
	if runtime.GOOS != "windows" {
		info, _ := os.Stat(script)
		if info.Mode().Perm()&0111 == 0 {
			t.Fatal("report.sh não tem permissão de execução")
		}
	}
}

func TestCopyEmbeddedSkillsSkipExisting(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "skills")

	n1, err := copyEmbeddedSkills(dest)
	if err != nil {
		t.Fatal(err)
	}

	existing := filepath.Join(dest, "deps-upgrade", "SKILL.md")
	custom := []byte("---\nname: deps-upgrade\n---\n\nCustom skill.\n")
	os.WriteFile(existing, custom, 0644)

	n2, err := copyEmbeddedSkills(dest)
	if err != nil {
		t.Fatal(err)
	}
	if n2 != 0 {
		t.Fatalf("esperava 0 cópias (já existia), obtive %d", n2)
	}

	data, _ := os.ReadFile(existing)
	if string(data) != string(custom) {
		t.Fatal("ficheiro existente foi sobrescrevido")
	}

	_ = n1
}

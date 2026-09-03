package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteConfigActivatesOnlyChosenAgent(t *testing.T) {
	dir := t.TempDir()

	if err := writeConfig(dir, "opencode"); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "config.toml"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)

	if !strings.Contains(content, "opencode = true") {
		t.Fatalf("agente escolhido não ativado: %s", content)
	}

	for _, name := range agentNames {
		if name == "opencode" {
			continue
		}
		if strings.Contains(content, name+" = true") {
			t.Fatalf("agente %s não deveria estar ativo: %s", name, content)
		}
	}
}

func TestWriteConfigRejectsUnknownAgent(t *testing.T) {
	// writeConfig ativa apenas o agente passado; um nome desconhecido
	// não ativa nenhum, mas não deve quebrar a escrita.
	dir := t.TempDir()

	if err := writeConfig(dir, "desconhecido"); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "config.toml"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)

	for _, name := range agentNames {
		if strings.Contains(content, name+" = true") {
			t.Fatalf("agente %s deveria estar desativado: %s", name, content)
		}
	}
}

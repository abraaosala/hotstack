package skills

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Skill struct {
	Name        string
	Directory   string
	Description string
	DisableModelInvocation bool
	Instructions string
	Scripts     []string
	References  []string
}

func Load(dir string) ([]Skill, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler diretório de skills: %w", err)
	}

	var skills []Skill
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		skill, err := loadSkill(filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, err
		}
		if skill != nil {
			skills = append(skills, *skill)
		}
	}

	return skills, nil
}

func LoadOne(skillsDir, name string) (*Skill, error) {
	path := filepath.Join(skillsDir, name)
	return loadSkill(path)
}

func loadSkill(dir string) (*Skill, error) {
	skillFile := filepath.Join(dir, "SKILL.md")
	content, err := os.ReadFile(skillFile)
	if err != nil {
		return nil, nil
	}

	meta, body := parseFrontmatter(string(content))

	name := dirName(dir)
	if meta["name"] != "" {
		name = meta["name"]
	}

	skill := &Skill{
		Name:                     name,
		Directory:                dir,
		Description:              meta["description"],
		DisableModelInvocation:   meta["disable-model-invocation"] == "true",
		Instructions:             body,
		Scripts:                  listDir(filepath.Join(dir, "scripts")),
		References:               listDir(filepath.Join(dir, "references")),
	}

	return skill, nil
}

func parseFrontmatter(content string) (map[string]string, string) {
	meta := make(map[string]string)

	content = strings.TrimLeft(content, "\ufeff \t\r\n")
	if !strings.HasPrefix(content, "---") {
		return meta, content
	}

	newline := strings.IndexByte(content, '\n')
	if newline == -1 {
		return meta, content
	}

	rest := content[newline+1:]
	end := strings.Index(rest, "\n---")
	if end == -1 {
		return meta, content
	}

	front := rest[:end]
	body := strings.TrimLeft(rest[end+4:], "\r\n")

	for _, line := range strings.Split(front, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		sep := strings.Index(line, ":")
		if sep == -1 {
			continue
		}
		key := strings.TrimSpace(line[:sep])
		value := strings.Trim(strings.TrimSpace(line[sep+1:]), `"'`)
		meta[key] = value
	}

	return meta, body
}

func dirName(path string) string {
	path = filepath.ToSlash(path)
	trimmed := strings.TrimRight(path, "/")
	idx := strings.LastIndex(trimmed, "/")
	if idx == -1 {
		return trimmed
	}
	return trimmed[idx+1:]
}

func listDir(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names
}

package cli

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/abraaosala/hotstack/internal/skills"
)

const localSkillsDir = ".hot/skills"

var errNoSkills = errors.New("nenhuma skill encontrada: cria .hot/skills/<name>/SKILL.md no projeto ou coloca a skill em HOTSTACK_HOME/skills (ou ~/.hotstack/skills)")

func globalSkillsDir() string {
	if home := os.Getenv("HOTSTACK_HOME"); home != "" {
		return filepath.Join(home, "skills")
	}
	userHome, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(userHome, ".hotstack", "skills")
}

func listSkills() ([]skills.Skill, error) {
	var all []skills.Skill

	for _, dir := range []string{localSkillsDir, globalSkillsDir()} {
		if dir == "" {
			continue
		}
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue
		}
		loaded, err := skills.Load(dir)
		if err != nil {
			return nil, err
		}
		for i := range loaded {
			loaded[i].Directory = absPath(loaded[i].Directory)
		}
		all = append(all, loaded...)
	}

	return dedupeSkills(all), nil
}

func findSkill(name string) (*skills.Skill, error) {
	all, err := listSkills()
	if err != nil {
		return nil, err
	}

	// skills locais (projeto) têm precedência sobre globais
	for _, dir := range []string{localSkillsDir, globalSkillsDir()} {
		if dir == "" {
			continue
		}
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue
		}
		skill, err := skills.LoadOne(dir, name)
		if err != nil {
			return nil, err
		}
		if skill != nil {
			skill.Directory = absPath(skill.Directory)
			return skill, nil
		}
	}

	if len(all) == 0 {
		return nil, errNoSkills
	}
	return nil, nil
}

func dedupeSkills(all []skills.Skill) []skills.Skill {
	seen := make(map[string]bool)
	out := make([]skills.Skill, 0, len(all))
	for _, s := range all {
		if seen[s.Name] {
			continue
		}
		seen[s.Name] = true
		out = append(out, s)
	}
	return out
}

func absPath(p string) string {
	if abs, err := filepath.Abs(p); err == nil {
		return abs
	}
	return p
}

package cli

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

//go:embed embedded/skills
var embeddedSkills embed.FS

func copyEmbeddedSkills(destDir string) (int, error) {
	entries, err := fs.ReadDir(embeddedSkills, "embedded/skills")
	if err != nil {
		return 0, fmt.Errorf("erro ao ler skills embutidas: %w", err)
	}

	copied := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		skillName := entry.Name()
		dest := filepath.Join(destDir, skillName)

		if _, err := os.Stat(dest); err == nil {
			continue
		}

		if err := copyEmbeddedSkill(skillName, dest); err != nil {
			return copied, err
		}
		copied++
	}

	return copied, nil
}

func copyEmbeddedSkill(name, dest string) error {
	src := "embedded/skills/" + name
	return fs.WalkDir(embeddedSkills, src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, _ := filepath.Rel(src, path)
		target := filepath.Join(dest, rel)

		if d.IsDir() {
			return os.MkdirAll(target, 0755)
		}

		data, err := fs.ReadFile(embeddedSkills, path)
		if err != nil {
			return err
		}

		perm := fs.FileMode(0644)
		if filepath.Base(path) == "report.sh" || filepath.Ext(path) == ".sh" {
			perm = 0755
		}

		return os.WriteFile(target, data, perm)
	})
}

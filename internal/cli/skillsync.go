package cli

import (
	"os"
	"path/filepath"
)

// copyLocalSkills copies every skill from srcSkillsDir (.hot/skills) into
// destBase/<name>, preserving SKILL.md, scripts and references. Existing
// destinations are left untouched.
func copyLocalSkills(srcSkillsDir, destBase string) (int, error) {
	entries, err := os.ReadDir(srcSkillsDir)
	if err != nil {
		return 0, nil
	}

	copied := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		skillName := entry.Name()
		dest := filepath.Join(destBase, skillName)

		if existing, err := os.Stat(dest); err == nil && existing.IsDir() {
			continue
		}

		if err := copyLocalSkill(filepath.Join(srcSkillsDir, skillName), dest); err != nil {
			return copied, err
		}
		copied++
	}

	return copied, nil
}

func copyLocalSkill(src, dest string) error {
	return filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, rerr := filepath.Rel(src, path)
		if rerr != nil {
			return rerr
		}
		target := filepath.Join(dest, rel)

		if d.IsDir() {
			return os.MkdirAll(target, 0755)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		perm := os.FileMode(0644)
		if filepath.Ext(path) == ".sh" {
			perm = 0755
		}

		return os.WriteFile(target, data, perm)
	})
}

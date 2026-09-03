package evals

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// SetupBase copies a fixture base project (or base_dir) into the working dir.
// basesDir is the path to the bases/ folder for the evals (relative to skill eval dir).
func SetupBase(basesDir, baseDir, name string, workDir string) error {
	var src string

	if baseDir != "" {
		src = baseDir
	} else {
		src = filepath.Join(basesDir, name)
	}

	info, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("base não encontrada: %s (%s): %w", name, src, err)
	}

	if err := os.MkdirAll(workDir, 0755); err != nil {
		return err
	}

	if !info.IsDir() {
		return copyFile(src, filepath.Join(workDir, filepath.Base(src)))
	}

	return copyDir(src, workDir)
}

func copyDir(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

// CopyFixtures copies the extra fixtures/ folder of a test case into workDir.
func CopyFixtures(caseDir, workDir string) error {
	fixtures := filepath.Join(caseDir, "fixtures")
	if _, err := os.Stat(fixtures); os.IsNotExist(err) {
		return nil
	}
	return copyDir(fixtures, workDir)
}

func InitGit(dir string) error {
	g := execGit(dir, "init", "-q")
	if err := g.Run(); err != nil {
		return fmt.Errorf("git init: %w", err)
	}

	execGit(dir, "config", "user.email", "eval@hotstack.local").Run()
	execGit(dir, "config", "user.name", "HotStack Eval").Run()

	_ = os.WriteFile(filepath.Join(dir, ".gitignore"), []byte(""), 0644)
	c := execGit(dir, "add", ".")
	_ = c.Run()
	d := execGit(dir, "commit", "-q", "-m", "base", "--no-verify")
	return d.Run()
}

func IsGitDirty(dir string) bool {
	g := execGit(dir, "status", "--porcelain")
	buf := &strings.Builder{}
	g.Stdout = buf
	rErr := g.Run()
	if rErr != nil {
		return false
	}
	return strings.TrimSpace(buf.String()) != ""
}

func execGit(dir string, args ...string) *exec.Cmd {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Stderr = os.Stderr
	return cmd
}
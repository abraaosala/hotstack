package evals

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type Runner struct {
	ProjectRoot string
	AgentPath   string
}

func NewRunner(projectRoot, agent string) *Runner {
	return &Runner{ProjectRoot: projectRoot, AgentPath: agent}
}

func (r *Runner) RunCase(skillDir, caseName string) (Result, error) {
	_, c, err := LoadCase(skillDir, caseName)
	if err != nil {
		return Result{}, err
	}

	tmpDir, err := os.MkdirTemp("", "hotstack-eval-*")
	if err != nil {
		return Result{}, err
	}
	defer os.RemoveAll(tmpDir)

	basesDir := filepath.Join(skillDir, "..", "bases")
	if err := SetupBase(basesDir, c.BaseDir, c.Base, tmpDir); err != nil {
		return Result{}, fmt.Errorf("setup base: %w", err)
	}

	caseDir := filepath.Join(skillDir, caseName)
	if err := CopyFixtures(caseDir, tmpDir); err != nil {
		return Result{}, fmt.Errorf("copy fixtures: %w", err)
	}

	if err := InitGit(tmpDir); err != nil {
		return Result{}, fmt.Errorf("init git: %w", err)
	}

	log := r.capture(tmpDir, c.Intent)

	ctx := &EvalContext{
		Dir: tmpDir,
		Log: log,
		Env: os.Environ(),
	}

	result := Result{CaseName: caseName}

	for _, g := range c.Graders {
		passed, msg := g.Evaluate(ctx)
		if passed {
			result.Passed = append(result.Passed, msg)
		} else {
			result.Failed = append(result.Failed, msg)
		}
	}

	return result, nil
}

func (r *Runner) capture(workDir, intent string) *bytes.Buffer {
	log := &bytes.Buffer{}

	skillMd, _ := os.ReadFile(filepath.Join(r.ProjectRoot, ".hot", "skills", "hotstack", "SKILL.md"))
	if skillMd == nil {
		skillMd, _ = os.ReadFile("SKILL.md")
	}

	prompt := intent
	if skillMd != nil {
		prompt = string(skillMd) + "\n\n---\n\n" + intent
	}

	var cmd *exec.Cmd
	switch {
	case r.AgentPath != "":
		cmd = exec.Command(r.AgentPath, prompt)
	case runtime.GOOS == "windows":
		cmd = exec.Command("cmd", "/c", "echo", prompt)
	default:
		cmd = exec.Command("sh", "-c", prompt)
	}

	cmd.Dir = workDir
	cmd.Stdout = log
	cmd.Stderr = log
	cmd.Env = os.Environ()
	cmd.Run()

	return log
}

func (r *Runner) RunAll(skillDir string) ([]Result, error) {
	entries, err := os.ReadDir(skillDir)
	if err != nil {
		return nil, err
	}

	var results []Result
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		testFile := filepath.Join(skillDir, entry.Name(), "test.md")
		if _, err := os.Stat(testFile); os.IsNotExist(err) {
			continue
		}
		result, err := r.RunCase(skillDir, entry.Name())
		if err != nil {
			result = Result{
				CaseName: entry.Name(),
				Failed:   []string{err.Error()},
			}
		}
		results = append(results, result)
	}
	return results, nil
}

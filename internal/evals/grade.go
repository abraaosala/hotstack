package evals

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

type EvalContext struct {
	Dir     string
	Log     *bytes.Buffer
	Env     []string
}

func NewEvalContext(dir string) *EvalContext {
	return &EvalContext{
		Dir: dir,
		Log: &bytes.Buffer{},
		Env: os.Environ(),
	}
}

func (g Grader) Evaluate(ctx *EvalContext) (passed bool, message string) {
	switch g.Type {
	case "tool_used":
		return g.evalToolUsed(ctx)
	case "file_content":
		return g.evalFileContent(ctx)
	case "git_dirty":
		return g.evalGitDirty(ctx)
	case "file_exists":
		return g.evalFileExists(ctx)
	case "file_not_exists":
		ok, msg := g.evalFileExists(ctx)
		return !ok, msg
	case "exit_code":
		return g.evalExitCode(ctx)
	case "output_contains":
		return g.evalOutputContains(ctx)
	case "command_exists":
		return g.evalCommandExists(ctx)
	case "snapshot":
		return g.evalSnapshot(ctx)
	default:
		return false, "grader desconhecido: " + g.Type
	}
}

func (g Grader) evalToolUsed(ctx *EvalContext) (bool, string) {
	log := ctx.Log.String()
	if g.Tool != "" {
		lines := strings.Split(log, "\n")
		found := false
		for _, line := range lines {
			if strings.Contains(line, g.Tool) {
				// guard counter
				if g.Min > 0 && countSubstrings(log, g.Tool) < g.Min {
					return false, "tool " + g.Tool + " usada menos que min(" + itoa(g.Min) + ")"
				}
				found = true
				break
			}
		}
		if g.Min > 0 && !found {
			return false, "tool " + g.Tool + " não usada"
		}
		if g.Min > 0 && countSubstrings(log, g.Tool) > g.Max && g.Max > 0 {
			return false, "tool " + g.Tool + " usada mais que max(" + itoa(g.Max) + ")"
		}
	}

	if g.InputMatch != "" {
		if !strings.Contains(log, g.InputMatch) {
			return false, "input não encontrado: " + g.InputMatch
		}
	}

	if g.Min > 0 && countSubstrings(log, g.Tool) <= 0 {
		return false, "tool " + g.Tool + " não usada"
	}

	return true, ""
}

func (g Grader) evalFileContent(ctx *EvalContext) (bool, string) {
	path := filepath.Join(ctx.Dir, g.Path)
	content, err := os.ReadFile(path)
	if err != nil {
		return false, "não consegui ler " + g.Path + ": " + err.Error()
	}

	re, rerr := regexp.Compile(g.Pattern)
	if rerr == nil {
		matched := re.Match(content)
		switch g.Match {
		case "not_contains":
			return !matched, "padrão " + g.Pattern + " encontrado (não devia)"
		default:
			return matched, "padrão não encontrado: " + g.Pattern
		}
	}

	contains := strings.Contains(string(content), g.Pattern)
	switch g.Match {
	case "not_contains":
		return !contains, "padrão " + g.Pattern + " encontrado (não devia)"
	default:
		return contains, "padrão não encontrado: " + g.Pattern
	}
}

func (g Grader) evalGitDirty(ctx *EvalContext) (bool, string) {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = ctx.Dir
	cmd.Env = ctx.Env

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Run(); err != nil {
		return false, "erro ao correr git status: " + err.Error()
	}

	lines := nonEmptyLines(out.String())
	count := len(lines)

	if g.Min > 0 && count < g.Min {
		return false, "working tree tem " + itoa(count) + " mudanças, min é " + itoa(g.Min)
	}
	if g.Max > 0 && count > g.Max {
		return false, "working tree tem " + itoa(count) + " mudanças, max é " + itoa(g.Max)
	}
	return true, ""
}

func (g Grader) evalFileExists(ctx *EvalContext) (bool, string) {
	path := filepath.Join(ctx.Dir, g.Path)
	_, err := os.Stat(path)
	exists := err == nil
	base := filepath.Base(path)
	if !exists {
		return false, base + " não existe"
	}
	return true, ""
}

func (g Grader) evalOutputContains(ctx *EvalContext) (bool, string) {
	output := ctx.Log.String()

	if g.Pattern != "" {
		re, rerr := regexp.Compile(g.Pattern)
		if rerr == nil {
			matched := re.MatchString(output)
			switch g.Match {
			case "not_contains":
				return !matched, "padrão " + g.Pattern + " encontrado no output (não devia)"
			default:
				return matched, "padrão não encontrado no output: " + g.Pattern
			}
		}

		contains := strings.Contains(output, g.Pattern)
		switch g.Match {
		case "not_contains":
			return !contains, "padrão " + g.Pattern + " encontrado no output (não devia)"
		default:
			return contains, "padrão não encontrado no output: " + g.Pattern
		}
	}

	if g.Output != "" {
		contains := strings.Contains(output, g.Output)
		switch g.Match {
		case "not_contains":
			return !contains, "output contém " + g.Output + " (não devia)"
		default:
			return contains, "output não contém: " + g.Output
		}
	}

	return true, ""
}

func (g Grader) evalCommandExists(ctx *EvalContext) (bool, string) {
	if g.Command == "" {
		return false, "comando não especificado para command_exists"
	}

	_, err := exec.LookPath(g.Command)
	if err != nil {
		return false, "comando " + g.Command + " não encontrado no PATH"
	}

	return true, ""
}

func (g Grader) evalSnapshot(ctx *EvalContext) (bool, string) {
	if g.Snapshot == "" {
		return false, "snapshot path não especificado"
	}

	snapshotPath := filepath.Join(ctx.Dir, g.Snapshot)
	currentOutput := ctx.Log.String()

	// Se o snapshot não existe, criar com o output atual
	if _, err := os.Stat(snapshotPath); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(snapshotPath), 0755); err != nil {
			return false, "erro ao criar diretório do snapshot: " + err.Error()
		}
		if err := os.WriteFile(snapshotPath, []byte(currentOutput), 0644); err != nil {
			return false, "erro ao criar snapshot: " + err.Error()
		}
		return true, "snapshot criado: " + g.Snapshot
	}

	// Ler snapshot existente
	snapshotContent, err := os.ReadFile(snapshotPath)
	if err != nil {
		return false, "erro ao ler snapshot: " + err.Error()
	}

	if string(snapshotContent) == currentOutput {
		return true, ""
	}

	return false, "output difere do snapshot " + g.Snapshot
}

func (g Grader) evalExitCode(ctx *EvalContext) (bool, string) {
	if g.Command == "" {
		return false, "comando não especificado para exit_code"
	}

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", g.Command)
	} else {
		cmd = exec.Command("sh", "-c", g.Command)
	}
	cmd.Dir = ctx.Dir
	cmd.Env = ctx.Env

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return false, "erro ao executar comando: " + err.Error()
		}
	}

	if exitCode != g.Code {
		return false, "exit code " + strconv.Itoa(exitCode) + ", esperado " + strconv.Itoa(g.Code)
	}

	return true, ""
}

func nonEmptyLines(s string) []string {
	var lines []string
	for _, l := range strings.Split(s, "\n") {
		if strings.TrimSpace(l) != "" {
			lines = append(lines, l)
		}
	}
	return lines
}

func countSubstrings(s, sub string) int {
	if sub == "" {
		return 0
	}
	return strings.Count(s, sub)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}
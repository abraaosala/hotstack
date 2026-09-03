package skills

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

func (s *Skill) RunScript(name string) error {
	for _, script := range s.Scripts {
		if script == name {
			path := filepath.Join(s.Directory, "scripts", name)
			return runScript(path)
		}
	}
	return fmt.Errorf("script não encontrado: %s", name)
}

func runScript(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	var cmd *exec.Cmd

	switch {
	case info.Mode()&0111 != 0:
		cmd = exec.Command(path)
	case runtime.GOOS == "windows" && (filepath.Ext(path) == ".bat" || filepath.Ext(path) == ".cmd"):
		cmd = exec.Command("cmd", "/c", path)
	case filepath.Ext(path) == ".py":
		cmd = exec.Command("python", path)
	case filepath.Ext(path) == ".sh":
		abs, aerr := filepath.Abs(path)
		if aerr != nil {
			abs = path
		}
		cmd = exec.Command("bash", filepath.ToSlash(abs))
	default:
		cmd = exec.Command(path)
	}

	// Scripts rodam na raiz do projeto (cwd atual), não na pasta da skill,
	// para que consigam detetar go.mod, package.json, composer.json, etc.
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

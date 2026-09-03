package cli

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/abraa/hotstack/internal/skills"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var scriptFlag string

var runCmd = &cobra.Command{
	Use:   "run <skill> [--script <name>]",
	Short: "Executa uma skill",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSkill(args[0])
	},
}

func init() {
	runCmd.Flags().StringVar(&scriptFlag, "script", "", "Nome do script bundlado para executar")
}

func runSkill(name string) error {
	skill, err := findSkill(name)
	if err != nil {
		return err
	}
	if skill == nil {
		return fmt.Errorf("skill não encontrada: %s", name)
	}

	if scriptFlag != "" {
		return runSkillScript(skill, scriptFlag)
	}

	fmt.Printf("Executando skill: %s\n", skill.Name)
	fmt.Println(strings.Repeat("-", 40))

	if skill.DisableModelInvocation {
		color.Yellow("Esta skill desativa invocação de modelo. Use --script para executar um script.")
		return nil
	}

	agent := detectAgent()
	if agent == "" {
		color.Yellow("Nenhum agent detectado. Use --script <nome> para rodar um script diretamente.")
		return nil
	}

	return invokeAgent(agent, skill.Instructions)
}

func runSkillScript(skill *skills.Skill, script string) error {
	if !contains(skill.Scripts, script) {
		return fmt.Errorf("script não encontrado em %s: %s (disponíveis: %s)",
			skill.Name, script, strings.Join(skill.Scripts, ", "))
	}
	fmt.Printf("Executando script %s/%s\n", skill.Name, script)
	return skill.RunScript(script)
}

func detectAgent() string {
	agents := []string{"claude", "opencode", "cursor-agent", "codex", "gemini"}
	for _, agent := range agents {
		if _, err := exec.LookPath(agent); err == nil {
			return agent
		}
	}
	return ""
}

func invokeAgent(agent, prompt string) error {
	cmd := exec.Command(agent, prompt)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

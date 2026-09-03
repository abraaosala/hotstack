package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/abraa/hotstack/internal/evals"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var evalCmd = &cobra.Command{
	Use:   "eval <skill> [case]",
	Short: "Executa evals para validar uma skill",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		skill := args[0]
		testCase := ""
		if len(args) > 1 {
			testCase = args[1]
		}
		return runEval(skill, testCase)
	},
}

func runEval(skill, testCase string) error {
	skillDir := filepath.Join(".hot", "evals", skill)

	if _, err := os.Stat(skillDir); os.IsNotExist(err) {
		return fmt.Errorf("evals não encontrados para: %s", skill)
	}

	projectRoot, _ := os.Getwd()
	runner := evals.NewRunner(projectRoot, "")

	if testCase != "" {
		return runSingleEval(runner, skillDir, testCase)
	}

	return runAllEvals(runner, skillDir, skill)
}

func runSingleEval(runner *evals.Runner, skillDir, caseName string) error {
	fmt.Printf("Evaluando: %s\n\n", caseName)

	result, err := runner.RunCase(skillDir, caseName)
	if err != nil {
		return err
	}

	printResult(result)
	if !result.OK() {
		return fmt.Errorf("%d graders falharam", result.FailedCount())
	}
	return nil
}

func runAllEvals(runner *evals.Runner, skillDir, skill string) error {
	results, err := runner.RunAll(skillDir)
	if err != nil {
		return err
	}

	if len(results) == 0 {
		color.Yellow("Nenhum eval encontrado em %s", skillDir)
		return nil
	}

	passed := 0
	failed := 0

	for _, r := range results {
		fmt.Println(strings.Repeat("-", 60))
		printResult(r)
		if r.OK() {
			passed++
		} else {
			failed++
		}
	}

	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("Resultado: %d passou, %d falhou (%d total)\n", passed, failed, len(results))

	if failed > 0 {
		return fmt.Errorf("%d evals falharam", failed)
	}
	return nil
}

func printResult(r evals.Result) {
	if r.OK() {
		color.Green("✓ %s", r.CaseName)
	} else {
		color.Red("✗ %s", r.CaseName)
	}
	for _, m := range r.Failed {
		if m != "" {
			fmt.Printf("    FAIL: %s\n", m)
		}
	}
}

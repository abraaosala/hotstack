package cli

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Lista as skills disponíveis",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runList()
	},
}

func runList() error {
	all, err := listSkills()
	if err != nil {
		return err
	}

	if len(all) == 0 {
		color.Yellow("Nenhuma skill encontrada")
		color.Yellow("  Coloca skills em %s", localSkillsDir)
		color.Yellow("  ou em %s (ou define HOTSTACK_HOME)", globalSkillsDir())
		return nil
	}

	color.Cyan("Skills disponíveis (%d):\n", len(all))
	for _, s := range all {
		color.Green("  %s", s.Name)
		if s.Description != "" {
			fmt.Printf("    %s\n", s.Description)
		}
		if len(s.Scripts) > 0 {
			fmt.Printf("    scripts: %s\n", joinNames(s.Scripts))
		}
		if len(s.References) > 0 {
			fmt.Printf("    references: %s\n", joinNames(s.References))
		}
	}

	return nil
}

func joinNames(names []string) string {
	out := ""
	for i, n := range names {
		if i > 0 {
			out += ", "
		}
		out += n
	}
	return out
}

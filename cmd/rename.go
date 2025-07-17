package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/kjunh972/loex/internal/config"
)

var renameCmd = &cobra.Command{
	Use:   "rename [old-name] [new-name]",
	Short: "Rename a project",
	Long:  `Rename an existing project. This updates the project name and moves all associated files.`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		oldName := args[0]
		newName := args[1]
		
		configManager, err := config.NewManager()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		if err := configManager.RenameProject(oldName, newName); err != nil {
			fmt.Printf("Failed to rename project: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Project renamed from '%s' to '%s'\n", oldName, newName)
	},
}
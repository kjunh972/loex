package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/kjunh972/loex/internal/config"
)

var (
	forceFlag bool
)

var removeCmd = &cobra.Command{
	Use:   "remove [project]",
	Short: "Remove a project",
	Long:  `Remove a project and all its configuration, logs, and PID files. This action cannot be undone.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		projectName := args[0]
		
		configManager, err := config.NewManager()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		if !configManager.ProjectExists(projectName) {
			fmt.Printf("Project '%s' not found.\n", projectName)
			os.Exit(1)
		}

		if !forceFlag {
			fmt.Printf(" Are you sure you want to remove project '%s'?\n", projectName)
			fmt.Printf("   This will delete all configuration, logs, and PID files.\n")
			fmt.Printf("   This action cannot be undone.\n\n")
			fmt.Print("Type 'yes' to confirm: ")

			reader := bufio.NewReader(os.Stdin)
			response, err := reader.ReadString('\n')
			if err != nil {
				fmt.Printf("Error reading input: %v\n", err)
				os.Exit(1)
			}

			response = strings.TrimSpace(strings.ToLower(response))
			if response != "yes" {
				fmt.Println("Operation cancelled.")
				return
			}
		}

		if err := configManager.DeleteProject(projectName); err != nil {
			fmt.Printf("Failed to remove project: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Project '%s' removed successfully\n", projectName)
	},
}

func init() {
	removeCmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "Skip confirmation prompt")
}
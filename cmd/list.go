package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/kjunh972/loex/internal/config"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all projects",
	Long:  `Display a list of all configured projects.`,
	Run: func(cmd *cobra.Command, args []string) {
		configManager, err := config.NewManager()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		projects, err := configManager.ListProjects()
		if err != nil {
			fmt.Printf("Failed to list projects: %v\n", err)
			os.Exit(1)
		}

		if len(projects) == 0 {
			fmt.Println("No projects found")
			fmt.Println("Use 'loex init [project]' to create your first project")
			return
		}

		fmt.Printf("Found %d project(s):\n\n", len(projects))
		
		for _, projectName := range projects {
			project, err := configManager.LoadProject(projectName)
			if err != nil {
				fmt.Printf("  %s (error loading)\n", projectName)
				continue
			}

			serviceCount := len(project.Services)
			fmt.Printf("  %s (%d service(s))\n", projectName, serviceCount)
			
			if serviceCount > 0 {
				for serviceType := range project.Services {
					fmt.Printf("    - %s\n", serviceType)
				}
			}
			fmt.Println()
		}

		fmt.Printf("Use 'loex status [project]' to check service status\n")
		fmt.Printf("Use 'loex start [project]' to start services\n")
	},
}
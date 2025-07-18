package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/kjunh972/loex/internal/config"
	"github.com/kjunh972/loex/internal/logger"
	"github.com/kjunh972/loex/internal/process"
)

var listCmd = &cobra.Command{
	Use:   "list [project]",
	Short: "List all projects or show details of a specific project",
	Long:  `Display a list of all configured projects, or show detailed information about a specific project.`,
	Args:  cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		configManager, err := config.NewManager()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		// If specific project is requested
		if len(args) == 1 {
			projectName := args[0]
			project, err := configManager.LoadProject(projectName)
			if err != nil {
				fmt.Printf("Project '%s' not found\n", projectName)
				os.Exit(1)
			}

			fmt.Printf("Project: %s\n", projectName)
			fmt.Printf("Created: %s\n", project.Created.Format("2006-01-02 15:04:05"))
			fmt.Printf("Updated: %s\n", project.Updated.Format("2006-01-02 15:04:05"))
			fmt.Printf("Services: %d\n\n", len(project.Services))

			if len(project.Services) > 0 {
				// Initialize process manager to check service status
				logManager := logger.NewManager(configManager)
				processManager := process.NewManager(configManager, logManager)
				
				for serviceType, service := range project.Services {
					// Get service status
					status, err := processManager.GetServiceStatus(projectName, serviceType)
					if err != nil {
						status = "unknown"
					}
					
					// Format status with indicator
					var statusDisplay string
					switch status {
					case "running":
						statusDisplay = "running ●"
					case "stopped":
						statusDisplay = "stopped ○"
					default:
						statusDisplay = "unknown ?"
					}
					
					fmt.Printf("  %s: %s\n", serviceType, statusDisplay)
					fmt.Printf("    Command: %s\n", service.Command)
					fmt.Printf("    Directory: %s\n", service.Dir)
					fmt.Println()
				}
			}
			return
		}

		// List all projects
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

		fmt.Printf("Use 'loex list [project]' to see detailed configuration\n")
		fmt.Printf("Use 'loex status [project]' to check service status\n")
		fmt.Printf("Use 'loex start [project]' to start services\n")
	},
}
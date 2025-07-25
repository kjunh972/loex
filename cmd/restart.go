package cmd

import (
	"fmt"
	"os"

	"github.com/kjunh972/loex/internal/config"
	"github.com/kjunh972/loex/internal/logger"
	"github.com/kjunh972/loex/internal/process"
	"github.com/kjunh972/loex/pkg/models"
	"github.com/spf13/cobra"
)

var restartCmd = &cobra.Command{
	Use:   "restart [project]",
	Short: "Restart all services for a project",
	Long:  `Stop and start all services for a project.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		projectName := args[0]

		configManager, err := config.NewManager()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		if !configManager.ProjectExists(projectName) {
			fmt.Printf("Project '%s' not found\n", projectName)
			fmt.Printf("Use 'loex list' to see available projects\n")
			os.Exit(1)
		}

		project, err := configManager.LoadProject(projectName)
		if err != nil {
			fmt.Printf("Failed to load project: %v\n", err)
			os.Exit(1)
		}

		if len(project.Services) == 0 {
			fmt.Printf("No services configured for project '%s'\n", projectName)
			fmt.Printf("Configure services first:\n")
			fmt.Printf("  loex config detect %s    # Auto-detect (recommended)\n", projectName)
			fmt.Printf("  loex config wizard %s    # Interactive setup\n", projectName)
			fmt.Printf("  loex config %s [service] [command]    # Manual setup\n", projectName)
			os.Exit(1)
		}

		loggerManager := logger.NewManager(configManager)
		processManager := process.NewManager(configManager, loggerManager)

		fmt.Printf("Restarting services for project '%s'...\n", projectName)

		// Stop all running services
		serviceOrder := []models.ServiceType{models.ServiceFrontend, models.ServiceBackend, models.ServiceDB}
		for _, serviceType := range serviceOrder {
			if _, exists := project.Services[serviceType]; exists {
				if isRunning, _ := processManager.IsServiceRunning(projectName, serviceType); isRunning {
					fmt.Printf("Stopping %s service...\n", serviceType)
					if err := processManager.StopService(projectName, serviceType); err != nil {
						fmt.Printf("Failed to stop %s service: %v\n", serviceType, err)
					}
				}
			}
		}

		// Start services in order
		serviceStartOrder := []models.ServiceType{models.ServiceDB, models.ServiceBackend, models.ServiceFrontend}
		var errors []string
		for _, serviceType := range serviceStartOrder {
			if _, exists := project.Services[serviceType]; exists {
				fmt.Printf("Starting %s service...\n", serviceType)
				if err := processManager.StartService(projectName, serviceType); err != nil {
					errors = append(errors, fmt.Sprintf("%s: %v", serviceType, err))
				}
			}
		}

		if len(errors) > 0 {
			fmt.Printf("Some services failed to start:\n")
			for _, err := range errors {
				fmt.Printf("   - %s\n", err)
			}
			os.Exit(1)
		}

		fmt.Printf("Project '%s' restarted\n", projectName)
	},
}


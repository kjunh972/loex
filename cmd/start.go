package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/kjunh972/loex/internal/config"
	"github.com/kjunh972/loex/internal/logger"
	"github.com/kjunh972/loex/internal/process"
	"github.com/kjunh972/loex/pkg/models"
)

var (
	serviceFlag string
)

var startCmd = &cobra.Command{
	Use:   "start [project] [service]",
	Short: "Start services for a project",
	Long:  `Start all services (frontend, backend, database) for the specified project, or start a specific service by providing the service name.`,
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		projectName := args[0]
		
		configManager, err := config.NewManager()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		if !configManager.ProjectExists(projectName) {
			fmt.Printf("Project '%s' not found. Use 'loex init %s' first.\n", projectName, projectName)
			os.Exit(1)
		}

		loggerManager := logger.NewManager(configManager)
		processManager := process.NewManager(configManager, loggerManager)

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

		var servicesToStart []models.ServiceType
		var specificService string

		if len(args) == 2 {
			specificService = args[1]
		} else if serviceFlag != "" {
			specificService = serviceFlag
		}

		if specificService != "" {
			serviceType := models.ServiceType(specificService)
			if _, exists := project.Services[serviceType]; !exists {
				fmt.Printf("Service '%s' not configured for project '%s'\n", specificService, projectName)
				os.Exit(1)
			}
			servicesToStart = []models.ServiceType{serviceType}
		} else {
			serviceOrder := []models.ServiceType{models.ServiceDB, models.ServiceBackend, models.ServiceFrontend}
			for _, serviceType := range serviceOrder {
				if _, exists := project.Services[serviceType]; exists {
					servicesToStart = append(servicesToStart, serviceType)
				}
			}
		}

		var errors []string
		for i, serviceType := range servicesToStart {
			if err := processManager.StartService(projectName, serviceType); err != nil {
				errors = append(errors, fmt.Sprintf("%s: %v", serviceType, err))
			} else {
				if i < len(servicesToStart)-1 {
					fmt.Printf("Waiting for %s to start...\n", serviceType)
					time.Sleep(3 * time.Second)
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

		if specificService != "" {
			fmt.Printf("Service '%s' started successfully for project '%s'\n", specificService, projectName)
		} else {
			fmt.Printf("All services started successfully for project '%s'\n", projectName)
		}
		
		fmt.Printf("Use 'loex status %s' to check service status\n", projectName)
		fmt.Printf("Use 'loex stop %s' to stop services\n", projectName)
	},
}

func init() {
	startCmd.Flags().StringVarP(&serviceFlag, "service", "s", "", "Start specific service (frontend, backend, db)")
}
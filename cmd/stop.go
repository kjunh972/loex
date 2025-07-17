package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/kjunh972/loex/internal/config"
	"github.com/kjunh972/loex/internal/logger"
	"github.com/kjunh972/loex/internal/process"
	"github.com/kjunh972/loex/pkg/models"
)

var stopCmd = &cobra.Command{
	Use:   "stop [project]",
	Short: "Stop services for a project",
	Long:  `Stop all running services for the specified project, or stop a specific service with --service flag.`,
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

		loggerManager := logger.NewManager(configManager)
		processManager := process.NewManager(configManager, loggerManager)

		if serviceFlag != "" {
			// Stop specific service
			serviceType := models.ServiceType(serviceFlag)
			if err := processManager.StopService(projectName, serviceType); err != nil {
				fmt.Printf("Failed to stop service '%s': %v\n", serviceFlag, err)
				os.Exit(1)
			}
			fmt.Printf("Service '%s' stopped for project '%s'\n", serviceFlag, projectName)
		} else {
			// Stop all services
			if err := processManager.StopAllServices(projectName); err != nil {
				fmt.Printf("Failed to stop services: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("All services stopped for project '%s'\n", projectName)
		}
	},
}

func init() {
	stopCmd.Flags().StringVarP(&serviceFlag, "service", "s", "", "Stop specific service (frontend, backend, db)")
}
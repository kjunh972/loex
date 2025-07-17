package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/kjunh972/loex/internal/config"
	"github.com/kjunh972/loex/internal/logger"
	"github.com/kjunh972/loex/internal/process"
)

var statusCmd = &cobra.Command{
	Use:   "status [project]",
	Short: "Check status of project services",
	Long:  `Display the current status of all services for the specified project.`,
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

		project, err := configManager.LoadProject(projectName)
		if err != nil {
			fmt.Printf("Failed to load project: %v\n", err)
			os.Exit(1)
		}

		if len(project.Services) == 0 {
			fmt.Printf("Project '%s' has no configured services\n", projectName)
			return
		}

		loggerManager := logger.NewManager(configManager)
		processManager := process.NewManager(configManager, loggerManager)

		status, err := processManager.GetAllServicesStatus(projectName)
		if err != nil {
			fmt.Printf("Failed to get status: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Status for project '%s':\n\n", projectName)
		
		for serviceType, service := range project.Services {
			serviceStatus := status[serviceType]
			statusIcon := getStatusIcon(serviceStatus)
			
			fmt.Printf("  %s %s\n", statusIcon, serviceType)
			fmt.Printf("    Command: %s\n", service.Command)
			fmt.Printf("    Directory: %s\n", service.Dir)
			fmt.Printf("    Status: %s\n", serviceStatus)
			
			if serviceStatus == "running" {
				if processInfo, err := processManager.GetProcessDetails(projectName, serviceType); err == nil {
					fmt.Printf("    PID: %d\n", processInfo.PID)
					fmt.Printf("    Started: %s\n", processInfo.StartTime.Format("2006-01-02 15:04:05"))
				}
			}
			fmt.Println()
		}

		runningCount := 0
		for _, s := range status {
			if s == "running" {
				runningCount++
			}
		}

		if runningCount == 0 {
			fmt.Printf("Use 'loex start %s' to start services\n", projectName)
		} else if runningCount < len(status) {
			fmt.Printf("Use 'loex start %s' to start remaining services\n", projectName)
		} else {
			fmt.Printf("Use 'loex stop %s' to stop services\n", projectName)
		}
	},
}

func getStatusIcon(status string) string {
	switch status {
	case "running":
		return "[RUNNING]"
	case "stopped":
		return "[STOPPED]"
	case "error":
		return "[ERROR]"
	default:
		return "[UNKNOWN]"
	}
}
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/kjunh972/loex/internal/config"
	"github.com/kjunh972/loex/internal/detector"
	"github.com/kjunh972/loex/internal/logger"
	"github.com/kjunh972/loex/internal/process"
	"github.com/kjunh972/loex/pkg/models"
)

var (
	dirFlag string
)

var configCmd = &cobra.Command{
	Use:   "config [project] [service] [command]",
	Short: "Configure project services", 
	Long:  `Configure services for projects using auto-detection or interactive wizard.`,
	Args:  cobra.RangeArgs(0, 3),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
			return
		}
		
		if len(args) < 3 {
			fmt.Printf("Configure Services:\n\n")
			fmt.Printf("Auto-detect services (recommended):\n")
			fmt.Printf("  cd /path/to/your/project\n")
			fmt.Printf("  loex config detect [project]\n\n")
			fmt.Printf("Interactive setup:\n")
			fmt.Printf("  loex config wizard [project]\n\n")
			fmt.Printf("Manual configuration:\n")
			fmt.Printf("  loex config [project] [service] [command]\n")
			return
		}
		
		projectName := args[0]
		serviceTypeStr := args[1]
		command := args[2]
		
		serviceType := models.ServiceType(serviceTypeStr)
		if serviceType != models.ServiceFrontend && serviceType != models.ServiceBackend && serviceType != models.ServiceDB {
			fmt.Printf("Invalid service type '%s'. Use: frontend, backend, db\n", serviceTypeStr)
			fmt.Printf("Configure Services:\n")
			fmt.Printf("  loex config detect %s    # Auto-detect (recommended)\n", projectName)
			fmt.Printf("  loex config wizard %s    # Interactive setup\n", projectName)
			fmt.Printf("  loex config %s [service] [command]    # Manual setup\n", projectName)
			os.Exit(1)
		}
		
		configManager, err := config.NewManager()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		var project *models.Project
		if configManager.ProjectExists(projectName) {
			project, err = configManager.LoadProject(projectName)
			if err != nil {
				fmt.Printf("Failed to load project: %v\n", err)
				os.Exit(1)
			}
		} else {
			project = &models.Project{
				Name:     projectName,
				Services: make(map[models.ServiceType]models.Service),
				Created:  time.Now(),
				Updated:  time.Now(),
			}
		}

		var serviceDir string
		if dirFlag != "" {
			absDir, err := filepath.Abs(dirFlag)
			if err != nil {
				fmt.Printf("Invalid directory path: %v\n", err)
				os.Exit(1)
			}
			serviceDir = absDir
		} else {
			cwd, err := os.Getwd()
			if err != nil {
				fmt.Printf("Failed to get current directory: %v\n", err)
				os.Exit(1)
			}
			serviceDir = cwd
		}

		if _, err := os.Stat(serviceDir); os.IsNotExist(err) {
			fmt.Printf("Directory does not exist: %s\n", serviceDir)
			os.Exit(1)
		}

		fmt.Printf("\nConfiguration Summary:\n")
		fmt.Printf("  Project: %s\n", projectName)
		fmt.Printf("  Service: %s\n", serviceType)
		fmt.Printf("  Command: %s\n", command)
		fmt.Printf("  Directory: %s\n", serviceDir)
		fmt.Print("\nSave this configuration? (Y/n): ")

		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error reading input: %v\n", err)
			os.Exit(1)
		}

		response = strings.TrimSpace(strings.ToLower(response))
		if response != "" && response != "y" && response != "yes" {
			fmt.Printf("Configuration cancelled\n")
			os.Exit(0)
		}

		project.Services[serviceType] = models.Service{
			Type:    serviceType,
			Command: command,
			Dir:     serviceDir,
		}

		if err := configManager.SaveProject(project); err != nil {
			fmt.Printf("Failed to save project: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Service '%s' configured for project '%s'\n", serviceType, projectName)
	},
}


var configWizardCmd = &cobra.Command{
	Use:   "wizard [project]",
	Short: "Interactive project configuration",
	Long:  `Interactive wizard to configure all services for a project.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		projectName := args[0]
		
		configManager, err := config.NewManager()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		var project *models.Project
		if configManager.ProjectExists(projectName) {
			project, err = configManager.LoadProject(projectName)
			if err != nil {
				fmt.Printf("Failed to load project: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Configuring existing project '%s'\n\n", projectName)
		} else {
			project = &models.Project{
				Name:     projectName,
				Services: make(map[models.ServiceType]models.Service),
				Created:  time.Now(),
				Updated:  time.Now(),
			}
			fmt.Printf("Creating new project '%s'\n\n", projectName)
		}

		detector := detector.New()
		reader := bufio.NewReader(os.Stdin)

		services := []models.ServiceType{models.ServiceFrontend, models.ServiceBackend, models.ServiceDB}

		for _, serviceType := range services {
			fmt.Printf("Configuring %s service:\n", serviceType)

			// Get directory
			fmt.Print("Enter directory path (press Enter for current directory): ")
			dirInput, _ := reader.ReadString('\n')
			dirInput = strings.TrimSpace(dirInput)
			
			var serviceDir string
			if dirInput == "" {
				cwd, _ := os.Getwd()
				serviceDir = cwd
			} else {
				absDir, err := filepath.Abs(dirInput)
				if err != nil {
					fmt.Printf("Invalid directory path, using current directory\n")
					cwd, _ := os.Getwd()
					serviceDir = cwd
				} else {
					serviceDir = absDir
				}
			}

			fmt.Printf("Using directory: %s\n", serviceDir)

			if _, err := os.Stat(serviceDir); os.IsNotExist(err) {
				fmt.Printf(" Directory does not exist, skipping %s service\n\n", serviceType)
				continue
			}

			results, err := detector.DetectServices(serviceDir)
			var command string

			if err == nil {
				for _, result := range results {
					if result.Service == serviceType {
						fmt.Printf("Auto-detected: %s\n", result.Command)
						fmt.Printf("   Reason: %s\n", result.DetectionReason)
						fmt.Print("Use this command? (Y/n): ")

						response, _ := reader.ReadString('\n')
						response = strings.TrimSpace(strings.ToLower(response))
						if response == "" || response == "y" || response == "yes" {
							command = result.Command
							break
						}
					}
				}
			}

			if command == "" {
				fmt.Printf("Enter command for %s service: ", serviceType)
				cmdInput, _ := reader.ReadString('\n')
				command = strings.TrimSpace(cmdInput)
			}

			if command == "" {
				fmt.Printf(" No command provided, skipping %s service\n\n", serviceType)
				continue
			}

			project.Services[serviceType] = models.Service{
				Type:    serviceType,
				Command: command,
				Dir:     serviceDir,
			}

			fmt.Printf("%s service configured\n\n", serviceType)
		}

		if len(project.Services) == 0 {
			fmt.Printf("No services configured\n")
			os.Exit(1)
		}

		if err := configManager.SaveProject(project); err != nil {
			fmt.Printf("Failed to save project: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Project '%s' configured successfully with %d service(s)\n", projectName, len(project.Services))
		fmt.Printf("Use 'loex start %s' to start all services\n", projectName)
	},
}

var configDetectCmd = &cobra.Command{
	Use:   "detect [project]",
	Short: "Auto-detect and configure services in current directory",
	Long:  `Automatically detect services in the current directory and configure them for the project.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		projectName := args[0]
		
		configManager, err := config.NewManager()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		var project *models.Project
		if configManager.ProjectExists(projectName) {
			project, err = configManager.LoadProject(projectName)
			if err != nil {
				fmt.Printf("Failed to load project: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Checking project '%s' for new services\n\n", projectName)
		} else {
			project = &models.Project{
				Name:     projectName,
				Services: make(map[models.ServiceType]models.Service),
				Created:  time.Now(),
				Updated:  time.Now(),
			}
			fmt.Printf("Creating new project '%s'\n\n", projectName)
		}

		cwd, _ := os.Getwd()
		fmt.Printf("Analyzing current directory: %s\n\n", cwd)

		detector := detector.New()
		results, err := detector.DetectServices(cwd)
		if err != nil {
			fmt.Printf("Failed to detect services: %v\n", err)
			os.Exit(1)
		}

		if len(results) == 0 {
			fmt.Printf("No services detected in current directory.\n")
			fmt.Printf("Make sure you're in a project directory (with package.json, go.mod, etc.)\n")
			return
		}

		var existingServices []models.ServiceType
		var hasNewServices bool
		
		for _, result := range results {
			if _, exists := project.Services[result.Service]; exists {
				existingServices = append(existingServices, result.Service)
			} else {
				hasNewServices = true
			}
		}

		if len(existingServices) > 0 {
			fmt.Printf("Already configured services in this project:\n")
			for _, serviceType := range existingServices {
				service := project.Services[serviceType]
				fmt.Printf("  - %s: %s\n", serviceType, service.Command)
			}
			fmt.Println()
		}

		if !hasNewServices {
			fmt.Printf("No new services detected. All detected services are already configured.\n")
			return
		}

		fmt.Printf("New services detected:\n")
		for _, result := range results {
			if _, exists := project.Services[result.Service]; !exists {
				fmt.Printf("  - %s: %s (%s)\n", result.Service, result.Command, result.DetectionReason)
			}
		}
		fmt.Println()

		reader := bufio.NewReader(os.Stdin)
		
		for _, result := range results {
			if _, exists := project.Services[result.Service]; exists {
				continue
			}
			fmt.Printf("Configuring %s service:\n", result.Service)
			fmt.Printf("Auto-detected command: %s\n", result.Command)
			fmt.Printf("Reason: %s\n", result.DetectionReason)
			fmt.Print("Use this command? (Y/n): ")

			response, _ := reader.ReadString('\n')
			response = strings.TrimSpace(strings.ToLower(response))
			
			var command string
			if response == "" || response == "y" || response == "yes" {
				command = result.Command
			} else {
				fmt.Printf("Enter custom command for %s service: ", result.Service)
				cmdInput, _ := reader.ReadString('\n')
				command = strings.TrimSpace(cmdInput)
				
				if command == "" {
					fmt.Printf("No command provided, skipping %s service\n\n", result.Service)
					continue
				}
			}

			project.Services[result.Service] = models.Service{
				Type:    result.Service,
				Command: command,
				Dir:     cwd,
			}

			fmt.Printf("%s service configured\n\n", result.Service)
		}

		if err := configManager.SaveProject(project); err != nil {
			fmt.Printf("Failed to save project: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Project '%s' configured successfully with %d service(s)\n", projectName, len(project.Services))
		fmt.Printf("Use 'loex start %s' to start services\n", projectName)
	},
}

var configEditCmd = &cobra.Command{
	Use:   "edit [project] [service]",
	Short: "Edit service configuration",
	Long:  `Edit the command and directory for an existing service.`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		projectName := args[0]
		serviceTypeStr := args[1]

		serviceType := models.ServiceType(serviceTypeStr)
		if serviceType != models.ServiceFrontend && serviceType != models.ServiceBackend && serviceType != models.ServiceDB {
			fmt.Printf("Invalid service type '%s'. Use: frontend, backend, db\n", serviceTypeStr)
			os.Exit(1)
		}

		configManager, err := config.NewManager()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		if !configManager.ProjectExists(projectName) {
			fmt.Printf("Project '%s' not found\n", projectName)
			os.Exit(1)
		}

		project, err := configManager.LoadProject(projectName)
		if err != nil {
			fmt.Printf("Failed to load project: %v\n", err)
			os.Exit(1)
		}

		service, exists := project.Services[serviceType]
		if !exists {
			fmt.Printf("Service '%s' not configured for project '%s'\n", serviceType, projectName)
			fmt.Printf("Configure it first:\n")
			fmt.Printf("  loex config %s %s [command]\n", projectName, serviceType)
			os.Exit(1)
		}

		fmt.Printf("Current configuration for %s service:\n", serviceType)
		fmt.Printf("  Command: %s\n", service.Command)
		fmt.Printf("  Directory: %s\n", service.Dir)
		fmt.Print("\nEnter new command (press Enter to keep current): ")

		reader := bufio.NewReader(os.Stdin)
		newCommand, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error reading input: %v\n", err)
			os.Exit(1)
		}
		newCommand = strings.TrimSpace(newCommand)
		if newCommand == "" {
			newCommand = service.Command
		}

		fmt.Print("Enter new directory (press Enter to keep current): ")
		newDir, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error reading input: %v\n", err)
			os.Exit(1)
		}
		newDir = strings.TrimSpace(newDir)
		if newDir == "" {
			newDir = service.Dir
		} else {
			absDir, err := filepath.Abs(newDir)
			if err != nil {
				fmt.Printf("Invalid directory path: %v\n", err)
				os.Exit(1)
			}
			if _, err := os.Stat(absDir); os.IsNotExist(err) {
				fmt.Printf("Directory does not exist: %s\n", absDir)
				os.Exit(1)
			}
			newDir = absDir
		}

		fmt.Printf("\nNew configuration:\n")
		fmt.Printf("  Command: %s\n", newCommand)
		fmt.Printf("  Directory: %s\n", newDir)
		fmt.Print("\nSave changes? (Y/n): ")

		response, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error reading input: %v\n", err)
			os.Exit(1)
		}

		response = strings.TrimSpace(strings.ToLower(response))
		if response != "" && response != "y" && response != "yes" {
			fmt.Printf("Changes discarded\n")
			os.Exit(0)
		}

		project.Services[serviceType] = models.Service{
			Type:    serviceType,
			Command: newCommand,
			Dir:     newDir,
		}
		project.Updated = time.Now()

		if err := configManager.SaveProject(project); err != nil {
			fmt.Printf("Failed to save project: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Service '%s' configuration updated for project '%s'\n", serviceType, projectName)
	},
}

var configDeleteCmd = &cobra.Command{
	Use:   "delete [project] [service]",
	Short: "Delete service configuration",
	Long:  `Delete a service configuration from a project.`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		projectName := args[0]
		serviceTypeStr := args[1]

		serviceType := models.ServiceType(serviceTypeStr)
		if serviceType != models.ServiceFrontend && serviceType != models.ServiceBackend && serviceType != models.ServiceDB {
			fmt.Printf("Invalid service type '%s'. Use: frontend, backend, db\n", serviceTypeStr)
			os.Exit(1)
		}

		configManager, err := config.NewManager()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		if !configManager.ProjectExists(projectName) {
			fmt.Printf("Project '%s' not found\n", projectName)
			os.Exit(1)
		}

		project, err := configManager.LoadProject(projectName)
		if err != nil {
			fmt.Printf("Failed to load project: %v\n", err)
			os.Exit(1)
		}

		service, exists := project.Services[serviceType]
		if !exists {
			fmt.Printf("Service '%s' not configured for project '%s'\n", serviceType, projectName)
			os.Exit(1)
		}

		loggerManager := logger.NewManager(configManager)
		processManager := process.NewManager(configManager, loggerManager)
		if isRunning, _ := processManager.IsServiceRunning(projectName, serviceType); isRunning {
			fmt.Printf("Service '%s' is currently running. Stop it first:\n", serviceType)
			fmt.Printf("  loex stop %s %s\n", projectName, serviceType)
			os.Exit(1)
		}

		fmt.Printf("Service configuration to delete:\n")
		fmt.Printf("  Project: %s\n", projectName)
		fmt.Printf("  Service: %s\n", serviceType)
		fmt.Printf("  Command: %s\n", service.Command)
		fmt.Printf("  Directory: %s\n", service.Dir)
		fmt.Print("\nAre you sure you want to delete this service configuration? (y/N): ")

		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error reading input: %v\n", err)
			os.Exit(1)
		}

		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			fmt.Printf("Operation cancelled\n")
			os.Exit(0)
		}

		delete(project.Services, serviceType)
		project.Updated = time.Now()

		if err := configManager.SaveProject(project); err != nil {
			fmt.Printf("Failed to save project: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Service '%s' deleted from project '%s'\n", serviceType, projectName)
	},
}

func init() {
	configCmd.AddCommand(configWizardCmd)
	configCmd.AddCommand(configDetectCmd)
	configCmd.AddCommand(configEditCmd)
	configCmd.AddCommand(configDeleteCmd)
	
	configCmd.Flags().StringVar(&dirFlag, "dir", "", "Directory path for the service")
}
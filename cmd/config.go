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
	"github.com/kjunh972/loex/pkg/models"
)

var (
	dirFlag string
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage project configuration",
	Long:  `Configure services for projects. Use subcommands to set, get, or interactively configure services.`,
}

var configSetCmd = &cobra.Command{
	Use:   "set [project] [service] [command]",
	Short: "Set service configuration",
	Long: `Set the command and directory for a service.
Services: frontend, backend, db
If --dir is not specified, current directory is used.`,
	Args: cobra.RangeArgs(2, 3),
	Run: func(cmd *cobra.Command, args []string) {
		projectName := args[0]
		serviceTypeStr := args[1]
		
		// Validate service type
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

		// Load or create project
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

		// Determine directory
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

		// Check if directory exists
		if _, err := os.Stat(serviceDir); os.IsNotExist(err) {
			fmt.Printf("Directory does not exist: %s\n", serviceDir)
			os.Exit(1)
		}

		var command string
		if len(args) == 3 {
			command = args[2]
		} else {
			// Auto-detect command
			detector := detector.New()
			results, err := detector.DetectServices(serviceDir)
			if err != nil {
				fmt.Printf("Failed to detect services: %v\n", err)
				os.Exit(1)
			}

			var detectedCmd string
			var detectedReason string
			for _, result := range results {
				if result.Service == serviceType {
					detectedCmd = result.Command
					detectedReason = result.DetectionReason
					break
				}
			}

			if detectedCmd != "" {
				fmt.Printf("Auto-detected command for %s: %s\n", serviceType, detectedCmd)
				fmt.Printf("   Reason: %s\n", detectedReason)
				fmt.Print("Use this command? (Y/n): ")

				reader := bufio.NewReader(os.Stdin)
				response, err := reader.ReadString('\n')
				if err != nil {
					fmt.Printf("Error reading input: %v\n", err)
					os.Exit(1)
				}

				response = strings.TrimSpace(strings.ToLower(response))
				if response == "" || response == "y" || response == "yes" {
					command = detectedCmd
				}
			}

			if command == "" {
				fmt.Printf("Enter command for %s service: ", serviceType)
				reader := bufio.NewReader(os.Stdin)
				input, err := reader.ReadString('\n')
				if err != nil {
					fmt.Printf("Error reading input: %v\n", err)
					os.Exit(1)
				}
				command = strings.TrimSpace(input)
			}
		}

		if command == "" {
			fmt.Printf("Command cannot be empty\n")
			os.Exit(1)
		}

		// Save service configuration
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
		fmt.Printf("   Command: %s\n", command)
		fmt.Printf("   Directory: %s\n", serviceDir)
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

		// Load or create project
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

			// Check if directory exists
			if _, err := os.Stat(serviceDir); os.IsNotExist(err) {
				fmt.Printf(" Directory does not exist, skipping %s service\n\n", serviceType)
				continue
			}

			// Auto-detect command
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

			// Save service
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

func init() {
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configWizardCmd)
	
	configSetCmd.Flags().StringVar(&dirFlag, "dir", "", "Directory path for the service")
}
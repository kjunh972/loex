package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/kjunh972/loex/internal/config"
	"github.com/kjunh972/loex/internal/detector"
	"github.com/kjunh972/loex/pkg/models"
)

var detectCmd = &cobra.Command{
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

		// Load or create project
		var project *models.Project
		if configManager.ProjectExists(projectName) {
			project, err = configManager.LoadProject(projectName)
			if err != nil {
				fmt.Printf("Failed to load project: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Adding services to existing project '%s'\n\n", projectName)
		} else {
			project = &models.Project{
				Name:     projectName,
				Services: make(map[models.ServiceType]models.Service),
				Created:  time.Now(),
				Updated:  time.Now(),
			}
			fmt.Printf("Creating new project '%s'\n\n", projectName)
		}

		// Get current directory
		cwd, _ := os.Getwd()
		fmt.Printf("Analyzing current directory: %s\n\n", cwd)

		// Auto-detect services in current directory
		detector := detector.New()
		results, err := detector.DetectServices(cwd)
		if err != nil {
			fmt.Printf("Failed to detect services: %v\n", err)
			os.Exit(1)
		}

		if len(results) == 0 {
			fmt.Printf("No services detected in current directory.\n")
			fmt.Printf("Make sure you're in a project directory (with package.json, go.mod, etc.)\n")
			os.Exit(1)
		}

		fmt.Printf("Detected services:\n")
		for _, result := range results {
			fmt.Printf("  - %s: %s (%s)\n", result.Service, result.Command, result.DetectionReason)
		}
		fmt.Println()

		reader := bufio.NewReader(os.Stdin)
		
		// Configure each detected service
		for _, result := range results {
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

			// Save service
			project.Services[result.Service] = models.Service{
				Type:    result.Service,
				Command: command,
				Dir:     cwd,
			}

			fmt.Printf("%s service configured\n\n", result.Service)
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
		fmt.Printf("Use 'loex start %s' to start services\n", projectName)
	},
}
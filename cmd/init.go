package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/kjunh972/loex/internal/config"
	"github.com/kjunh972/loex/pkg/models"
)

var initCmd = &cobra.Command{
	Use:   "init [project]",
	Short: "Initialize a new project",
	Long:  `Initialize a new project configuration. This creates an empty project that you can configure later.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		projectName := args[0]
		
		if projectName == "" {
			fmt.Printf("Project name cannot be empty\n")
			os.Exit(1)
		}
		
		if strings.ContainsAny(projectName, " \t\n\r/\\:<>|*?") {
			fmt.Printf("Project name contains invalid characters. Use only letters, numbers, hyphens, and underscores.\n")
			os.Exit(1)
		}
		
		configManager, err := config.NewManager()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		if configManager.ProjectExists(projectName) {
			fmt.Printf("Project '%s' already exists\n", projectName)
			fmt.Printf("Use 'loex list' to see all projects\n")
			os.Exit(1)
		}

		project := &models.Project{
			Name:     projectName,
			Services: make(map[models.ServiceType]models.Service),
			Created:  time.Now(),
			Updated:  time.Now(),
		}

		if err := configManager.SaveProject(project); err != nil {
			fmt.Printf("Failed to save project: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Project '%s' initialized successfully\n", projectName)
		fmt.Printf("Next steps - Configure Services:\n\n")
		fmt.Printf("Auto-detect services (recommended):\n")
		fmt.Printf("  cd /path/to/your/project\n")
		fmt.Printf("  loex config detect %s\n\n", projectName)
		fmt.Printf("Interactive setup:\n")
		fmt.Printf("  loex config wizard %s\n\n", projectName)
		fmt.Printf("Manual configuration:\n")
		fmt.Printf("  loex config %s [service] [command]\n", projectName)
	},
}
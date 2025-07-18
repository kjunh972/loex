package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "loex",
	Short: "Local development environment manager",
	Long: `Loex is a CLI tool for managing and running your local frontend, backend, and database services easily.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Check if version flag is set
		if versionFlag, _ := cmd.Flags().GetBool("version"); versionFlag {
			fmt.Printf("loex version %s\n", version)
			if commit != "unknown" {
				fmt.Printf("commit: %s\n", commit)
			}
			if date != "unknown" {
				fmt.Printf("built: %s\n", date)
			}
			return
		}
		
		fmt.Println("Loex - Local Execution Manager")
		fmt.Println("Use 'loex --help' or 'loex [command] --help' to see available commands.")
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("loex version %s\n", version)
		if commit != "unknown" {
			fmt.Printf("commit: %s\n", commit)
		}
		if date != "unknown" {
			fmt.Printf("built: %s\n", date)
		}
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func SetVersionInfo(v, c, d string) {
	version = v
	commit = c
	date = d
}

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(removeCmd)
	rootCmd.AddCommand(renameCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(detectCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(updateCmd)
	
	// Add version flag
	rootCmd.Flags().BoolP("version", "v", false, "Print version information")
}



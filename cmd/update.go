package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/kjunh972/loex/internal/updater"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update loex to the latest version",
	Long:  `Check for and install the latest version of loex from GitHub releases.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Checking for updates...")
		
		updater := updater.New()
		
		// Check for updates
		hasUpdate, latestVersion, err := updater.CheckForUpdate(version)
		if err != nil {
			fmt.Printf("Failed to check for updates: %v\n", err)
			os.Exit(1)
		}

		if !hasUpdate {
			fmt.Printf("You are already using the latest version (%s)\n", version)
			return
		}

		fmt.Printf("New version available: %s (current: %s)\n", latestVersion, version)
		fmt.Print("Do you want to update? (Y/n): ")

		var response string
		fmt.Scanln(&response)
		
		if response != "" && response != "y" && response != "Y" && response != "yes" && response != "Yes" {
			fmt.Println("Update cancelled")
			return
		}

		fmt.Printf("Updating to version %s...\n", latestVersion)
		
		if err := updater.Update(latestVersion); err != nil {
			if strings.Contains(err.Error(), "permission denied") {
				fmt.Printf("Update failed due to permission restrictions.\n")
				fmt.Printf("For Homebrew installations, please use:\n")
				fmt.Printf("  brew update && brew upgrade loex\n")
				fmt.Printf("\nAlternatively, run with sudo (not recommended):\n")
				fmt.Printf("  sudo loex update\n")
			} else {
				fmt.Printf("Failed to update: %v\n", err)
			}
			os.Exit(1)
		}

		fmt.Printf("Successfully updated to version %s\n", latestVersion)
		fmt.Println("Please restart your terminal or run 'loex -v' to verify the update")
	},
}
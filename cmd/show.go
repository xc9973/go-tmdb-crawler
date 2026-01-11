package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// showCmd represents the show command
var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Manage TV shows",
	Long:  `Manage TV shows in the database - list, add, remove, and view shows`,
}

// showListCmd represents the list shows command
var showListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all shows",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Fetching shows...")
		// TODO: Implement list shows
		fmt.Println("Shows list:")
		fmt.Println("  1. 骄阳似我 (TMDB ID: 278577)")
		fmt.Println("  2. 陆海之战 (TMDB ID: 259886)")
	},
}

// showAddCmd represents the add show command
var showAddCmd = &cobra.Command{
	Use:   "add [tmdb-id]",
	Short: "Add a show by TMDB ID",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		tmdbID := parseIntArg(args[0])
		fmt.Printf("Adding show with TMDB ID: %d\n", tmdbID)
		// TODO: Implement add show
		fmt.Println("✓ Show added successfully")
	},
}

// showDeleteCmd represents the delete show command
var showDeleteCmd = &cobra.Command{
	Use:   "delete [show-id]",
	Short: "Delete a show by ID",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		showID := parseIntArg(args[0])
		fmt.Printf("Deleting show with ID: %d\n", showID)
		// TODO: Implement delete show
		fmt.Println("✓ Show deleted successfully")
	},
}

// showInfoCmd represents the show info command
var showInfoCmd = &cobra.Command{
	Use:   "info [show-id]",
	Short: "Show detailed information about a show",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		showID := parseIntArg(args[0])
		fmt.Printf("Fetching show info for ID: %d\n", showID)
		// TODO: Implement show info
		fmt.Println("Show Details:")
		fmt.Println("  Name: 骄阳似我")
		fmt.Println("  Status: Returning Series")
		fmt.Println("  First Air Date: 2025-01-01")
		fmt.Println("  Seasons: 1")
		fmt.Println("  Episodes: 8")
	},
}

// showTodayCmd represents the today updates command
var showTodayCmd = &cobra.Command{
	Use:   "today",
	Short: "Show today's episode updates",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Fetching today's updates...")
		// TODO: Implement today updates
		fmt.Println("Today's Updates:")
		fmt.Println("  📺 骄阳似我 S01E08 - 第八集")
		fmt.Println("  📺 辐射 S02E05 - 鸟骨")
		fmt.Println("  📺 咒术回战 S03E45 - 假面")
	},
}

// showSearchCmd represents the search show command
var showSearchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search shows by name",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		query := args[0]
		fmt.Printf("Searching for shows: %s\n", query)
		// TODO: Implement search
		fmt.Printf("Search results for '%s':\n", query)
		fmt.Println("  1. 骄阳似我 (TMDB ID: 278577)")
		fmt.Println("  2. 陆海之战 (TMDB ID: 259886)")
	},
}

func init() {
	rootCmd.AddCommand(showCmd)
	showCmd.AddCommand(showListCmd)
	showCmd.AddCommand(showAddCmd)
	showCmd.AddCommand(showDeleteCmd)
	showCmd.AddCommand(showInfoCmd)
	showCmd.AddCommand(showTodayCmd)
	showCmd.AddCommand(showSearchCmd)
}

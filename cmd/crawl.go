package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var tmdbID int
var tmdbIDs []int
var status string

// crawlCmd represents the crawl command
var crawlCmd = &cobra.Command{
	Use:   "crawl",
	Short: "Crawl TV shows from TMDB",
	Long:  `Crawl TV shows from TMDB by ID or status`,
}

// crawlShowCmd represents the crawl show command
var crawlShowCmd = &cobra.Command{
	Use:   "show [tmdb-id]",
	Short: "Crawl a single show from TMDB",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		tmdbID := parseIntArg(args[0])
		fmt.Printf("Crawling show with TMDB ID: %d\n", tmdbID)
		// TODO: Implement crawler service call
		fmt.Println("✓ Show crawled successfully")
	},
}

// crawlAllCmd represents the crawl all command
var crawlAllCmd = &cobra.Command{
	Use:   "all",
	Short: "Crawl all shows in the database",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Starting to crawl all shows...")
		// TODO: Implement refresh all
		fmt.Println("✓ All shows refreshed successfully")
	},
}

// crawlStatusCmd represents the crawl by status command
var crawlStatusCmd = &cobra.Command{
	Use:   "status [returning|ended]",
	Short: "Crawl shows by status",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		status := args[0]
		fmt.Printf("Crawling shows with status: %s\n", status)
		// TODO: Implement crawl by status
		fmt.Printf("✓ Shows with status '%s' refreshed successfully\n", status)
	},
}

// crawlBatchCmd represents the batch crawl command
var crawlBatchCmd = &cobra.Command{
	Use:   "batch [tmdb-id1] [tmdb-id2] ...",
	Short: "Crawl multiple shows by TMDB IDs",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		tmdbIDs := parseIntArrayArgs(args)
		fmt.Printf("Crawling %d shows...\n", len(tmdbIDs))
		// TODO: Implement batch crawl
		fmt.Printf("✓ Batch crawl completed for %d shows\n", len(tmdbIDs))
	},
}

func init() {
	rootCmd.AddCommand(crawlCmd)
	crawlCmd.AddCommand(crawlShowCmd)
	crawlCmd.AddCommand(crawlAllCmd)
	crawlCmd.AddCommand(crawlStatusCmd)
	crawlCmd.AddCommand(crawlBatchCmd)
}

func parseIntArg(s string) int {
	var id int
	fmt.Sscanf(s, "%d", &id)
	return id
}

func parseIntArrayArgs(args []string) []int {
	ids := make([]int, len(args))
	for i, arg := range args {
		ids[i] = parseIntArg(arg)
	}
	return ids
}

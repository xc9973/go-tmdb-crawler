package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "tmdb-crawler",
	Short: "A TMDB TV show crawler and manager",
	Long: `TMDB Crawler is a CLI tool for managing TV shows from TMDB.
It can crawl show information, track updates, and generate update lists.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize()
}

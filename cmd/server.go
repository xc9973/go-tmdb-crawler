package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/xc9973/go-tmdb-crawler/api"
	"github.com/xc9973/go-tmdb-crawler/config"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the web server",
	Long:  `Start the web server for the TMDB crawler application`,
	Run: func(cmd *cobra.Command, args []string) {
		// Load configuration
		cfg, err := config.Load()
		if err != nil {
			log.Fatalf("Failed to load config: %v", err)
		}

		// Initialize router (SetupRouter handles all database and service initialization)
		router := api.SetupRouter(cfg)

		// Start server
		addr := fmt.Sprintf(":%d", cfg.App.Port)
		fmt.Printf("Server starting on http://localhost%s\n", addr)

		if err := router.Run(addr); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
}

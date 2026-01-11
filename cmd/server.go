package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/xc9973/go-tmdb-crawler/api"
	"github.com/xc9973/go-tmdb-crawler/config"
	"github.com/xc9973/go-tmdb-crawler/repositories"
	"github.com/xc9973/go-tmdb-crawler/services"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
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

		// Initialize database
		var db *gorm.DB
		if cfg.Database.Type == "sqlite" {
			db, err = gorm.Open(sqlite.Open(cfg.Database.Path), &gorm.Config{})
		} else {
			// For PostgreSQL support in the future
			log.Fatalf("PostgreSQL not yet implemented, use SQLite")
		}

		if err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}

		// Initialize repositories
		showRepo := repositories.NewShowRepository(db)
		episodeRepo := repositories.NewEpisodeRepository(db)
		crawlLogRepo := repositories.NewCrawlLogRepository(db)

		// Initialize services (placeholder for now)
		_ = services.NewTMDBService(cfg.TMDB.APIKey, cfg.TMDB.BaseURL, cfg.TMDB.Language)
		_ = services.NewCrawlerService(nil, showRepo, episodeRepo, crawlLogRepo)

		// Initialize router
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

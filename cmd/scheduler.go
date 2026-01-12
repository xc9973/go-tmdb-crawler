package cmd

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/xc9973/go-tmdb-crawler/config"
	"github.com/xc9973/go-tmdb-crawler/repositories"
	"github.com/xc9973/go-tmdb-crawler/services"
	"github.com/xc9973/go-tmdb-crawler/utils"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var schedulerCmd = &cobra.Command{
	Use:   "scheduler",
	Short: "Run the scheduler service",
	Long:  `Run the scheduler service to execute scheduled crawl and publish tasks`,
	Run: func(cmd *cobra.Command, args []string) {
		// Load configuration
		cfg, err := config.Load()
		if err != nil {
			log.Fatalf("Failed to load config: %v", err)
		}

		// Check if scheduler is enabled
		if !cfg.Scheduler.Enabled {
			log.Println("Scheduler is disabled in configuration")
			log.Println("Set ENABLE_SCHEDULER=true to enable")
			return
		}

		// Initialize database
		db, err := gorm.Open(sqlite.Open(cfg.Database.Path), &gorm.Config{})
		if err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}

		// Initialize repositories
		showRepo := repositories.NewShowRepository(db)
		episodeRepo := repositories.NewEpisodeRepository(db)
		crawlLogRepo := repositories.NewCrawlLogRepository(db)
		crawlTaskRepo := repositories.NewCrawlTaskRepository(db)
		telegraphPostRepo := repositories.NewTelegraphPostRepository(db)

		// Initialize services
		logger := utils.NewLogger(cfg.App.LogLevel, cfg.Paths.Log)
		tmdb := services.MustTMDBService(cfg.TMDB.APIKey, cfg.TMDB.BaseURL, cfg.TMDB.Language)
		crawler := services.NewCrawlerService(tmdb, showRepo, episodeRepo, crawlLogRepo, crawlTaskRepo)
		telegraph := services.NewTelegraphService(cfg.Telegraph.Token, cfg.Telegraph.AuthorName, cfg.Telegraph.AuthorURL)
		publisher := services.NewPublisherService(telegraph, showRepo, episodeRepo, telegraphPostRepo, nil)

		// Initialize scheduler
		scheduler := services.NewScheduler(crawler, publisher, logger)

		// Start scheduler
		log.Println("Starting scheduler service...")
		if err := scheduler.Start(); err != nil {
			log.Fatalf("Failed to start scheduler: %v", err)
		}

		log.Println("Scheduler started successfully")
		log.Println("Press Ctrl+C to stop")

		// Print next run times
		nextRuns := scheduler.GetNextRunTimes()
		log.Println("Scheduled tasks:")
		for job, nextRun := range nextRuns {
			log.Printf("  %s: %s", job, nextRun)
		}

		// Wait for interrupt signal
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan

		// Stop scheduler
		log.Println("Stopping scheduler...")
		scheduler.Stop()
		log.Println("Scheduler stopped")
	},
}

var schedulerRunOnceCmd = &cobra.Command{
	Use:   "run-once",
	Short: "Run crawl and publish jobs once",
	Long:  `Execute crawl and publish jobs immediately without scheduling`,
	Run: func(cmd *cobra.Command, args []string) {
		// Load configuration
		cfg, err := config.Load()
		if err != nil {
			log.Fatalf("Failed to load config: %v", err)
		}

		// Initialize database
		db, err := gorm.Open(sqlite.Open(cfg.Database.Path), &gorm.Config{})
		if err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}

		// Initialize repositories
		showRepo := repositories.NewShowRepository(db)
		episodeRepo := repositories.NewEpisodeRepository(db)
		crawlLogRepo := repositories.NewCrawlLogRepository(db)
		crawlTaskRepo := repositories.NewCrawlTaskRepository(db)
		telegraphPostRepo := repositories.NewTelegraphPostRepository(db)

		// Initialize services
		tmdb := services.MustTMDBService(cfg.TMDB.APIKey, cfg.TMDB.BaseURL, cfg.TMDB.Language)
		crawler := services.NewCrawlerService(tmdb, showRepo, episodeRepo, crawlLogRepo, crawlTaskRepo)
		telegraph := services.NewTelegraphService(cfg.Telegraph.Token, cfg.Telegraph.AuthorName, cfg.Telegraph.AuthorURL)
		publisher := services.NewPublisherService(telegraph, showRepo, episodeRepo, telegraphPostRepo, nil)

		// Run crawl job
		log.Println("Running crawl job...")
		if err := crawler.RefreshAll(); err != nil {
			log.Printf("Crawl job failed: %v", err)
		} else {
			log.Println("Crawl job completed successfully")
		}

		// Run publish job
		log.Println("Running publish job...")
		result, err := publisher.PublishTodayUpdates()
		if err != nil {
			log.Printf("Publish job failed: %v", err)
		} else if result.Success {
			log.Printf("Publish job completed successfully: %s", result.URL)
		} else {
			log.Printf("Publish job skipped: %v", result.Error)
		}
	},
}

var schedulerStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show scheduler status",
	Long:  `Display the current status of the scheduler and next run times`,
	Run: func(cmd *cobra.Command, args []string) {
		// Load configuration
		cfg, err := config.Load()
		if err != nil {
			log.Fatalf("Failed to load config: %v", err)
		}

		// Get default cron specs
		specs := services.GetDefaultCronSpecs()

		fmt.Println("\n=== Scheduler Configuration ===")
		fmt.Printf("Enabled: %v\n", cfg.Scheduler.Enabled)
		fmt.Printf("Timezone: %s\n", cfg.Scheduler.TZ)
		fmt.Println("\nScheduled Tasks:")
		for jobType, spec := range specs {
			fmt.Printf("  %s: %s\n", jobType, spec)
		}

		fmt.Println("\n=== Default Cron Specifications ===")
		fmt.Println("Daily Crawl: 0 0 8,12,20 * * * (8am, 12pm, 8pm)")
		fmt.Println("Daily Publish: 0 30 20 * * * (8:30pm)")
		fmt.Println("Weekly Crawl: 0 0 6 * * 1 (Monday 6am)")
		fmt.Println("Weekly Publish: 0 0 7 * * 1 (Monday 7am)")
	},
}

func init() {
	rootCmd.AddCommand(schedulerCmd)
	schedulerCmd.AddCommand(schedulerRunOnceCmd)
	schedulerCmd.AddCommand(schedulerStatusCmd)
}

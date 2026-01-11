package services

import (
	"fmt"
	"strings"
	"time"

	"github.com/xc9973/go-tmdb-crawler/models"
	"github.com/xc9973/go-tmdb-crawler/repositories"
)

// MarkdownService handles Markdown content generation
type MarkdownService struct {
	episodeRepo repositories.EpisodeRepository
	showRepo    repositories.ShowRepository
}

// NewMarkdownService creates a new Markdown service instance
func NewMarkdownService(
	episodeRepo repositories.EpisodeRepository,
	showRepo repositories.ShowRepository,
) *MarkdownService {
	return &MarkdownService{
		episodeRepo: episodeRepo,
		showRepo:    showRepo,
	}
}

// GenerateTodayUpdates generates Markdown content for today's updates
func (s *MarkdownService) GenerateTodayUpdates() (string, error) {
	episodes, err := s.episodeRepo.GetTodayUpdates()
	if err != nil {
		return "", fmt.Errorf("failed to get today's episodes: %w", err)
	}

	return s.GenerateUpdateList(episodes), nil
}

// GenerateUpdateList generates Markdown content for a list of episodes
func (s *MarkdownService) GenerateUpdateList(episodes []*models.Episode) string {
	var builder strings.Builder

	// Header
	today := time.Now().Format("2006年01月02日")
	builder.WriteString(fmt.Sprintf("# 📺 今日更新清单\n\n"))
	builder.WriteString(fmt.Sprintf("**📅 更新日期**: %s\n\n", today))
	builder.WriteString("---\n\n")

	// Group by show
	showMap := make(map[string][]*models.Episode)
	for _, episode := range episodes {
		showName := episode.Show.Name
		showMap[showName] = append(showMap[showName], episode)
	}

	// Generate content for each show
	for showName, showEpisodes := range showMap {
		builder.WriteString(fmt.Sprintf("## %s\n\n", showName))

		for _, ep := range showEpisodes {
			episodeCode := ep.GetEpisodeCode()
			airDate := ""
			if ep.AirDate != nil {
				airDate = ep.AirDate.Format("2006-01-02")
			}

			builder.WriteString(fmt.Sprintf("### %s - %s\n", episodeCode, ep.Name))
			if airDate != "" {
				builder.WriteString(fmt.Sprintf("**播出日期**: %s\n", airDate))
			}
			if ep.Overview != "" {
				builder.WriteString(fmt.Sprintf("**简介**: %s\n", ep.Overview))
			}
			builder.WriteString("\n")
		}

		builder.WriteString("---\n\n")
	}

	// Footer
	builder.WriteString(fmt.Sprintf("📊 **统计**: 共 %d 部剧集, %d 集更新\n\n", len(showMap), len(episodes)))
	builder.WriteString("*数据来源: TMDB*")

	return builder.String()
}

// GenerateShowDetail generates detailed Markdown content for a single show
func (s *MarkdownService) GenerateShowDetail(showID uint) (string, error) {
	show, err := s.showRepo.GetByID(showID)
	if err != nil {
		return "", fmt.Errorf("failed to get show: %w", err)
	}

	episodes, err := s.episodeRepo.GetByShowID(showID)
	if err != nil {
		return "", fmt.Errorf("failed to get episodes: %w", err)
	}

	return s.GenerateShowContent(show, episodes), nil
}

// GenerateShowContent generates Markdown content for a show with episodes
func (s *MarkdownService) GenerateShowContent(show *models.Show, episodes []*models.Episode) string {
	var builder strings.Builder

	// Header
	builder.WriteString(fmt.Sprintf("# %s\n\n", show.Name))
	if show.OriginalName != "" && show.OriginalName != show.Name {
		builder.WriteString(fmt.Sprintf("**原名**: %s\n\n", show.OriginalName))
	}

	// Show info
	builder.WriteString("## 📋 剧集信息\n\n")
	builder.WriteString(fmt.Sprintf("- **状态**: %s\n", show.GetDisplayStatus()))
	builder.WriteString(fmt.Sprintf("- **类型**: %s\n", show.GetDisplayType()))
	if show.FirstAirDate != nil {
		builder.WriteString(fmt.Sprintf("- **首播日期**: %s\n", show.FirstAirDate.Format("2006-01-02")))
	}
	builder.WriteString(fmt.Sprintf("- **评分**: %.1f/10 (%d票)\n", show.VoteAverage, show.VoteCount))
	builder.WriteString(fmt.Sprintf("- **语言**: %s\n", show.Language))
	builder.WriteString("\n")

	// Overview
	if show.Overview != "" {
		builder.WriteString("## 📖 简介\n\n")
		builder.WriteString(show.Overview)
		builder.WriteString("\n\n")
	}

	// Episodes
	if len(episodes) > 0 {
		builder.WriteString("## 🎬 剧集列表\n\n")

		// Group by season
		seasonMap := make(map[int][]*models.Episode)
		for _, ep := range episodes {
			seasonMap[int(ep.SeasonNumber)] = append(seasonMap[int(ep.SeasonNumber)], ep)
		}

		// Sort by season
		seasons := make([]int, 0, len(seasonMap))
		for season := range seasonMap {
			seasons = append(seasons, season)
		}

		for _, season := range seasons {
			seasonEpisodes := seasonMap[season]
			builder.WriteString(fmt.Sprintf("### 第%d季 (%d集)\n\n", season, len(seasonEpisodes)))

			for _, ep := range seasonEpisodes {
				episodeCode := ep.GetEpisodeCode()
				airDate := ""
				if ep.AirDate != nil {
					airDate = ep.AirDate.Format("2006-01-02")
				}

				builder.WriteString(fmt.Sprintf("**%s** - %s", episodeCode, ep.Name))
				if airDate != "" {
					builder.WriteString(fmt.Sprintf(" *(%s)*", airDate))
				}
				builder.WriteString("\n")

				if ep.Overview != "" {
					builder.WriteString(fmt.Sprintf("> %s\n", ep.Overview))
				}
				builder.WriteString("\n")
			}
		}
	}

	// Footer
	builder.WriteString("---\n\n")
	builder.WriteString(fmt.Sprintf("*TMDB ID: %d | 最后更新: %s*",
		show.TmdbID,
		time.Now().Format("2006-01-02 15:04:05")))

	return builder.String()
}

// GenerateWeeklyUpdates generates Markdown content for weekly updates
func (s *MarkdownService) GenerateWeeklyUpdates() (string, error) {
	today := time.Now().Truncate(24 * time.Hour)
	startDate := today.AddDate(0, 0, -7)

	episodes, err := s.episodeRepo.GetByDateRange(startDate, today)
	if err != nil {
		return "", fmt.Errorf("failed to get episodes: %w", err)
	}

	return s.GenerateDateRangeUpdates(startDate, today, episodes), nil
}

// GenerateDateRangeUpdates generates Markdown content for a date range
func (s *MarkdownService) GenerateDateRangeUpdates(startDate, endDate time.Time, episodes []*models.Episode) string {
	var builder strings.Builder

	// Header
	builder.WriteString(fmt.Sprintf("# 📺 更新清单\n\n"))
	builder.WriteString(fmt.Sprintf("**📅 日期范围**: %s 至 %s\n\n",
		startDate.Format("2006-01-02"),
		endDate.Format("2006-01-02")))
	builder.WriteString("---\n\n")

	// Group by date
	dateMap := make(map[string][]*models.Episode)
	for _, episode := range episodes {
		if episode.AirDate != nil {
			dateStr := episode.AirDate.Format("2006-01-02")
			dateMap[dateStr] = append(dateMap[dateStr], episode)
		}
	}

	// Generate content for each date
	for dateStr, dateEpisodes := range dateMap {
		builder.WriteString(fmt.Sprintf("## 📅 %s\n\n", dateStr))

		// Group by show
		showMap := make(map[string][]*models.Episode)
		for _, ep := range dateEpisodes {
			showName := ep.Show.Name
			showMap[showName] = append(showMap[showName], ep)
		}

		// List shows for this date
		for showName, showEpisodes := range showMap {
			builder.WriteString(fmt.Sprintf("### %s\n\n", showName))

			for _, ep := range showEpisodes {
				episodeCode := ep.GetEpisodeCode()
				builder.WriteString(fmt.Sprintf("- **%s** - %s\n", episodeCode, ep.Name))
			}
			builder.WriteString("\n")
		}

		builder.WriteString("---\n\n")
	}

	// Footer
	showCount := make(map[uint]bool)
	for _, ep := range episodes {
		showCount[ep.ShowID] = true
	}

	builder.WriteString(fmt.Sprintf("📊 **统计**: 共 %d 部剧集, %d 集更新\n\n", len(showCount), len(episodes)))
	builder.WriteString("*数据来源: TMDB*")

	return builder.String()
}

// SaveToFile saves Markdown content to a file
func (s *MarkdownService) SaveToFile(content, filename string) error {
	// This would typically use os.WriteFile
	// For now, return success as file writing will be handled by the caller
	return fmt.Errorf("file saving not implemented in service layer")
}

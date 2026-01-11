package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/xc9973/go-tmdb-crawler/config"
	"github.com/xc9973/go-tmdb-crawler/models"
	"github.com/xc9973/go-tmdb-crawler/repositories"
	"github.com/xc9973/go-tmdb-crawler/utils"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// MigrationStats 迁移统计
type MigrationStats struct {
	ShowsProcessed    int
	ShowsSuccess      int
	ShowsFailed       int
	EpisodesProcessed int
	EpisodesSuccess   int
	EpisodesFailed    int
	StartTime         time.Time
	EndTime           time.Time
}

func main() {
	stats := &MigrationStats{
		StartTime: time.Now(),
	}

	fmt.Println("============================================================")
	fmt.Println("数据导入脚本")
	fmt.Println("============================================================")
	fmt.Println()

	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("错误: 加载配置失败: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志
	logger := utils.NewLogger("import", cfg.App.LogLevel)

	// 连接数据库(使用SQLite)
	db, err := gorm.Open(sqlite.Open(cfg.GetSQLitePath()), &gorm.Config{})
	if err != nil {
		fmt.Printf("错误: 连接数据库失败: %v\n", err)
		os.Exit(1)
	}

	// 自动迁移表结构
	if err := db.AutoMigrate(&models.Show{}, &models.Episode{}); err != nil {
		fmt.Printf("错误: 数据库迁移失败: %v\n", err)
		os.Exit(1)
	}

	// 初始化仓储
	showRepo := repositories.NewShowRepository(db)
	episodeRepo := repositories.NewEpisodeRepository(db)

	// 导入剧集数据
	fmt.Println("开始导入剧集数据...")
	if err := importShows(showRepo, stats); err != nil {
		fmt.Printf("错误: 导入剧集失败: %v\n", err)
		os.Exit(1)
	}

	// 导入剧集详情
	fmt.Println("\n开始导入剧集详情...")
	if err := importEpisodes(episodeRepo, showRepo, stats); err != nil {
		fmt.Printf("错误: 导入剧集详情失败: %v\n", err)
		os.Exit(1)
	}

	stats.EndTime = time.Now()

	// 打印统计报告
	printStats(stats)
	logger.Info("数据导入完成")
}

// importShows 导入剧集数据
func importShows(showRepo repositories.ShowRepository, stats *MigrationStats) error {
	file, err := os.Open("scripts/migrate/output/shows.csv")
	if err != nil {
		return fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = ','

	// 读取表头
	headers, err := reader.Read()
	if err != nil {
		return fmt.Errorf("读取表头失败: %w", err)
	}

	// 处理UTF-8 BOM
	if len(headers) > 0 && len(headers[0]) > 0 {
		firstRune := []rune(headers[0])[0]
		if firstRune == '\ufeff' {
			headers[0] = string([]rune(headers[0])[1:])
		}
	}

	fmt.Printf("找到列: %v\n", headers)

	// 创建列索引映射
	colIndex := make(map[string]int)
	for i, col := range headers {
		colIndex[col] = i
	}

	// 读取数据
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("读取数据失败: %w", err)
	}

	stats.ShowsProcessed = len(records)

	for i, record := range records {
		fmt.Printf("处理剧集 %d/%d...\n", i+1, len(records))

		// 提取TMDB ID
		tmdbIDStr := getFieldValue(record, colIndex, "TMDB_ID")
		tmdbID, err := strconv.Atoi(tmdbIDStr)
		if err != nil {
			fmt.Printf("  跳过: 无效的TMDB ID: %s\n", tmdbIDStr)
			stats.ShowsFailed++
			continue
		}

		// 检查是否已存在
		existing, err := showRepo.GetByTmdbID(tmdbID)
		if err == nil && existing != nil {
			fmt.Printf("  跳过: 剧集已存在 (TMDB ID: %d)\n", tmdbID)
			stats.ShowsSuccess++
			continue
		}

		// 创建剧集对象
		show := &models.Show{
			TmdbID:       tmdbID,
			Name:         getFieldValue(record, colIndex, "名称"),
			OriginalName: getFieldValue(record, colIndex, "原名"),
			Status:       getFieldValue(record, colIndex, "状态"),
			Overview:     getFieldValue(record, colIndex, "简介"),
			PosterPath:   getFieldValue(record, colIndex, "海报路径"),
			BackdropPath: getFieldValue(record, colIndex, "背景路径"),
			Genres:       getFieldValue(record, colIndex, "类型"),
		}

		// 解析评分
		if voteAvgStr := getFieldValue(record, colIndex, "评分"); voteAvgStr != "" {
			if voteAvg, err := strconv.ParseFloat(voteAvgStr, 32); err == nil {
				show.VoteAverage = float32(voteAvg)
			}
		}

		// 保存到数据库
		if err := showRepo.Create(show); err != nil {
			fmt.Printf("  错误: 保存失败: %v\n", err)
			stats.ShowsFailed++
			continue
		}

		fmt.Printf("  ✅ 成功导入: %s (ID: %d)\n", show.Name, show.ID)
		stats.ShowsSuccess++
	}

	return nil
}

// importEpisodes 导入剧集详情
func importEpisodes(
	episodeRepo repositories.EpisodeRepository,
	showRepo repositories.ShowRepository,
	stats *MigrationStats,
) error {
	file, err := os.Open("scripts/migrate/output/episodes.csv")
	if err != nil {
		return fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = ','

	// 读取表头
	headers, err := reader.Read()
	if err != nil {
		return fmt.Errorf("读取表头失败: %w", err)
	}

	// 处理UTF-8 BOM
	if len(headers) > 0 && len(headers[0]) > 0 {
		firstRune := []rune(headers[0])[0]
		if firstRune == '\ufeff' {
			headers[0] = string([]rune(headers[0])[1:])
		}
	}

	fmt.Printf("找到列: %v\n", headers)

	// 创建列索引映射
	colIndex := make(map[string]int)
	for i, col := range headers {
		colIndex[col] = i
	}

	// 读取数据
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("读取数据失败: %w", err)
	}

	stats.EpisodesProcessed = len(records)

	for i, record := range records {
		fmt.Printf("处理剧集 %d/%d...\n", i+1, len(records))

		// 提取TMDB ID
		tmdbIDStr := getFieldValue(record, colIndex, "TMDB_ID")
		tmdbID, err := strconv.Atoi(tmdbIDStr)
		if err != nil {
			fmt.Printf("  跳过: 无效的TMDB ID: %s\n", tmdbIDStr)
			stats.EpisodesFailed++
			continue
		}

		// 查找对应的剧集
		show, err := showRepo.GetByTmdbID(tmdbID)
		if err != nil {
			fmt.Printf("  跳过: 找不到剧集 (TMDB ID: %d)\n", tmdbID)
			stats.EpisodesFailed++
			continue
		}

		// 提取季数和集数
		seasonStr := getFieldValue(record, colIndex, "季数")
		episodeStr := getFieldValue(record, colIndex, "集数")

		seasonNumber, _ := strconv.Atoi(seasonStr)
		episodeNumber, _ := strconv.Atoi(episodeStr)

		if seasonNumber == 0 || episodeNumber == 0 {
			fmt.Printf("  跳过: 无效的季数或集数 (S%dE%d)\n", seasonNumber, episodeNumber)
			stats.EpisodesFailed++
			continue
		}

		// 检查是否已存在(通过查询该季度所有剧集)
		episodes, err := episodeRepo.GetBySeason(show.ID, seasonNumber)
		if err == nil {
			for _, ep := range episodes {
				if ep.EpisodeNumber == episodeNumber {
					fmt.Printf("  跳过: 剧集已存在 (S%dE%d)\n", seasonNumber, episodeNumber)
					stats.EpisodesSuccess++
					continue
				}
			}
		}

		// 解析播出日期
		var airDate *time.Time
		if airDateStr := getFieldValue(record, colIndex, "播出日期"); airDateStr != "" {
			if t, err := time.Parse("2006-01-02", airDateStr); err == nil {
				airDate = &t
			}
		}

		// 创建剧集对象
		episode := &models.Episode{
			ShowID:        show.ID,
			SeasonNumber:  seasonNumber,
			EpisodeNumber: episodeNumber,
			Name:          getFieldValue(record, colIndex, "名称"),
			Overview:      getFieldValue(record, colIndex, "简介"),
			StillPath:     getFieldValue(record, colIndex, "剧照路径"),
			AirDate:       airDate,
		}

		// 解析评分
		if voteAvgStr := getFieldValue(record, colIndex, "评分"); voteAvgStr != "" {
			if voteAvg, err := strconv.ParseFloat(voteAvgStr, 32); err == nil {
				episode.VoteAverage = float32(voteAvg)
			}
		}

		// 保存到数据库
		if err := episodeRepo.Create(episode); err != nil {
			fmt.Printf("  错误: 保存失败: %v\n", err)
			stats.EpisodesFailed++
			continue
		}

		fmt.Printf("  ✅ 成功导入: S%dE%d - %s\n", seasonNumber, episodeNumber, episode.Name)
		stats.EpisodesSuccess++
	}

	return nil
}

// getFieldValue 获取字段值
func getFieldValue(record []string, colIndex map[string]int, fieldName string) string {
	if idx, ok := colIndex[fieldName]; ok && idx < len(record) {
		return record[idx]
	}
	return ""
}

// printStats 打印统计信息
func printStats(stats *MigrationStats) {
	fmt.Println("\n============================================================")
	fmt.Println("迁移统计报告")
	fmt.Println("============================================================")
	fmt.Printf("开始时间: %s\n", stats.StartTime.Format("2006-01-02 15:04:05"))
	fmt.Printf("结束时间: %s\n", stats.EndTime.Format("2006-01-02 15:04:05"))
	fmt.Printf("总耗时: %v\n\n", stats.EndTime.Sub(stats.StartTime))

	fmt.Println("剧集数据:")
	fmt.Printf("  处理: %d\n", stats.ShowsProcessed)
	fmt.Printf("  成功: %d\n", stats.ShowsSuccess)
	fmt.Printf("  失败: %d\n\n", stats.ShowsFailed)

	fmt.Println("剧集详情:")
	fmt.Printf("  处理: %d\n", stats.EpisodesProcessed)
	fmt.Printf("  成功: %d\n", stats.EpisodesSuccess)
	fmt.Printf("  失败: %d\n\n", stats.EpisodesFailed)

	fmt.Println("============================================================")

	if stats.ShowsFailed > 0 || stats.EpisodesFailed > 0 {
		fmt.Println("⚠️  部分数据导入失败,请检查日志")
	} else {
		fmt.Println("✅ 所有数据导入成功!")
	}
}

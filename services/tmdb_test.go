package services

import (
	"testing"
)

// TestTMDBService_GetShowDetails 测试获取剧集详情
func TestTMDBService_GetShowDetails(t *testing.T) {
	// 跳过实际API调用测试,需要有效的API密钥
	t.Skip("需要TMDB API密钥")

	/*
		service := NewTMDBService("your-api-key")

		// 测试获取已知剧集
		show, err := service.GetShowDetails(1668) // Friends
		if err != nil {
			t.Fatalf("Failed to get show details: %v", err)
		}

		if show.Name != "Friends" {
			t.Errorf("Expected name 'Friends', got '%s'", show.Name)
		}

		if show.ID != 1668 {
			t.Errorf("Expected ID 1668, got %d", show.ID)
		}
	*/
}

// TestTMDBService_SearchShow 测试搜索剧集
func TestTMDBService_SearchShow(t *testing.T) {
	t.Skip("需要TMDB API密钥")

	/*
		service := NewTMDBService("your-api-key")

		results, err := service.SearchShow("Breaking Bad")
		if err != nil {
			t.Fatalf("Failed to search show: %v", err)
		}

		if len(results) == 0 {
			t.Error("Expected at least one result")
		}
	*/
}

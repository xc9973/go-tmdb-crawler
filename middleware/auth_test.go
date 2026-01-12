package middleware

import (
	"testing"
)

// TestFailedAttemptsCleanup 测试失败记录的惰性清理功能
func TestFailedAttemptsCleanup(t *testing.T) {
	auth := NewAdminAuth("test-key", true)

	// 模拟多个不同 IP 的失败尝试
	ips := []string{
		"192.168.1.1",
		"192.168.1.2",
		"192.168.1.3",
		"192.168.1.4",
		"192.168.1.5",
	}

	// 为每个 IP 记录失败
	for _, ip := range ips {
		auth.recordFailure(ip)
	}

	// 检查记录数
	stats := auth.GetFailedAttemptsStats()
	if stats["total_records"] != len(ips) {
		t.Errorf("Expected %d records, got %d", len(ips), stats["total_records"])
	}

	// 检查活跃记录数
	if stats["active_count"] != len(ips) {
		t.Errorf("Expected %d active records, got %d", len(ips), stats["active_count"])
	}

	t.Logf("Stats: %+v", stats)
}

// TestFailedAttemptsStats 测试统计信息
func TestFailedAttemptsStats(t *testing.T) {
	auth := NewAdminAuth("test-key", true)

	// 初始状态应该没有记录
	stats := auth.GetFailedAttemptsStats()
	if stats["total_records"] != 0 {
		t.Errorf("Expected 0 records initially, got %d", stats["total_records"])
	}

	// 添加一些失败记录
	auth.recordFailure("192.168.1.1")
	auth.recordFailure("192.168.1.1")
	auth.recordFailure("192.168.1.2")

	stats = auth.GetFailedAttemptsStats()
	if stats["total_records"] != 2 {
		t.Errorf("Expected 2 records, got %d", stats["total_records"])
	}

	if stats["active_count"] != 2 {
		t.Errorf("Expected 2 active records, got %d", stats["active_count"])
	}

	t.Logf("Stats: %+v", stats)
}

// TestFailedAttemptsBan 测试封禁功能
func TestFailedAttemptsBan(t *testing.T) {
	auth := NewAdminAuth("test-key", true)

	ip := "192.168.1.1"

	// 记录 5 次失败（应该被封禁）
	for i := 0; i < 5; i++ {
		auth.recordFailure(ip)
	}

	// 检查统计信息
	stats := auth.GetFailedAttemptsStats()
	if stats["blocked_count"] != 1 {
		t.Errorf("Expected 1 blocked record, got %d", stats["blocked_count"])
	}

	t.Logf("Stats after 5 failures: %+v", stats)
}

// TestFailedAttemptsClear 测试清除失败记录
func TestFailedAttemptsClear(t *testing.T) {
	auth := NewAdminAuth("test-key", true)

	ip := "192.168.1.1"

	// 记录失败
	auth.recordFailure(ip)
	auth.recordFailure(ip)

	// 检查记录存在
	stats := auth.GetFailedAttemptsStats()
	if stats["active_count"] != 1 {
		t.Errorf("Expected 1 active record, got %d", stats["active_count"])
	}

	// 清除失败记录
	auth.clearFailure(ip)

	// 检查记录已清除
	stats = auth.GetFailedAttemptsStats()
	if stats["active_count"] != 0 {
		t.Errorf("Expected 0 active records after clear, got %d", stats["active_count"])
	}

	// 但记录仍然存在（只是计数被清零）
	if stats["total_records"] != 1 {
		t.Errorf("Expected 1 total record (count cleared but record exists), got %d", stats["total_records"])
	}

	t.Logf("Stats after clear: %+v", stats)
}

// TestCleanupExpiredAttemptsLocked 测试过期记录清理
func TestCleanupExpiredAttemptsLocked(t *testing.T) {
	auth := NewAdminAuth("test-key", true)

	// 添加一些失败记录
	ips := []string{
		"192.168.1.1",
		"192.168.1.2",
		"192.168.1.3",
	}

	for _, ip := range ips {
		auth.recordFailure(ip)
	}

	// 检查初始记录数
	stats := auth.GetFailedAttemptsStats()
	if stats["total_records"] != len(ips) {
		t.Errorf("Expected %d records, got %d", len(ips), stats["total_records"])
	}

	// 手动调用清理（限制清理 2 条）
	auth.mu.Lock()
	auth.cleanupExpiredAttemptsLocked(2)
	auth.mu.Unlock()

	// 检查记录数（应该没有变化，因为记录还未过期）
	stats = auth.GetFailedAttemptsStats()
	if stats["total_records"] != len(ips) {
		t.Errorf("Expected %d records (not expired yet), got %d", len(ips), stats["total_records"])
	}

	t.Logf("Stats after cleanup attempt: %+v", stats)
}

// TestLazyCleanup 测试惰性清理机制
func TestLazyCleanup(t *testing.T) {
	auth := NewAdminAuth("test-key", true)

	// 模拟大量不同 IP 的失败尝试
	// 每次调用 recordFailure 都会触发惰性清理（最多清理 10 条）
	for i := 0; i < 100; i++ {
		ip := "192.168.1." + string(rune('1'+i%10))
		auth.recordFailure(ip)
	}

	// 检查记录数
	stats := auth.GetFailedAttemptsStats()
	t.Logf("Stats after 100 failures: %+v", stats)

	// 记录数应该远小于 100，因为每次都会清理一些过期记录
	// 但由于所有记录都是新创建的，所以不会被清理
	// 这个测试主要验证代码不会崩溃
	if stats["total_records"] == 0 {
		t.Error("Expected some records to exist")
	}
}

// BenchmarkFailedAttempts 性能基准测试
func BenchmarkFailedAttempts(b *testing.B) {
	auth := NewAdminAuth("test-key", true)

	// 预先添加一些记录
	for i := 0; i < 1000; i++ {
		ip := "192.168.1." + string(rune('1'+i%10))
		auth.recordFailure(ip)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		auth.recordFailure("192.168.1.1")
	}
}

// BenchmarkGetFailedAttemptsStats 统计信息性能基准测试
func BenchmarkGetFailedAttemptsStats(b *testing.B) {
	auth := NewAdminAuth("test-key", true)

	// 预先添加一些记录
	for i := 0; i < 1000; i++ {
		ip := "192.168.1." + string(rune('1'+i%10))
		auth.recordFailure(ip)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		auth.GetFailedAttemptsStats()
	}
}

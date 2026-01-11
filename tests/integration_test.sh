#!/bin/bash

# TMDB剧集爬取系统 - 集成测试脚本
# 测试日期: 2026-01-11

set -e  # 遇到错误立即退出

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 测试配置
PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_DIR"

SERVER_PID=""
TEST_REPORT="$PROJECT_DIR/tests/test_report.md"
TEST_RESULTS=()

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# 初始化测试报告
init_report() {
    cat > "$TEST_REPORT" << EOF
# 集成测试报告

**测试日期**: $(date +"%Y-%m-%d %H:%M:%S")
**测试环境**: 开发环境
**测试人员**: 自动化测试脚本

---

## 测试结果总结

| 场景 | 状态 | 备注 |
|------|------|------|
EOF
}

# 添加测试结果到报告
add_result() {
    local scenario="$1"
    local status="$2"
    local notes="$3"
    
    echo "| $scenario | $status | $notes |" >> "$TEST_REPORT"
    
    if [ "$status" = "✅ 通过" ]; then
        TEST_RESULTS+=("✅ $scenario")
        log_success "$scenario - $notes"
    else
        TEST_RESULTS+=("❌ $scenario")
        log_error "$scenario - $notes"
    fi
}

# 环境检查
check_environment() {
    log_info "检查测试环境..."
    
    # 检查Go是否安装
    if ! command -v go &> /dev/null; then
        log_error "Go未安装"
        exit 1
    fi
    
    # 检查.postgreSQL是否安装
    if ! command -v psql &> /dev/null; then
        log_warning "PostgreSQL未安装,部分测试可能失败"
    fi
    
    # 检查.env文件
    if [ ! -f "$PROJECT_DIR/.env" ]; then
        log_error ".env文件不存在,请先创建"
        exit 1
    fi
    
    log_success "环境检查通过"
}

# 编译项目
build_project() {
    log_info "编译项目..."
    
    if ! go build -o tmdb-crawler main.go 2>&1; then
        log_error "编译失败"
        add_result "项目编译" "❌ 失败" "编译错误"
        return 1
    fi
    
    log_success "项目编译成功"
    add_result "项目编译" "✅ 通过" "-"
}

# 启动服务器
start_server() {
    log_info "启动服务器..."
    
    # 检查是否已在运行
    if pgrep -f "tmdb-crawler server" > /dev/null; then
        log_warning "服务器已在运行,停止旧进程..."
        pkill -f "tmdb-crawler server" || true
        sleep 2
    fi
    
    # 启动服务器
    ./tmdb-crawler server > /tmp/tmdb-server.log 2>&1 &
    SERVER_PID=$!
    
    # 等待服务器启动
    sleep 5
    
    # 验证服务器是否运行
    if ! kill -0 $SERVER_PID 2>/dev/null; then
        log_error "服务器启动失败"
        cat /tmp/tmdb-server.log
        add_result "服务器启动" "❌ 失败" "查看 /tmp/tmdb-server.log"
        return 1
    fi
    
    log_success "服务器启动成功 (PID: $SERVER_PID)"
    add_result "服务器启动" "✅ 通过" "-"
}

# 停止服务器
stop_server() {
    if [ -n "$SERVER_PID" ]; then
        log_info "停止服务器..."
        kill $SERVER_PID 2>/dev/null || true
        wait $SERVER_PID 2>/dev/null || true
        log_success "服务器已停止"
    fi
}

# 测试场景1: API测试
test_api() {
    log_info "测试场景1: API集成测试"
    
    local api_base="http://localhost:8080/api/v1"
    local failed=0
    
    # 测试1: 获取剧集列表
    log_info "  测试1.1: 获取剧集列表"
    response=$(curl -s "$api_base/shows?page=1&page_size=10" || echo "")
    if echo "$response" | grep -q '"code":0'; then
        log_success "  获取剧集列表成功"
    else
        log_error "  获取剧集列表失败"
        failed=1
    fi
    
    # 测试2: 获取爬取日志
    log_info "  测试1.2: 获取爬取日志"
    response=$(curl -s "$api_base/crawler/logs?page=1&page_size=10" || echo "")
    if echo "$response" | grep -q '"code":0'; then
        log_success "  获取爬取日志成功"
    else
        log_error "  获取爬取日志失败"
        failed=1
    fi
    
    # 测试3: 获取今日更新
    log_info "  测试1.3: 获取今日更新"
    response=$(curl -s "$api_base/calendar/today" || echo "")
    if echo "$response" | grep -q '"code":0'; then
        log_success "  获取今日更新成功"
    else
        log_error "  获取今日更新失败"
        failed=1
    fi
    
    if [ $failed -eq 0 ]; then
        add_result "API集成测试" "✅ 通过" "所有API端点响应正常"
    else
        add_result "API集成测试" "❌ 失败" "部分API测试失败"
    fi
}

# 测试场景2: 爬取测试
test_crawl() {
    log_info "测试场景2: 爬取功能测试"
    
    # 使用一个测试用的TMDB ID (1668 = Gravity)
    local test_tmdb_id=1668
    
    log_info "  测试2.1: 爬取剧集 (TMDB ID: $test_tmdb_id)"
    if ./tmdb-crawler crawl $test_tmdb_id 2>&1 | tee /tmp/crawl_test.log; then
        log_success "  剧集爬取成功"
        add_result "爬取功能测试" "✅ 通过" "成功爬取TMDB ID: $test_tmdb_id"
    else
        log_error "  剧集爬取失败"
        cat /tmp/crawl_test.log
        add_result "爬取功能测试" "❌ 失败" "爬取失败,查看 /tmp/crawl_test.log"
    fi
}

# 测试场景3: Web界面测试
test_web_interface() {
    log_info "测试场景3: Web界面测试"
    
    local failed=0
    
    # 测试1: 主页
    log_info "  测试3.1: 访问主页"
    response=$(curl -s "http://localhost:8080/" || echo "")
    if echo "$response" | grep -q "TMDB"; then
        log_success "  主页访问成功"
    else
        log_error "  主页访问失败"
        failed=1
    fi
    
    # 测试2: 今日更新页面
    log_info "  测试3.2: 访问今日更新页面"
    response=$(curl -s "http://localhost:8080/today.html" || echo "")
    if echo "$response" | grep -q "今日更新"; then
        log_success "  今日更新页面访问成功"
    else
        log_error "  今日更新页面访问失败"
        failed=1
    fi
    
    # 测试3: 日志页面
    log_info "  测试3.3: 访问日志页面"
    response=$(curl -s "http://localhost:8080/logs.html" || echo "")
    if echo "$response" | grep -q "爬取日志"; then
        log_success "  日志页面访问成功"
    else
        log_error "  日志页面访问失败"
        failed=1
    fi
    
    if [ $failed -eq 0 ]; then
        add_result "Web界面测试" "✅ 通过" "所有页面可访问"
    else
        add_result "Web界面测试" "❌ 失败" "部分页面访问失败"
    fi
}

# 测试场景4: 错误处理测试
test_error_handling() {
    log_info "测试场景4: 错误处理测试"
    
    local failed=0
    
    # 测试1: 无效的TMDB ID
    log_info "  测试4.1: 无效的TMDB ID"
    if ./tmdb-crawler crawl 999999999 2>&1 | grep -qi "error\|failed\|not found"; then
        log_success "  正确处理无效TMDB ID"
    else
        log_error "  未正确处理无效TMDB ID"
        failed=1
    fi
    
    # 测试2: 无效的API端点
    log_info "  测试4.2: 无效的API端点"
    response=$(curl -s "http://localhost:8080/api/v1/invalid-endpoint" || echo "")
    if echo "$response" | grep -q '"404"\|"error"\|"not found"'; then
        log_success "  正确处理无效端点"
    else
        log_error "  未正确处理无效端点"
        failed=1
    fi
    
    if [ $failed -eq 0 ]; then
        add_result "错误处理测试" "✅ 通过" "错误处理正确"
    else
        add_result "错误处理测试" "❌ 失败" "部分错误处理失败"
    fi
}

# 生成测试总结
generate_summary() {
    cat >> "$TEST_REPORT" << EOF

### 发现的问题

EOF
    
    # 统计测试结果
    local passed=0
    local failed=0
    
    for result in "${TEST_RESULTS[@]}"; do
        if [[ $result == ✅* ]]; then
            ((passed++))
        else
            ((failed++))
        fi
    done
    
    cat >> "$TEST_REPORT" << EOF
**测试统计**:
- 通过: $passed
- 失败: $failed
- 总计: $((passed + failed))

### 测试结论

EOF
    
    if [ $failed -eq 0 ]; then
        cat >> "$TEST_REPORT" << EOF
✅ **所有测试通过**

系统功能完整,API响应正常,Web界面可访问,错误处理完善。可以投入使用。
EOF
    else
        cat >> "$TEST_REPORT" << EOF
⚠️ **部分测试失败**

系统存在 $failed 个测试失败,需要修复以下问题后再次测试。

建议:
1. 查看失败测试的日志
2. 修复发现的问题
3. 重新运行测试
EOF
    fi
    
    echo "" >> "$TEST_REPORT"
    echo "---" >> "$TEST_REPORT"
    echo "**测试日志位置**: /tmp/tmdb-server.log, /tmp/crawl_test.log" >> "$TEST_REPORT"
}

# 主测试流程
main() {
    echo "================================"
    echo "  TMDB剧集爬取系统 - 集成测试"
    echo "================================"
    echo ""
    
    # 初始化
    init_report
    
    # 环境检查
    check_environment
    
    # 编译项目
    build_project
    if [ $? -ne 0 ]; then
        log_error "编译失败,终止测试"
        generate_summary
        cat "$TEST_REPORT"
        exit 1
    fi
    
    # 启动服务器
    start_server
    if [ $? -ne 0 ]; then
        log_error "服务器启动失败,终止测试"
        generate_summary
        cat "$TEST_REPORT"
        exit 1
    fi
    
    # 捕获退出信号,确保服务器被停止
    trap stop_server EXIT INT TERM
    
    # 执行测试
    echo ""
    echo "开始执行测试..."
    echo ""
    
    test_api
    test_crawl
    test_web_interface
    test_error_handling
    
    # 生成总结
    echo ""
    generate_summary
    
    # 显示结果
    echo ""
    echo "================================"
    echo "  测试结果"
    echo "================================"
    echo ""
    
    for result in "${TEST_RESULTS[@]}"; do
        echo "$result"
    done
    
    echo ""
    echo "详细报告已生成: $TEST_REPORT"
    echo ""
    
    # 显示报告内容
    cat "$TEST_REPORT"
}

# 运行主流程
main

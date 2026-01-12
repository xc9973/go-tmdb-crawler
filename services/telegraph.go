package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/xc9973/go-tmdb-crawler/models"
)

// TelegraphService handles Telegraph API operations
type TelegraphService struct {
	apiURL     string
	shortName  string
	authorName string
	authorURL  string
	httpClient *http.Client
}

// NewTelegraphService creates a new Telegraph service instance
func NewTelegraphService(shortName, authorName, authorURL string) *TelegraphService {
	return &TelegraphService{
		apiURL:     "https://api.telegra.ph",
		shortName:  shortName,
		authorName: authorName,
		authorURL:  authorURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// TelegraphPage represents a Telegraph page
type TelegraphPage struct {
	Path        string `json:"path"`
	URL         string `json:"url"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Views       int    `json:"views"`
}

// TelegraphCreateRequest represents the create page request
type TelegraphCreateRequest struct {
	ShortName  string   `json:"short_name"`
	AuthorName string   `json:"author_name,omitempty"`
	AuthorURL  string   `json:"author_url,omitempty"`
	Title      string   `json:"title"`
	Content    []Node   `json:"content"`
	Tags       []string `json:"tags,omitempty"`
}

// TelegraphCreateResponse represents the create page response
type TelegraphCreateResponse struct {
	OK     bool          `json:"ok"`
	Error  string        `json:"error,omitempty"`
	Result TelegraphPage `json:"result,omitempty"`
}

// TelegraphEditRequest represents the edit page request
type TelegraphEditRequest struct {
	Path       string   `json:"path"`
	ShortName  string   `json:"short_name"`
	AuthorName string   `json:"author_name,omitempty"`
	AuthorURL  string   `json:"author_url,omitempty"`
	Title      string   `json:"title"`
	Content    []Node   `json:"content"`
	Tags       []string `json:"tags,omitempty"`
}

// TelegraphEditResponse represents the edit page response
type TelegraphEditResponse struct {
	OK     bool          `json:"ok"`
	Error  string        `json:"error,omitempty"`
	Result TelegraphPage `json:"result,omitempty"`
}

// Node represents a Telegraph content node
type Node map[string]interface{}

// NewTextNode creates a text node
func NewTextNode(text string) Node {
	return Node{
		"tag": "p",
		"children": []interface{}{
			text,
		},
	}
}

// NewBoldNode creates a bold text node
func NewBoldNode(text string) Node {
	return Node{
		"tag": "b",
		"children": []interface{}{
			text,
		},
	}
}

// NewHeaderNode creates a header node
func NewHeaderNode(level string, text string) Node {
	return Node{
		"tag": level,
		"children": []interface{}{
			text,
		},
	}
}

// NewListNode creates a list item node
func NewListNode(text string) Node {
	return Node{
		"tag": "li",
		"children": []interface{}{
			text,
		},
	}
}

// NewLinkNode creates a link node
func NewLinkNode(text, url string) Node {
	return Node{
		"tag": "a",
		"attrs": map[string]string{
			"href": url,
		},
		"children": []interface{}{
			text,
		},
	}
}

// NewHrNode creates a horizontal rule node
func NewHrNode() Node {
	return Node{
		"tag": "hr",
	}
}

// NewBrNode creates a line break node
func NewBrNode() Node {
	return Node{
		"tag": "br",
	}
}

// CreatePage creates a new Telegraph page
func (s *TelegraphService) CreatePage(title string, content []Node, tags []string) (*TelegraphPage, error) {
	req := TelegraphCreateRequest{
		ShortName:  s.shortName,
		AuthorName: s.authorName,
		AuthorURL:  s.authorURL,
		Title:      title,
		Content:    content,
		Tags:       tags,
	}

	resp, err := s.doRequest("createPage", req)
	if err != nil {
		return nil, err
	}

	var result TelegraphCreateResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !result.OK {
		return nil, fmt.Errorf("Telegraph API error: %s", result.Error)
	}

	return &result.Result, nil
}

// EditPage edits an existing Telegraph page
func (s *TelegraphService) EditPage(path, title string, content []Node, tags []string) (*TelegraphPage, error) {
	req := TelegraphEditRequest{
		Path:       path,
		ShortName:  s.shortName,
		AuthorName: s.authorName,
		AuthorURL:  s.authorURL,
		Title:      title,
		Content:    content,
		Tags:       tags,
	}

	resp, err := s.doRequest("editPage", req)
	if err != nil {
		return nil, err
	}

	var result TelegraphEditResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !result.OK {
		return nil, fmt.Errorf("Telegraph API error: %s", result.Error)
	}

	return &result.Result, nil
}

// doRequest performs an HTTP request to Telegraph API
func (s *TelegraphService) doRequest(method string, data interface{}) ([]byte, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/%s", s.apiURL, method)
	resp, err := s.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %d - %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// GenerateUpdateListContent generates content for update list
func (s *TelegraphService) GenerateUpdateListContent(episodes []*models.Episode) []Node {
	content := []Node{}

	// Title
	content = append(content, NewHeaderNode("h3", "📺 今日更新"))
	content = append(content, NewHrNode())
	content = append(content, NewBrNode())

	// Date
	today := time.Now().Format("2006年01月02日")
	content = append(content, NewTextNode(fmt.Sprintf("📅 更新日期: %s", today)))
	content = append(content, NewBrNode())
	content = append(content, NewBrNode())

	// Group by show
	showMap := make(map[string][]*models.Episode)
	for _, episode := range episodes {
		if episode == nil {
			continue
		}
		showName := "未知剧集"
		if episode.Show != nil && episode.Show.Name != "" {
			showName = episode.Show.Name
		} else if episode.ShowID != 0 {
			showName = fmt.Sprintf("ShowID:%d", episode.ShowID)
		}
		showMap[showName] = append(showMap[showName], episode)
	}

	// List episodes by show
	for showName, showEpisodes := range showMap {
		content = append(content, NewBoldNode(showName))

		for _, ep := range showEpisodes {
			if ep == nil {
				continue
			}
			episodeCode := ep.GetEpisodeCode()
			text := fmt.Sprintf("%s - %s", episodeCode, ep.Name)
			content = append(content, NewListNode(text))
		}

		content = append(content, NewBrNode())
	}

	// Footer
	content = append(content, NewHrNode())
	content = append(content, NewTextNode(fmt.Sprintf("📊 共 %d 部剧集更新", len(showMap))))
	content = append(content, NewBrNode())
	content = append(content, NewTextNode("数据来源: TMDB"))

	return content
}

// GenerateShowContent generates content for a single show
func (s *TelegraphService) GenerateShowContent(show *models.Show, episodes []*models.Episode) []Node {
	content := []Node{}

	// Title
	content = append(content, NewHeaderNode("h3", "📺 "+show.Name))
	content = append(content, NewHrNode())
	content = append(content, NewBrNode())

	// Show info
	content = append(content, NewTextNode(fmt.Sprintf("原名: %s", show.OriginalName)))
	content = append(content, NewBrNode())
	content = append(content, NewTextNode(fmt.Sprintf("状态: %s", show.GetDisplayStatus())))
	content = append(content, NewBrNode())
	content = append(content, NewTextNode(fmt.Sprintf("评分: %.1f/10 (%d票)", show.VoteAverage, show.VoteCount)))
	content = append(content, NewBrNode())
	content = append(content, NewBrNode())

	// Overview
	if show.Overview != "" {
		content = append(content, NewBoldNode("简介:"))
		content = append(content, NewTextNode(show.Overview))
		content = append(content, NewBrNode())
		content = append(content, NewBrNode())
	}

	// Episodes
	content = append(content, NewBoldNode("剧集列表:"))
	for _, ep := range episodes {
		episodeCode := ep.GetEpisodeCode()
		text := fmt.Sprintf("%s - %s", episodeCode, ep.Name)
		if ep.AirDate != nil {
			text += fmt.Sprintf(" (%s)", ep.AirDate.Format("2006-01-02"))
		}
		content = append(content, NewListNode(text))
	}

	return content
}

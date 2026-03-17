package rag

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/google/uuid"
	"github.com/walterfan/lazy-ai-coder/pkg/database"
	"github.com/walterfan/lazy-ai-coder/pkg/models"
	"gorm.io/gorm"
)

// URLLoaderConfig contains configuration for URL loading
type URLLoaderConfig struct {
	ProjectID    string
	RealmID      string
	UserID       string
	ChunkSize    int
	ChunkOverlap int
	Timeout      time.Duration
	UserAgent    string
}

// URLLoader handles loading documents from URLs
type URLLoader struct {
	config           URLLoaderConfig
	textSplitter     *TextSplitter
	embeddingService *EmbeddingService
	db               *gorm.DB
	httpClient       *http.Client
}

// NewURLLoader creates a new URL loader
func NewURLLoader(config URLLoaderConfig, embeddingConfig EmbeddingConfig) *URLLoader {
	timeout := config.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	userAgent := config.UserAgent
	if userAgent == "" {
		userAgent = "Mozilla/5.0 (compatible; AsyncLLMAgent/1.0; +https://github.com/walterfan/lazy-ai-coder)"
	}

	return &URLLoader{
		config:           config,
		textSplitter:     NewTextSplitter(config.ChunkSize, config.ChunkOverlap),
		embeddingService: NewEmbeddingService(embeddingConfig),
		db:               database.GetDB(),
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// LoadURL loads a document from a URL
func (ul *URLLoader) LoadURL(targetURL string) (*LoaderStats, error) {
	stats := &LoaderStats{
		Errors: []string{},
	}

	// Validate URL
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return stats, fmt.Errorf("invalid URL: %w", err)
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return stats, fmt.Errorf("unsupported URL scheme: %s (only http and https supported)", parsedURL.Scheme)
	}

	// Fetch and parse content
	content, title, err := ul.fetchAndExtractContent(targetURL)
	if err != nil {
		stats.Errors = append(stats.Errors, fmt.Sprintf("Error fetching URL: %v", err))
		return stats, err
	}

	if strings.TrimSpace(content) == "" {
		return stats, fmt.Errorf("no content extracted from URL")
	}

	// Split into chunks
	chunks := ul.textSplitter.SplitDocument(content, targetURL)
	stats.DocumentChunks = len(chunks)
	stats.TotalChunks = len(chunks)
	stats.FilesProcessed = 1

	// Store chunks
	_, err = ul.storeDocumentChunks(chunks, targetURL, title)
	if err != nil {
		stats.Errors = append(stats.Errors, fmt.Sprintf("Error storing chunks: %v", err))
		return stats, err
	}

	stats.EmbeddingsStored = len(chunks)
	return stats, nil
}

// fetchAndExtractContent fetches HTML and extracts main content
func (ul *URLLoader) fetchAndExtractContent(targetURL string) (content, title string, err error) {
	// Create request with custom user agent
	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		return "", "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", ul.config.UserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	// Fetch URL
	resp, err := ul.httpClient.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("HTTP error: %d %s", resp.StatusCode, resp.Status)
	}

	// Check content type
	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/html") && !strings.Contains(contentType, "application/xhtml") {
		// Try to read as plain text
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", "", fmt.Errorf("failed to read response body: %w", err)
		}
		return string(bodyBytes), targetURL, nil
	}

	// Parse HTML
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Extract title
	title = doc.Find("title").First().Text()
	if title == "" {
		title = targetURL
	}

	// Extract main content using article extraction algorithm
	content = ul.extractArticleContent(doc)

	return content, title, nil
}

// extractArticleContent extracts the main article content from HTML
func (ul *URLLoader) extractArticleContent(doc *goquery.Document) string {
	// Strategy 1: Look for common article containers
	articleSelectors := []string{
		"article",
		"[role='main']",
		"main",
		".post-content",
		".article-content",
		".entry-content",
		".content",
		"#content",
		".markdown-body", // GitHub, GitLab
		".post-body",
	}

	for _, selector := range articleSelectors {
		selection := doc.Find(selector).First()
		if selection.Length() > 0 {
			text := ul.extractTextFromSelection(selection)
			if len(text) > 200 { // Minimum content threshold
				return text
			}
		}
	}

	// Strategy 2: Find largest text block (fallback)
	var largestText string
	var largestLength int

	doc.Find("div, section, article").Each(func(i int, s *goquery.Selection) {
		// Skip navigation, footer, sidebar, etc.
		class, _ := s.Attr("class")
		id, _ := s.Attr("id")
		skipPatterns := []string{"nav", "menu", "sidebar", "footer", "header", "comment", "ad", "widget"}

		for _, pattern := range skipPatterns {
			if strings.Contains(strings.ToLower(class), pattern) || strings.Contains(strings.ToLower(id), pattern) {
				return // Skip this element
			}
		}

		text := ul.extractTextFromSelection(s)
		if len(text) > largestLength {
			largestLength = len(text)
			largestText = text
		}
	})

	if len(largestText) > 100 {
		return largestText
	}

	// Strategy 3: Just get body text (last resort)
	return ul.extractTextFromSelection(doc.Find("body"))
}

// extractTextFromSelection extracts clean text from a goquery selection
func (ul *URLLoader) extractTextFromSelection(s *goquery.Selection) string {
	// Remove script, style, noscript tags
	s.Find("script, style, noscript, iframe, svg").Remove()

	var contentBuilder strings.Builder

	// Extract text from paragraphs, headings, lists, blockquotes, pre/code
	s.Find("h1, h2, h3, h4, h5, h6, p, li, blockquote, pre, code, td, th").Each(func(i int, elem *goquery.Selection) {
		text := strings.TrimSpace(elem.Text())
		if text != "" {
			contentBuilder.WriteString(text)
			contentBuilder.WriteString("\n\n")
		}
	})

	content := contentBuilder.String()

	// Clean up extra whitespace
	content = strings.Join(strings.Fields(content), " ")

	// Restore paragraph breaks
	content = strings.ReplaceAll(content, ". ", ".\n")

	return strings.TrimSpace(content)
}

// storeDocumentChunks generates embeddings and stores document chunks
func (ul *URLLoader) storeDocumentChunks(chunks []TextChunk, sourceURL, title string) ([]TextChunk, error) {
	// Generate embeddings
	embeddings, err := ul.embeddingService.GenerateChunkEmbeddings(chunks)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embeddings: %w", err)
	}

	// Store in database
	for i, chunk := range chunks {
		embeddingJSON, err := EmbeddingToJSON(embeddings[i])
		if err != nil {
			return nil, fmt.Errorf("failed to convert embedding to JSON: %w", err)
		}

		doc := models.Document{
			ID:              uuid.New().String(),
			RealmID:         ul.config.RealmID,
			ProjectID:       ul.config.ProjectID,
			Name:            title,
			Path:            sourceURL,
			Content:         chunk.Content,
			VectorEmbedding: embeddingJSON,
			CreatedBy:       ul.config.UserID,
			UpdatedBy:       ul.config.UserID,
		}

		if err := ul.db.Create(&doc).Error; err != nil {
			return nil, fmt.Errorf("failed to store document chunk: %w", err)
		}
	}

	fmt.Printf("Stored %d document chunks from URL: %s\n", len(chunks), sourceURL)
	return chunks, nil
}

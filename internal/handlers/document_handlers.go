package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/walterfan/lazy-ai-coder/internal/rag"
	"github.com/walterfan/lazy-ai-coder/internal/services"
)

// DocumentHandlers handles document CRUD and loading endpoints
type DocumentHandlers struct {
	documentService *services.DocumentService
	db              *gorm.DB
}

// NewDocumentHandlers creates new document handlers
func NewDocumentHandlers(db *gorm.DB) *DocumentHandlers {
	return &DocumentHandlers{
		documentService: services.NewDocumentService(db),
		db:              db,
	}
}

// getEmbeddingConfig returns embedding configuration from environment variables
func getEmbeddingConfig() rag.EmbeddingConfig {
	model := os.Getenv("EMBEDDING_MODEL")
	if model == "" {
		model = "text-embedding-ada-002" // Default to OpenAI's model
	}

	apiKey := os.Getenv("EMBEDDING_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("LLM_API_KEY") // Fallback to LLM_API_KEY
	}

	baseURL := os.Getenv("EMBEDDING_URL")
	if baseURL == "" {
		baseURL = os.Getenv("LLM_BASE_URL") // Fallback to LLM_BASE_URL for backwards compatibility
	}

	return rag.EmbeddingConfig{
		APIKey:  apiKey,
		BaseURL: baseURL,
		Model:   model,
	}
}

// ListDocuments godoc
// @Summary List documents
// @Description Get list of documents with filtering and pagination
// @Tags documents
// @Produce json
// @Param scope query string false "Scope: all, personal, shared" default(all)
// @Param project_id query string false "Filter by project ID"
// @Param q query string false "Search in name, path, and content"
// @Param sort query string false "Sort by: created_at, updated_at, name" default(created_at)
// @Param limit query int false "Limit results" default(50)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/documents [get]
func (h *DocumentHandlers) ListDocuments(c *gin.Context) {
	// Get user context
	_, userIDStr, realmIDStr, _, _ := GetUserContext(c)

	// Parse query parameters
	scopeStr := c.DefaultQuery("scope", "all")
	scope := services.DocumentScope(scopeStr)
	projectID := c.Query("project_id")
	nameFilter := c.Query("q")
	sortBy := c.DefaultQuery("sort", "created_at")
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	// Convert user context to pointers
	var userID *string
	if userIDStr != "" {
		userID = &userIDStr
	}

	var projID *string
	if projectID != "" {
		projID = &projectID
	}

	// Get documents
	documents, total, err := h.documentService.ListDocuments(userID, realmIDStr, projID, scope, nameFilter, sortBy, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   documents,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// GetDocument godoc
// @Summary Get document by ID
// @Description Get a single document with full content
// @Tags documents
// @Produce json
// @Param id path string true "Document ID"
// @Success 200 {object} models.Document
// @Router /api/v1/documents/{id} [get]
func (h *DocumentHandlers) GetDocument(c *gin.Context) {
	id := c.Param("id")

	// Get user context
	_, userIDStr, realmIDStr, _, _ := GetUserContext(c)

	var userID *string
	if userIDStr != "" {
		userID = &userIDStr
	}

	document, err := h.documentService.GetDocumentByID(id, userID, realmIDStr)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, document)
}

// GetDocumentChunks godoc
// @Summary Get all chunks for a document
// @Description Get all chunks for a specific document by path and project
// @Tags documents
// @Produce json
// @Param project_id query string true "Project ID"
// @Param path query string true "Document path"
// @Success 200 {array} models.Document
// @Router /api/v1/documents/chunks [get]
func (h *DocumentHandlers) GetDocumentChunks(c *gin.Context) {
	projectID := c.Query("project_id")
	path := c.Query("path")

	if projectID == "" || path == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project_id and path are required"})
		return
	}

	// Get user context
	_, userIDStr, realmIDStr, _, _ := GetUserContext(c)

	var userID *string
	if userIDStr != "" {
		userID = &userIDStr
	}

	chunks, err := h.documentService.GetDocumentChunks(projectID, path, userID, realmIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"chunks": chunks,
		"total":  len(chunks),
	})
}

// LoadFromURL godoc
// @Summary Load document from URL
// @Description Fetch content from a URL, extract text, and store as document chunks
// @Tags documents
// @Accept json
// @Produce json
// @Param request body map[string]interface{} true "URL load request"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/documents/load-url [post]
func (h *DocumentHandlers) LoadFromURL(c *gin.Context) {
	var req struct {
		URL          string `json:"url" binding:"required"`
		ProjectID    string `json:"project_id" binding:"required"`
		ChunkSize    int    `json:"chunk_size"`
		ChunkOverlap int    `json:"chunk_overlap"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user context
	_, userIDStr, realmIDStr, _, _ := GetUserContext(c)

	// Validate URL format
	if !strings.HasPrefix(req.URL, "http://") && !strings.HasPrefix(req.URL, "https://") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid URL: must start with http:// or https://"})
		return
	}

	// Set defaults
	if req.ChunkSize == 0 {
		req.ChunkSize = 1000
	}
	if req.ChunkOverlap == 0 {
		req.ChunkOverlap = 200
	}

	// Get embedding config from environment
	embeddingConfig := getEmbeddingConfig()

	// Create URL loader
	urlLoaderConfig := rag.URLLoaderConfig{
		ProjectID:    req.ProjectID,
		RealmID:      realmIDStr,
		UserID:       userIDStr,
		ChunkSize:    req.ChunkSize,
		ChunkOverlap: req.ChunkOverlap,
	}

	loader := rag.NewURLLoader(urlLoaderConfig, embeddingConfig)

	// Load from URL
	stats, err := loader.LoadURL(req.URL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"stats":   stats,
			"message": "Failed to load document from URL",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":           "Document loaded successfully",
		"url":               req.URL,
		"project_id":        req.ProjectID,
		"chunks_created":    stats.DocumentChunks,
		"embeddings_stored": stats.EmbeddingsStored,
		"stats":             stats,
	})
}

// CreateFromText godoc
// @Summary Create document from text
// @Description Create a document from user-provided text content
// @Tags documents
// @Accept json
// @Produce json
// @Param request body map[string]interface{} true "Text document request"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/documents/create-from-text [post]
func (h *DocumentHandlers) CreateFromText(c *gin.Context) {
	var req struct {
		Name         string `json:"name" binding:"required"`
		Content      string `json:"content" binding:"required"`
		ProjectID    string `json:"project_id"`   // Optional: existing project ID
		ProjectName  string `json:"project_name"` // Optional: new project name
		ChunkSize    int    `json:"chunk_size"`
		ChunkOverlap int    `json:"chunk_overlap"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user context
	_, userIDStr, realmIDStr, _, _ := GetUserContext(c)

	// Validate content is not empty
	if strings.TrimSpace(req.Content) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "content cannot be empty"})
		return
	}

	// Validate that either project_id or project_name is provided
	if req.ProjectID == "" && req.ProjectName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "either project_id or project_name must be provided"})
		return
	}

	// If project_name is provided, create a new project
	projectID := req.ProjectID
	if req.ProjectName != "" {
		projectService := services.NewProjectService(h.db)

		// Check if project with this name already exists
		existingProjects, _, err := projectService.ListProjects(
			&userIDStr,
			realmIDStr,
			services.ProjectScopeAll,
			req.ProjectName, // nameFilter
			"",              // languageFilter
			"name",          // sortBy
			10,              // limit
			0,               // offset
		)
		if err == nil && len(existingProjects) > 0 {
			// Use existing project if name matches exactly
			for _, proj := range existingProjects {
				if proj.Name == req.ProjectName {
					projectID = proj.ID
					break
				}
			}
		}

		// If no existing project found, create new one
		if projectID == "" {
			newProject, err := projectService.CreateProject(
				req.ProjectName, // name
				"",              // description
				"",              // gitURL
				"",              // gitRepo
				"",              // gitBranch
				"",              // language
				"",              // entryPoint
				&userIDStr,      // userID
				realmIDStr,      // realmID
				userIDStr,       // createdBy
			)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to create project: %v", err)})
				return
			}
			projectID = newProject.ID
		}
	}

	// Set defaults
	if req.ChunkSize == 0 {
		req.ChunkSize = 1000
	}
	if req.ChunkOverlap == 0 {
		req.ChunkOverlap = 200
	}

	// Create temporary file with the text content
	tempDir, err := os.MkdirTemp("", "text-input-*")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to create temp directory: %v", err)})
		return
	}
	defer os.RemoveAll(tempDir) // Clean up

	// Save content to temp file
	tempFile := filepath.Join(tempDir, req.Name)
	if err := os.WriteFile(tempFile, []byte(req.Content), 0644); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to write temp file: %v", err)})
		return
	}

	// Get embedding config from environment
	embeddingConfig := getEmbeddingConfig()

	// Create document loader
	loaderConfig := rag.LoaderConfig{
		ProjectID:    projectID,
		RealmID:      realmIDStr,
		UserID:       userIDStr,
		ChunkSize:    req.ChunkSize,
		ChunkOverlap: req.ChunkOverlap,
		Recursive:    false,
		DryRun:       false,
	}

	loader := rag.NewDocumentLoader(loaderConfig, embeddingConfig)

	// Process the text file
	stats, err := loader.LoadPath(tempFile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"stats":   stats,
			"message": "Failed to create document from text",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":           "Document created successfully",
		"name":              req.Name,
		"project_id":        projectID,
		"project_name":      req.ProjectName,
		"chunks_created":    stats.DocumentChunks,
		"embeddings_stored": stats.EmbeddingsStored,
		"stats":             stats,
	})
}

// UploadFiles godoc
// @Summary Upload files
// @Description Upload one or more files to be processed and stored as document chunks
// @Tags documents
// @Accept multipart/form-data
// @Produce json
// @Param files formData file true "Files to upload (multiple allowed)"
// @Param project_id formData string true "Project ID"
// @Param chunk_size formData int false "Chunk size" default(1000)
// @Param chunk_overlap formData int false "Chunk overlap" default(200)
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/documents/upload [post]
func (h *DocumentHandlers) UploadFiles(c *gin.Context) {
	// Get user context
	_, userIDStr, realmIDStr, _, _ := GetUserContext(c)

	// Get form values
	projectID := c.PostForm("project_id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project_id is required"})
		return
	}

	chunkSizeStr := c.DefaultPostForm("chunk_size", "1000")
	chunkOverlapStr := c.DefaultPostForm("chunk_overlap", "200")

	chunkSize, _ := strconv.Atoi(chunkSizeStr)
	chunkOverlap, _ := strconv.Atoi(chunkOverlapStr)

	// Get multipart form
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("failed to parse multipart form: %v", err)})
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no files provided"})
		return
	}

	// Create temporary directory for uploads
	tempDir, err := os.MkdirTemp("", "upload-*")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to create temp directory: %v", err)})
		return
	}
	defer os.RemoveAll(tempDir) // Clean up

	// Save uploaded files
	var savedFiles []string
	for _, file := range files {
		filePath := filepath.Join(tempDir, file.Filename)
		if err := c.SaveUploadedFile(file, filePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("failed to save file %s: %v", file.Filename, err),
			})
			return
		}
		savedFiles = append(savedFiles, filePath)
	}

	// Get embedding config from environment
	embeddingConfig := getEmbeddingConfig()

	// Create document loader
	loaderConfig := rag.LoaderConfig{
		ProjectID:    projectID,
		RealmID:      realmIDStr,
		UserID:       userIDStr,
		ChunkSize:    chunkSize,
		ChunkOverlap: chunkOverlap,
		Recursive:    false,
		DryRun:       false,
	}

	loader := rag.NewDocumentLoader(loaderConfig, embeddingConfig)

	// Process each file
	totalStats := &rag.LoaderStats{Errors: []string{}}
	for _, filePath := range savedFiles {
		stats, err := loader.LoadPath(filePath)
		if err != nil {
			totalStats.Errors = append(totalStats.Errors, fmt.Sprintf("%s: %v", filepath.Base(filePath), err))
		} else {
			totalStats.FilesProcessed += stats.FilesProcessed
			totalStats.DocumentChunks += stats.DocumentChunks
			totalStats.CodeChunks += stats.CodeChunks
			totalStats.TotalChunks += stats.TotalChunks
			if len(stats.Errors) == 0 && stats.FilesProcessed > 0 {
				totalStats.EmbeddingsStored++
			}
		}
	}

	// Return results
	c.JSON(http.StatusOK, gin.H{
		"message":         "Files processed successfully",
		"files_uploaded":  len(savedFiles),
		"files_processed": totalStats.FilesProcessed,
		"chunks_created":  totalStats.TotalChunks,
		"code_chunks":     totalStats.CodeChunks,
		"document_chunks": totalStats.DocumentChunks,
		"errors":          totalStats.Errors,
	})
}

// DeleteDocument godoc
// @Summary Delete document
// @Description Soft-delete a document by ID
// @Tags documents
// @Param id path string true "Document ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/documents/{id} [delete]
func (h *DocumentHandlers) DeleteDocument(c *gin.Context) {
	id := c.Param("id")

	// Get user context
	_, userIDStr, realmIDStr, _, _ := GetUserContext(c)

	var userID *string
	if userIDStr != "" {
		userID = &userIDStr
	}

	err := h.documentService.DeleteDocument(id, userID, realmIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Document deleted successfully"})
}

// DeleteDocumentByPath godoc
// @Summary Delete all chunks of a document
// @Description Soft-delete all chunks of a document by path and project
// @Tags documents
// @Accept json
// @Produce json
// @Param request body map[string]interface{} true "Delete request"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/documents/delete-by-path [post]
func (h *DocumentHandlers) DeleteDocumentByPath(c *gin.Context) {
	var req struct {
		ProjectID string `json:"project_id" binding:"required"`
		Path      string `json:"path" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user context
	_, userIDStr, realmIDStr, _, _ := GetUserContext(c)

	var userID *string
	if userIDStr != "" {
		userID = &userIDStr
	}

	rowsAffected, err := h.documentService.DeleteDocumentsByPath(req.ProjectID, req.Path, userID, realmIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        "Document chunks deleted successfully",
		"chunks_deleted": rowsAffected,
	})
}

// GetDocumentStats godoc
// @Summary Get document statistics
// @Description Get statistics about documents (total chunks, unique documents, etc.)
// @Tags documents
// @Produce json
// @Param project_id query string false "Filter by project ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/documents/stats [get]
func (h *DocumentHandlers) GetDocumentStats(c *gin.Context) {
	projectID := c.Query("project_id")

	// Get user context
	_, userIDStr, realmIDStr, _, _ := GetUserContext(c)

	var userID *string
	if userIDStr != "" {
		userID = &userIDStr
	}

	var projID *string
	if projectID != "" {
		projID = &projectID
	}

	stats, err := h.documentService.GetDocumentStats(projID, userID, realmIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

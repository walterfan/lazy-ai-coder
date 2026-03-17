package rag

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/walterfan/lazy-ai-coder/pkg/database"
	"github.com/walterfan/lazy-ai-coder/pkg/models"
	"gorm.io/gorm"
)

// LoaderConfig contains configuration for the document/code loader
type LoaderConfig struct {
	ProjectID       string
	RealmID         string
	UserID          string
	ChunkSize       int
	ChunkOverlap    int
	IncludeTypes    []string
	ExcludePatterns []string
	Recursive       bool
	DryRun          bool
	BatchSize       int
}

// LoaderStats contains statistics from a load operation
type LoaderStats struct {
	FilesProcessed   int
	FilesSkipped     int
	CodeChunks       int
	DocumentChunks   int
	TotalChunks      int
	EmbeddingsStored int
	Errors           []string
}

// DocumentLoader handles the RAG loading pipeline
type DocumentLoader struct {
	config           LoaderConfig
	docParser        *DocumentParser
	codeParser       *CodeParser
	textSplitter     *TextSplitter
	embeddingService *EmbeddingService
	db               *gorm.DB
}

// NewDocumentLoader creates a new document loader
func NewDocumentLoader(config LoaderConfig, embeddingConfig EmbeddingConfig) *DocumentLoader {
	return &DocumentLoader{
		config:           config,
		docParser:        NewDocumentParser(),
		codeParser:       NewCodeParser(),
		textSplitter:     NewTextSplitter(config.ChunkSize, config.ChunkOverlap),
		embeddingService: NewEmbeddingService(embeddingConfig),
		db:               database.GetDB(),
	}
}

// LoadPath loads documents and code from a file or directory
func (dl *DocumentLoader) LoadPath(path string) (*LoaderStats, error) {
	stats := &LoaderStats{
		Errors: []string{},
	}

	// Check if path exists
	fileInfo, err := os.Stat(path)
	if err != nil {
		return stats, fmt.Errorf("path not found: %w", err)
	}

	// Load single file or directory
	if fileInfo.IsDir() {
		return dl.loadDirectory(path, stats)
	}

	return dl.loadFile(path, stats)
}

// loadDirectory loads all supported files from a directory
func (dl *DocumentLoader) loadDirectory(dirPath string, stats *LoaderStats) (*LoaderStats, error) {
	walkFunc := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			stats.Errors = append(stats.Errors, fmt.Sprintf("Error accessing %s: %v", path, err))
			return nil // Continue walking
		}

		// Skip directories
		if info.IsDir() {
			// Check if should skip this directory
			for _, pattern := range dl.config.ExcludePatterns {
				if strings.Contains(path, pattern) {
					return filepath.SkipDir
				}
			}
			return nil
		}

		// Check if file should be processed
		if !ShouldProcessFile(path, dl.config.IncludeTypes, dl.config.ExcludePatterns) {
			stats.FilesSkipped++
			return nil
		}

		// Process file
		_, err = dl.processFile(path, stats)
		if err != nil {
			stats.Errors = append(stats.Errors, fmt.Sprintf("Error processing %s: %v", path, err))
		}

		return nil
	}

	// Walk directory
	if dl.config.Recursive {
		err := filepath.Walk(dirPath, walkFunc)
		if err != nil {
			return stats, fmt.Errorf("error walking directory: %w", err)
		}
	} else {
		// Only process files in the immediate directory
		entries, err := os.ReadDir(dirPath)
		if err != nil {
			return stats, fmt.Errorf("error reading directory: %w", err)
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			path := filepath.Join(dirPath, entry.Name())
			if !ShouldProcessFile(path, dl.config.IncludeTypes, dl.config.ExcludePatterns) {
				stats.FilesSkipped++
				continue
			}

			_, err = dl.processFile(path, stats)
			if err != nil {
				stats.Errors = append(stats.Errors, fmt.Sprintf("Error processing %s: %v", path, err))
			}
		}
	}

	return stats, nil
}

// loadFile loads a single file
func (dl *DocumentLoader) loadFile(path string, stats *LoaderStats) (*LoaderStats, error) {
	if !ShouldProcessFile(path, dl.config.IncludeTypes, dl.config.ExcludePatterns) {
		stats.FilesSkipped++
		return stats, nil
	}

	_, err := dl.processFile(path, stats)
	if err != nil {
		stats.Errors = append(stats.Errors, fmt.Sprintf("Error processing %s: %v", path, err))
		return stats, err
	}

	return stats, nil
}

// processFile processes a single file through the RAG pipeline
func (dl *DocumentLoader) processFile(filePath string, stats *LoaderStats) ([]TextChunk, error) {
	// Detect file type
	fileInfo := DetectFileType(filePath)

	var chunks []TextChunk
	var err error

	// Parse based on file type
	switch fileInfo.Type {
	case FileTypeDocument, FileTypeConfig:
		chunks, err = dl.processDocument(filePath, fileInfo, stats)
	case FileTypeCode:
		chunks, err = dl.processCode(filePath, fileInfo, stats)
	default:
		stats.FilesSkipped++
		return nil, fmt.Errorf("unsupported file type: %s", filePath)
	}

	if err != nil {
		return nil, err
	}

	stats.FilesProcessed++
	return chunks, nil
}

// processDocument processes a document file
func (dl *DocumentLoader) processDocument(filePath string, fileInfo FileInfo, stats *LoaderStats) ([]TextChunk, error) {
	// Parse document
	content, err := dl.docParser.ParseDocument(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse document: %w", err)
	}

	// Split into chunks
	chunks := dl.textSplitter.SplitDocument(content, filePath)
	stats.DocumentChunks += len(chunks)
	stats.TotalChunks += len(chunks)

	if dl.config.DryRun {
		fmt.Printf("[DRY RUN] Document: %s - %d chunks\n", filePath, len(chunks))
		return chunks, nil
	}

	// Generate embeddings and store
	return dl.storeDocumentChunks(chunks, filePath)
}

// processCode processes a code file
func (dl *DocumentLoader) processCode(filePath string, fileInfo FileInfo, stats *LoaderStats) ([]TextChunk, error) {
	// Parse code
	metadata, err := dl.codeParser.ParseCode(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse code: %w", err)
	}

	// Split into chunks using hybrid approach
	chunks := dl.textSplitter.SplitCode(metadata)
	stats.CodeChunks += len(chunks)
	stats.TotalChunks += len(chunks)

	if dl.config.DryRun {
		fmt.Printf("[DRY RUN] Code: %s - %d chunks (functions: %d, classes: %d)\n",
			filePath, len(chunks), len(metadata.Functions), len(metadata.Classes))
		return chunks, nil
	}

	// Generate embeddings and store
	return dl.storeCodeChunks(chunks, filePath, metadata)
}

// storeDocumentChunks generates embeddings and stores document chunks
func (dl *DocumentLoader) storeDocumentChunks(chunks []TextChunk, filePath string) ([]TextChunk, error) {
	// Generate embeddings
	embeddings, err := dl.embeddingService.GenerateChunkEmbeddings(chunks)
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
			RealmID:         dl.config.RealmID,
			ProjectID:       dl.config.ProjectID,
			Name:            filepath.Base(filePath),
			Path:            filePath,
			Content:         chunk.Content,
			VectorEmbedding: embeddingJSON,
			CreatedBy:       dl.config.UserID,
			UpdatedBy:       dl.config.UserID,
		}

		if err := dl.db.Create(&doc).Error; err != nil {
			return nil, fmt.Errorf("failed to store document chunk: %w", err)
		}
	}

	fmt.Printf("Stored %d document chunks from: %s\n", len(chunks), filePath)
	return chunks, nil
}

// storeCodeChunks generates embeddings and stores code chunks
func (dl *DocumentLoader) storeCodeChunks(chunks []TextChunk, filePath string, metadata *CodeMetadata) ([]TextChunk, error) {
	// Generate embeddings
	embeddings, err := dl.embeddingService.GenerateChunkEmbeddings(chunks)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embeddings: %w", err)
	}

	// Store in database
	for i, chunk := range chunks {
		embeddingJSON, err := EmbeddingToJSON(embeddings[i])
		if err != nil {
			return nil, fmt.Errorf("failed to convert embedding to JSON: %w", err)
		}

		code := models.Code{
			ID:              uuid.New().String(),
			RealmID:         dl.config.RealmID,
			ProjectID:       dl.config.ProjectID,
			Path:            filePath,
			Code:            chunk.Content,
			VectorEmbedding: embeddingJSON,
			CreatedBy:       dl.config.UserID,
			UpdatedBy:       dl.config.UserID,
		}

		if err := dl.db.Create(&code).Error; err != nil {
			return nil, fmt.Errorf("failed to store code chunk: %w", err)
		}
	}

	fmt.Printf("Stored %d code chunks from: %s (language: %s)\n", len(chunks), filePath, metadata.Language)
	return chunks, nil
}

// PrintStats prints loader statistics
func (stats *LoaderStats) PrintStats() {
	fmt.Println("\n=== RAG Loading Summary ===")
	fmt.Printf("Files Processed: %d\n", stats.FilesProcessed)
	fmt.Printf("Files Skipped: %d\n", stats.FilesSkipped)
	fmt.Printf("Code Chunks: %d\n", stats.CodeChunks)
	fmt.Printf("Document Chunks: %d\n", stats.DocumentChunks)
	fmt.Printf("Total Chunks: %d\n", stats.TotalChunks)

	if len(stats.Errors) > 0 {
		fmt.Printf("\nErrors (%d):\n", len(stats.Errors))
		for _, err := range stats.Errors {
			fmt.Printf("  - %s\n", err)
		}
	}
}

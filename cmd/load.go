package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/walterfan/lazy-ai-coder/internal/rag"
	"github.com/walterfan/lazy-ai-coder/pkg/database"
)

var (
	loadPath         string
	loadProjectID    string
	loadRealmID      string
	loadUserID       string
	loadRecursive    bool
	loadDryRun       bool
	loadChunkSize    int
	loadChunkOverlap int
	loadTypes        string
	loadExclude      string
	loadBatchSize    int
)

var loadCmd = &cobra.Command{
	Use:   "load",
	Short: "Load documents and code files into RAG database",
	Long: `Load documents (PDF, DOCX, PPTX, MD, RST) and code files (Go, Python, Java, C++, JavaScript, etc.)
into the RAG (Retrieval-Augmented Generation) database with vector embeddings.

The loader will:
1. Auto-detect file types (documents vs code)
2. Parse files and extract content
3. Split content into chunks (semantic for code, token-based for documents)
4. Generate embeddings using OpenAI API
5. Store chunks and embeddings in PostgreSQL with pgvector

Examples:
  # Load a single file
  ./lazy-ai-coder load --path notes.md --project-id my-project

  # Load all documents in a directory recursively
  ./lazy-ai-coder load -p ~/Documents/notes -r --project-id my-project

  # Load code repository
  ./lazy-ai-coder load -p ~/myproject/src -r --project-id my-project

  # Load only specific file types
  ./lazy-ai-coder load -p ~/workspace -r --types "go,py,md" --project-id my-project

  # Exclude vendor/build directories
  ./lazy-ai-coder load -p ~/repo -r --exclude "vendor/,build/,node_modules/" --project-id my-project

  # Dry run to preview what would be loaded
  ./lazy-ai-coder load -p ~/notes --dry-run --project-id my-project

  # Custom chunk size
  ./lazy-ai-coder load -p notes.pdf --chunk-size 500 --chunk-overlap 100 --project-id my-project
`,
	Run: func(cmd *cobra.Command, args []string) {
		// Validate required parameters
		if loadPath == "" {
			fmt.Fprintf(os.Stderr, "Error: --path is required\n")
			cmd.Usage()
			os.Exit(1)
		}

		if loadProjectID == "" {
			fmt.Fprintf(os.Stderr, "Error: --project-id is required\n")
			cmd.Usage()
			os.Exit(1)
		}

		// Default values
		if loadRealmID == "" {
			loadRealmID = "default"
		}
		if loadUserID == "" {
			loadUserID = "system"
		}

		// Parse file types filter
		var includeTypes []string
		if loadTypes != "" {
			includeTypes = strings.Split(loadTypes, ",")
			for i, t := range includeTypes {
				includeTypes[i] = strings.TrimSpace(t)
			}
		}

		// Parse exclude patterns
		var excludePatterns []string
		if loadExclude != "" {
			excludePatterns = strings.Split(loadExclude, ",")
			for i, p := range excludePatterns {
				excludePatterns[i] = strings.TrimSpace(p)
			}
		}

		// Get embedding configuration from environment
		embeddingAPIKey := os.Getenv("EMBEDDING_API_KEY")
		if embeddingAPIKey == "" {
			embeddingAPIKey = os.Getenv("LLM_API_KEY") // Fallback to LLM_API_KEY
		}

		embeddingBaseURL := os.Getenv("EMBEDDING_URL")
		if embeddingBaseURL == "" {
			embeddingBaseURL = os.Getenv("LLM_BASE_URL") // Fallback to LLM_BASE_URL
		}

		embeddingModel := os.Getenv("EMBEDDING_MODEL")

		if embeddingAPIKey == "" {
			fmt.Fprintf(os.Stderr, "Error: EMBEDDING_API_KEY or LLM_API_KEY environment variable is required\n")
			os.Exit(1)
		}

		if embeddingBaseURL == "" {
			embeddingBaseURL = "https://api.openai.com/v1"
		}

		if embeddingModel == "" {
			embeddingModel = "text-embedding-ada-002"
		}

		// Initialize database (only if not dry run)
		if !loadDryRun {
			if err := database.InitDB(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: Failed to initialize database: %v\n", err)
				os.Exit(1)
			}
			defer database.CloseDB()
		}

		// Create loader configuration
		loaderConfig := rag.LoaderConfig{
			ProjectID:       loadProjectID,
			RealmID:         loadRealmID,
			UserID:          loadUserID,
			ChunkSize:       loadChunkSize,
			ChunkOverlap:    loadChunkOverlap,
			IncludeTypes:    includeTypes,
			ExcludePatterns: excludePatterns,
			Recursive:       loadRecursive,
			DryRun:          loadDryRun,
			BatchSize:       loadBatchSize,
		}

		embeddingConfig := rag.EmbeddingConfig{
			APIKey:  embeddingAPIKey,
			BaseURL: embeddingBaseURL,
			Model:   embeddingModel,
		}

		// Create document loader
		loader := rag.NewDocumentLoader(loaderConfig, embeddingConfig)

		// Display configuration
		fmt.Println("=== RAG Loader Configuration ===")
		fmt.Printf("Path: %s\n", loadPath)
		fmt.Printf("Project ID: %s\n", loadProjectID)
		fmt.Printf("Realm ID: %s\n", loadRealmID)
		fmt.Printf("Recursive: %v\n", loadRecursive)
		fmt.Printf("Chunk Size: %d tokens\n", loadChunkSize)
		fmt.Printf("Chunk Overlap: %d tokens\n", loadChunkOverlap)
		if len(includeTypes) > 0 {
			fmt.Printf("Include Types: %s\n", strings.Join(includeTypes, ", "))
		}
		if len(excludePatterns) > 0 {
			fmt.Printf("Exclude Patterns: %s\n", strings.Join(excludePatterns, ", "))
		}
		fmt.Printf("Embedding Model: %s\n", embeddingModel)
		if loadDryRun {
			fmt.Println("DRY RUN MODE - No data will be stored")
		}
		fmt.Println()

		// Load documents/code
		fmt.Println("Loading files...")
		stats, err := loader.LoadPath(loadPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// Print statistics
		stats.PrintStats()

		if len(stats.Errors) > 0 {
			fmt.Println("\n⚠️  Completed with errors")
			os.Exit(1)
		}

		fmt.Println("\n✅ Loading completed successfully!")
	},
}

func init() {
	rootCmd.AddCommand(loadCmd)

	// Required flags
	loadCmd.Flags().StringVarP(&loadPath, "path", "p", "", "File or directory path to load (required)")
	loadCmd.Flags().StringVar(&loadProjectID, "project-id", "", "Project ID to associate documents/code with (required)")

	// Optional flags
	loadCmd.Flags().StringVar(&loadRealmID, "realm-id", "default", "Realm ID for multi-tenancy")
	loadCmd.Flags().StringVar(&loadUserID, "user-id", "system", "User ID for created_by field")
	loadCmd.Flags().BoolVarP(&loadRecursive, "recursive", "r", false, "Recursively process directories")
	loadCmd.Flags().BoolVar(&loadDryRun, "dry-run", false, "Preview what would be loaded without storing")
	loadCmd.Flags().IntVar(&loadChunkSize, "chunk-size", 1000, "Maximum chunk size in tokens")
	loadCmd.Flags().IntVar(&loadChunkOverlap, "chunk-overlap", 200, "Overlap between chunks in tokens")
	loadCmd.Flags().StringVar(&loadTypes, "types", "", "Comma-separated list of file types to include (e.g., 'go,py,md')")
	loadCmd.Flags().StringVar(&loadExclude, "exclude", "", "Comma-separated list of patterns to exclude (e.g., 'vendor/,node_modules/')")
	loadCmd.Flags().IntVar(&loadBatchSize, "batch-size", 10, "Batch size for embedding generation")

	// Mark required flags
	loadCmd.MarkFlagRequired("path")
	loadCmd.MarkFlagRequired("project-id")
}

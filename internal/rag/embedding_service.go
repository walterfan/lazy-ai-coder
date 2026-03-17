package rag

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// EmbeddingService handles generating embeddings via OpenAI API
type EmbeddingService struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

// EmbeddingConfig contains configuration for embedding service
type EmbeddingConfig struct {
	APIKey  string
	BaseURL string
	Model   string
}

// EmbeddingRequest represents the OpenAI embeddings API request
type EmbeddingRequest struct {
	Input          interface{} `json:"input"` // Can be string or []string
	Model          string      `json:"model"`
	EncodingFormat string      `json:"encoding_format,omitempty"`
}

// EmbeddingResponse represents the OpenAI embeddings API response
type EmbeddingResponse struct {
	Object string `json:"object"`
	Data   []struct {
		Object    string    `json:"object"`
		Embedding []float32 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Model string `json:"model"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
}

// NewEmbeddingService creates a new embedding service
func NewEmbeddingService(config EmbeddingConfig) *EmbeddingService {
	if config.BaseURL == "" {
		config.BaseURL = "https://api.openai.com/v1"
	}
	if config.Model == "" {
		config.Model = "text-embedding-ada-002" // Default model, produces 1536-dim vectors
	}

	return &EmbeddingService{
		apiKey:  config.APIKey,
		baseURL: config.BaseURL,
		model:   config.Model,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// GenerateEmbedding generates embedding for a single text
func (es *EmbeddingService) GenerateEmbedding(text string) ([]float32, error) {
	if text == "" {
		return nil, fmt.Errorf("text cannot be empty")
	}

	embeddings, err := es.GenerateEmbeddings([]string{text})
	if err != nil {
		return nil, err
	}

	if len(embeddings) == 0 {
		return nil, fmt.Errorf("no embeddings returned")
	}

	return embeddings[0], nil
}

// GenerateEmbeddings generates embeddings for multiple texts in batch
func (es *EmbeddingService) GenerateEmbeddings(texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, fmt.Errorf("texts cannot be empty")
	}

	// OpenAI API allows max 2048 texts per request, but we'll use smaller batches
	const maxBatchSize = 100
	var allEmbeddings [][]float32

	for i := 0; i < len(texts); i += maxBatchSize {
		end := i + maxBatchSize
		if end > len(texts) {
			end = len(texts)
		}

		batch := texts[i:end]
		embeddings, err := es.generateBatch(batch)
		if err != nil {
			return nil, fmt.Errorf("failed to generate batch %d-%d: %w", i, end, err)
		}

		allEmbeddings = append(allEmbeddings, embeddings...)
	}

	return allEmbeddings, nil
}

// generateBatch generates embeddings for a single batch
func (es *EmbeddingService) generateBatch(texts []string) ([][]float32, error) {
	// Prepare request
	reqBody := EmbeddingRequest{
		Input: texts,
		Model: es.model,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/embeddings", es.baseURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", es.apiKey))

	// Send request
	resp, err := es.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var embResp EmbeddingResponse
	if err := json.Unmarshal(body, &embResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Extract embeddings in order
	embeddings := make([][]float32, len(embResp.Data))
	for _, item := range embResp.Data {
		if item.Index >= len(embeddings) {
			return nil, fmt.Errorf("invalid embedding index: %d", item.Index)
		}
		embeddings[item.Index] = item.Embedding
	}

	return embeddings, nil
}

// GenerateChunkEmbeddings generates embeddings for text chunks
func (es *EmbeddingService) GenerateChunkEmbeddings(chunks []TextChunk) ([][]float32, error) {
	if len(chunks) == 0 {
		return nil, fmt.Errorf("no chunks provided")
	}

	// Extract text from chunks
	texts := make([]string, len(chunks))
	for i, chunk := range chunks {
		texts[i] = chunk.Content
	}

	return es.GenerateEmbeddings(texts)
}

// ValidateEmbedding checks if embedding has the expected dimensions
func ValidateEmbedding(embedding []float32, expectedDim int) error {
	if len(embedding) == 0 {
		return fmt.Errorf("embedding is empty")
	}

	if expectedDim > 0 && len(embedding) != expectedDim {
		return fmt.Errorf("embedding dimension mismatch: got %d, expected %d", len(embedding), expectedDim)
	}

	return nil
}

// EmbeddingToJSON converts float32 slice to JSON string for database storage
func EmbeddingToJSON(embedding []float32) (string, error) {
	data, err := json.Marshal(embedding)
	if err != nil {
		return "", fmt.Errorf("failed to marshal embedding: %w", err)
	}
	return string(data), nil
}

// JSONToEmbedding converts JSON string to float32 slice
func JSONToEmbedding(jsonStr string) ([]float32, error) {
	var embedding []float32
	if err := json.Unmarshal([]byte(jsonStr), &embedding); err != nil {
		return nil, fmt.Errorf("failed to unmarshal embedding: %w", err)
	}
	return embedding, nil
}

// EmbeddingToPostgresArray converts float32 slice to Postgres array format for pgvector
// Format: [0.1,0.2,0.3,...] or use the float32 slice directly with pgx driver
func EmbeddingToPostgresArray(embedding []float32) string {
	// For pgvector, we can use the format: [val1,val2,val3,...]
	var buf bytes.Buffer
	buf.WriteString("[")
	for i, val := range embedding {
		if i > 0 {
			buf.WriteString(",")
		}
		buf.WriteString(fmt.Sprintf("%f", val))
	}
	buf.WriteString("]")
	return buf.String()
}

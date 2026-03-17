package rag

import (
	"strings"
	"unicode"
)

// TextChunk represents a chunk of text
type TextChunk struct {
	Content   string
	Metadata  map[string]interface{}
	StartLine int
	EndLine   int
	Tokens    int
}

// SplitterConfig contains configuration for text splitting
type SplitterConfig struct {
	ChunkSize    int // Max tokens per chunk
	ChunkOverlap int // Overlap between chunks in tokens
}

// TextSplitter handles text chunking
type TextSplitter struct {
	config SplitterConfig
}

// NewTextSplitter creates a new text splitter
func NewTextSplitter(chunkSize, chunkOverlap int) *TextSplitter {
	if chunkSize <= 0 {
		chunkSize = 1000
	}
	if chunkOverlap < 0 {
		chunkOverlap = 200
	}
	if chunkOverlap >= chunkSize {
		chunkOverlap = chunkSize / 5
	}

	return &TextSplitter{
		config: SplitterConfig{
			ChunkSize:    chunkSize,
			ChunkOverlap: chunkOverlap,
		},
	}
}

// SplitDocument splits document text into chunks using token-based splitting
func (ts *TextSplitter) SplitDocument(content string, filePath string) []TextChunk {
	return ts.splitByTokens(content, filePath, nil)
}

// SplitCode splits code using hybrid approach: semantic (by function/class) with token-based fallback
func (ts *TextSplitter) SplitCode(metadata *CodeMetadata) []TextChunk {
	var chunks []TextChunk

	// Strategy 1: Try semantic splitting by functions and classes
	semanticChunks := ts.splitCodeSemantic(metadata)

	// Check if semantic chunks are too large and need further splitting
	for _, chunk := range semanticChunks {
		if chunk.Tokens <= ts.config.ChunkSize {
			chunks = append(chunks, chunk)
		} else {
			// Chunk is too large, split it using token-based approach
			subChunks := ts.splitByTokens(chunk.Content, metadata.FilePath, chunk.Metadata)
			chunks = append(chunks, subChunks...)
		}
	}

	// If no semantic chunks were produced, fall back to token-based splitting
	if len(chunks) == 0 {
		chunks = ts.splitByTokens(metadata.Content, metadata.FilePath, map[string]interface{}{
			"language": metadata.Language,
		})
	}

	return chunks
}

// splitCodeSemantic splits code by functions and classes (semantic splitting)
func (ts *TextSplitter) splitCodeSemantic(metadata *CodeMetadata) []TextChunk {
	var chunks []TextChunk

	// Create chunks from functions
	for _, funcInfo := range metadata.Functions {
		tokens := ts.estimateTokens(funcInfo.Content)
		chunk := TextChunk{
			Content:   funcInfo.Content,
			StartLine: funcInfo.StartLine,
			EndLine:   funcInfo.EndLine,
			Tokens:    tokens,
			Metadata: map[string]interface{}{
				"type":     "function",
				"name":     funcInfo.Name,
				"language": metadata.Language,
				"file":     metadata.FilePath,
			},
		}
		chunks = append(chunks, chunk)
	}

	// Create chunks from classes
	for _, classInfo := range metadata.Classes {
		tokens := ts.estimateTokens(classInfo.Content)
		chunk := TextChunk{
			Content:   classInfo.Content,
			StartLine: classInfo.StartLine,
			EndLine:   classInfo.EndLine,
			Tokens:    tokens,
			Metadata: map[string]interface{}{
				"type":     "class",
				"name":     classInfo.Name,
				"language": metadata.Language,
				"file":     metadata.FilePath,
			},
		}
		chunks = append(chunks, chunk)
	}

	// If we have very few or no chunks, include the full file context
	if len(chunks) < 2 && len(metadata.Content) > 0 {
		// File is small or doesn't have clear structure, treat as single chunk
		tokens := ts.estimateTokens(metadata.Content)
		if tokens <= ts.config.ChunkSize*2 {
			chunk := TextChunk{
				Content: metadata.Content,
				Tokens:  tokens,
				Metadata: map[string]interface{}{
					"type":     "file",
					"language": metadata.Language,
					"file":     metadata.FilePath,
				},
			}
			return []TextChunk{chunk}
		}
	}

	return chunks
}

// splitByTokens splits text into chunks based on token count
func (ts *TextSplitter) splitByTokens(content string, filePath string, baseMetadata map[string]interface{}) []TextChunk {
	var chunks []TextChunk

	if len(content) == 0 {
		return chunks
	}

	// Split by paragraphs first (double newline)
	paragraphs := strings.Split(content, "\n\n")

	var currentChunk strings.Builder
	var currentTokens int
	var chunkStartLine int = 1
	var currentLine int = 1

	for _, para := range paragraphs {
		paraTokens := ts.estimateTokens(para)

		// If paragraph itself is too large, split it further
		if paraTokens > ts.config.ChunkSize {
			// Save current chunk if not empty
			if currentChunk.Len() > 0 {
				chunk := ts.createChunk(currentChunk.String(), chunkStartLine, currentLine, filePath, baseMetadata)
				chunks = append(chunks, chunk)
				currentChunk.Reset()
				currentTokens = 0
				chunkStartLine = currentLine
			}

			// Split large paragraph by sentences
			sentences := ts.splitIntoSentences(para)
			for _, sent := range sentences {
				sentTokens := ts.estimateTokens(sent)

				if currentTokens+sentTokens > ts.config.ChunkSize && currentChunk.Len() > 0 {
					// Create chunk with overlap
					chunk := ts.createChunk(currentChunk.String(), chunkStartLine, currentLine, filePath, baseMetadata)
					chunks = append(chunks, chunk)

					// Keep overlap
					overlapText := ts.getOverlap(currentChunk.String())
					currentChunk.Reset()
					currentChunk.WriteString(overlapText)
					currentTokens = ts.estimateTokens(overlapText)
					chunkStartLine = currentLine
				}

				currentChunk.WriteString(sent)
				currentChunk.WriteString(" ")
				currentTokens += sentTokens
			}

			currentLine += strings.Count(para, "\n") + 2
			continue
		}

		// Check if adding this paragraph would exceed chunk size
		if currentTokens+paraTokens > ts.config.ChunkSize && currentChunk.Len() > 0 {
			// Create chunk
			chunk := ts.createChunk(currentChunk.String(), chunkStartLine, currentLine-1, filePath, baseMetadata)
			chunks = append(chunks, chunk)

			// Keep overlap
			overlapText := ts.getOverlap(currentChunk.String())
			currentChunk.Reset()
			currentChunk.WriteString(overlapText)
			currentChunk.WriteString("\n\n")
			currentTokens = ts.estimateTokens(overlapText)
			chunkStartLine = currentLine
		}

		// Add paragraph to current chunk
		if currentChunk.Len() > 0 {
			currentChunk.WriteString("\n\n")
		}
		currentChunk.WriteString(para)
		currentTokens += paraTokens
		currentLine += strings.Count(para, "\n") + 2
	}

	// Add remaining content as final chunk
	if currentChunk.Len() > 0 {
		chunk := ts.createChunk(currentChunk.String(), chunkStartLine, currentLine, filePath, baseMetadata)
		chunks = append(chunks, chunk)
	}

	return chunks
}

// createChunk creates a TextChunk with metadata
func (ts *TextSplitter) createChunk(content string, startLine, endLine int, filePath string, baseMetadata map[string]interface{}) TextChunk {
	metadata := make(map[string]interface{})
	if baseMetadata != nil {
		for k, v := range baseMetadata {
			metadata[k] = v
		}
	}
	metadata["file"] = filePath

	return TextChunk{
		Content:   strings.TrimSpace(content),
		StartLine: startLine,
		EndLine:   endLine,
		Tokens:    ts.estimateTokens(content),
		Metadata:  metadata,
	}
}

// getOverlap extracts the overlap portion from the end of text
func (ts *TextSplitter) getOverlap(text string) string {
	if ts.config.ChunkOverlap <= 0 {
		return ""
	}

	// Split into words and take last N words for overlap
	words := strings.Fields(text)
	overlapWords := ts.config.ChunkOverlap / 4 // Rough estimate: 4 chars per token

	if overlapWords >= len(words) {
		return text
	}

	if overlapWords > 0 {
		return strings.Join(words[len(words)-overlapWords:], " ")
	}

	return ""
}

// splitIntoSentences splits text into sentences
func (ts *TextSplitter) splitIntoSentences(text string) []string {
	var sentences []string
	var current strings.Builder

	runes := []rune(text)
	for i := 0; i < len(runes); i++ {
		current.WriteRune(runes[i])

		// Check for sentence endings
		if runes[i] == '.' || runes[i] == '!' || runes[i] == '?' || runes[i] == '\n' {
			// Look ahead to see if next char is space or end of text
			if i+1 >= len(runes) || unicode.IsSpace(runes[i+1]) {
				sentence := strings.TrimSpace(current.String())
				if len(sentence) > 0 {
					sentences = append(sentences, sentence)
				}
				current.Reset()
			}
		}
	}

	// Add any remaining text
	if current.Len() > 0 {
		sentence := strings.TrimSpace(current.String())
		if len(sentence) > 0 {
			sentences = append(sentences, sentence)
		}
	}

	return sentences
}

// estimateTokens estimates the number of tokens in text
// Using rough estimate: 1 token ≈ 4 characters (common for English)
func (ts *TextSplitter) estimateTokens(text string) int {
	// Remove extra whitespace
	text = strings.TrimSpace(text)
	if len(text) == 0 {
		return 0
	}

	// Rough estimation: average of character count and word count
	charCount := len(text)
	wordCount := len(strings.Fields(text))

	// Tokens are usually between words and chars/4
	// Use average: (chars/4 + words) / 2
	tokens := (charCount/4 + wordCount) / 2

	if tokens < 1 {
		tokens = 1
	}

	return tokens
}

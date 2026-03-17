package memory

import (
	"context"

	"github.com/walterfan/lazy-ai-coder/pkg/models"
)

// SimilarityRetriever defines the interface for finding similar learning records
type SimilarityRetriever interface {
	// FindSimilar finds learning records similar to the given input
	FindSimilar(ctx context.Context, userID, input string, limit int) ([]models.ChatRecord, error)
}

// Repository defines the subset of repository methods needed by the similarity retriever
type Repository interface {
	FindSimilar(ctx context.Context, userID string, input string, limit int) ([]models.ChatRecord, error)
}

// FuzzySimilarityRetriever implements SimilarityRetriever using fuzzy string matching
type FuzzySimilarityRetriever struct {
	repo      Repository
	threshold float64 // minimum similarity threshold (0.0 to 1.0)
}

// NewFuzzySimilarityRetriever creates a new fuzzy similarity retriever
func NewFuzzySimilarityRetriever(repo Repository) *FuzzySimilarityRetriever {
	return &FuzzySimilarityRetriever{
		repo:      repo,
		threshold: 0.3, // default threshold
	}
}

// NewFuzzySimilarityRetrieverWithThreshold creates a retriever with custom threshold
func NewFuzzySimilarityRetrieverWithThreshold(repo Repository, threshold float64) *FuzzySimilarityRetriever {
	if threshold < 0 {
		threshold = 0
	}
	if threshold > 1 {
		threshold = 1
	}
	return &FuzzySimilarityRetriever{
		repo:      repo,
		threshold: threshold,
	}
}

// FindSimilar finds similar records using the repository's fuzzy search
func (r *FuzzySimilarityRetriever) FindSimilar(ctx context.Context, userID, input string, limit int) ([]models.ChatRecord, error) {
	if limit < 1 {
		limit = 3
	}

	// Use repository's FindSimilar which does LIKE-based fuzzy matching
	return r.repo.FindSimilar(ctx, userID, input, limit)
}

// SimilarityResult wraps a record with its similarity score
type SimilarityResult struct {
	Record    models.ChatRecord `json:"record"`
	Score     float64           `json:"score"`      // Similarity score (0.0 to 1.0)
	MatchedOn string            `json:"matched_on"` // What field matched (user_input, response)
}

// FormatSimilarRecordsForPrompt formats similar records for inclusion in a prompt
func FormatSimilarRecordsForPrompt(records []models.ChatRecord, maxRecords int) string {
	if len(records) == 0 {
		return ""
	}

	if maxRecords > 0 && len(records) > maxRecords {
		records = records[:maxRecords]
	}

	var result string
	result = "You previously recorded similar items:\n"
	for i, record := range records {
		payload, _ := record.GetResponsePayload()
		summary := ""
		if payload != nil {
			if payload.Explanation != "" {
				summary = payload.Explanation
			} else if payload.Answer != "" {
				summary = payload.Answer
			} else if len(payload.Plan) > 0 {
				summary = payload.Plan[0]
			} else if payload.Introduction != "" {
				summary = payload.Introduction
			}
		}
		// Truncate summary
		if len(summary) > 100 {
			summary = summary[:100] + "..."
		}
		result += "\n" + string(rune('1'+i)) + ". \"" + record.UserInput + "\" (" + record.InputType + ")"
		if summary != "" {
			result += ": " + summary
		}
	}
	return result
}

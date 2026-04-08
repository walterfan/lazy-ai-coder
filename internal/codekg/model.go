package codekg

import (
	"database/sql"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type Repository struct {
	ID         string         `gorm:"primaryKey;size:64" json:"id"`
	Name       string         `gorm:"size:255;not null" json:"name"`
	URL        string         `gorm:"size:1024" json:"url"`
	LocalPath  string         `gorm:"size:1024" json:"local_path"`
	Branch     string         `gorm:"size:128;default:main" json:"branch"`
	LastCommit string         `gorm:"size:64" json:"last_commit"`
	LastSync   *time.Time     `json:"last_sync"`
	Status     string         `gorm:"size:32;default:idle" json:"status"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Repository) TableName() string { return "codekg_repositories" }

type Entity struct {
	ID         string         `gorm:"primaryKey;size:128" json:"id"`
	RepoID     string         `gorm:"size:64;index;not null" json:"repo_id"`
	EntityType string         `gorm:"size:32;index;not null" json:"entity_type"`
	Name       string         `gorm:"size:512;index;not null" json:"name"`
	FilePath   string         `gorm:"size:1024;index;not null" json:"file_path"`
	StartLine  int            `json:"start_line"`
	EndLine    int            `json:"end_line"`
	Signature  string         `gorm:"type:text" json:"signature"`
	DocString  string         `gorm:"type:text" json:"doc_string"`
	Body       string         `gorm:"type:text" json:"body"`
	Summary    string         `gorm:"type:text" json:"summary"`
	Language   string         `gorm:"size:32" json:"language"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Entity) TableName() string { return "codekg_entities" }

// KnowledgeDoc stores generated PKB-style documents (repo-map, overview, architecture)
type KnowledgeDoc struct {
	ID        string         `gorm:"primaryKey;size:128" json:"id"`
	RepoID    string         `gorm:"size:64;index;not null" json:"repo_id"`
	DocType   string         `gorm:"size:32;index;not null" json:"doc_type"`
	Title     string         `gorm:"size:255" json:"title"`
	Content   string         `gorm:"type:text" json:"content"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (KnowledgeDoc) TableName() string { return "codekg_knowledge_docs" }

type SyncStatus struct {
	JobID           string `json:"job_id"`
	Status          string `json:"status"`
	TotalFiles      int    `json:"total_files"`
	ProcessedFiles  int    `json:"processed_files"`
	EntitiesCreated int    `json:"entities_created"`
	EntitiesUpdated int    `json:"entities_updated"`
	EntitiesDeleted int    `json:"entities_deleted"`
	Error           string `json:"error,omitempty"`
}

type SearchRequest struct {
	Query      string `json:"query" binding:"required"`
	TopK       int    `json:"top_k"`
	EntityType string `json:"entity_type"`
	RepoID     string `json:"repo_id"`
}

type SearchResult struct {
	Entities []Entity `json:"entities"`
	Answer   string   `json:"answer"`
}

func AutoMigrate(db *gorm.DB) error {
	if err := db.AutoMigrate(&Repository{}, &Entity{}, &KnowledgeDoc{}); err != nil {
		return err
	}
	return initVecTable(db)
}

// initVecTable creates the sqlite-vec virtual table for KNN search.
// The embedding dimension is configurable; defaults to 1536 (text-embedding-3-small).
func initVecTable(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("get underlying sql.DB: %w", err)
	}
	return InitVecTableWithDim(sqlDB, 1536)
}

func InitVecTableWithDim(sqlDB *sql.DB, dim int) error {
	query := fmt.Sprintf(`CREATE VIRTUAL TABLE IF NOT EXISTS codekg_entity_vec USING vec0(
		entity_id TEXT PRIMARY KEY,
		embedding float[%d] distance_metric=cosine
	)`, dim)
	_, err := sqlDB.Exec(query)
	return err
}

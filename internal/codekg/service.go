package codekg

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	sqlite_vec "github.com/asg017/sqlite-vec-go-bindings/cgo"
	"github.com/google/uuid"
	"github.com/walterfan/lazy-ai-coder/internal/llm"
	ilog "github.com/walterfan/lazy-ai-coder/internal/log"
	"github.com/walterfan/lazy-ai-coder/internal/rag"
	"gorm.io/gorm"
)

func init() {
	sqlite_vec.Auto()
}

type Service struct {
	db       *gorm.DB
	sqlDB    *sql.DB
	parser   *rag.CodeParser
	embedder *rag.EmbeddingService
	mu       sync.Mutex
	syncJobs map[string]*SyncStatus
}

func NewService(db *gorm.DB) *Service {
	if err := AutoMigrate(db); err != nil {
		ilog.GetLogger().Errorf("Failed to auto-migrate codekg tables: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		ilog.GetLogger().Errorf("Failed to get sql.DB for sqlite-vec: %v", err)
	}

	var embedder *rag.EmbeddingService
	apiKey := os.Getenv("LLM_API_KEY")
	if apiKey != "" {
		embedder = rag.NewEmbeddingService(rag.EmbeddingConfig{
			APIKey:  apiKey,
			BaseURL: os.Getenv("LLM_BASE_URL"),
			Model:   getEnvOrDefault("CODEKG_EMBEDDING_MODEL", "text-embedding-3-small"),
		})
	}

	return &Service{
		db:       db,
		sqlDB:    sqlDB,
		parser:   rag.NewCodeParser(),
		embedder: embedder,
		syncJobs: make(map[string]*SyncStatus),
	}
}

func (s *Service) RegisterRepo(repo *Repository) error {
	if repo.ID == "" {
		repo.ID = uuid.New().String()[:8]
	}
	if repo.Branch == "" {
		repo.Branch = "main"
	}
	return s.db.Create(repo).Error
}

func (s *Service) ListRepos() ([]Repository, error) {
	var repos []Repository
	err := s.db.Order("created_at desc").Find(&repos).Error
	return repos, err
}

func (s *Service) GetRepo(id string) (*Repository, error) {
	var repo Repository
	err := s.db.First(&repo, "id = ?", id).Error
	return &repo, err
}

func (s *Service) DeleteRepo(id string) error {
	s.deleteVecEntries(id)
	s.db.Where("repo_id = ?", id).Delete(&KnowledgeDoc{})
	s.db.Where("repo_id = ?", id).Delete(&Entity{})
	return s.db.Delete(&Repository{}, "id = ?", id).Error
}

func (s *Service) GetSyncStatus(repoID string) *SyncStatus {
	s.mu.Lock()
	defer s.mu.Unlock()
	if status, ok := s.syncJobs[repoID]; ok {
		return status
	}
	return &SyncStatus{Status: "idle"}
}

func (s *Service) TriggerSync(repoID string) (string, error) {
	repo, err := s.GetRepo(repoID)
	if err != nil {
		return "", fmt.Errorf("repository not found: %w", err)
	}

	repoPath := repo.LocalPath
	if repoPath == "" {
		repoPath = repo.URL
	}
	if repoPath == "" {
		return "", fmt.Errorf("no local_path or url configured for repo %s", repo.Name)
	}

	jobID := uuid.New().String()[:8]
	status := &SyncStatus{
		JobID:  jobID,
		Status: "running",
	}
	s.mu.Lock()
	s.syncJobs[repoID] = status
	s.mu.Unlock()

	s.db.Model(repo).Update("status", "syncing")

	go s.runSync(repo, repoPath, status)

	return jobID, nil
}

func (s *Service) runSync(repo *Repository, repoPath string, status *SyncStatus) {
	defer func() {
		if r := recover(); r != nil {
			status.Status = "failed"
			status.Error = fmt.Sprintf("panic: %v", r)
		}
	}()

	codeFiles, err := s.collectCodeFiles(repoPath)
	if err != nil {
		status.Status = "failed"
		status.Error = fmt.Sprintf("failed to scan directory: %v", err)
		s.db.Model(repo).Updates(map[string]interface{}{"status": "failed"})
		return
	}

	status.TotalFiles = len(codeFiles)
	if len(codeFiles) == 0 {
		status.Status = "completed"
		now := time.Now()
		s.db.Model(repo).Updates(map[string]interface{}{"status": "idle", "last_sync": &now})
		return
	}

	s.db.Where("repo_id = ?", repo.ID).Delete(&Entity{})
	s.deleteVecEntries(repo.ID)

	var allEntities []Entity
	for _, filePath := range codeFiles {
		entities := s.parseFile(repo.ID, repoPath, filePath)
		allEntities = append(allEntities, entities...)
		status.ProcessedFiles++
		status.EntitiesCreated += len(entities)
	}

	batchSize := 100
	for i := 0; i < len(allEntities); i += batchSize {
		end := i + batchSize
		if end > len(allEntities) {
			end = len(allEntities)
		}
		if err := s.db.CreateInBatches(allEntities[i:end], batchSize).Error; err != nil {
			ilog.GetLogger().Errorf("Failed to save entities batch: %v", err)
		}
	}

	if s.embedder != nil && len(allEntities) > 0 {
		s.generateAndStoreEmbeddings(allEntities)
	}

	s.generateKnowledgeDocs(repo, repoPath, allEntities, codeFiles)

	now := time.Now()
	s.db.Model(repo).Updates(map[string]interface{}{
		"status":    "idle",
		"last_sync": &now,
	})
	status.Status = "completed"
}

func (s *Service) collectCodeFiles(rootPath string) ([]string, error) {
	var files []string
	skipDirs := map[string]bool{
		"vendor": true, "node_modules": true, ".git": true,
		"__pycache__": true, ".idea": true, ".vscode": true,
		"dist": true, "build": true, "target": true, ".next": true,
	}

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			if skipDirs[info.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		fileInfo := rag.DetectFileType(path)
		if fileInfo.Type == rag.FileTypeCode {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func (s *Service) parseFile(repoID, repoRoot, filePath string) []Entity {
	metadata, err := s.parser.ParseCode(filePath)
	if err != nil {
		ilog.GetLogger().Warnf("Failed to parse %s: %v", filePath, err)
		return nil
	}

	relPath, _ := filepath.Rel(repoRoot, filePath)
	if relPath == "" {
		relPath = filePath
	}

	var entities []Entity

	for _, fn := range metadata.Functions {
		entities = append(entities, Entity{
			ID:         entityID(repoID, relPath, "function", fn.Name, fn.StartLine),
			RepoID:     repoID,
			EntityType: "function",
			Name:       fn.Name,
			FilePath:   relPath,
			StartLine:  fn.StartLine,
			EndLine:    fn.EndLine,
			Signature:  extractSignature(fn.Content),
			DocString:  fn.Comment,
			Body:       truncate(fn.Content, 4000),
			Language:   string(metadata.Language),
		})
	}

	for _, cls := range metadata.Classes {
		entities = append(entities, Entity{
			ID:         entityID(repoID, relPath, "class", cls.Name, cls.StartLine),
			RepoID:     repoID,
			EntityType: "class",
			Name:       cls.Name,
			FilePath:   relPath,
			StartLine:  cls.StartLine,
			EndLine:    cls.EndLine,
			Signature:  extractSignature(cls.Content),
			DocString:  cls.Comment,
			Body:       truncate(cls.Content, 4000),
			Language:   string(metadata.Language),
		})
	}

	return entities
}

func (s *Service) generateAndStoreEmbeddings(entities []Entity) {
	if s.sqlDB == nil {
		ilog.GetLogger().Warn("sql.DB not available, skipping embedding storage")
		return
	}

	const batchSize = 50
	for i := 0; i < len(entities); i += batchSize {
		end := i + batchSize
		if end > len(entities) {
			end = len(entities)
		}

		var texts []string
		for _, e := range entities[i:end] {
			input := fmt.Sprintf("Language: %s\nType: %s\nName: %s\nSignature: %s\nDoc: %s",
				e.Language, e.EntityType, e.Name, e.Signature, e.DocString)
			texts = append(texts, input)
		}

		embeddings, err := s.embedder.GenerateEmbeddings(texts)
		if err != nil {
			ilog.GetLogger().Warnf("Failed to generate embeddings for batch %d: %v", i/batchSize, err)
			continue
		}

		for j, emb := range embeddings {
			blob, err := sqlite_vec.SerializeFloat32(emb)
			if err != nil {
				ilog.GetLogger().Warnf("Failed to serialize embedding for %s: %v", entities[i+j].ID, err)
				continue
			}
			eid := entities[i+j].ID
			s.sqlDB.Exec("DELETE FROM codekg_entity_vec WHERE entity_id = ?", eid)
			_, err = s.sqlDB.Exec(
				"INSERT INTO codekg_entity_vec(entity_id, embedding) VALUES (?, ?)",
				eid, blob)
			if err != nil {
				ilog.GetLogger().Warnf("Failed to store embedding for %s: %v", eid, err)
			}
		}
	}
}

func (s *Service) Search(req SearchRequest) (*SearchResult, error) {
	if req.TopK <= 0 {
		req.TopK = 10
	}

	var entities []Entity

	if s.embedder != nil && s.sqlDB != nil {
		ranked, err := s.searchByVec(req)
		if err != nil {
			ilog.GetLogger().Warnf("Vec search failed, falling back to keyword: %v", err)
			entities = s.keywordFallback(req)
		} else {
			entities = ranked
		}
	} else {
		entities = s.keywordFallback(req)
	}

	answer, err := s.generateAnswer(req.Query, entities)
	if err != nil {
		ilog.GetLogger().Warnf("Failed to generate answer: %v", err)
		answer = "Failed to generate answer. See matched code entities below."
	}

	return &SearchResult{
		Entities: entities,
		Answer:   answer,
	}, nil
}

// searchByVec uses sqlite-vec KNN to find the closest entity embeddings.
func (s *Service) searchByVec(req SearchRequest) ([]Entity, error) {
	queryEmb, err := s.embedder.GenerateEmbedding(req.Query)
	if err != nil {
		return nil, fmt.Errorf("embed query: %w", err)
	}
	blob, err := sqlite_vec.SerializeFloat32(queryEmb)
	if err != nil {
		return nil, fmt.Errorf("serialize query vec: %w", err)
	}

	rows, err := s.sqlDB.Query(
		"SELECT entity_id, distance FROM codekg_entity_vec WHERE embedding MATCH ? AND k = ?",
		blob, req.TopK)
	if err != nil {
		return nil, fmt.Errorf("vec knn query: %w", err)
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		var dist float64
		if err := rows.Scan(&id, &dist); err != nil {
			continue
		}
		ids = append(ids, id)
	}
	if len(ids) == 0 {
		return nil, nil
	}

	var entities []Entity
	q := s.db.Where("id IN ?", ids)
	if req.RepoID != "" {
		q = q.Where("repo_id = ?", req.RepoID)
	}
	if req.EntityType != "" {
		q = q.Where("entity_type = ?", req.EntityType)
	}
	q.Select("id, repo_id, entity_type, name, file_path, start_line, end_line, signature, doc_string, body, summary, language, created_at").
		Find(&entities)

	idOrder := make(map[string]int, len(ids))
	for i, id := range ids {
		idOrder[id] = i
	}
	sortedEntities := make([]Entity, len(entities))
	copy(sortedEntities, entities)
	for i := 0; i < len(sortedEntities); i++ {
		for j := i + 1; j < len(sortedEntities); j++ {
			if idOrder[sortedEntities[j].ID] < idOrder[sortedEntities[i].ID] {
				sortedEntities[i], sortedEntities[j] = sortedEntities[j], sortedEntities[i]
			}
		}
	}
	return sortedEntities, nil
}

func (s *Service) keywordFallback(req SearchRequest) []Entity {
	var entities []Entity
	query := s.db.Where("1=1")
	if req.RepoID != "" {
		query = query.Where("repo_id = ?", req.RepoID)
	}
	if req.EntityType != "" {
		query = query.Where("entity_type = ?", req.EntityType)
	}
	query.Find(&entities)
	return s.rankByKeyword(req.Query, entities, req.TopK)
}

func (s *Service) rankByKeyword(query string, entities []Entity, topK int) []Entity {
	queryLower := strings.ToLower(query)
	words := strings.Fields(queryLower)

	type scored struct {
		entity Entity
		score  int
	}
	var results []scored

	for _, e := range entities {
		score := 0
		nameLower := strings.ToLower(e.Name)
		bodyLower := strings.ToLower(e.Body)
		for _, w := range words {
			if strings.Contains(nameLower, w) {
				score += 3
			}
			if strings.Contains(bodyLower, w) {
				score += 1
			}
		}
		if score > 0 {
			results = append(results, scored{entity: e, score: score})
		}
	}

	for i := 0; i < len(results); i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].score > results[i].score {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	var out []Entity
	for i := 0; i < len(results) && i < topK; i++ {
		out = append(out, results[i].entity)
	}
	return out
}

func (s *Service) generateAnswer(query string, entities []Entity) (string, error) {
	settings := llm.LLMSettings{
		BaseUrl:     os.Getenv("LLM_BASE_URL"),
		ApiKey:      os.Getenv("LLM_API_KEY"),
		Model:       getEnvOrDefault("LLM_MODEL", "gpt-4o-mini"),
		Temperature: 0.1,
	}
	if settings.ApiKey == "" {
		return "LLM API key not configured. Cannot generate answer.", nil
	}

	systemPrompt := `You are a code knowledge base expert. Answer the user's question based on the code entities provided. 
Always cite specific function/file locations. If the code context is insufficient, say so honestly.`

	var contextParts []string
	for i, e := range entities {
		part := fmt.Sprintf("### [%d] %s `%s` (%s:%d-%d)\n```\n%s\n```",
			i+1, e.EntityType, e.Name, e.FilePath, e.StartLine, e.EndLine,
			truncate(e.Body, 1500))
		contextParts = append(contextParts, part)
	}

	userPrompt := fmt.Sprintf("## Code Context\n%s\n\n## Question\n%s",
		strings.Join(contextParts, "\n\n"), query)

	return llm.AskLLM(systemPrompt, userPrompt, settings)
}

func (s *Service) GetEntities(repoID, entityType string, page, perPage int) ([]Entity, int64, error) {
	if page <= 0 {
		page = 1
	}
	if perPage <= 0 {
		perPage = 20
	}

	query := s.db.Model(&Entity{})
	if repoID != "" {
		query = query.Where("repo_id = ?", repoID)
	}
	if entityType != "" {
		query = query.Where("entity_type = ?", entityType)
	}

	var total int64
	query.Count(&total)

	var entities []Entity
	err := query.Select("id, repo_id, entity_type, name, file_path, start_line, end_line, signature, doc_string, summary, language, created_at").
		Offset((page - 1) * perPage).Limit(perPage).
		Order("name asc").Find(&entities).Error
	return entities, total, err
}

// --- PKB-style knowledge doc generation ---

func (s *Service) generateKnowledgeDocs(repo *Repository, repoPath string, entities []Entity, codeFiles []string) {
	s.db.Where("repo_id = ?", repo.ID).Delete(&KnowledgeDoc{})

	repoMap := s.buildRepoMap(repo, repoPath, entities, codeFiles)
	s.db.Create(&repoMap)

	overview := s.buildProjectOverview(repo, repoPath, entities, codeFiles)
	s.db.Create(&overview)
}

func (s *Service) buildRepoMap(repo *Repository, repoPath string, entities []Entity, codeFiles []string) KnowledgeDoc {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Repository Map: %s\n\n", repo.Name))

	// Directory structure
	sb.WriteString("## Directory Structure\n\n```\n")
	dirTree := buildDirTree(repoPath, codeFiles)
	sb.WriteString(dirTree)
	sb.WriteString("```\n\n")

	// Language breakdown
	langCount := map[string]int{}
	for _, e := range entities {
		langCount[e.Language]++
	}
	sb.WriteString("## Language Breakdown\n\n")
	sb.WriteString("| Language | Entities |\n|---|---|\n")
	for lang, count := range langCount {
		sb.WriteString(fmt.Sprintf("| %s | %d |\n", lang, count))
	}

	// Entity type summary
	typeCount := map[string]int{}
	for _, e := range entities {
		typeCount[e.EntityType]++
	}
	sb.WriteString("\n## Entity Summary\n\n")
	sb.WriteString("| Type | Count |\n|---|---|\n")
	for t, count := range typeCount {
		sb.WriteString(fmt.Sprintf("| %s | %d |\n", t, count))
	}

	// Key entry points (functions named main, init, New*, Handle*, Setup*)
	sb.WriteString("\n## Key Entry Points\n\n")
	for _, e := range entities {
		if e.EntityType == "function" && isEntryPoint(e.Name) {
			sb.WriteString(fmt.Sprintf("- **%s** — `%s:%d`\n", e.Name, e.FilePath, e.StartLine))
		}
	}

	sb.WriteString(fmt.Sprintf("\n## Stats\n\n- **Total files**: %d\n- **Total entities**: %d\n", len(codeFiles), len(entities)))

	return KnowledgeDoc{
		ID:      entityID(repo.ID, "repo-map", "doc", "repo-map", 0),
		RepoID:  repo.ID,
		DocType: "repo-map",
		Title:   "Repository Map",
		Content: sb.String(),
	}
}

func (s *Service) buildProjectOverview(repo *Repository, repoPath string, entities []Entity, codeFiles []string) KnowledgeDoc {
	settings := llm.LLMSettings{
		BaseUrl:     os.Getenv("LLM_BASE_URL"),
		ApiKey:      os.Getenv("LLM_API_KEY"),
		Model:       getEnvOrDefault("LLM_MODEL", "gpt-4o-mini"),
		Temperature: 0.3,
	}

	if settings.ApiKey == "" {
		return KnowledgeDoc{
			ID:      entityID(repo.ID, "overview", "doc", "overview", 0),
			RepoID:  repo.ID,
			DocType: "overview",
			Title:   "Project Overview",
			Content: fmt.Sprintf("# Project Overview: %s\n\nLLM API key not configured. Run with LLM_API_KEY to generate an AI-powered overview.\n\n**Files**: %d | **Entities**: %d",
				repo.Name, len(codeFiles), len(entities)),
		}
	}

	systemPrompt := `You are a senior engineer writing a concise project overview for an AI-readable knowledge base.
Follow this structure:
1. Purpose — what problem the project solves (1-2 sentences)
2. Technology Stack — languages, frameworks, databases
3. Architecture — high-level module structure
4. Key Components — most important modules/packages and their roles
5. Entry Points — where the app starts, main routes, CLI commands

Be specific. Reference actual file paths and function names from the provided data.
Keep it under 500 words. Use Markdown.`

	var entitiesSummary strings.Builder
	entitiesSummary.WriteString(fmt.Sprintf("Repository: %s\nPath: %s\nFiles: %d\n\n", repo.Name, repoPath, len(codeFiles)))

	shown := 0
	for _, e := range entities {
		if shown >= 80 {
			break
		}
		entitiesSummary.WriteString(fmt.Sprintf("[%s] %s — %s:%d %s\n", e.EntityType, e.Name, e.FilePath, e.StartLine, e.Signature))
		shown++
	}

	answer, err := llm.AskLLM(systemPrompt, entitiesSummary.String(), settings)
	if err != nil {
		answer = fmt.Sprintf("# Project Overview: %s\n\nFailed to generate overview: %v", repo.Name, err)
	}

	return KnowledgeDoc{
		ID:      entityID(repo.ID, "overview", "doc", "overview", 0),
		RepoID:  repo.ID,
		DocType: "overview",
		Title:   "Project Overview",
		Content: answer,
	}
}

func (s *Service) GetKnowledgeDocs(repoID string) ([]KnowledgeDoc, error) {
	var docs []KnowledgeDoc
	err := s.db.Where("repo_id = ?", repoID).Order("doc_type asc").Find(&docs).Error
	return docs, err
}

func buildDirTree(rootPath string, codeFiles []string) string {
	dirs := map[string]int{}
	for _, f := range codeFiles {
		relPath, _ := filepath.Rel(rootPath, f)
		if relPath == "" {
			relPath = f
		}
		dir := filepath.Dir(relPath)
		dirs[dir]++
	}

	var lines []string
	for dir, count := range dirs {
		lines = append(lines, fmt.Sprintf("%-50s (%d files)", dir+"/", count))
	}

	// Simple alphabetical sort
	for i := 0; i < len(lines); i++ {
		for j := i + 1; j < len(lines); j++ {
			if lines[j] < lines[i] {
				lines[i], lines[j] = lines[j], lines[i]
			}
		}
	}

	return strings.Join(lines, "\n")
}

func isEntryPoint(name string) bool {
	lower := strings.ToLower(name)
	prefixes := []string{"main", "init", "new", "handle", "setup", "run", "start", "serve"}
	for _, p := range prefixes {
		if strings.HasPrefix(lower, p) {
			return true
		}
	}
	return false
}

// deleteVecEntries removes sqlite-vec rows for all entities belonging to a repo.
func (s *Service) deleteVecEntries(repoID string) {
	if s.sqlDB == nil {
		return
	}
	_, err := s.sqlDB.Exec(
		`DELETE FROM codekg_entity_vec WHERE entity_id IN (
			SELECT id FROM codekg_entities WHERE repo_id = ?
		)`, repoID)
	if err != nil {
		ilog.GetLogger().Warnf("Failed to delete vec entries for repo %s: %v", repoID, err)
	}
}

// --- helpers ---

func entityID(repoID, filePath, entityType, name string, startLine int) string {
	h := sha256.Sum256([]byte(fmt.Sprintf("%s:%s:%s:%s:%d", repoID, filePath, entityType, name, startLine)))
	return fmt.Sprintf("%x", h[:8])
}

func extractSignature(content string) string {
	lines := strings.SplitN(content, "\n", 3)
	if len(lines) > 0 {
		return strings.TrimSpace(lines[0])
	}
	return ""
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func getEnvOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

package codekg

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	return db
}

func setupTestService(t *testing.T) *Service {
	t.Helper()
	os.Unsetenv("LLM_API_KEY")
	db := setupTestDB(t)
	svc := NewService(db)
	require.NotNil(t, svc)
	return svc
}

func TestNewService_MigratesTablesAndVec(t *testing.T) {
	svc := setupTestService(t)

	var count int64
	svc.db.Raw("SELECT count(*) FROM sqlite_master WHERE type='table' AND name='codekg_repositories'").Scan(&count)
	assert.Equal(t, int64(1), count, "codekg_repositories table should exist")

	svc.db.Raw("SELECT count(*) FROM sqlite_master WHERE type='table' AND name='codekg_entities'").Scan(&count)
	assert.Equal(t, int64(1), count, "codekg_entities table should exist")

	svc.db.Raw("SELECT count(*) FROM sqlite_master WHERE type='table' AND name='codekg_entity_vec'").Scan(&count)
	assert.Equal(t, int64(1), count, "codekg_entity_vec virtual table should exist")
}

func TestRegisterAndListRepos(t *testing.T) {
	svc := setupTestService(t)

	repo := &Repository{Name: "test-repo", LocalPath: "/tmp/test-repo"}
	err := svc.RegisterRepo(repo)
	require.NoError(t, err)
	assert.NotEmpty(t, repo.ID)

	repos, err := svc.ListRepos()
	require.NoError(t, err)
	assert.Len(t, repos, 1)
	assert.Equal(t, "test-repo", repos[0].Name)
}

func TestGetRepo(t *testing.T) {
	svc := setupTestService(t)

	repo := &Repository{Name: "find-me", LocalPath: "/tmp/find-me"}
	require.NoError(t, svc.RegisterRepo(repo))

	found, err := svc.GetRepo(repo.ID)
	require.NoError(t, err)
	assert.Equal(t, "find-me", found.Name)

	_, err = svc.GetRepo("nonexistent")
	assert.Error(t, err)
}

func TestParseFile_GoSource(t *testing.T) {
	svc := setupTestService(t)

	tmpDir := t.TempDir()
	goFile := filepath.Join(tmpDir, "hello.go")
	err := os.WriteFile(goFile, []byte(`package main

import "fmt"

// Greet returns a greeting message.
func Greet(name string) string {
	return fmt.Sprintf("Hello, %s!", name)
}

func main() {
	fmt.Println(Greet("world"))
}
`), 0644)
	require.NoError(t, err)

	entities := svc.parseFile("repo1", tmpDir, goFile)
	assert.NotEmpty(t, entities, "should extract at least one entity from a Go file")

	var names []string
	for _, e := range entities {
		names = append(names, e.Name)
		assert.Equal(t, "repo1", e.RepoID)
		assert.Equal(t, "function", e.EntityType)
		assert.NotEmpty(t, e.Body)
		assert.Equal(t, "hello.go", e.FilePath)
	}
	assert.Contains(t, names, "Greet")
	assert.Contains(t, names, "main")
}

func TestKeywordSearch_WithoutEmbeddings(t *testing.T) {
	svc := setupTestService(t)

	repo := &Repository{Name: "kw-repo", LocalPath: "/tmp/kw"}
	require.NoError(t, svc.RegisterRepo(repo))

	entities := []Entity{
		{ID: "e1", RepoID: repo.ID, EntityType: "function", Name: "HandleRequest", Body: "func HandleRequest(w http.ResponseWriter, r *http.Request) {}", Language: "go"},
		{ID: "e2", RepoID: repo.ID, EntityType: "function", Name: "ParseConfig", Body: "func ParseConfig(path string) (*Config, error) {}", Language: "go"},
		{ID: "e3", RepoID: repo.ID, EntityType: "function", Name: "ConnectDB", Body: "func ConnectDB() (*sql.DB, error) {}", Language: "go"},
	}
	require.NoError(t, svc.db.Create(&entities).Error)

	result, err := svc.Search(SearchRequest{
		Query:  "handle request",
		RepoID: repo.ID,
		TopK:   5,
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.NotEmpty(t, result.Entities)
	assert.Equal(t, "HandleRequest", result.Entities[0].Name)
}

func TestKeywordSearch_FilterByEntityType(t *testing.T) {
	svc := setupTestService(t)

	repo := &Repository{Name: "filter-repo", LocalPath: "/tmp/filter"}
	require.NoError(t, svc.RegisterRepo(repo))

	entities := []Entity{
		{ID: "f1", RepoID: repo.ID, EntityType: "function", Name: "NewService", Body: "func NewService() *Service {}", Language: "go"},
		{ID: "c1", RepoID: repo.ID, EntityType: "class", Name: "Service", Body: "type Service struct {}", Language: "go"},
	}
	require.NoError(t, svc.db.Create(&entities).Error)

	result, err := svc.Search(SearchRequest{
		Query:      "service",
		RepoID:     repo.ID,
		EntityType: "class",
		TopK:       5,
	})
	require.NoError(t, err)
	assert.Len(t, result.Entities, 1)
	assert.Equal(t, "Service", result.Entities[0].Name)
}

func TestKeywordSearch_NoMatch(t *testing.T) {
	svc := setupTestService(t)

	repo := &Repository{Name: "empty-repo", LocalPath: "/tmp/empty"}
	require.NoError(t, svc.RegisterRepo(repo))

	entities := []Entity{
		{ID: "x1", RepoID: repo.ID, EntityType: "function", Name: "Alpha", Body: "func Alpha() {}", Language: "go"},
	}
	require.NoError(t, svc.db.Create(&entities).Error)

	result, err := svc.Search(SearchRequest{
		Query:  "zzzznotfound",
		RepoID: repo.ID,
		TopK:   5,
	})
	require.NoError(t, err)
	assert.Empty(t, result.Entities)
}

func TestGetEntities_Pagination(t *testing.T) {
	svc := setupTestService(t)

	repo := &Repository{Name: "page-repo", LocalPath: "/tmp/page"}
	require.NoError(t, svc.RegisterRepo(repo))

	for i := 0; i < 15; i++ {
		e := Entity{
			ID: entityID(repo.ID, "file.go", "function", string(rune('A'+i)), i),
			RepoID: repo.ID, EntityType: "function",
			Name: string(rune('A' + i)), FilePath: "file.go", Language: "go",
		}
		require.NoError(t, svc.db.Create(&e).Error)
	}

	entities, total, err := svc.GetEntities(repo.ID, "", 1, 5)
	require.NoError(t, err)
	assert.Equal(t, int64(15), total)
	assert.Len(t, entities, 5)

	entities2, _, err := svc.GetEntities(repo.ID, "", 2, 5)
	require.NoError(t, err)
	assert.Len(t, entities2, 5)
	assert.NotEqual(t, entities[0].ID, entities2[0].ID)
}

func TestSyncStatus_DefaultIdle(t *testing.T) {
	svc := setupTestService(t)
	status := svc.GetSyncStatus("nonexistent")
	assert.Equal(t, "idle", status.Status)
}

func TestCollectCodeFiles(t *testing.T) {
	svc := setupTestService(t)
	tmpDir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "readme.md"), []byte("# Readme"), 0644))
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "node_modules"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "node_modules", "dep.js"), []byte("module.exports={}"), 0644))

	files, err := svc.collectCodeFiles(tmpDir)
	require.NoError(t, err)
	assert.Len(t, files, 1, "should find main.go only, skipping .md and node_modules")
	assert.Contains(t, files[0], "main.go")
}

func TestRankByKeyword_Ordering(t *testing.T) {
	svc := setupTestService(t)

	entities := []Entity{
		{Name: "alpha", Body: "does something"},
		{Name: "beta", Body: "alpha beta gamma"},
		{Name: "alpha_handler", Body: "handles alpha requests"},
	}

	ranked := svc.rankByKeyword("alpha", entities, 10)
	require.Len(t, ranked, 3)
	assert.Equal(t, "alpha_handler", ranked[0].Name, "name match (3) + body match (1) = 4 should rank first")
	assert.Equal(t, "alpha", ranked[1].Name, "name match (3) should rank second")
	assert.Equal(t, "beta", ranked[2].Name, "body-only match (1) should rank last")
}

func TestBuildRepoMap(t *testing.T) {
	svc := setupTestService(t)
	tmpDir := t.TempDir()

	repo := &Repository{ID: "rm1", Name: "map-repo"}
	entities := []Entity{
		{EntityType: "function", Name: "main", FilePath: "cmd/main.go", StartLine: 5, Language: "go"},
		{EntityType: "function", Name: "NewService", FilePath: "internal/svc.go", StartLine: 10, Language: "go"},
		{EntityType: "class", Name: "Config", FilePath: "internal/config.go", StartLine: 1, Language: "go"},
	}
	codeFiles := []string{
		filepath.Join(tmpDir, "cmd/main.go"),
		filepath.Join(tmpDir, "internal/svc.go"),
		filepath.Join(tmpDir, "internal/config.go"),
	}

	doc := svc.buildRepoMap(repo, tmpDir, entities, codeFiles)
	assert.Equal(t, "repo-map", doc.DocType)
	assert.Contains(t, doc.Content, "map-repo")
	assert.Contains(t, doc.Content, "go")
	assert.Contains(t, doc.Content, "function")
	assert.Contains(t, doc.Content, "main")
	assert.Contains(t, doc.Content, "NewService")
}

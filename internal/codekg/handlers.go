package codekg

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handlers struct {
	service *Service
}

func NewHandlers(service *Service) *Handlers {
	return &Handlers{service: service}
}

func (h *Handlers) RegisterRoutes(r *gin.Engine) {
	g := r.Group("/api/v1/codekg")
	{
		g.GET("/repos", h.listRepos)
		g.POST("/repos", h.registerRepo)
		g.DELETE("/repos/:id", h.deleteRepo)
		g.POST("/repos/:id/sync", h.triggerSync)
		g.GET("/repos/:id/status", h.getSyncStatus)
		g.POST("/search", h.search)
		g.GET("/entities", h.listEntities)
		g.GET("/repos/:id/knowledge", h.getKnowledgeDocs)
	}
}

type registerRepoRequest struct {
	Name      string `json:"name" binding:"required"`
	URL       string `json:"url"`
	LocalPath string `json:"local_path"`
	Branch    string `json:"branch"`
}

func (h *Handlers) registerRepo(c *gin.Context) {
	var req registerRepoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	repo := &Repository{
		Name:      req.Name,
		URL:       req.URL,
		LocalPath: req.LocalPath,
		Branch:    req.Branch,
	}

	if repo.LocalPath == "" && repo.URL != "" {
		repo.LocalPath = repo.URL
	}

	if err := h.service.RegisterRepo(repo); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, repo)
}

func (h *Handlers) deleteRepo(c *gin.Context) {
	repoID := c.Param("id")
	if err := h.service.DeleteRepo(repoID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "repository deleted"})
}

func (h *Handlers) listRepos(c *gin.Context) {
	repos, err := h.service.ListRepos()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, repos)
}

func (h *Handlers) triggerSync(c *gin.Context) {
	repoID := c.Param("id")
	jobID, err := h.service.TriggerSync(repoID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"job_id": jobID})
}

func (h *Handlers) getSyncStatus(c *gin.Context) {
	repoID := c.Param("id")
	status := h.service.GetSyncStatus(repoID)
	c.JSON(http.StatusOK, status)
}

func (h *Handlers) search(c *gin.Context) {
	var req SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.service.Search(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handlers) getKnowledgeDocs(c *gin.Context) {
	repoID := c.Param("id")
	docs, err := h.service.GetKnowledgeDocs(repoID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, docs)
}

func (h *Handlers) listEntities(c *gin.Context) {
	repoID := c.Query("repo_id")
	entityType := c.Query("type")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

	entities, total, err := h.service.GetEntities(repoID, entityType, page, perPage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":  entities,
		"total": total,
		"page":  page,
	})
}

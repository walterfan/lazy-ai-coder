package handlers

import (
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/walterfan/lazy-ai-coder/internal/assets"
)

// AssetsHandlers serves search and download for assets (commands, rules, skills)
type AssetsHandlers struct {
	loader *assets.Loader
}

// NewAssetsHandlers creates handlers with the given assets loader
func NewAssetsHandlers(loader *assets.Loader) *AssetsHandlers {
	return &AssetsHandlers{loader: loader}
}

// ListRequest query params: type (command|rule|skill|all), q (search), category
// ListAssets godoc
// @Summary List assets (commands, rules, skills)
// @Description List and search commands, rules, and skills from the assets folder
// @Tags assets
// @Produce json
// @Param type query string false "Filter: command, rule, skill, or all" default(all)
// @Param q query string false "Search query (name, path, content snippet)"
// @Param category query string false "Filter by category (e.g. golang, awesome)"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/assets [get]
func (h *AssetsHandlers) ListAssets(c *gin.Context) {
	typ := strings.ToLower(c.DefaultQuery("type", "all"))
	q := strings.TrimSpace(c.Query("q"))
	category := strings.TrimSpace(c.Query("category"))

	var list []assets.Item
	var err error

	switch typ {
	case "command", "commands":
		list, err = h.loader.ListCommands()
	case "rule", "rules":
		list, err = h.loader.ListRules()
	case "skill", "skills":
		list, err = h.loader.ListSkills()
	case "all", "":
		var commands, rules, skills []assets.Item
		if commands, err = h.loader.ListCommands(); err != nil {
			break
		}
		if rules, err = h.loader.ListRules(); err != nil {
			break
		}
		if skills, err = h.loader.ListSkills(); err != nil {
			break
		}
		list = append(append(commands, rules...), skills...)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid type: use command, rule, skill, or all"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if q != "" {
		list = assets.Search(list, q)
	}
	if category != "" {
		var filtered []assets.Item
		cat := strings.ToLower(category)
		for _, it := range list {
			if strings.ToLower(it.Category) == cat {
				filtered = append(filtered, it)
			}
		}
		list = filtered
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  list,
		"total": len(list),
	})
}

// DownloadAsset godoc
// @Summary Download an asset file
// @Description Get full content of an asset by path (e.g. commands/bug.fix.md). Use download=1 to force attachment.
// @Tags assets
// @Produce plain
// @Param path query string true "Relative path, e.g. commands/bug.fix.md, rules/golang/go.mdc, skills/awesome/golang-patterns/SKILL.md"
// @Param download query bool false "If true, respond with Content-Disposition attachment"
// @Success 200 {string} string "File content"
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/assets/download [get]
func (h *AssetsHandlers) DownloadAsset(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "path is required"})
		return
	}

	data, filename, err := h.loader.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrPermission) {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	if c.Query("download") == "1" || c.Query("download") == "true" {
		c.Header("Content-Disposition", "attachment; filename=\""+filename+"\"")
	}
	c.Data(http.StatusOK, "text/plain; charset=utf-8", data)
}

// DownloadSkillZip godoc
// @Summary Download a skill folder as a zip
// @Description Zip the entire skill directory (SKILL.md + subfolders) and return it
// @Tags assets
// @Produce application/zip
// @Param path query string true "Relative path to skill, e.g. skills/awesome/continuous-learning-v2/SKILL.md"
// @Success 200 {file} file "Zip archive"
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/assets/download-skill [get]
func (h *AssetsHandlers) DownloadSkillZip(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "path is required"})
		return
	}

	skillDir := assets.SkillFolderFromPath(path)

	data, filename, err := h.loader.ZipSkillFolder(skillDir)
	if err != nil {
		if errors.Is(err, os.ErrPermission) {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
		if os.IsNotExist(err) || errors.Is(err, os.ErrInvalid) {
			c.JSON(http.StatusNotFound, gin.H{"error": "skill folder not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Disposition", "attachment; filename=\""+filename+"\"")
	c.Data(http.StatusOK, "application/zip", data)
}

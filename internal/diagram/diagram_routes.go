package diagram

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/walterfan/lazy-ai-coder/internal/log"
	"github.com/walterfan/lazy-ai-coder/internal/models"
)

// DiagramHandlers handles diagram-related HTTP requests
type DiagramHandlers struct {
	service *DiagramService
}

// NewDiagramHandlers creates a new DiagramHandlers
func NewDiagramHandlers(service *DiagramService) *DiagramHandlers {
	return &DiagramHandlers{
		service: service,
	}
}

// HandleDrawRequest godoc
// @Summary Generate diagrams with PlantUML
// @Description Generate PlantUML diagrams (UML and mindmap) from text
// @Tags diagrams
// @Accept json
// @Produce json
// @Param request body models.DrawAPIRequest true "Draw Request"
// @Success 200 {array} models.DrawAPIResponse
// @Failure 400 {object} map[string]string
// @Router /draw [post]
func (h *DiagramHandlers) HandleDrawRequest(c *gin.Context) {
	logger := log.GetLogger()
	var req models.DrawAPIRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Invalid request body", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	logger.Infof("Received draw request: %s", req.AssistantPrompt)

	arrResp := h.service.GenerateDiagrams(req.AssistantPrompt)
	c.JSON(http.StatusOK, arrResp)
}


package smartprompt

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/walterfan/lazy-ai-coder/internal/log"
	"github.com/walterfan/lazy-ai-coder/internal/models"
)

// SmartPromptHandlers handles smart prompt-related HTTP requests
type SmartPromptHandlers struct {
	service *SmartPromptService
}

// NewSmartPromptHandlers creates a new SmartPromptHandlers
func NewSmartPromptHandlers(service *SmartPromptService) *SmartPromptHandlers {
	return &SmartPromptHandlers{
		service: service,
	}
}

// HandleSmartPromptGenerate godoc
// @Summary Generate smart prompt with context detection
// @Description Generate CAR format prompt with GitLab context analysis, code examples, and quality scoring
// @Tags smart-prompt
// @Accept json
// @Produce json
// @Param request body models.SmartPromptRequest true "Smart Prompt Request"
// @Success 200 {object} models.SmartPromptResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/smart-prompt/generate [post]
func (h *SmartPromptHandlers) HandleSmartPromptGenerate(c *gin.Context) {
	logger := log.GetLogger()
	var req models.SmartPromptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	logger.Infof("Received smart prompt request: %s", req.Input)

	response, err := h.service.GenerateSmartPrompt(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// HandleGetPresets godoc
// @Summary Get available presets
// @Description Get list of available prompt generation presets
// @Tags smart-prompt
// @Produce json
// @Success 200 {array} models.Preset
// @Router /api/v1/smart-prompt/presets [get]
func (h *SmartPromptHandlers) HandleGetPresets(c *gin.Context) {
	c.JSON(http.StatusOK, DefaultPresets)
}

// HandleGetFrameworks godoc
// @Summary Get all prompt engineering frameworks
// @Description Get list of all available prompt engineering frameworks (CRISPE, RISEN, CO-STAR, APE, CAR)
// @Tags smart-prompt
// @Produce json
// @Success 200 {array} Framework
// @Router /api/v1/smart-prompt/frameworks [get]
func (h *SmartPromptHandlers) HandleGetFrameworks(c *gin.Context) {
	c.JSON(http.StatusOK, AllFrameworks)
}

// HandleGetFramework godoc
// @Summary Get framework by ID
// @Description Get details of a specific prompt engineering framework
// @Tags smart-prompt
// @Produce json
// @Param id path string true "Framework ID"
// @Success 200 {object} Framework
// @Failure 404 {object} map[string]string
// @Router /api/v1/smart-prompt/frameworks/{id} [get]
func (h *SmartPromptHandlers) HandleGetFramework(c *gin.Context) {
	frameworkID := c.Param("id")
	framework := GetFrameworkByID(frameworkID)
	if framework == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Framework not found"})
		return
	}
	c.JSON(http.StatusOK, framework)
}

// HandleGetTemplateCategories godoc
// @Summary Get template categories
// @Description Get list of all template categories
// @Tags smart-prompt
// @Produce json
// @Success 200 {array} TemplateCategory
// @Router /api/v1/smart-prompt/templates/categories [get]
func (h *SmartPromptHandlers) HandleGetTemplateCategories(c *gin.Context) {
	c.JSON(http.StatusOK, AllCategories)
}

// HandleGetTemplates godoc
// @Summary Get all prompt templates
// @Description Get list of all available prompt templates, optionally filtered by category
// @Tags smart-prompt
// @Produce json
// @Param category query string false "Filter by category"
// @Success 200 {array} PromptTemplate
// @Router /api/v1/smart-prompt/templates [get]
func (h *SmartPromptHandlers) HandleGetTemplates(c *gin.Context) {
	category := c.Query("category")
	if category != "" {
		templates := GetTemplatesByCategory(category)
		c.JSON(http.StatusOK, templates)
		return
	}
	c.JSON(http.StatusOK, AllTemplates)
}

// HandleGetTemplate godoc
// @Summary Get template by ID
// @Description Get details of a specific prompt template
// @Tags smart-prompt
// @Produce json
// @Param id path string true "Template ID"
// @Success 200 {object} PromptTemplate
// @Failure 404 {object} map[string]string
// @Router /api/v1/smart-prompt/templates/{id} [get]
func (h *SmartPromptHandlers) HandleGetTemplate(c *gin.Context) {
	templateID := c.Param("id")
	template := GetTemplateByID(templateID)
	if template == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
		return
	}
	c.JSON(http.StatusOK, template)
}

// RefinePromptRequest represents a prompt refinement request
type RefinePromptRequest struct {
	Prompt   string         `json:"prompt" binding:"required"`
	Settings models.Settings `json:"settings"`
}

// HandleRefinePrompt godoc
// @Summary Refine prompt with LLM suggestions
// @Description Use LLM to analyze prompt and provide improvement suggestions
// @Tags smart-prompt
// @Accept json
// @Produce json
// @Param request body RefinePromptRequest true "Refinement Request"
// @Success 200 {object} RefinementResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/smart-prompt/refine [post]
func (h *SmartPromptHandlers) HandleRefinePrompt(c *gin.Context) {
	logger := log.GetLogger()
	var req RefinePromptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	logger.Infof("Received refinement request for prompt of length %d", len(req.Prompt))

	response, err := RefinePrompt(req.Prompt, req.Settings)
	if err != nil {
		logger.Errorf("Failed to refine prompt: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// RefineWithRequirementsRequest represents a prompt refinement request with user requirements
type RefineWithRequirementsRequest struct {
	SystemPrompt string          `json:"system_prompt" binding:"required"`
	UserPrompt   string          `json:"user_prompt" binding:"required"`
	Requirements string          `json:"requirements" binding:"required"`
	Settings     models.Settings `json:"settings"`
}

// RefineWithRequirementsResponse represents the refinement response
type RefineWithRequirementsResponse struct {
	SystemPrompt string `json:"system_prompt"`
	UserPrompt   string `json:"user_prompt"`
}

// HandleRefineWithRequirements godoc
// @Summary Refine prompt with user requirements
// @Description Use LLM to refine system and user prompts based on user's specific requirements
// @Tags smart-prompt
// @Accept json
// @Produce json
// @Param request body RefineWithRequirementsRequest true "Refinement Request with Requirements"
// @Success 200 {object} RefineWithRequirementsResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/smart-prompt/refine-with-requirements [post]
func (h *SmartPromptHandlers) HandleRefineWithRequirements(c *gin.Context) {
	logger := log.GetLogger()
	var req RefineWithRequirementsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	logger.Infof("Received refinement request with requirements: %s", req.Requirements)

	response, err := RefinePromptWithRequirements(req.SystemPrompt, req.UserPrompt, req.Requirements, req.Settings)
	if err != nil {
		logger.Errorf("Failed to refine prompt with requirements: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// QuickRefineRequest represents a quick refinement request
type QuickRefineRequest struct {
	Prompt      string `json:"prompt" binding:"required"`
	FrameworkID string `json:"framework_id"`
}

// HandleQuickRefine godoc
// @Summary Get quick refinement suggestions
// @Description Get framework-specific quick suggestions without calling LLM
// @Tags smart-prompt
// @Accept json
// @Produce json
// @Param request body QuickRefineRequest true "Quick Refinement Request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /api/v1/smart-prompt/quick-refine [post]
func (h *SmartPromptHandlers) HandleQuickRefine(c *gin.Context) {
	var req QuickRefineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	suggestions := QuickRefineWithFramework(req.Prompt, req.FrameworkID)
	qualityScore := EstimatePromptQuality(req.Prompt)

	c.JSON(http.StatusOK, gin.H{
		"suggestions":    suggestions,
		"quality_score":  qualityScore,
		"max_score":      10.0,
	})
}

// GenerateFromFrameworkRequest represents a framework-based generation request
type GenerateFromFrameworkRequest struct {
	FrameworkID string            `json:"framework_id" binding:"required"`
	Fields      map[string]string `json:"fields" binding:"required"`
}

// HandleGenerateFromFramework godoc
// @Summary Generate prompt from framework
// @Description Generate a prompt using a specific framework and field values
// @Tags smart-prompt
// @Accept json
// @Produce json
// @Param request body GenerateFromFrameworkRequest true "Generation Request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/smart-prompt/generate-from-framework [post]
func (h *SmartPromptHandlers) HandleGenerateFromFramework(c *gin.Context) {
	logger := log.GetLogger()
	var req GenerateFromFrameworkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	framework := GetFrameworkByID(req.FrameworkID)
	if framework == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Framework not found"})
		return
	}

	logger.Infof("Generating prompt using framework: %s", framework.Name)

	// Generate prompt using the framework template (returns system + user prompts)
	promptPair, err := GeneratePromptFromFramework(req.FrameworkID, req.Fields)
	if err != nil {
		logger.Errorf("Failed to generate prompt: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Calculate quality score based on full prompt
	qualityScore := EstimatePromptQuality(promptPair.FullPrompt)

	c.JSON(http.StatusOK, gin.H{
		"system_prompt":  promptPair.SystemPrompt,
		"user_prompt":    promptPair.UserPrompt,
		"full_prompt":    promptPair.FullPrompt,
		"framework_id":   req.FrameworkID,
		"framework_name": framework.Name,
		"quality_score":  qualityScore,
		"max_score":      10.0,
	})
}

// HandleUseTemplate godoc
// @Summary Use a template
// @Description Increment use count for a template and return it
// @Tags smart-prompt
// @Param id path string true "Template ID"
// @Produce json
// @Success 200 {object} PromptTemplate
// @Failure 404 {object} map[string]string
// @Router /api/v1/smart-prompt/templates/{id}/use [post]
func (h *SmartPromptHandlers) HandleUseTemplate(c *gin.Context) {
	templateID := c.Param("id")
	template := GetTemplateByID(templateID)
	if template == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
		return
	}

	// Increment use count
	IncrementTemplateUseCount(templateID)

	c.JSON(http.StatusOK, template)
}

// AutoFillFieldsRequest represents an auto-fill request
type AutoFillFieldsRequest struct {
	FrameworkID string          `json:"framework_id" binding:"required"`
	UserInput   string          `json:"user_input" binding:"required"`
	Settings    models.Settings `json:"settings"`
}

// AutoFillFieldsResponse represents the auto-fill response
type AutoFillFieldsResponse struct {
	Fields map[string]string `json:"fields"`
}

// HandleAutoFillFields godoc
// @Summary Auto-fill framework fields with LLM
// @Description Use LLM to automatically fill framework fields based on user input
// @Tags smart-prompt
// @Accept json
// @Produce json
// @Param request body AutoFillFieldsRequest true "Auto-fill Request"
// @Success 200 {object} AutoFillFieldsResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/smart-prompt/auto-fill-fields [post]
func (h *SmartPromptHandlers) HandleAutoFillFields(c *gin.Context) {
	logger := log.GetLogger()
	var req AutoFillFieldsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	framework := GetFrameworkByID(req.FrameworkID)
	if framework == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Framework not found"})
		return
	}

	logger.Infof("Auto-filling fields for framework: %s with input: %s", framework.Name, req.UserInput)

	fields, err := AutoFillFrameworkFields(framework, req.UserInput, req.Settings)
	if err != nil {
		logger.Errorf("Failed to auto-fill fields: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, AutoFillFieldsResponse{
		Fields: fields,
	})
}

// HandleCreateFramework godoc
// @Summary Create a new framework
// @Description Create a custom prompt engineering framework
// @Tags smart-prompt
// @Accept json
// @Produce json
// @Param framework body Framework true "Framework data"
// @Success 201 {object} Framework
// @Failure 400 {object} map[string]string
// @Router /api/v1/smart-prompt/frameworks [post]
func (h *SmartPromptHandlers) HandleCreateFramework(c *gin.Context) {
	logger := log.GetLogger()
	var framework Framework
	if err := c.ShouldBindJSON(&framework); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate framework ID doesn't already exist
	if GetFrameworkByID(framework.ID) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Framework with this ID already exists"})
		return
	}

	logger.Infof("Creating new framework: %s", framework.ID)

	// Add to AllFrameworks (in-memory)
	AllFrameworks = append(AllFrameworks, framework)
	c.JSON(http.StatusCreated, framework)
}

// HandleUpdateFramework godoc
// @Summary Update an existing framework
// @Description Update a prompt engineering framework by ID
// @Tags smart-prompt
// @Accept json
// @Produce json
// @Param id path string true "Framework ID"
// @Param framework body Framework true "Updated framework data"
// @Success 200 {object} Framework
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/smart-prompt/frameworks/{id} [put]
func (h *SmartPromptHandlers) HandleUpdateFramework(c *gin.Context) {
	logger := log.GetLogger()
	frameworkID := c.Param("id")
	var updatedFramework Framework
	if err := c.ShouldBindJSON(&updatedFramework); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	logger.Infof("Updating framework: %s", frameworkID)

	// Find and update in AllFrameworks
	for i, fw := range AllFrameworks {
		if fw.ID == frameworkID {
			// Preserve the ID from URL parameter
			updatedFramework.ID = frameworkID
			AllFrameworks[i] = updatedFramework
			c.JSON(http.StatusOK, updatedFramework)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Framework not found"})
}

// HandleDeleteFramework godoc
// @Summary Delete a framework
// @Description Delete a prompt engineering framework by ID
// @Tags smart-prompt
// @Param id path string true "Framework ID"
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/smart-prompt/frameworks/{id} [delete]
func (h *SmartPromptHandlers) HandleDeleteFramework(c *gin.Context) {
	logger := log.GetLogger()
	frameworkID := c.Param("id")

	logger.Infof("Deleting framework: %s", frameworkID)

	// Find and remove from AllFrameworks
	for i, fw := range AllFrameworks {
		if fw.ID == frameworkID {
			AllFrameworks = append(AllFrameworks[:i], AllFrameworks[i+1:]...)
			c.JSON(http.StatusOK, gin.H{"message": "Framework deleted successfully"})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Framework not found"})
}

// HandleCreateTemplate godoc
// @Summary Create a new template
// @Description Create a custom prompt template
// @Tags smart-prompt
// @Accept json
// @Produce json
// @Param template body PromptTemplate true "Template data"
// @Success 201 {object} PromptTemplate
// @Failure 400 {object} map[string]string
// @Router /api/v1/smart-prompt/templates [post]
func (h *SmartPromptHandlers) HandleCreateTemplate(c *gin.Context) {
	logger := log.GetLogger()
	var template PromptTemplate
	if err := c.ShouldBindJSON(&template); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate template ID doesn't already exist
	if GetTemplateByID(template.ID) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Template with this ID already exists"})
		return
	}

	logger.Infof("Creating new template: %s", template.ID)

	// Initialize use count
	template.UseCount = 0

	// Add to AllTemplates (in-memory)
	AllTemplates = append(AllTemplates, template)
	c.JSON(http.StatusCreated, template)
}

// HandleUpdateTemplate godoc
// @Summary Update an existing template
// @Description Update a prompt template by ID
// @Tags smart-prompt
// @Accept json
// @Produce json
// @Param id path string true "Template ID"
// @Param template body PromptTemplate true "Updated template data"
// @Success 200 {object} PromptTemplate
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/smart-prompt/templates/{id} [put]
func (h *SmartPromptHandlers) HandleUpdateTemplate(c *gin.Context) {
	logger := log.GetLogger()
	templateID := c.Param("id")
	var updatedTemplate PromptTemplate
	if err := c.ShouldBindJSON(&updatedTemplate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	logger.Infof("Updating template: %s", templateID)

	// Find and update in AllTemplates
	for i, tmpl := range AllTemplates {
		if tmpl.ID == templateID {
			// Preserve the ID and use count
			updatedTemplate.ID = templateID
			updatedTemplate.UseCount = tmpl.UseCount
			AllTemplates[i] = updatedTemplate
			c.JSON(http.StatusOK, updatedTemplate)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
}

// HandleDeleteTemplate godoc
// @Summary Delete a template
// @Description Delete a prompt template by ID
// @Tags smart-prompt
// @Param id path string true "Template ID"
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/smart-prompt/templates/{id} [delete]
func (h *SmartPromptHandlers) HandleDeleteTemplate(c *gin.Context) {
	logger := log.GetLogger()
	templateID := c.Param("id")

	logger.Infof("Deleting template: %s", templateID)

	// Find and remove from AllTemplates
	for i, tmpl := range AllTemplates {
		if tmpl.ID == templateID {
			AllTemplates = append(AllTemplates[:i], AllTemplates[i+1:]...)
			c.JSON(http.StatusOK, gin.H{"message": "Template deleted successfully"})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
}


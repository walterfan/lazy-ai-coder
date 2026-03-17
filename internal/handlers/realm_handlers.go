package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	apimodels "github.com/walterfan/lazy-ai-coder/internal/models"
	"github.com/walterfan/lazy-ai-coder/internal/services"
)

// RealmHandlers handles realm management operations (for super_admin)
type RealmHandlers struct {
	db           *gorm.DB
	realmService *services.RealmService
}

// NewRealmHandlers creates a new realm handlers
func NewRealmHandlers(db *gorm.DB) *RealmHandlers {
	return &RealmHandlers{
		db:           db,
		realmService: services.NewRealmService(db),
	}
}

// CreateRealmRequest represents the request body for creating a realm
type CreateRealmRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

// UpdateRealmRequest represents the request body for updating a realm
type UpdateRealmRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// RealmResponse represents a realm with user count
type RealmResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	UserCount   int    `json:"user_count"`
	CreatedBy   string `json:"created_by"`
	CreatedAt   string `json:"created_at"`
}

// GetAllRealms lists all realms with user counts
// @Summary List all realms
// @Description Get all realms in the system with user counts (super_admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} RealmResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/admin/realms [get]
func (h *RealmHandlers) GetAllRealms(c *gin.Context) {
	realms, err := h.realmService.ListRealmsWithUserCount()
	if err != nil {
		c.JSON(http.StatusInternalServerError, apimodels.ErrorResponse{
			Error: "Failed to fetch realms: " + err.Error(),
		})
		return
	}

	response := make([]RealmResponse, len(realms))
	for i, realm := range realms {
		response[i] = RealmResponse{
			ID:          realm.ID,
			Name:        realm.Name,
			Description: realm.Description,
			UserCount:   realm.UserCount,
			CreatedBy:   realm.CreatedBy,
			CreatedAt:   realm.CreatedTime.Format("2006-01-02 15:04:05"),
		}
	}

	c.JSON(http.StatusOK, response)
}

// GetRealmByID retrieves a specific realm by ID
// @Summary Get realm details
// @Description Get details of a specific realm (super_admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Realm ID"
// @Success 200 {object} models.Realm
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/admin/realms/{id} [get]
func (h *RealmHandlers) GetRealmByID(c *gin.Context) {
	realmID := c.Param("id")

	realm, err := h.realmService.GetRealmByID(realmID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, apimodels.ErrorResponse{
				Error: "Realm not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, apimodels.ErrorResponse{
				Error: "Failed to fetch realm: " + err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, realm)
}

// CreateRealm creates a new realm
// @Summary Create realm
// @Description Create a new realm (super_admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateRealmRequest true "Realm creation request"
// @Success 201 {object} models.Realm
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/admin/realms [post]
func (h *RealmHandlers) CreateRealm(c *gin.Context) {
	var req CreateRealmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, apimodels.ErrorResponse{
			Error: "Invalid request body: " + err.Error(),
		})
		return
	}

	adminUsername := c.GetString("username")

	realm, err := h.realmService.CreateRealm(req.Name, req.Description, adminUsername)
	if err != nil {
		c.JSON(http.StatusInternalServerError, apimodels.ErrorResponse{
			Error: "Failed to create realm: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, realm)
}

// UpdateRealm updates an existing realm
// @Summary Update realm
// @Description Update an existing realm (super_admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Realm ID"
// @Param request body UpdateRealmRequest true "Realm update request"
// @Success 200 {object} models.Realm
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/admin/realms/{id} [put]
func (h *RealmHandlers) UpdateRealm(c *gin.Context) {
	realmID := c.Param("id")

	var req UpdateRealmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, apimodels.ErrorResponse{
			Error: "Invalid request body: " + err.Error(),
		})
		return
	}

	adminUsername := c.GetString("username")

	realm, err := h.realmService.UpdateRealm(realmID, req.Name, req.Description, adminUsername)
	if err != nil {
		c.JSON(http.StatusInternalServerError, apimodels.ErrorResponse{
			Error: "Failed to update realm: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, realm)
}

// DeleteRealm soft deletes a realm
// @Summary Delete realm
// @Description Soft delete a realm (super_admin only, realm must have no users)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Realm ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/admin/realms/{id} [delete]
func (h *RealmHandlers) DeleteRealm(c *gin.Context) {
	realmID := c.Param("id")

	if err := h.realmService.DeleteRealm(realmID); err != nil {
		c.JSON(http.StatusBadRequest, apimodels.ErrorResponse{
			Error: "Failed to delete realm: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Realm deleted successfully",
	})
}

// GetUsersInRealm lists all active users in a specific realm
// @Summary Get users in realm
// @Description List all active users in a realm (super_admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Realm ID"
// @Success 200 {array} models.User
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/admin/realms/{id}/users [get]
func (h *RealmHandlers) GetUsersInRealm(c *gin.Context) {
	realmID := c.Param("id")

	// Verify realm exists
	if _, err := h.realmService.GetRealmByID(realmID); err != nil {
		c.JSON(http.StatusNotFound, apimodels.ErrorResponse{
			Error: "Realm not found",
		})
		return
	}

	users, err := h.realmService.ListUsersInRealm(realmID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, apimodels.ErrorResponse{
			Error: "Failed to fetch users: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, users)
}

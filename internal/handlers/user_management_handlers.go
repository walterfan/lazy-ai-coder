package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	apimodels "github.com/walterfan/lazy-ai-coder/internal/models"
	"github.com/walterfan/lazy-ai-coder/internal/services"
	"github.com/walterfan/lazy-ai-coder/pkg/models"
)

// UserManagementHandlers handles user management operations (for super_admin)
type UserManagementHandlers struct {
	db           *gorm.DB
	roleService  *services.RoleService
	realmService *services.RealmService
}

// NewUserManagementHandlers creates a new user management handlers
func NewUserManagementHandlers(db *gorm.DB) *UserManagementHandlers {
	return &UserManagementHandlers{
		db:           db,
		roleService:  services.NewRoleService(db),
		realmService: services.NewRealmService(db),
	}
}

// UserApprovalRequest represents the request body for approving a user
type UserApprovalRequest struct {
	RealmID string `json:"realm_id" binding:"required"`
	RoleID  string `json:"role_id" binding:"required"`
}

// UserUpdateRealmRequest represents the request body for updating user realm
type UserUpdateRealmRequest struct {
	RealmID string `json:"realm_id" binding:"required"`
}

// UserUpdateRoleRequest represents the request body for updating user role
type UserUpdateRoleRequest struct {
	RoleID string `json:"role_id" binding:"required"`
}

// UserListResponse represents a user with role and realm information
type UserListResponse struct {
	ID         string     `json:"id"`
	Username   string     `json:"username"`
	Email      string     `json:"email"`
	Name       string     `json:"name"`
	RealmID    string     `json:"realm_id"`
	RealmName  string     `json:"realm_name"`
	IsActive   bool       `json:"is_active"`
	Roles      []RoleInfo `json:"roles"`
	CreatedAt  time.Time  `json:"created_at"`
	LastLogin  *time.Time `json:"last_login"`
	AvatarURL  string     `json:"avatar_url,omitempty"`
}

// RoleInfo represents minimal role information
type RoleInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// GetPendingUsers lists all users pending approval (is_active=false)
// @Summary List pending users
// @Description Get all users pending admin approval (super_admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} UserListResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/admin/pending-users [get]
func (h *UserManagementHandlers) GetPendingUsers(c *gin.Context) {
	// Get pending users from database
	var users []models.User
	if err := h.db.Where("is_active = ?", false).Order("created_time DESC").Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, apimodels.ErrorResponse{
			Error: "Failed to fetch pending users",
		})
		return
	}

	// Convert to response format
	response := make([]UserListResponse, len(users))
	for i, user := range users {
		response[i] = UserListResponse{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			Name:      user.Name,
			RealmID:   user.RealmID,
			RealmName: "",
			IsActive:  user.IsActive,
			Roles:     []RoleInfo{},
			CreatedAt: user.CreatedTime,
			LastLogin: user.LastLoginAt,
			AvatarURL: user.AvatarURL,
		}
	}

	c.JSON(http.StatusOK, response)
}

// ApproveUser approves a pending user, assigns realm and role
// @Summary Approve user
// @Description Approve a pending user and assign realm and role (super_admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Param request body UserApprovalRequest true "Approval request"
// @Success 200 {object} UserListResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/admin/users/{id}/approve [post]
func (h *UserManagementHandlers) ApproveUser(c *gin.Context) {
	userID := c.Param("id")
	adminUsername := c.GetString("username")

	var req UserApprovalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, apimodels.ErrorResponse{
			Error: "Invalid request body: " + err.Error(),
		})
		return
	}

	// Check if user exists and is pending
	var user models.User
	if err := h.db.Where("id = ? AND is_active = ?", userID, false).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, apimodels.ErrorResponse{
				Error: "User not found or already approved",
			})
		} else {
			c.JSON(http.StatusInternalServerError, apimodels.ErrorResponse{
				Error: "Failed to find user",
			})
		}
		return
	}

	// Verify realm exists
	if _, err := h.realmService.GetRealmByID(req.RealmID); err != nil {
		c.JSON(http.StatusBadRequest, apimodels.ErrorResponse{
			Error: "Invalid realm: " + err.Error(),
		})
		return
	}

	// Verify role exists
	if _, err := h.roleService.GetRoleByID(req.RoleID); err != nil {
		c.JSON(http.StatusBadRequest, apimodels.ErrorResponse{
			Error: "Invalid role: " + err.Error(),
		})
		return
	}

	// Start transaction
	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Update user: set active, assign realm
	if err := tx.Model(&user).Updates(map[string]interface{}{
		"is_active":  true,
		"realm_id":   req.RealmID,
		"updated_by": adminUsername,
	}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, apimodels.ErrorResponse{
			Error: "Failed to approve user",
		})
		return
	}

	// Assign role to user using the transaction
	roleService := services.NewRoleService(tx)
	if err := roleService.AssignRoleToUser(userID, req.RoleID, adminUsername); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, apimodels.ErrorResponse{
			Error: "Failed to assign role: " + err.Error(),
		})
		return
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, apimodels.ErrorResponse{
			Error: "Failed to commit approval",
		})
		return
	}

	// Fetch updated user with roles
	roles, _ := h.roleService.GetUserRoles(userID)
	realm, _ := h.realmService.GetRealmByID(req.RealmID)

	roleInfos := make([]RoleInfo, len(roles))
	for i, role := range roles {
		roleInfos[i] = RoleInfo{ID: role.ID, Name: role.Name}
	}

	response := UserListResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Name:      user.Name,
		RealmID:   user.RealmID,
		RealmName: realm.Name,
		IsActive:  true,
		Roles:     roleInfos,
		CreatedAt: user.CreatedTime,
		LastLogin: user.LastLoginAt,
		AvatarURL: user.AvatarURL,
	}

	c.JSON(http.StatusOK, response)
}

// RejectUser rejects a pending user (soft delete or deactivate)
// @Summary Reject user
// @Description Reject a pending user application (super_admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Success 200 {object} map[string]string
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/admin/users/{id}/reject [post]
func (h *UserManagementHandlers) RejectUser(c *gin.Context) {
	userID := c.Param("id")

	// Check if user exists and is pending
	var user models.User
	if err := h.db.Where("id = ? AND is_active = ?", userID, false).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, apimodels.ErrorResponse{
				Error: "User not found or already approved",
			})
		} else {
			c.JSON(http.StatusInternalServerError, apimodels.ErrorResponse{
				Error: "Failed to find user",
			})
		}
		return
	}

	// Soft delete the user
	if err := h.db.Delete(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, apimodels.ErrorResponse{
			Error: "Failed to reject user",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User rejected successfully",
	})
}

// GetAllUsers lists all users with their roles and realms (for super_admin)
// @Summary List all users
// @Description Get all users in the system (super_admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param is_active query boolean false "Filter by active status"
// @Param realm_id query string false "Filter by realm ID"
// @Success 200 {array} UserListResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/admin/users [get]
func (h *UserManagementHandlers) GetAllUsers(c *gin.Context) {
	// Build query
	query := h.db.Model(&models.User{})

	// Filter by active status if provided
	if isActiveStr := c.Query("is_active"); isActiveStr != "" {
		query = query.Where("is_active = ?", isActiveStr == "true")
	}

	// Filter by realm if provided
	if realmID := c.Query("realm_id"); realmID != "" {
		query = query.Where("realm_id = ?", realmID)
	}

	// Get users
	var users []models.User
	if err := query.Order("created_time DESC").Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, apimodels.ErrorResponse{
			Error: "Failed to fetch users",
		})
		return
	}

	// Convert to response format with roles and realm info
	response := make([]UserListResponse, len(users))
	for i, user := range users {
		roles, _ := h.roleService.GetUserRoles(user.ID)
		roleInfos := make([]RoleInfo, len(roles))
		for j, role := range roles {
			roleInfos[j] = RoleInfo{ID: role.ID, Name: role.Name}
		}

		realmName := ""
		if user.RealmID != "" {
			if realm, err := h.realmService.GetRealmByID(user.RealmID); err == nil {
				realmName = realm.Name
			}
		}

		response[i] = UserListResponse{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			Name:      user.Name,
			RealmID:   user.RealmID,
			RealmName: realmName,
			IsActive:  user.IsActive,
			Roles:     roleInfos,
			CreatedAt: user.CreatedTime,
			LastLogin: user.LastLoginAt,
			AvatarURL: user.AvatarURL,
		}
	}

	c.JSON(http.StatusOK, response)
}

// UpdateUserRealm changes a user's realm
// @Summary Update user realm
// @Description Change a user's realm assignment (super_admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Param request body UserUpdateRealmRequest true "Realm update request"
// @Success 200 {object} map[string]string
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/admin/users/{id}/realm [put]
func (h *UserManagementHandlers) UpdateUserRealm(c *gin.Context) {
	userID := c.Param("id")
	adminUsername := c.GetString("username")

	var req UserUpdateRealmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, apimodels.ErrorResponse{
			Error: "Invalid request body: " + err.Error(),
		})
		return
	}

	// Assign user to new realm
	if err := h.realmService.AssignUserToRealm(userID, req.RealmID, adminUsername); err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, apimodels.ErrorResponse{
				Error: "User or realm not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, apimodels.ErrorResponse{
				Error: "Failed to update realm: " + err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User realm updated successfully",
	})
}

// UpdateUserRole changes a user's role (replaces existing roles with new role)
// @Summary Update user role
// @Description Change a user's role assignment (super_admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Param request body UserUpdateRoleRequest true "Role update request"
// @Success 200 {object} map[string]string
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/admin/users/{id}/role [put]
func (h *UserManagementHandlers) UpdateUserRole(c *gin.Context) {
	userID := c.Param("id")
	adminUsername := c.GetString("username")

	var req UserUpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, apimodels.ErrorResponse{
			Error: "Invalid request body: " + err.Error(),
		})
		return
	}

	// Verify role exists
	if _, err := h.roleService.GetRoleByID(req.RoleID); err != nil {
		c.JSON(http.StatusBadRequest, apimodels.ErrorResponse{
			Error: "Invalid role: " + err.Error(),
		})
		return
	}

	// Start transaction
	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Remove all existing roles for this user
	if err := tx.Where("user_id = ?", userID).Delete(&models.UserRole{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, apimodels.ErrorResponse{
			Error: "Failed to remove existing roles",
		})
		return
	}

	// Assign new role
	roleService := services.NewRoleService(tx)
	if err := roleService.AssignRoleToUser(userID, req.RoleID, adminUsername); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, apimodels.ErrorResponse{
			Error: "Failed to assign new role: " + err.Error(),
		})
		return
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, apimodels.ErrorResponse{
			Error: "Failed to commit role update",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User role updated successfully",
	})
}

// DeactivateUser deactivates an active user
// @Summary Deactivate user
// @Description Deactivate an active user (super_admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Success 200 {object} map[string]string
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/admin/users/{id}/deactivate [post]
func (h *UserManagementHandlers) DeactivateUser(c *gin.Context) {
	userID := c.Param("id")
	adminUsername := c.GetString("username")

	// Check if user exists
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, apimodels.ErrorResponse{
				Error: "User not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, apimodels.ErrorResponse{
				Error: "Failed to find user",
			})
		}
		return
	}

	// Prevent deactivating super_admin users
	isSuperAdmin, _ := h.roleService.IsSuperAdmin(userID)
	if isSuperAdmin {
		c.JSON(http.StatusForbidden, apimodels.ErrorResponse{
			Error: "Cannot deactivate super admin users",
		})
		return
	}

	// Deactivate user
	if err := h.db.Model(&user).Updates(map[string]interface{}{
		"is_active":  false,
		"updated_by": adminUsername,
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, apimodels.ErrorResponse{
			Error: "Failed to deactivate user",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User deactivated successfully",
	})
}

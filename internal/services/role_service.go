package services

import (
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/walterfan/lazy-ai-coder/pkg/models"
)

// RoleService handles role management operations
type RoleService struct {
	db *gorm.DB
}

// NewRoleService creates a new role service
func NewRoleService(db *gorm.DB) *RoleService {
	return &RoleService{db: db}
}

// Role name constants
const (
	RoleSuperAdmin = "super_admin"
	RoleAdmin      = "admin"
	RoleUser       = "user"
)

// System role IDs (match migration seed data)
const (
	RoleIDSuperAdmin = "role_super_admin"
	RoleIDAdmin      = "role_admin"
	RoleIDUser       = "role_user"
)

// AssignRoleToUser assigns a role to a user
func (s *RoleService) AssignRoleToUser(userID, roleID, assignedBy string) error {
	// Check if role exists
	var role models.Role
	if err := s.db.First(&role, "id = ?", roleID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("role not found: %s", roleID)
		}
		return fmt.Errorf("failed to find role: %w", err)
	}

	// Check if user exists
	var user models.User
	if err := s.db.First(&user, "id = ?", userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("user not found: %s", userID)
		}
		return fmt.Errorf("failed to find user: %w", err)
	}

	// Check if assignment already exists (including soft deleted)
	var existingUserRole models.UserRole
	err := s.db.Unscoped().Where("user_id = ? AND role_id = ?", userID, roleID).First(&existingUserRole).Error
	if err == nil {
		// Assignment exists
		if existingUserRole.DeletedAt.Valid {
			// Restore soft deleted assignment
			return s.db.Model(&existingUserRole).Update("deleted_at", nil).Error
		}
		// Already assigned
		return fmt.Errorf("role already assigned to user")
	} else if err != gorm.ErrRecordNotFound {
		return fmt.Errorf("failed to check existing role assignment: %w", err)
	}

	// Create new assignment
	userRole := models.UserRole{
		ID:        uuid.New().String(),
		UserID:    userID,
		RoleID:    roleID,
		CreatedBy: assignedBy,
		UpdatedBy: assignedBy,
	}

	if err := s.db.Create(&userRole).Error; err != nil {
		return fmt.Errorf("failed to assign role: %w", err)
	}

	return nil
}

// AssignRoleByName assigns a role to a user by role name
func (s *RoleService) AssignRoleByName(userID, roleName, assignedBy string) error {
	// Find role by name in system realm
	var role models.Role
	if err := s.db.Where("name = ? AND realm_id = ?", roleName, "system").First(&role).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("role not found: %s", roleName)
		}
		return fmt.Errorf("failed to find role: %w", err)
	}

	return s.AssignRoleToUser(userID, role.ID, assignedBy)
}

// RemoveRoleFromUser removes a role from a user (soft delete)
func (s *RoleService) RemoveRoleFromUser(userID, roleID string) error {
	result := s.db.Where("user_id = ? AND role_id = ?", userID, roleID).Delete(&models.UserRole{})
	if result.Error != nil {
		return fmt.Errorf("failed to remove role: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("role assignment not found")
	}
	return nil
}

// GetUserRoles returns all roles assigned to a user
func (s *RoleService) GetUserRoles(userID string) ([]models.Role, error) {
	var roles []models.Role
	err := s.db.
		Joins("JOIN user_role ON user_role.role_id = role.id").
		Where("user_role.user_id = ? AND user_role.deleted_at IS NULL", userID).
		Find(&roles).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}
	return roles, nil
}

// GetUserRoleNames returns role names for a user
func (s *RoleService) GetUserRoleNames(userID string) ([]string, error) {
	roles, err := s.GetUserRoles(userID)
	if err != nil {
		return nil, err
	}

	roleNames := make([]string, len(roles))
	for i, role := range roles {
		roleNames[i] = role.Name
	}
	return roleNames, nil
}

// HasRole checks if a user has a specific role by role ID
func (s *RoleService) HasRole(userID, roleID string) (bool, error) {
	var count int64
	err := s.db.Model(&models.UserRole{}).
		Where("user_id = ? AND role_id = ?", userID, roleID).
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check role: %w", err)
	}
	return count > 0, nil
}

// HasRoleName checks if a user has a specific role by role name
func (s *RoleService) HasRoleName(userID, roleName string) (bool, error) {
	var count int64
	err := s.db.Table("user_role").
		Joins("JOIN role ON role.id = user_role.role_id").
		Where("user_role.user_id = ? AND role.name = ? AND user_role.deleted_at IS NULL", userID, roleName).
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check role: %w", err)
	}
	return count > 0, nil
}

// IsSuperAdmin checks if a user is a super admin
func (s *RoleService) IsSuperAdmin(userID string) (bool, error) {
	return s.HasRoleName(userID, RoleSuperAdmin)
}

// IsAdmin checks if a user is an admin (admin or super_admin)
func (s *RoleService) IsAdmin(userID string) (bool, error) {
	isSuperAdmin, err := s.IsSuperAdmin(userID)
	if err != nil {
		return false, err
	}
	if isSuperAdmin {
		return true, nil
	}

	return s.HasRoleName(userID, RoleAdmin)
}

// IsUser checks if a user has the user role
func (s *RoleService) IsUser(userID string) (bool, error) {
	return s.HasRoleName(userID, RoleUser)
}

// GetRoleByName gets a role by name from system realm
func (s *RoleService) GetRoleByName(roleName string) (*models.Role, error) {
	var role models.Role
	err := s.db.Where("name = ? AND realm_id = ?", roleName, "system").First(&role).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("role not found: %s", roleName)
		}
		return nil, fmt.Errorf("failed to get role: %w", err)
	}
	return &role, nil
}

// GetRoleByID gets a role by ID
func (s *RoleService) GetRoleByID(roleID string) (*models.Role, error) {
	var role models.Role
	err := s.db.First(&role, "id = ?", roleID).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("role not found: %s", roleID)
		}
		return nil, fmt.Errorf("failed to get role: %w", err)
	}
	return &role, nil
}

// ListAllRoles lists all roles in the system
func (s *RoleService) ListAllRoles() ([]models.Role, error) {
	var roles []models.Role
	err := s.db.Order("realm_id, name").Find(&roles).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list roles: %w", err)
	}
	return roles, nil
}

// CreateRole creates a new role
func (s *RoleService) CreateRole(realmID, name, description, createdBy string) (*models.Role, error) {
	// Check if role already exists in realm
	var existing models.Role
	err := s.db.Where("realm_id = ? AND name = ?", realmID, name).First(&existing).Error
	if err == nil {
		return nil, fmt.Errorf("role already exists: %s in realm %s", name, realmID)
	} else if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to check existing role: %w", err)
	}

	role := models.Role{
		ID:          uuid.New().String(),
		RealmID:     realmID,
		Name:        name,
		Description: description,
		CreatedBy:   createdBy,
		UpdatedBy:   createdBy,
	}

	if err := s.db.Create(&role).Error; err != nil {
		return nil, fmt.Errorf("failed to create role: %w", err)
	}

	return &role, nil
}

// UpdateRole updates a role's description
func (s *RoleService) UpdateRole(roleID, description, updatedBy string) error {
	result := s.db.Model(&models.Role{}).
		Where("id = ?", roleID).
		Updates(map[string]interface{}{
			"description": description,
			"updated_by":  updatedBy,
		})
	if result.Error != nil {
		return fmt.Errorf("failed to update role: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("role not found: %s", roleID)
	}
	return nil
}

// DeleteRole soft deletes a role
func (s *RoleService) DeleteRole(roleID string) error {
	// Prevent deletion of system roles
	systemRoleIDs := []string{RoleIDSuperAdmin, RoleIDAdmin, RoleIDUser}
	for _, sysRoleID := range systemRoleIDs {
		if roleID == sysRoleID {
			return fmt.Errorf("cannot delete system role: %s", roleID)
		}
	}

	result := s.db.Delete(&models.Role{}, "id = ?", roleID)
	if result.Error != nil {
		return fmt.Errorf("failed to delete role: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("role not found: %s", roleID)
	}
	return nil
}

// GetUsersWithRole returns all users with a specific role
func (s *RoleService) GetUsersWithRole(roleID string) ([]models.User, error) {
	var users []models.User
	err := s.db.
		Joins("JOIN user_role ON user_role.user_id = app_user.id").
		Where("user_role.role_id = ? AND user_role.deleted_at IS NULL", roleID).
		Find(&users).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get users with role: %w", err)
	}
	return users, nil
}

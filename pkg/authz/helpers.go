// pkg/authz/helpers.go
package authz

import (
	"github.com/walterfan/lazy-ai-coder/pkg/database"
	"gorm.io/gorm"
)

// IsSuperAdmin checks if a user has the super_admin role
// Super admins can access resources across all realms
func IsSuperAdmin(db *gorm.DB, userID string) bool {
	if db == nil || userID == "" {
		return false
	}

	// Query user_roles table to check if user has super_admin role
	var count int64
	err := db.Table("user_roles").
		Where("user_id = ? AND role_id = ? AND deleted_at IS NULL", userID, "role_super_admin").
		Count(&count).Error

	if err != nil {
		return false
	}

	return count > 0
}

// IsAdmin checks if a user has the admin role
// Admins have full access within their own realm
func IsAdmin(db *gorm.DB, userID string) bool {
	if db == nil || userID == "" {
		return false
	}

	// Query user_roles table to check if user has admin role
	var count int64
	err := db.Table("user_roles").
		Where("user_id = ? AND role_id = ? AND deleted_at IS NULL", userID, "role_admin").
		Count(&count).Error

	if err != nil {
		return false
	}

	return count > 0
}

// GetUserRoles returns all role names for a user
func GetUserRoles(db *gorm.DB, userID string) ([]string, error) {
	if db == nil || userID == "" {
		return nil, nil
	}

	var roleIDs []string
	err := db.Table("user_roles").
		Select("role_id").
		Where("user_id = ? AND deleted_at IS NULL", userID).
		Pluck("role_id", &roleIDs).Error

	if err != nil {
		return nil, err
	}

	// Convert role IDs to role names by removing "role_" prefix
	roleNames := make([]string, len(roleIDs))
	for i, roleID := range roleIDs {
		// role_super_admin -> super_admin
		if len(roleID) > 5 && roleID[:5] == "role_" {
			roleNames[i] = roleID[5:]
		} else {
			roleNames[i] = roleID
		}
	}

	return roleNames, nil
}

// ApplyRealmFilter applies realm-based filtering to a query
// If user is super_admin, no filter is applied (can see all realms)
// Otherwise, query is filtered to user's realm only
func ApplyRealmFilter(query *gorm.DB, userID string, realmID string) *gorm.DB {
	if query == nil {
		return query
	}

	// Get database from query
	db := query.Statement.DB

	// Check if user is super_admin
	if IsSuperAdmin(db, userID) {
		// Super admin can access all realms - no filter
		return query
	}

	// For non-super-admin users, filter by realm_id
	if realmID != "" && realmID != "guest" {
		return query.Where("realm_id = ?", realmID)
	}

	// Guest users or users without realm: filter to empty set (no access)
	return query.Where("1 = 0")
}

// CanModifyResource checks if a user can modify a specific resource
// Super admins can modify any resource
// Admins can modify resources in their realm
// Users can only modify resources they created (user_id match) in their realm
func CanModifyResource(db *gorm.DB, userID string, realmID string, resourceRealmID string, resourceOwnerID string) bool {
	if db == nil || userID == "" {
		return false
	}

	// Super admins can modify anything
	if IsSuperAdmin(db, userID) {
		return true
	}

	// Resource must be in user's realm
	if resourceRealmID != realmID {
		return false
	}

	// Admins can modify any resource in their realm
	if IsAdmin(db, userID) {
		return true
	}

	// Regular users can only modify their own resources
	return resourceOwnerID == userID
}

// CanAccessRealm checks if a user can access a specific realm
// Super admins can access any realm
// Other users can only access their own realm
func CanAccessRealm(db *gorm.DB, userID string, userRealmID string, targetRealmID string) bool {
	if db == nil || userID == "" {
		return false
	}

	// Super admins can access any realm
	if IsSuperAdmin(db, userID) {
		return true
	}

	// Users can only access their own realm
	return userRealmID == targetRealmID
}

// InitAuthzHelpers initializes the authorization helpers
// This should be called after database initialization
func InitAuthzHelpers() {
	// Initialize with the global database instance
	if database.DB != nil {
		// Perform any necessary initialization
		// For now, just verify database connection
	}
}

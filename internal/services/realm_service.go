package services

import (
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/walterfan/lazy-ai-coder/pkg/models"
)

// RealmService handles realm management operations
type RealmService struct {
	db *gorm.DB
}

// NewRealmService creates a new realm service
func NewRealmService(db *gorm.DB) *RealmService {
	return &RealmService{db: db}
}

// System realm constants
const (
	SystemRealmID   = "system"
	SystemRealmName = "System"
)

// CreateRealm creates a new realm
func (s *RealmService) CreateRealm(name, description, createdBy string) (*models.Realm, error) {
	// Check if realm name already exists
	var existing models.Realm
	err := s.db.Where("name = ?", name).First(&existing).Error
	if err == nil {
		return nil, fmt.Errorf("realm name already exists: %s", name)
	} else if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to check existing realm: %w", err)
	}

	realm := models.Realm{
		ID:          uuid.New().String(),
		Name:        name,
		Description: description,
		CreatedBy:   createdBy,
		UpdatedBy:   createdBy,
	}

	if err := s.db.Create(&realm).Error; err != nil {
		return nil, fmt.Errorf("failed to create realm: %w", err)
	}

	return &realm, nil
}

// GetRealmByID retrieves a realm by ID
func (s *RealmService) GetRealmByID(realmID string) (*models.Realm, error) {
	var realm models.Realm
	err := s.db.First(&realm, "id = ?", realmID).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("realm not found: %s", realmID)
		}
		return nil, fmt.Errorf("failed to get realm: %w", err)
	}
	return &realm, nil
}

// GetRealmByName retrieves a realm by name
func (s *RealmService) GetRealmByName(name string) (*models.Realm, error) {
	var realm models.Realm
	err := s.db.Where("name = ?", name).First(&realm).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("realm not found: %s", name)
		}
		return nil, fmt.Errorf("failed to get realm: %w", err)
	}
	return &realm, nil
}

// UpdateRealm updates a realm's name and/or description
func (s *RealmService) UpdateRealm(realmID, name, description, updatedBy string) (*models.Realm, error) {
	// Prevent updating system realm
	if realmID == SystemRealmID {
		return nil, fmt.Errorf("cannot update system realm")
	}

	// Check if realm exists
	realm, err := s.GetRealmByID(realmID)
	if err != nil {
		return nil, err
	}

	// Check if new name conflicts with existing realm
	if name != "" && name != realm.Name {
		var existing models.Realm
		err := s.db.Where("name = ? AND id != ?", name, realmID).First(&existing).Error
		if err == nil {
			return nil, fmt.Errorf("realm name already exists: %s", name)
		} else if err != gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("failed to check existing realm: %w", err)
		}
		realm.Name = name
	}

	if description != "" {
		realm.Description = description
	}
	realm.UpdatedBy = updatedBy

	if err := s.db.Save(realm).Error; err != nil {
		return nil, fmt.Errorf("failed to update realm: %w", err)
	}

	return realm, nil
}

// DeleteRealm soft deletes a realm
func (s *RealmService) DeleteRealm(realmID string) error {
	// Prevent deletion of system realm
	if realmID == SystemRealmID {
		return fmt.Errorf("cannot delete system realm")
	}

	// Check if realm has users
	var userCount int64
	if err := s.db.Model(&models.User{}).Where("realm_id = ?", realmID).Count(&userCount).Error; err != nil {
		return fmt.Errorf("failed to check realm users: %w", err)
	}
	if userCount > 0 {
		return fmt.Errorf("cannot delete realm with existing users (count: %d)", userCount)
	}

	// Soft delete the realm
	result := s.db.Delete(&models.Realm{}, "id = ?", realmID)
	if result.Error != nil {
		return fmt.Errorf("failed to delete realm: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("realm not found: %s", realmID)
	}

	return nil
}

// ListAllRealms lists all realms (for super admin)
func (s *RealmService) ListAllRealms() ([]models.Realm, error) {
	var realms []models.Realm
	err := s.db.Order("name").Find(&realms).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list realms: %w", err)
	}
	return realms, nil
}

// RealmWithUserCount represents a realm with user count
type RealmWithUserCount struct {
	models.Realm
	UserCount int `json:"user_count"`
}

// ListRealmsWithUserCount lists all realms with user counts (for super admin)
func (s *RealmService) ListRealmsWithUserCount() ([]RealmWithUserCount, error) {
	var realms []models.Realm
	if err := s.db.Order("name").Find(&realms).Error; err != nil {
		return nil, fmt.Errorf("failed to list realms: %w", err)
	}

	result := make([]RealmWithUserCount, len(realms))
	for i, realm := range realms {
		var userCount int64
		if err := s.db.Model(&models.User{}).Where("realm_id = ? AND is_active = ?", realm.ID, true).Count(&userCount).Error; err != nil {
			return nil, fmt.Errorf("failed to count users for realm %s: %w", realm.ID, err)
		}
		result[i] = RealmWithUserCount{
			Realm:     realm,
			UserCount: int(userCount),
		}
	}

	return result, nil
}

// GetUserRealm gets the realm for a specific user
func (s *RealmService) GetUserRealm(userID string) (*models.Realm, error) {
	var user models.User
	if err := s.db.First(&user, "id = ?", userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found: %s", userID)
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if user.RealmID == "" {
		return nil, fmt.Errorf("user has no realm assigned")
	}

	return s.GetRealmByID(user.RealmID)
}

// AssignUserToRealm assigns a user to a realm
func (s *RealmService) AssignUserToRealm(userID, realmID, updatedBy string) error {
	// Check if realm exists
	if _, err := s.GetRealmByID(realmID); err != nil {
		return err
	}

	// Update user's realm
	result := s.db.Model(&models.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"realm_id":   realmID,
			"updated_by": updatedBy,
		})
	if result.Error != nil {
		return fmt.Errorf("failed to assign user to realm: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found: %s", userID)
	}

	return nil
}

// ListUsersInRealm lists all active users in a realm
func (s *RealmService) ListUsersInRealm(realmID string) ([]models.User, error) {
	var users []models.User
	err := s.db.Where("realm_id = ? AND is_active = ?", realmID, true).
		Order("username").
		Find(&users).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list users in realm: %w", err)
	}
	return users, nil
}

// CreateRealmForUser creates a realm for a specific user (admin's personal realm)
func (s *RealmService) CreateRealmForUser(userID, username, createdBy string) (*models.Realm, error) {
	// Create realm name from username
	realmName := fmt.Sprintf("%s_realm", username)
	description := fmt.Sprintf("Personal realm for %s", username)

	realm, err := s.CreateRealm(realmName, description, createdBy)
	if err != nil {
		return nil, err
	}

	// Assign user to this realm
	if err := s.AssignUserToRealm(userID, realm.ID, createdBy); err != nil {
		// Rollback: delete the realm
		s.db.Unscoped().Delete(&models.Realm{}, "id = ?", realm.ID)
		return nil, fmt.Errorf("failed to assign user to realm: %w", err)
	}

	return realm, nil
}

// EnsureSystemRealm ensures the system realm exists
func (s *RealmService) EnsureSystemRealm() error {
	var realm models.Realm
	err := s.db.Where("id = ?", SystemRealmID).First(&realm).Error
	if err == nil {
		// System realm already exists
		return nil
	}
	if err != gorm.ErrRecordNotFound {
		return fmt.Errorf("failed to check system realm: %w", err)
	}

	// Create system realm
	systemRealm := models.Realm{
		ID:          SystemRealmID,
		Name:        SystemRealmName,
		Description: "System realm for super administrators",
		CreatedBy:   "system",
		UpdatedBy:   "system",
	}

	if err := s.db.Create(&systemRealm).Error; err != nil {
		return fmt.Errorf("failed to create system realm: %w", err)
	}

	return nil
}

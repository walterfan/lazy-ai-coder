package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/term"
	"gorm.io/gorm"

	"github.com/walterfan/lazy-ai-coder/internal/auth"
	"github.com/walterfan/lazy-ai-coder/pkg/database"
	"github.com/walterfan/lazy-ai-coder/pkg/models"
)

var (
	passwordUser     string // username or email
	passwordNew      string // new password
	passwordOld      string // old password (optional for admin override)
	passwordForce    bool   // force update without old password (admin only)
	passwordInteractive bool // interactive password input
)

var passwordCmd = &cobra.Command{
	Use:   "password",
	Short: "Manage user passwords",
	Long:  `Manage user passwords from the command line`,
}

var passwordUpdateCmd = &cobra.Command{
	Use:   "update [username|email]",
	Short: "Update a user's password",
	Long: `Update a user's password from the command line.

You can specify the user by username or email. The command supports:
- Interactive password input (default)
- Non-interactive mode with flags
- Admin override mode (skip old password verification)

Examples:
  # Interactive mode (recommended)
  lazy-ai-coder password update john@example.com

  # Non-interactive mode
  lazy-ai-coder password update john@example.com --new-password "NewPass123!" --old-password "OldPass123!"

  # Admin override (skip old password check)
  lazy-ai-coder password update john@example.com --new-password "NewPass123!" --force`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize config
		if err := InitConfig(); err != nil {
			fmt.Printf("❌ Failed to initialize config: %v\n", err)
			os.Exit(1)
		}

		// Initialize database
		if err := database.InitDB(); err != nil {
			fmt.Printf("❌ Failed to initialize database: %v\n", err)
			fmt.Println("Hint: Make sure no other processes (like the web server) are using the database")
			os.Exit(1)
		}
		defer database.CloseDB()

		db := database.GetDB()

		// Get user identifier from args or flag
		userIdentifier := passwordUser
		if len(args) > 0 {
			userIdentifier = args[0]
		}
		if userIdentifier == "" {
			fmt.Println("❌ Error: username or email is required")
			fmt.Println("Usage: lazy-ai-coder password update [username|email]")
			os.Exit(1)
		}

		// Find user by username or email
		user, err := findUserByIdentifier(db, userIdentifier)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				fmt.Printf("❌ User not found: %s\n", userIdentifier)
			} else {
				fmt.Printf("❌ Failed to find user: %v\n", err)
			}
			os.Exit(1)
		}

		// Check if user has a password (OAuth-only users don't have passwords)
		if user.HashedPassword == nil {
			fmt.Printf("❌ User '%s' uses OAuth authentication and has no password to update\n", userIdentifier)
			fmt.Println("   OAuth users should manage their passwords through their OAuth provider (e.g., GitLab)")
			os.Exit(1)
		}

		// Get passwords
		var oldPassword, newPassword string

		if passwordInteractive || (passwordNew == "" && passwordOld == "" && !passwordForce) {
			// Interactive mode
			var err error
			if !passwordForce {
				oldPassword, err = readPassword("Enter current password: ")
				if err != nil {
					fmt.Printf("❌ Failed to read password: %v\n", err)
					os.Exit(1)
				}
			}

			newPassword, err = readPassword("Enter new password: ")
			if err != nil {
				fmt.Printf("❌ Failed to read password: %v\n", err)
				os.Exit(1)
			}

			confirmPassword, err := readPassword("Confirm new password: ")
			if err != nil {
				fmt.Printf("❌ Failed to read password: %v\n", err)
				os.Exit(1)
			}

			if newPassword != confirmPassword {
				fmt.Println("❌ Error: Passwords do not match")
				os.Exit(1)
			}
		} else {
			// Non-interactive mode
			if passwordNew == "" {
				fmt.Println("❌ Error: new password is required (use --new-password or --interactive)")
				os.Exit(1)
			}
			newPassword = passwordNew

			if !passwordForce {
				if passwordOld == "" {
					fmt.Println("❌ Error: old password is required (use --old-password, --force, or --interactive)")
					os.Exit(1)
				}
				oldPassword = passwordOld
			}
		}

		// Initialize password service
		passwordService := auth.NewPasswordAuthService(db)

		// Update password
		if passwordForce {
			// Admin override: update password directly without old password verification
			if err := updatePasswordDirectly(db, user.ID, newPassword); err != nil {
				fmt.Printf("❌ Failed to update password: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("✅ Password updated successfully for user: %s (%s)\n", user.Username, user.Email)
		} else {
			// Normal password change with old password verification
			if err := passwordService.ChangePassword(user.ID, oldPassword, newPassword); err != nil {
				if errors.Is(err, auth.ErrInvalidCredentials) {
					fmt.Println("❌ Error: Current password is incorrect")
				} else if errors.Is(err, auth.ErrWeakPassword) {
					fmt.Printf("❌ Error: %v\n", err)
				} else if errors.Is(err, auth.ErrUserNotFound) {
					fmt.Println("❌ Error: User not found")
				} else {
					fmt.Printf("❌ Failed to update password: %v\n", err)
				}
				os.Exit(1)
			}
			fmt.Printf("✅ Password updated successfully for user: %s (%s)\n", user.Username, user.Email)
		}
	},
}

// findUserByIdentifier finds a user by username or email
func findUserByIdentifier(db *gorm.DB, identifier string) (*models.User, error) {
	var user models.User

	// Try username first (case-insensitive)
	result := db.Where("LOWER(username) = ?", strings.ToLower(identifier)).First(&user)
	if result.Error == nil {
		return &user, nil
	}

	// Try email (case-insensitive)
	result = db.Where("LOWER(email) = ?", strings.ToLower(identifier)).First(&user)
	if result.Error == nil {
		return &user, nil
	}

	return nil, result.Error
}

// updatePasswordDirectly updates password without old password verification (admin only)
func updatePasswordDirectly(db *gorm.DB, userID, newPassword string) error {
	// Validate new password
	if err := auth.ValidatePassword(newPassword); err != nil {
		return err
	}

	// Hash new password
	hashedPassword, err := auth.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password directly
	result := db.Model(&models.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"hashed_password": hashedPassword,
			"updated_time":    gorm.Expr("CURRENT_TIMESTAMP"),
		})

	if result.Error != nil {
		return fmt.Errorf("failed to update password: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return auth.ErrUserNotFound
	}

	return nil
}

// readPassword reads a password from stdin without echoing
func readPassword(prompt string) (string, error) {
	fmt.Print(prompt)

	// Read password from stdin
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", err
	}

	fmt.Println() // New line after password input
	return string(passwordBytes), nil
}


func init() {
	// Add password command to root
	rootCmd.AddCommand(passwordCmd)

	// Add update subcommand
	passwordCmd.AddCommand(passwordUpdateCmd)

	// Flags for password update command
	passwordUpdateCmd.Flags().StringVarP(&passwordUser, "user", "u", "", "Username or email of the user")
	passwordUpdateCmd.Flags().StringVarP(&passwordNew, "new-password", "n", "", "New password (non-interactive mode)")
	passwordUpdateCmd.Flags().StringVarP(&passwordOld, "old-password", "o", "", "Old password (non-interactive mode)")
	passwordUpdateCmd.Flags().BoolVarP(&passwordForce, "force", "f", false, "Force update without old password verification (admin only)")
	passwordUpdateCmd.Flags().BoolVarP(&passwordInteractive, "interactive", "i", true, "Interactive password input (default: true)")
}


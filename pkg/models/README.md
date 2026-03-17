# Models Package

This package implements a comprehensive authentication and authorization system based on AWS IAM principles, along with project and content management capabilities.

## 🏗️ Architecture Overview

The models are designed around the concept of **Realms** (multi-tenancy) with **AWS-style IAM** for fine-grained permissions.

### Core Entities

```
Realm (Tenant)
├── Users (app_user table)
├── Roles 
├── Policies
│   └── Statements (AWS-style)
└── Projects
    ├── Code Files
    └── Documents
```

## 📋 Model Definitions

### Authentication & Authorization

#### Realm
- **Purpose**: Top-level organizational unit (tenant isolation)
- **Key Fields**: `ID`, `Name`, `Description`
- **Table**: `realm`

#### User  
- **Purpose**: Application users with hashed passwords
- **Key Fields**: `ID`, `RealmID`, `Username`, `Email`, `HashedPassword`, `IsActive`
- **Table**: `app_user` (avoids SQL reserved keyword conflicts)
- **Security**: Password field is excluded from JSON serialization

#### Role
- **Purpose**: Named collections of permissions
- **Key Fields**: `ID`, `RealmID`, `Name`, `Description`
- **Table**: `role`

#### Policy
- **Purpose**: Permission documents following AWS IAM format
- **Key Fields**: `ID`, `RealmID`, `Name`, `Version` (defaults to '2012-10-17')
- **Table**: `policy`

#### Statement
- **Purpose**: Individual permission rules within policies
- **Key Fields**: `PolicyID`, `Effect` (Allow/Deny), `Actions`, `Resources`, `Conditions`
- **Table**: `statement`
- **Format**: Actions, Resources, and Conditions stored as JSON strings

### Join Tables (Many-to-Many Relationships)

- **UserRole**: Users ↔ Roles
- **UserPolicy**: Users ↔ Policies (direct assignment)
- **RolePolicy**: Roles ↔ Policies
- **ResourcePolicy**: Resource-based policy attachments

### Project & Content Management

#### Project
- **Purpose**: Code projects with Git integration
- **Key Fields**: `ID`, `RealmID`, `Name`, `GitURL`, `GitRepo`, `GitBranch`, `Language`
- **Table**: `project`

#### Code
- **Purpose**: Source code files with vector embeddings
- **Key Fields**: `ID`, `ProjectID`, `Path`, `Code`, `VectorEmbedding`
- **Table**: `code`

#### Document
- **Purpose**: Documentation files with vector embeddings
- **Key Fields**: `ID`, `ProjectID`, `Name`, `Path`, `Content`, `VectorEmbedding`
- **Table**: `document`

#### Prompt
- **Purpose**: AI prompt templates (existing functionality updated)
- **Key Fields**: `ID`, `Name`, `Description`, `SystemPrompt`, `UserPrompt`, `Tags`
- **Table**: `prompt`

## 🔐 Permission System

### AWS-Style Evaluation Logic

1. **Explicit DENY always wins** - Any deny statement blocks access
2. **Explicit ALLOW required** - Default is implicit deny
3. **Evaluation order**: Resource-based → User policies → Role policies
4. **Variable substitution** in conditions (e.g., `${user:id}`)

### Action Examples
```
read:project, write:project, delete:project
list:documents, create:document
admin:users, modify:roles
```

### Resource Examples
```
project:*                    # All projects
project:123                  # Specific project
document:project:123/*       # All documents in project 123
user:${user:id}             # Current user
```

### Condition Examples
```json
{
  "StringEquals": {
    "project:owner": "${user:id}",
    "realm:id": "${user:realm_id}"
  },
  "DateGreaterThan": {
    "current:time": "2024-01-01T00:00:00Z"
  }
}
```

## 💾 Database Compatibility

### Field Types
- **IDs**: `string` (TEXT in SQLite, VARCHAR/UUID in others)
- **JSON**: Stored as `string`, parsed at application level
- **Timestamps**: Go `time.Time` with GORM auto-management
- **Booleans**: Native `bool` (converted to INTEGER in SQLite)

### Soft Deletes
All models support soft deletes via `gorm.DeletedAt`

### Indexes
- Realm, User, Project IDs
- User email and username
- Foreign key relationships
- Statement effect and policy ID

## 🚀 Usage Examples

### Auto-Migration
```go
import "github.com/walterfan/lazy-ai-coder/pkg/models"

// Migrate all models
err := db.AutoMigrate(models.GetAllModels()...)
```

### Creating a User
```go
user := models.User{
    ID:             "user-123",
    RealmID:        "default",
    Username:       "walter.fan",
    Email:          "walter.fan@example.com",
    HashedPassword: hashedPwd,
    IsActive:       true,
    CreatedBy:      "system",
}
db.Create(&user)
```

### Policy with Statements
```go
policy := models.Policy{
    ID:          "read-only-policy",
    RealmID:     "default",
    Name:        "ReadOnlyAccess",
    Description: "Read-only access to all resources",
}
db.Create(&policy)

statement := models.Statement{
    ID:        "read-stmt",
    PolicyID:  "read-only-policy",
    Effect:    "Allow",
    Actions:   `["read:*", "list:*"]`,
    Resources: `["*"]`,
}
db.Create(&statement)
```

### Querying with Relationships
```go
// Find user with their roles
var user models.User
db.Preload("UserRoles").First(&user, "id = ?", "user-123")

// Find policies for a user
var policies []models.Policy
db.Joins("JOIN user_policy ON policy.id = user_policy.policy_id").
   Where("user_policy.user_id = ?", "user-123").
   Find(&policies)
```

## 🔧 Integration with Existing Code

The updated models maintain backward compatibility where possible:
- `Prompt` model updated but keeps core functionality
- `User` model significantly changed (requires database migration)
- New models add IAM capabilities without affecting existing features

## 📚 Related Files

- `deploy/db/schema_sqlite.sql` - Complete database schema
- `pkg/database/database.go` - Database initialization
- `cmd/web.go` - May need updates for new authentication system

## 🛠️ Migration Notes

When upgrading from the old model structure:
1. **Backup existing database**
2. **Run auto-migration** (will create new tables)
3. **Migrate existing user data** to new structure
4. **Update application code** to use new field names
5. **Set up default realm** and admin policies 
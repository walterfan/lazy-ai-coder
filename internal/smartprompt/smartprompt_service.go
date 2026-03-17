package smartprompt

import (
	"fmt"
	"strings"

	"github.com/walterfan/lazy-ai-coder/internal/log"
	"github.com/walterfan/lazy-ai-coder/internal/models"
	"github.com/walterfan/lazy-ai-coder/internal/util"
)

// SmartPromptService handles smart prompt generation logic
type SmartPromptService struct{}

// NewSmartPromptService creates a new SmartPromptService
func NewSmartPromptService() *SmartPromptService {
	return &SmartPromptService{}
}

// AnalyzeGitLabProject detects tech stack from GitLab project files
func (s *SmartPromptService) AnalyzeGitLabProject(gitlabUrl, projectID, branch, privateToken string) (models.ProjectContext, error) {
	logger := log.GetLogger()
	ctx := models.ProjectContext{}

	// Config files to check and their handlers
	configFiles := map[string]func(string, *models.ProjectContext) error{
		"pom.xml":          ParsePomXml,
		"package.json":     ParsePackageJson,
		"requirements.txt": ParseRequirementsTxt,
		"go.mod":           ParseGoMod,
		"build.gradle":     ParseBuildGradle,
		"Cargo.toml":       ParseCargoToml,
	}

	// Try to read each config file
	for filename, parser := range configFiles {
		content, err := util.GetGitLabFileContent(gitlabUrl, projectID, filename, branch, privateToken)
		if err == nil {
			logger.Infof("Found %s, parsing...", filename)
			if err := parser(content, &ctx); err != nil {
				logger.Warnf("Failed to parse %s: %v", filename, err)
			}
			break // Found a config file, stop searching
		}
	}

	// Check for test files
	testFiles := []string{"src/test", "test", "tests", "__tests__"}
	for _, testPath := range testFiles {
		_, err := util.GetGitLabFileContent(gitlabUrl, projectID, testPath, branch, privateToken)
		if err == nil {
			ctx.HasTests = true
			break
		}
	}

	return ctx, nil
}

// GenerateExamples creates relevant code examples based on context
func (s *SmartPromptService) GenerateExamples(ctx models.ProjectContext, action string) []models.CodeExample {
	examples := []models.CodeExample{}

	// Generate examples based on detected context
	if ctx.Framework == "Spring Boot" {
		examples = append(examples, models.CodeExample{
			Title:    "Spring Boot REST Controller Example",
			Language: "java",
			Code: `@RestController
@RequestMapping("/api/users")
public class UserController {
    @Autowired
    private UserService userService;

    @GetMapping("/{id}")
    public ResponseEntity<User> getUser(@PathVariable Long id) {
        return userService.findById(id)
            .map(ResponseEntity::ok)
            .orElse(ResponseEntity.notFound().build());
    }

    @PostMapping
    public ResponseEntity<User> createUser(@Valid @RequestBody UserDTO userDTO) {
        User user = userService.create(userDTO);
        return ResponseEntity.status(HttpStatus.CREATED).body(user);
    }
}`,
			Description: "Example REST controller with proper annotations and error handling",
		})
	} else if ctx.Framework == "FastAPI" {
		examples = append(examples, models.CodeExample{
			Title:    "FastAPI Endpoint Example",
			Language: "python",
			Code: `from fastapi import APIRouter, HTTPException, Depends
from pydantic import BaseModel

router = APIRouter()

class UserCreate(BaseModel):
    name: str
    email: str

@router.post("/users", response_model=User)
async def create_user(user: UserCreate, db: Session = Depends(get_db)):
    """Create a new user"""
    try:
        new_user = User(**user.dict())
        db.add(new_user)
        await db.commit()
        return new_user
    except Exception as e:
        raise HTTPException(status_code=400, detail=str(e))`,
			Description: "Example FastAPI endpoint with Pydantic validation and async/await",
		})
	} else if ctx.Framework == "React" {
		examples = append(examples, models.CodeExample{
			Title:    "React Functional Component Example",
			Language: "typescript",
			Code: `import React, { useState, useEffect } from 'react';

interface User {
  id: number;
  name: string;
  email: string;
}

const UserList: React.FC = () => {
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchUsers = async () => {
      try {
        const response = await fetch('/api/users');
        const data = await response.json();
        setUsers(data);
      } catch (error) {
        console.error('Failed to fetch users:', error);
      } finally {
        setLoading(false);
      }
    };

    fetchUsers();
  }, []);

  if (loading) return <div>Loading...</div>;

  return (
    <ul>
      {users.map(user => (
        <li key={user.id}>{user.name} - {user.email}</li>
      ))}
    </ul>
  );
};

export default UserList;`,
			Description: "Example React component with hooks, TypeScript, and async data fetching",
		})
	} else if ctx.Framework == "Gin" {
		examples = append(examples, models.CodeExample{
			Title:    "Gin HTTP Handler Example",
			Language: "go",
			Code: `package handlers

import (
    "net/http"
    "github.com/gin-gonic/gin"
)

type UserRequest struct {
    Name  string ` + "`json:\"name\" binding:\"required\"`" + `
    Email string ` + "`json:\"email\" binding:\"required,email\"`" + `
}

func CreateUser(c *gin.Context) {
    var req UserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Create user logic here
    user := User{
        Name:  req.Name,
        Email: req.Email,
    }

    c.JSON(http.StatusCreated, user)
}`,
			Description: "Example Gin handler with input validation and proper error handling",
		})
	}

	return examples
}

// ScorePromptQuality evaluates the quality of the generated prompt
func (s *SmartPromptService) ScorePromptQuality(context, action, result string, detectedCtx models.ProjectContext) models.QualityScore {
	score := 0.0
	maxScore := 10.0
	feedback := []string{}
	suggestions := []string{}

	// Check Context completeness (0-3 points)
	contextScore := 0.0
	if len(context) > 50 {
		contextScore += 1.0
	}
	if detectedCtx.Language != "" {
		contextScore += 0.5
		feedback = append(feedback, "✓ Programming language detected")
	} else {
		suggestions = append(suggestions, "Add programming language information")
	}
	if detectedCtx.Framework != "" {
		contextScore += 0.5
		feedback = append(feedback, "✓ Framework detected")
	} else {
		suggestions = append(suggestions, "Specify framework or library")
	}
	if len(detectedCtx.Dependencies) > 0 {
		contextScore += 1.0
		feedback = append(feedback, fmt.Sprintf("✓ %d dependencies found", len(detectedCtx.Dependencies)))
	}
	score += contextScore

	// Check Action clarity (0-3 points)
	actionScore := 0.0
	if len(action) > 20 {
		actionScore += 1.0
	}
	if len(action) > 50 {
		actionScore += 1.0
		feedback = append(feedback, "✓ Action is detailed")
	} else {
		suggestions = append(suggestions, "Provide more specific action details")
	}
	if strings.Contains(strings.ToLower(action), "create") ||
		strings.Contains(strings.ToLower(action), "implement") ||
		strings.Contains(strings.ToLower(action), "add") {
		actionScore += 1.0
		feedback = append(feedback, "✓ Action verb is clear")
	}
	score += actionScore

	// Check Result/Acceptance Criteria (0-4 points)
	resultScore := 0.0
	if len(result) > 50 {
		resultScore += 1.0
	}
	if len(result) > 100 {
		resultScore += 1.0
		feedback = append(feedback, "✓ Detailed acceptance criteria")
	} else {
		suggestions = append(suggestions, "Add more detailed acceptance criteria")
	}
	if strings.Contains(strings.ToLower(result), "test") {
		resultScore += 1.0
		feedback = append(feedback, "✓ Testing requirements included")
	} else {
		suggestions = append(suggestions, "Include testing requirements")
	}
	if strings.Contains(strings.ToLower(result), "error") ||
		strings.Contains(strings.ToLower(result), "exception") ||
		strings.Contains(strings.ToLower(result), "validation") {
		resultScore += 1.0
		feedback = append(feedback, "✓ Error handling mentioned")
	} else {
		suggestions = append(suggestions, "Consider error handling requirements")
	}
	score += resultScore

	return models.QualityScore{
		Score:       score,
		MaxScore:    maxScore,
		Feedback:    feedback,
		Suggestions: suggestions,
	}
}

// GenerateSmartPrompt generates a CAR format prompt with context detection
func (s *SmartPromptService) GenerateSmartPrompt(req *models.SmartPromptRequest) (*models.SmartPromptResponse, error) {
	logger := log.GetLogger()

	// Get preset if specified
	var preset *models.Preset
	if req.PresetID != "" {
		preset = GetPresetByID(req.PresetID)
	}

	// Analyze GitLab project if requested
	detectedCtx := models.ProjectContext{}
	if req.AnalyzeContext && req.GitlabProject != "" {
		// Token must be passed from frontend settings
		privateToken := req.Settings.GitlabToken
		if privateToken == "" {
			logger.Warn("GitLab token not configured in settings, skipping project analysis")
		}
		branch := req.GitlabBranch
		if branch == "" {
			branch = "main"
		}

		ctx, err := s.AnalyzeGitLabProject(req.Settings.GitlabUrl, req.GitlabProject, branch, privateToken)
		if err != nil {
			logger.Warnf("Failed to analyze GitLab project: %v", err)
		} else {
			detectedCtx = ctx
			logger.Infof("Detected context: Language=%s, Framework=%s", ctx.Language, ctx.Framework)
		}
	}

	// Build Context section
	contextParts := []string{}
	if detectedCtx.Language != "" {
		contextParts = append(contextParts, fmt.Sprintf("- Language: %s", detectedCtx.Language))
	} else if preset != nil {
		contextParts = append(contextParts, fmt.Sprintf("- Language: %s", preset.Language))
	}
	if detectedCtx.Framework != "" {
		frameworkInfo := detectedCtx.Framework
		if detectedCtx.FrameworkVersion != "" {
			frameworkInfo += " " + detectedCtx.FrameworkVersion
		}
		contextParts = append(contextParts, fmt.Sprintf("- Framework: %s", frameworkInfo))
	} else if preset != nil && preset.Framework != "" {
		contextParts = append(contextParts, fmt.Sprintf("- Framework: %s", preset.Framework))
	}
	if detectedCtx.BuildTool != "" {
		contextParts = append(contextParts, fmt.Sprintf("- Build Tool: %s", detectedCtx.BuildTool))
	}
	if detectedCtx.Database != "" {
		contextParts = append(contextParts, fmt.Sprintf("- Database: %s", detectedCtx.Database))
	}
	if len(detectedCtx.Dependencies) > 0 {
		depsStr := strings.Join(detectedCtx.Dependencies[:min(5, len(detectedCtx.Dependencies))], ", ")
		contextParts = append(contextParts, fmt.Sprintf("- Key Dependencies: %s", depsStr))
	}
	if detectedCtx.HasTests {
		contextParts = append(contextParts, fmt.Sprintf("- Test Framework: %s", detectedCtx.TestFramework))
	}
	if preset != nil {
		contextParts = append(contextParts, fmt.Sprintf("- Architecture: %s", preset.ContextHints))
	}

	context := strings.Join(contextParts, "\n")

	// Build Action section
	action := req.Input

	// Build Result section (acceptance criteria)
	resultParts := []string{}
	resultParts = append(resultParts, "Acceptance Criteria:")
	resultParts = append(resultParts, "1. Code should follow best practices and conventions for "+detectedCtx.Language)
	if detectedCtx.HasTests {
		resultParts = append(resultParts, fmt.Sprintf("2. Include unit tests using %s", detectedCtx.TestFramework))
	} else {
		resultParts = append(resultParts, "2. Include appropriate unit tests")
	}
	resultParts = append(resultParts, "3. Handle errors appropriately with proper error messages")
	resultParts = append(resultParts, "4. Add comments for complex logic")
	resultParts = append(resultParts, "5. Ensure code is maintainable and follows SOLID principles")

	if preset != nil && preset.ResultHints != "" {
		resultParts = append(resultParts, "\nAdditional Requirements:")
		resultParts = append(resultParts, preset.ResultHints)
	}

	result := strings.Join(resultParts, "\n")

	// Generate full prompt
	fullPrompt := fmt.Sprintf(`**Context:**
%s

**Action:**
%s

**Result:**
%s`, context, action, result)

	// Generate code examples
	examples := s.GenerateExamples(detectedCtx, action)

	// Score the prompt quality
	qualityScore := s.ScorePromptQuality(context, action, result, detectedCtx)

	response := &models.SmartPromptResponse{
		Context:      context,
		Action:       action,
		Result:       result,
		FullPrompt:   fullPrompt,
		DetectedCtx:  detectedCtx,
		Examples:     examples,
		QualityScore: qualityScore,
	}

	return response, nil
}

// min helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}


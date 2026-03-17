package smartprompt

import (
	"fmt"
	"strings"

	"github.com/walterfan/lazy-ai-coder/internal/models"
)

// Framework represents a prompt engineering framework
type Framework struct {
	ID          string                  `json:"id"`
	Name        string                  `json:"name"`
	Description string                  `json:"description"`
	Fields      []FrameworkField        `json:"fields"`
	Template    string                  `json:"template"`
	BestFor     string                  `json:"best_for"`
	Example     models.FrameworkExample `json:"example"`
}

// FrameworkField represents a single field in a framework
type FrameworkField struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	Description string `json:"description"`
	Placeholder string `json:"placeholder"`
	Required    bool   `json:"required"`
	Type        string `json:"type"` // text, textarea, select
	Options     []string `json:"options,omitempty"`
}

// AllFrameworks contains all available prompt engineering frameworks
var AllFrameworks = []Framework{
	{
		ID:          "crispe",
		Name:        "CRISPE",
		Description: "Capacity-Request-Input-Steps-Performance-Example. Most versatile framework for detailed, structured prompts.",
		BestFor:     "Complex coding tasks requiring detailed specifications and examples",
		Template: `You are a {{.Capacity}}.

Task: {{.Request}}

Context and Input:
{{.Input}}

Steps to follow:
{{.Steps}}

Performance Requirements:
{{.Performance}}

{{if .Example}}Example:
{{.Example}}{{end}}`,
		Fields: []FrameworkField{
			{
				ID:          "capacity",
				Label:       "Capacity/Role",
				Description: "Define the role or expertise level ChatGPT should assume",
				Placeholder: "senior Python developer with 10 years of experience in building REST APIs",
				Required:    true,
				Type:        "text",
			},
			{
				ID:          "request",
				Label:       "Request",
				Description: "Clearly state what you want ChatGPT to do",
				Placeholder: "Create a secure user authentication endpoint with JWT tokens",
				Required:    true,
				Type:        "textarea",
			},
			{
				ID:          "input",
				Label:       "Input/Context",
				Description: "Provide necessary background information, tech stack, and constraints",
				Placeholder: "Using FastAPI framework, PostgreSQL database, and bcrypt for password hashing. Include rate limiting and brute force protection.",
				Required:    true,
				Type:        "textarea",
			},
			{
				ID:          "steps",
				Label:       "Steps",
				Description: "Define the methodology or steps to follow",
				Placeholder: "1. Define Pydantic models\n2. Create database schema\n3. Implement password hashing\n4. Add JWT token generation\n5. Create endpoint logic\n6. Add comprehensive tests",
				Required:    false,
				Type:        "textarea",
			},
			{
				ID:          "performance",
				Label:       "Performance Expectations",
				Description: "Specify output quality, format, style, and other requirements",
				Placeholder: "Production-ready code with error handling, type hints, docstrings, and unit tests. Follow PEP 8 style guide.",
				Required:    true,
				Type:        "textarea",
			},
			{
				ID:          "example",
				Label:       "Example (Optional)",
				Description: "Provide a sample input/output or similar code for reference",
				Placeholder: "Similar to how we handle user registration endpoint...",
				Required:    false,
				Type:        "textarea",
			},
		},
		Example: models.FrameworkExample{
			UseCase:     "FastAPI Authentication Endpoint",
			Input:       "Create a login endpoint with JWT authentication",
			GeneratedPrompt: `You are a senior Python developer with 10 years of experience in building REST APIs.

Task: Create a secure user authentication endpoint with JWT tokens

Context and Input:
Using FastAPI framework, PostgreSQL database, and bcrypt for password hashing. Include rate limiting and brute force protection.

Steps to follow:
1. Define Pydantic models for login request and response
2. Create database schema for users table
3. Implement password hashing using bcrypt
4. Add JWT token generation with expiration
5. Create POST /auth/login endpoint
6. Add comprehensive unit tests

Performance Requirements:
Production-ready code with:
- Proper error handling for invalid credentials
- Type hints throughout
- Docstrings for all functions
- Unit tests with pytest
- Follow PEP 8 style guide
- Return appropriate HTTP status codes (200, 401, 429)`,
		},
	},
	{
		ID:          "risen",
		Name:        "RISEN",
		Description: "Role-Instructions-Steps-End Goal-Narrowing. Perfect for complex, multi-step tasks with clear deliverables.",
		BestFor:     "Complex refactoring, architecture design, or multi-file implementations",
		Template: `Role: {{.Role}}

Instructions:
{{.Instructions}}

Steps:
{{.Steps}}

End Goal:
{{.EndGoal}}

Constraints and Narrowing:
{{.Narrowing}}`,
		Fields: []FrameworkField{
			{
				ID:          "role",
				Label:       "Role",
				Description: "What role should the AI assume?",
				Placeholder: "You are a senior software architect specializing in microservices design",
				Required:    true,
				Type:        "text",
			},
			{
				ID:          "instructions",
				Label:       "Instructions",
				Description: "Clear, detailed instructions for the task",
				Placeholder: "Design a microservices architecture for an e-commerce platform. Consider user service, product catalog, shopping cart, and payment processing.",
				Required:    true,
				Type:        "textarea",
			},
			{
				ID:          "steps",
				Label:       "Steps",
				Description: "Break down the task into specific steps",
				Placeholder: "1. Identify bounded contexts\n2. Define service responsibilities\n3. Design inter-service communication\n4. Define data models\n5. Specify API contracts",
				Required:    true,
				Type:        "textarea",
			},
			{
				ID:          "end_goal",
				Label:       "End Goal",
				Description: "What should the final output achieve?",
				Placeholder: "A complete architecture diagram with service definitions, API contracts, and deployment considerations",
				Required:    true,
				Type:        "textarea",
			},
			{
				ID:          "narrowing",
				Label:       "Narrowing/Constraints",
				Description: "Specify constraints, style, format, limitations",
				Placeholder: "Use REST APIs, PostgreSQL for data storage, Docker containers, maximum 5 services, include authentication strategy",
				Required:    false,
				Type:        "textarea",
			},
		},
		Example: models.FrameworkExample{
			UseCase:     "Microservices Architecture Design",
			Input:       "Design microservices for e-commerce platform",
			GeneratedPrompt: `Role: You are a senior software architect specializing in microservices design with 15 years of experience

Instructions:
Design a scalable microservices architecture for an e-commerce platform. The system needs to handle user management, product catalog, shopping cart, order processing, and payments.

Steps:
1. Identify bounded contexts using Domain-Driven Design principles
2. Define service responsibilities and boundaries
3. Design inter-service communication (sync vs async)
4. Define data models and database per service
5. Specify API contracts using OpenAPI
6. Address cross-cutting concerns (auth, logging, monitoring)

End Goal:
A complete architecture document including:
- Service diagram showing all microservices and their relationships
- API contracts for each service
- Data models and database schemas
- Communication patterns and message flows
- Deployment and scaling considerations

Constraints and Narrowing:
- Use REST APIs for synchronous communication
- Use RabbitMQ for asynchronous messaging
- PostgreSQL for transactional data
- Redis for caching
- Maximum 6 microservices
- Include JWT-based authentication strategy
- Consider GDPR compliance for user data`,
		},
	},
	{
		ID:          "costar",
		Name:        "CO-STAR",
		Description: "Context-Objective-Style-Tone-Audience-Response. Ideal for precision and specific output requirements.",
		BestFor:     "Documentation, code comments, or when output format/style is critical",
		Template: `Context:
{{.Context}}

Objective:
{{.Objective}}

Style:
{{.Style}}

Tone:
{{.Tone}}

Audience:
{{.Audience}}

Response Format:
{{.Response}}`,
		Fields: []FrameworkField{
			{
				ID:          "context",
				Label:       "Context",
				Description: "Provide background information and current situation",
				Placeholder: "We have a legacy REST API built with Flask that needs comprehensive API documentation for external developers",
				Required:    true,
				Type:        "textarea",
			},
			{
				ID:          "objective",
				Label:       "Objective",
				Description: "What you want to achieve with this prompt",
				Placeholder: "Generate OpenAPI 3.0 specification documentation for all endpoints with examples and error responses",
				Required:    true,
				Type:        "textarea",
			},
			{
				ID:          "style",
				Label:       "Style",
				Description: "Writing style or format preference",
				Placeholder: "Professional, technical, following OpenAPI 3.0 standards with clear descriptions",
				Required:    false,
				Type:        "text",
			},
			{
				ID:          "tone",
				Label:       "Tone",
				Description: "Emotional tone (professional, casual, friendly, etc.)",
				Placeholder: "Professional and developer-friendly",
				Required:    false,
				Type:        "text",
			},
			{
				ID:          "audience",
				Label:       "Audience",
				Description: "Who is the output for?",
				Placeholder: "External API consumers and integration developers",
				Required:    false,
				Type:        "text",
			},
			{
				ID:          "response",
				Label:       "Response Format",
				Description: "Expected response structure and format",
				Placeholder: "OpenAPI 3.0 YAML format with complete paths, schemas, responses, and authentication details",
				Required:    true,
				Type:        "textarea",
			},
		},
		Example: models.FrameworkExample{
			UseCase:     "API Documentation Generation",
			Input:       "Generate OpenAPI docs for Flask API",
			GeneratedPrompt: `Context:
We have a legacy REST API built with Flask (Python 3.9) that has been serving internal applications for 2 years. It includes user management, data processing, and reporting endpoints. We're now opening it to external partners and need professional documentation.

Objective:
Generate comprehensive OpenAPI 3.0 specification documentation for all endpoints. Include detailed descriptions, request/response schemas, example payloads, error responses, and authentication requirements.

Style:
Professional, technical writing following OpenAPI 3.0 standards. Use clear, concise descriptions with proper technical terminology.

Tone:
Professional and developer-friendly, balancing technical accuracy with accessibility.

Audience:
External API consumers, integration developers, and technical partners who will integrate with our API.

Response Format:
OpenAPI 3.0 specification in YAML format including:
- Complete API info (title, version, description, contact)
- All endpoint paths with HTTP methods
- Detailed request/response schemas with data types
- Example requests and responses
- Error responses (400, 401, 403, 404, 500)
- Authentication (Bearer token)
- Common reusable components`,
		},
	},
	{
		ID:          "ape",
		Name:        "APE",
		Description: "Action-Purpose-Expectation. Simple but effective for creative and exploratory tasks.",
		BestFor:     "Quick prototypes, brainstorming code solutions, or exploring approaches",
		Template: `Action: {{.Action}}

Purpose: {{.Purpose}}

Expectation: {{.Expectation}}`,
		Fields: []FrameworkField{
			{
				ID:          "action",
				Label:       "Action",
				Description: "What should be done?",
				Placeholder: "Implement a caching layer for database queries using Redis",
				Required:    true,
				Type:        "textarea",
			},
			{
				ID:          "purpose",
				Label:       "Purpose",
				Description: "Why are we doing this? What problem does it solve?",
				Placeholder: "To reduce database load and improve API response times by 70%. Current average response time is 800ms, target is 240ms.",
				Required:    true,
				Type:        "textarea",
			},
			{
				ID:          "expectation",
				Label:       "Expectation",
				Description: "What does success look like? What should the output be?",
				Placeholder: "Working Redis cache implementation with TTL configuration, cache invalidation strategy, and fallback to database if cache misses. Include unit tests.",
				Required:    true,
				Type:        "textarea",
			},
		},
		Example: models.FrameworkExample{
			UseCase:     "Redis Caching Implementation",
			Input:       "Add caching to speed up API",
			GeneratedPrompt: `Action: Implement a caching layer for database queries using Redis in our Node.js Express API

Purpose: To significantly reduce database load and improve API response times. Current average response time for product listing is 800ms with direct PostgreSQL queries. Our target is to achieve sub-240ms response times and reduce database queries by 80%. This will improve user experience and reduce infrastructure costs.

Expectation: Deliver a production-ready Redis caching solution including:
- Redis client configuration with connection pooling
- Cache wrapper functions for common query patterns
- TTL configuration based on data volatility (products: 5min, categories: 1hour)
- Cache invalidation strategy on data updates
- Automatic fallback to database if cache misses or Redis unavailable
- Monitoring hooks for cache hit/miss rates
- Unit tests with 80% coverage
- Documentation for configuration and usage`,
		},
	},
	{
		ID:          "rcar",
		Name:        "R-CAR",
		Description: "Role-Context-Action-Result. Enhanced simple format with role definition for better results.",
		BestFor:     "Simple, well-defined tasks that benefit from clear role specification",
		Template: `Role: {{.Role}}

Context:
{{.Context}}

Action:
{{.Action}}

Result:
{{.Result}}`,
		Fields: []FrameworkField{
			{
				ID:          "role",
				Label:       "Role",
				Description: "Define the AI's role and expertise level",
				Placeholder: "You are a senior Python developer with expertise in FastAPI and REST API development",
				Required:    true,
				Type:        "text",
			},
			{
				ID:          "context",
				Label:       "Context",
				Description: "Provide background and technical details",
				Placeholder: "Python 3.9, FastAPI, PostgreSQL database with users table",
				Required:    true,
				Type:        "textarea",
			},
			{
				ID:          "action",
				Label:       "Action",
				Description: "What needs to be implemented?",
				Placeholder: "Create a GET endpoint to retrieve user profile by ID",
				Required:    true,
				Type:        "textarea",
			},
			{
				ID:          "result",
				Label:       "Result/Acceptance Criteria",
				Description: "What should the output include?",
				Placeholder: "Return user data as JSON, handle 404 for missing users, include error handling and tests",
				Required:    true,
				Type:        "textarea",
			},
		},
		Example: models.FrameworkExample{
			UseCase:     "Simple API Endpoint",
			Input:       "Create GET user endpoint",
			GeneratedPrompt: `Role: You are a senior Python developer with 8 years of experience in FastAPI and REST API development

Context:
Python 3.9 application using FastAPI framework with PostgreSQL database. We have a users table with id, name, email, created_at columns.

Action:
Create a GET /api/users/{user_id} endpoint to retrieve user profile information by ID.

Result:
- Return user data as JSON (id, name, email, created_at)
- Return 404 with error message if user not found
- Include proper error handling for database errors
- Add input validation for user_id parameter
- Include unit tests with pytest`,
		},
	},
	{
		ID:          "race",
		Name:        "RACE",
		Description: "Role-Action-Context-Example. Optimized framework for clear, example-driven prompt engineering.",
		BestFor:     "Tasks where providing examples significantly improves output quality",
		Template: `Role: {{.Role}}

Action:
{{.Action}}

Context:
{{.Context}}

{{if .Example}}Example:
{{.Example}}{{end}}`,
		Fields: []FrameworkField{
			{
				ID:          "role",
				Label:       "Role",
				Description: "Define the AI's role, expertise, and perspective",
				Placeholder: "You are an expert software engineer specializing in clean code and best practices",
				Required:    true,
				Type:        "text",
			},
			{
				ID:          "action",
				Label:       "Action",
				Description: "Clearly state what you want the AI to do",
				Placeholder: "Refactor this function to improve readability and follow SOLID principles",
				Required:    true,
				Type:        "textarea",
			},
			{
				ID:          "context",
				Label:       "Context",
				Description: "Provide relevant background, constraints, and requirements",
				Placeholder: "This is a Node.js application using TypeScript. The function handles user authentication and should maintain backward compatibility. Performance is critical as it's called on every request.",
				Required:    true,
				Type:        "textarea",
			},
			{
				ID:          "example",
				Label:       "Example",
				Description: "Provide concrete examples of input/output or similar solutions",
				Placeholder: "Similar to how we refactored the authorization middleware:\n\nBefore:\nfunction checkAuth(req, res, next) { ... }\n\nAfter:\nclass AuthValidator {\n  validate(request: Request): boolean { ... }\n}",
				Required:    false,
				Type:        "textarea",
			},
		},
		Example: models.FrameworkExample{
			UseCase:     "Code Refactoring",
			Input:       "Refactor messy authentication function",
			GeneratedPrompt: `Role: You are an expert software engineer specializing in clean code, TypeScript, and security best practices

Action:
Refactor the provided user authentication function to improve readability, maintainability, and follow SOLID principles. Extract separate concerns into smaller, testable functions.

Context:
This is a Node.js Express application using TypeScript 4.9. The function handles JWT token validation and user authentication. It should maintain backward compatibility with existing API contracts. Performance is critical as this runs on every authenticated request (approx 10,000 req/sec). Must handle edge cases: expired tokens, malformed tokens, missing tokens, and revoked tokens.

Example:
Similar to how we refactored the authorization middleware:

Before:
function checkAuth(req, res, next) {
  const token = req.headers.authorization;
  if (!token) return res.status(401).json({error: 'No token'});
  jwt.verify(token, SECRET, (err, decoded) => {
    if (err) return res.status(401).json({error: 'Invalid'});
    req.user = decoded;
    next();
  });
}

After:
class AuthValidator {
  validate(request: Request): AuthResult {
    const token = this.extractToken(request);
    return this.verifyToken(token);
  }

  private extractToken(request: Request): string | null {
    return request.headers.authorization?.replace('Bearer ', '') ?? null;
  }

  private verifyToken(token: string | null): AuthResult {
    // Clear separation of concerns
  }
}`,
		},
	},
}

// GetFrameworkByID retrieves a framework by its ID
func GetFrameworkByID(id string) *Framework {
	for i := range AllFrameworks {
		if AllFrameworks[i].ID == id {
			return &AllFrameworks[i]
		}
	}
	return nil
}

// PromptPair represents system and user prompts
type PromptPair struct {
	SystemPrompt string `json:"system_prompt"`
	UserPrompt   string `json:"user_prompt"`
	FullPrompt   string `json:"full_prompt"` // For display/legacy
}

// GeneratePromptFromFramework creates a prompt using the selected framework and field values
// Returns system_prompt and user_prompt separately
func GeneratePromptFromFramework(frameworkID string, fields map[string]string) (*PromptPair, error) {
	framework := GetFrameworkByID(frameworkID)
	if framework == nil {
		return nil, fmt.Errorf("framework not found")
	}

	// Replace template variables with user input
	fullPrompt := framework.Template
	for fieldID, value := range fields {
		placeholder := "{{." + capitalize(fieldID) + "}}"
		if value != "" {
			fullPrompt = replaceTemplate(fullPrompt, placeholder, value)
		} else {
			// Remove optional sections if empty
			fullPrompt = replaceTemplate(fullPrompt, placeholder, "")
		}
	}

	// Generate system prompt and user prompt from framework fields
	systemPrompt, userPrompt := generateSystemAndUserPrompts(framework, fields)

	return &PromptPair{
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
		FullPrompt:   fullPrompt,
	}, nil
}

// generateSystemAndUserPrompts creates separate system and user prompts based on framework
func generateSystemAndUserPrompts(framework *Framework, fields map[string]string) (string, string) {
	// System prompt: Define the role and expectations
	systemPrompt := ""

	// User prompt: The actual request with context
	userPrompt := ""

	switch framework.ID {
	case "crispe":
		// Capacity becomes system role
		if capacity, ok := fields["capacity"]; ok && capacity != "" {
			systemPrompt = fmt.Sprintf("You are %s.", capacity)
		}

		// Performance expectations add to system prompt
		if performance, ok := fields["performance"]; ok && performance != "" {
			if systemPrompt != "" {
				systemPrompt += "\n\n"
			}
			systemPrompt += fmt.Sprintf("Output Requirements: %s", performance)
		}

		// Request, Input, Steps, Example become user prompt
		parts := []string{}
		if request, ok := fields["request"]; ok && request != "" {
			parts = append(parts, request)
		}
		if input, ok := fields["input"]; ok && input != "" {
			parts = append(parts, fmt.Sprintf("\nContext: %s", input))
		}
		if steps, ok := fields["steps"]; ok && steps != "" {
			parts = append(parts, fmt.Sprintf("\nSteps:\n%s", steps))
		}
		if example, ok := fields["example"]; ok && example != "" {
			parts = append(parts, fmt.Sprintf("\nExample: %s", example))
		}
		userPrompt = strings.Join(parts, "\n")

	case "risen":
		// Role becomes system prompt
		if role, ok := fields["role"]; ok && role != "" {
			systemPrompt = fmt.Sprintf("You are %s.", role)
		}

		// End goal and narrowing add to system
		if endGoal, ok := fields["end_goal"]; ok && endGoal != "" {
			if systemPrompt != "" {
				systemPrompt += "\n\n"
			}
			systemPrompt += fmt.Sprintf("Goal: %s", endGoal)
		}
		if narrowing, ok := fields["narrowing"]; ok && narrowing != "" {
			if systemPrompt != "" {
				systemPrompt += "\n"
			}
			systemPrompt += fmt.Sprintf("Constraints: %s", narrowing)
		}

		// Instructions and steps become user prompt
		parts := []string{}
		if instructions, ok := fields["instructions"]; ok && instructions != "" {
			parts = append(parts, instructions)
		}
		if steps, ok := fields["steps"]; ok && steps != "" {
			parts = append(parts, fmt.Sprintf("\nSteps:\n%s", steps))
		}
		userPrompt = strings.Join(parts, "\n")

	case "costar":
		// Style, Tone become system attributes
		systemParts := []string{}
		if style, ok := fields["style"]; ok && style != "" {
			systemParts = append(systemParts, fmt.Sprintf("Style: %s", style))
		}
		if tone, ok := fields["tone"]; ok && tone != "" {
			systemParts = append(systemParts, fmt.Sprintf("Tone: %s", tone))
		}
		if audience, ok := fields["audience"]; ok && audience != "" {
			systemParts = append(systemParts, fmt.Sprintf("Audience: %s", audience))
		}
		if response, ok := fields["response"]; ok && response != "" {
			systemParts = append(systemParts, fmt.Sprintf("Response Format: %s", response))
		}
		systemPrompt = strings.Join(systemParts, "\n")

		// Context and objective become user prompt
		userParts := []string{}
		if context, ok := fields["context"]; ok && context != "" {
			userParts = append(userParts, fmt.Sprintf("Context: %s", context))
		}
		if objective, ok := fields["objective"]; ok && objective != "" {
			userParts = append(userParts, objective)
		}
		userPrompt = strings.Join(userParts, "\n\n")

	case "ape":
		// Purpose becomes system context
		if purpose, ok := fields["purpose"]; ok && purpose != "" {
			systemPrompt = fmt.Sprintf("Purpose: %s", purpose)
		}
		if expectation, ok := fields["expectation"]; ok && expectation != "" {
			if systemPrompt != "" {
				systemPrompt += "\n"
			}
			systemPrompt += fmt.Sprintf("Expected Output: %s", expectation)
		}

		// Action becomes user prompt
		if action, ok := fields["action"]; ok && action != "" {
			userPrompt = action
		}

	case "rcar":
		// Role becomes system prompt
		if role, ok := fields["role"]; ok && role != "" {
			systemPrompt = role
		}

		// Context provides background (add to system)
		if context, ok := fields["context"]; ok && context != "" {
			if systemPrompt != "" {
				systemPrompt += "\n\n"
			}
			systemPrompt += fmt.Sprintf("Context: %s", context)
		}
		if result, ok := fields["result"]; ok && result != "" {
			if systemPrompt != "" {
				systemPrompt += "\n"
			}
			systemPrompt += fmt.Sprintf("Expected Result: %s", result)
		}

		// Action is the user request
		if action, ok := fields["action"]; ok && action != "" {
			userPrompt = action
		}

	case "race":
		// Role becomes system prompt
		if role, ok := fields["role"]; ok && role != "" {
			systemPrompt = role
		}

		// Context provides additional system context
		if context, ok := fields["context"]; ok && context != "" {
			if systemPrompt != "" {
				systemPrompt += "\n\n"
			}
			systemPrompt += fmt.Sprintf("Context: %s", context)
		}

		// Action becomes user prompt
		if action, ok := fields["action"]; ok && action != "" {
			userPrompt = action
		}

		// Example adds to user prompt
		if example, ok := fields["example"]; ok && example != "" {
			if userPrompt != "" {
				userPrompt += "\n\n"
			}
			userPrompt += fmt.Sprintf("Example:\n%s", example)
		}
	}

	// Fallback if either is empty
	if systemPrompt == "" {
		systemPrompt = "You are a helpful AI assistant."
	}
	if userPrompt == "" {
		userPrompt = "Please help with the task described above."
	}

	return systemPrompt, userPrompt
}

// Helper to capitalize first letter
func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	return string(s[0]-32) + s[1:]
}

// Helper to replace template placeholders
func replaceTemplate(template, placeholder, value string) string {
	// Basic template replacement (simplified version)
	// In production, use text/template package
	return template // This is a placeholder, actual implementation would use text/template
}

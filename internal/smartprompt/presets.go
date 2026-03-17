package smartprompt

import "github.com/walterfan/lazy-ai-coder/internal/models"

// DefaultPresets contains default presets for different tech stacks
var DefaultPresets = []models.Preset{
	{
		ID:           "java-spring",
		Name:         "Java Spring Boot",
		Language:     "Java",
		Framework:    "Spring Boot",
		ContextHints: "RESTful API, Maven/Gradle build, JPA/Hibernate ORM, Spring Security",
		ResultHints:  "Follow Spring Boot conventions, use proper annotations (@RestController, @Service, @Repository), include unit tests with JUnit 5 and Mockito, handle exceptions with @ControllerAdvice, use Lombok for boilerplate reduction",
	},
	{
		ID:           "python-fastapi",
		Name:         "Python FastAPI",
		Language:     "Python",
		Framework:    "FastAPI",
		ContextHints: "Async REST API, Pydantic models, SQLAlchemy ORM, Poetry/pip dependencies",
		ResultHints:  "Use type hints, async/await patterns, Pydantic for validation, include pytest tests, follow PEP 8 style guide, use dependency injection, document with OpenAPI/Swagger",
	},
	{
		ID:           "typescript-react",
		Name:         "TypeScript React",
		Language:     "TypeScript",
		Framework:    "React",
		ContextHints: "Frontend SPA, npm/yarn packages, Component-based architecture, State management (Redux/Context)",
		ResultHints:  "Use functional components with hooks, TypeScript interfaces for props, include Jest/RTL tests, follow React best practices, use CSS modules or styled-components, ensure accessibility",
	},
	{
		ID:           "golang",
		Name:         "Go",
		Language:     "Go",
		Framework:    "Standard Library / Gin",
		ContextHints: "Backend service, go.mod dependencies, Goroutines for concurrency, HTTP server",
		ResultHints:  "Follow Go conventions (gofmt, effective Go), use interfaces for abstraction, include table-driven tests, handle errors explicitly, use context for cancellation, document with godoc comments",
	},
	{
		ID:           "general",
		Name:         "General Programming",
		Language:     "Any",
		Framework:    "Any",
		ContextHints: "General software development",
		ResultHints:  "Write clean, maintainable code, include tests, follow SOLID principles, add comments for complex logic, consider edge cases, ensure error handling",
	},
}

// GetPresetByID returns a preset by ID
func GetPresetByID(presetID string) *models.Preset {
	for _, p := range DefaultPresets {
		if p.ID == presetID {
			return &p
		}
	}
	return nil
}


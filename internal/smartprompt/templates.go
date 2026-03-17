package smartprompt

// PromptTemplate represents a reusable prompt template
type PromptTemplate struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Category    string            `json:"category"`
	Framework   string            `json:"framework"` // Recommended framework ID
	Fields      map[string]string `json:"fields"`    // Pre-filled field values
	Tags        []string          `json:"tags"`
	UseCount    int               `json:"use_count"`
}

// TemplateCategory represents a category of templates
type TemplateCategory struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

// AllCategories contains all template categories
var AllCategories = []TemplateCategory{
	{
		ID:          "code-generation",
		Name:        "Code Generation",
		Description: "Generate new code, classes, functions, and modules",
		Icon:        "code",
	},
	{
		ID:          "code-review",
		Name:        "Code Review",
		Description: "Review code for quality, security, and best practices",
		Icon:        "search",
	},
	{
		ID:          "debugging",
		Name:        "Debugging",
		Description: "Analyze errors, find bugs, and suggest fixes",
		Icon:        "bug",
	},
	{
		ID:          "testing",
		Name:        "Testing",
		Description: "Generate tests, test data, and testing strategies",
		Icon:        "check-circle",
	},
	{
		ID:          "refactoring",
		Name:        "Refactoring",
		Description: "Improve code structure, performance, and maintainability",
		Icon:        "refresh",
	},
	{
		ID:          "documentation",
		Name:        "Documentation",
		Description: "Create API docs, README files, and code comments",
		Icon:        "book",
	},
	{
		ID:          "architecture",
		Name:        "Architecture & Design",
		Description: "Design systems, databases, and APIs",
		Icon:        "sitemap",
	},
}

// AllTemplates contains all available prompt templates
var AllTemplates = []PromptTemplate{
	// CODE GENERATION TEMPLATES
	{
		ID:          "rest-api-endpoint",
		Name:        "REST API Endpoint",
		Description: "Generate a complete REST API endpoint with validation, error handling, and tests",
		Category:    "code-generation",
		Framework:   "crispe",
		Fields: map[string]string{
			"capacity": "senior backend developer experienced in RESTful API design",
			"request":  "Create a {{METHOD}} {{ENDPOINT}} endpoint",
			"input":    "Framework: {{FRAMEWORK}}, Database: {{DATABASE}}, Authentication: {{AUTH}}",
			"steps": `1. Define request/response models with validation
2. Create database query/mutation
3. Implement endpoint handler with error handling
4. Add authentication/authorization checks
5. Write comprehensive unit tests
6. Add API documentation`,
			"performance": `Production-ready code with:
- Input validation and sanitization
- Proper HTTP status codes (200, 201, 400, 401, 403, 404, 500)
- Comprehensive error handling
- Type hints/annotations
- Unit tests with 80%+ coverage
- OpenAPI/Swagger documentation`,
		},
		Tags:    []string{"api", "rest", "endpoint", "backend"},
		UseCount: 0,
	},
	{
		ID:          "database-model",
		Name:        "Database Model/Schema",
		Description: "Create database models with relationships, constraints, and migrations",
		Category:    "code-generation",
		Framework:   "crispe",
		Fields: map[string]string{
			"capacity": "database architect with expertise in {{ORM}} and {{DATABASE}}",
			"request":  "Create a database model for {{ENTITY}}",
			"input":    "ORM: {{ORM}}, Database: {{DATABASE}}, Relationships: {{RELATIONSHIPS}}",
			"steps": `1. Define model class with all fields and types
2. Add constraints (unique, non-null, default values)
3. Define relationships (one-to-many, many-to-many)
4. Create indexes for query optimization
5. Generate migration script
6. Add model validation`,
			"performance": `Include:
- All field definitions with proper types
- Primary key and foreign keys
- Indexes for frequently queried fields
- Constraints and validations
- Relationship definitions
- Migration script for schema changes
- Example queries`,
		},
		Tags:    []string{"database", "model", "schema", "orm"},
		UseCount: 0,
	},
	{
		ID:          "cli-tool",
		Name:        "CLI Tool/Command",
		Description: "Create a command-line tool with arguments, options, and help text",
		Category:    "code-generation",
		Framework:   "crispe",
		Fields: map[string]string{
			"capacity": "senior developer experienced in building CLI tools",
			"request":  "Create a CLI tool for {{PURPOSE}}",
			"input":    "Language: {{LANGUAGE}}, CLI Library: {{CLI_LIB}}, Features: {{FEATURES}}",
			"steps": `1. Define command structure and subcommands
2. Add arguments and options with validation
3. Implement command logic
4. Add help text and examples
5. Include error handling
6. Write tests for commands`,
			"performance": `Deliver:
- Clean command interface with intuitive arguments
- Comprehensive help text with examples
- Input validation and error messages
- Exit codes for success/failure
- Configuration file support if needed
- Unit tests`,
		},
		Tags:    []string{"cli", "command-line", "tool"},
		UseCount: 0,
	},
	{
		ID:          "data-processor",
		Name:        "Data Processing Pipeline",
		Description: "Build a data processing pipeline for ETL, transformations, or batch processing",
		Category:    "code-generation",
		Framework:   "risen",
		Fields: map[string]string{
			"role":         "data engineer specializing in ETL pipelines",
			"instructions": "Build a data processing pipeline for {{PURPOSE}}. Input: {{INPUT_FORMAT}}, Output: {{OUTPUT_FORMAT}}, Processing: {{TRANSFORMATIONS}}",
			"steps": `1. Create data reader for input format
2. Implement validation and cleaning
3. Apply transformations and business logic
4. Handle errors and logging
5. Write to output format
6. Add monitoring and metrics
7. Create tests with sample data`,
			"end_goal": "A robust, production-ready data pipeline that processes {{INPUT_FORMAT}} to {{OUTPUT_FORMAT}} with proper error handling, logging, and monitoring",
			"narrowing": `Constraints:
- Handle large datasets efficiently (streaming/batching)
- Include data validation
- Log processing statistics
- Graceful error handling and retries
- Idempotent operations`,
		},
		Tags:    []string{"data", "etl", "pipeline", "processing"},
		UseCount: 0,
	},

	// CODE REVIEW TEMPLATES
	{
		ID:          "security-review",
		Name:        "Security Code Review",
		Description: "Comprehensive security analysis for vulnerabilities and best practices",
		Category:    "code-review",
		Framework:   "costar",
		Fields: map[string]string{
			"context":   "We have {{CODE_TYPE}} code that handles {{FUNCTIONALITY}}. Need security review before production deployment.",
			"objective": "Perform comprehensive security analysis identifying vulnerabilities, security anti-patterns, and providing remediation recommendations",
			"style":     "Technical security audit format with severity ratings",
			"tone":      "Professional, security-focused, actionable",
			"audience":  "Development team and security stakeholders",
			"response": `Structured report with:
1. Executive Summary
2. Critical Vulnerabilities (with CVSS scores)
3. Security Anti-patterns
4. OWASP Top 10 checklist
5. Remediation recommendations with code examples
6. Secure coding guidelines`,
		},
		Tags:    []string{"security", "review", "vulnerabilities", "owasp"},
		UseCount: 0,
	},
	{
		ID:          "performance-review",
		Name:        "Performance Analysis",
		Description: "Analyze code for performance bottlenecks and optimization opportunities",
		Category:    "code-review",
		Framework:   "costar",
		Fields: map[string]string{
			"context":   "{{CODE_TYPE}} experiencing performance issues. Current metrics: {{CURRENT_METRICS}}. Target: {{TARGET_METRICS}}",
			"objective": "Identify performance bottlenecks, inefficient algorithms, and provide optimization recommendations",
			"style":     "Technical performance analysis with metrics and benchmarks",
			"tone":      "Technical, data-driven, solution-oriented",
			"audience":  "Development team and technical leads",
			"response": `Analysis including:
1. Performance Bottlenecks (ranked by impact)
2. Time/Space Complexity Analysis
3. Database Query Optimization
4. Memory Usage Issues
5. Optimization Recommendations with expected improvements
6. Before/After code examples
7. Profiling suggestions`,
		},
		Tags:    []string{"performance", "optimization", "review", "bottlenecks"},
		UseCount: 0,
	},
	{
		ID:          "best-practices-review",
		Name:        "Best Practices Review",
		Description: "Review code against language-specific best practices and design patterns",
		Category:    "code-review",
		Framework:   "ape",
		Fields: map[string]string{
			"action":      "Review {{CODE_TYPE}} code against {{LANGUAGE}} best practices, design patterns, and coding standards",
			"purpose":     "Ensure code quality, maintainability, and adherence to team/industry standards before merging to main branch",
			"expectation": "Detailed review identifying violations of best practices, suggesting improvements, and providing refactored code examples. Include ratings for readability, maintainability, and adherence to patterns.",
		},
		Tags:    []string{"best-practices", "review", "patterns", "standards"},
		UseCount: 0,
	},

	// DEBUGGING TEMPLATES
	{
		ID:          "error-analysis",
		Name:        "Error Analysis & Fix",
		Description: "Analyze error messages/stack traces and provide detailed fix recommendations",
		Category:    "debugging",
		Framework:   "risen",
		Fields: map[string]string{
			"role":         "senior debugging specialist with deep knowledge of {{LANGUAGE}} and {{FRAMEWORK}}",
			"instructions": "Analyze the following error and provide root cause analysis with fix: {{ERROR_MESSAGE}}",
			"steps": `1. Parse error message and stack trace
2. Identify root cause (logic error, missing dependency, config issue, etc.)
3. Explain why the error occurs
4. Provide step-by-step fix instructions
5. Show corrected code
6. Suggest preventive measures`,
			"end_goal":  "Complete understanding of error cause and working fix that resolves the issue permanently",
			"narrowing": "Focus on {{ERROR_TYPE}} errors. Consider {{ENVIRONMENT}} environment constraints.",
		},
		Tags:    []string{"debugging", "error", "fix", "troubleshooting"},
		UseCount: 0,
	},
	{
		ID:          "bug-investigation",
		Name:        "Bug Investigation",
		Description: "Investigate unexpected behavior and find the bug causing it",
		Category:    "debugging",
		Framework:   "risen",
		Fields: map[string]string{
			"role":         "debugging expert specializing in {{DOMAIN}}",
			"instructions": "Investigate bug: {{BUG_DESCRIPTION}}. Expected: {{EXPECTED}}. Actual: {{ACTUAL}}",
			"steps": `1. Reproduce the issue with minimal test case
2. Add logging/debugging statements
3. Trace execution flow
4. Identify where behavior diverges
5. Determine root cause
6. Provide fix with explanation
7. Add regression test`,
			"end_goal":  "Identify root cause, provide fix, and add test to prevent regression",
			"narrowing": "Code area: {{CODE_AREA}}. Constraints: {{CONSTRAINTS}}",
		},
		Tags:    []string{"bug", "investigation", "debugging", "troubleshooting"},
		UseCount: 0,
	},
	{
		ID:          "crash-analysis",
		Name:        "Crash/Exception Analysis",
		Description: "Analyze crash dumps or exception logs to find the cause",
		Category:    "debugging",
		Framework:   "ape",
		Fields: map[string]string{
			"action":      "Analyze crash/exception: {{CRASH_LOG}} occurring in {{COMPONENT}}",
			"purpose":     "Determine root cause of application crash and prevent future occurrences",
			"expectation": "Root cause analysis explaining why crash occurs, fix to prevent it, and monitoring to detect similar issues early",
		},
		Tags:    []string{"crash", "exception", "analysis", "debugging"},
		UseCount: 0,
	},

	// TESTING TEMPLATES
	{
		ID:          "unit-tests",
		Name:        "Unit Tests Suite",
		Description: "Generate comprehensive unit tests with edge cases and mocking",
		Category:    "testing",
		Framework:   "crispe",
		Fields: map[string]string{
			"capacity": "QA engineer expert in {{TESTING_FRAMEWORK}} and {{LANGUAGE}}",
			"request":  "Create comprehensive unit tests for {{FUNCTION_OR_CLASS}}",
			"input":    "Testing Framework: {{TESTING_FRAMEWORK}}, Mocking Library: {{MOCKING_LIB}}, Code to test: {{CODE}}",
			"steps": `1. Identify test scenarios (happy path, edge cases, errors)
2. Set up test fixtures and mocks
3. Write positive test cases
4. Write negative test cases
5. Write edge case tests
6. Add assertions for all return values and side effects
7. Ensure 80%+ code coverage`,
			"performance": `Deliver:
- Test setup and teardown
- Mocks for external dependencies
- Happy path tests
- Error condition tests
- Edge case tests (null, empty, large inputs)
- Clear test names describing what is tested
- AAA pattern (Arrange-Act-Assert)
- 80%+ code coverage`,
		},
		Tags:    []string{"testing", "unit-tests", "test-driven"},
		UseCount: 0,
	},
	{
		ID:          "integration-tests",
		Name:        "Integration Tests",
		Description: "Create integration tests for API endpoints, database operations, or services",
		Category:    "testing",
		Framework:   "crispe",
		Fields: map[string]string{
			"capacity": "QA automation engineer with integration testing expertise",
			"request":  "Create integration tests for {{INTEGRATION_POINT}}",
			"input":    "System: {{SYSTEM}}, Components: {{COMPONENTS}}, Test Environment: {{ENV}}",
			"steps": `1. Set up test database/environment
2. Create test data fixtures
3. Write tests for successful integration
4. Test error scenarios and rollbacks
5. Verify data consistency
6. Test concurrent operations if applicable
7. Add cleanup after tests`,
			"performance": `Include:
- Environment setup and cleanup
- Test data creation
- Success scenario tests
- Failure scenario tests
- Data verification
- Transaction/rollback tests
- Clear test descriptions`,
		},
		Tags:    []string{"testing", "integration", "api-testing"},
		UseCount: 0,
	},
	{
		ID:          "test-data-generator",
		Name:        "Test Data Generator",
		Description: "Create realistic test data for testing and development",
		Category:    "testing",
		Framework:   "ape",
		Fields: map[string]string{
			"action":      "Generate realistic test data for {{ENTITY_TYPE}} with {{QUANTITY}} records",
			"purpose":     "Provide representative test data for development, testing, and demos without using production data",
			"expectation": "Script that generates realistic, varied test data matching schema {{SCHEMA}} with configurable quantity and customization options",
		},
		Tags:    []string{"testing", "test-data", "generator", "fixtures"},
		UseCount: 0,
	},
	{
		ID:          "e2e-tests",
		Name:        "End-to-End Tests",
		Description: "Create E2E tests for complete user workflows",
		Category:    "testing",
		Framework:   "risen",
		Fields: map[string]string{
			"role":         "QA automation engineer specializing in E2E testing with {{E2E_FRAMEWORK}}",
			"instructions": "Create E2E tests for user workflow: {{WORKFLOW}}",
			"steps": `1. Break down user journey into steps
2. Set up test environment and data
3. Implement page objects/selectors
4. Write test for happy path
5. Add tests for alternative paths
6. Test error handling
7. Add waits and retries for stability
8. Clean up test data`,
			"end_goal":  "Reliable E2E tests covering complete user workflows that can run in CI/CD pipeline",
			"narrowing": "Framework: {{E2E_FRAMEWORK}}, Target: {{TARGET_ENV}}, Constraints: {{CONSTRAINTS}}",
		},
		Tags:    []string{"e2e", "testing", "automation", "workflows"},
		UseCount: 0,
	},

	// REFACTORING TEMPLATES
	{
		ID:          "clean-code-refactor",
		Name:        "Clean Code Refactoring",
		Description: "Refactor code to improve readability, maintainability, and follow clean code principles",
		Category:    "refactoring",
		Framework:   "costar",
		Fields: map[string]string{
			"context":   "Legacy code in {{CODE_AREA}} that is difficult to understand and maintain: {{CODE}}",
			"objective": "Refactor to clean code following SOLID principles, improving readability and maintainability without changing behavior",
			"style":     "Clean code, following {{LANGUAGE}} conventions and best practices",
			"tone":      "Professional, educational",
			"audience":  "Development team maintaining this codebase",
			"response": `Provide:
1. Refactored code with explanations
2. Changes made (rename, extract method, remove duplication, etc.)
3. Before/after comparison
4. How it improves maintainability
5. Tests to verify behavior unchanged`,
		},
		Tags:    []string{"refactoring", "clean-code", "maintainability"},
		UseCount: 0,
	},
	{
		ID:          "performance-optimization",
		Name:        "Performance Optimization",
		Description: "Optimize code for better performance (speed, memory, database queries)",
		Category:    "refactoring",
		Framework:   "risen",
		Fields: map[string]string{
			"role":         "performance optimization specialist",
			"instructions": "Optimize {{CODE_AREA}} to improve {{PERFORMANCE_METRIC}}. Current: {{CURRENT}}, Target: {{TARGET}}",
			"steps": `1. Profile code to identify bottlenecks
2. Analyze algorithm complexity
3. Optimize data structures
4. Reduce unnecessary operations
5. Add caching where appropriate
6. Optimize database queries
7. Benchmark improvements`,
			"end_goal":  "Optimized code meeting target performance metrics with benchmark data proving improvements",
			"narrowing": "Focus on {{OPTIMIZATION_TYPE}}. Constraints: {{CONSTRAINTS}}. Don't sacrifice readability for micro-optimizations.",
		},
		Tags:    []string{"refactoring", "performance", "optimization"},
		UseCount: 0,
	},
	{
		ID:          "design-pattern",
		Name:        "Apply Design Pattern",
		Description: "Refactor code to implement specific design pattern",
		Category:    "refactoring",
		Framework:   "crispe",
		Fields: map[string]string{
			"capacity":    "software architect with expertise in design patterns and {{LANGUAGE}}",
			"request":     "Refactor {{CODE_AREA}} to implement {{PATTERN}} design pattern",
			"input":       "Current code: {{CODE}}, Pattern: {{PATTERN}}, Reason: {{REASON}}",
			"steps":       "1. Analyze current code structure\n2. Identify pattern participants\n3. Refactor to pattern structure\n4. Update related code\n5. Verify behavior unchanged\n6. Add documentation",
			"performance": "Refactored code implementing {{PATTERN}} pattern correctly with clear class/function responsibilities, improved flexibility, and documentation explaining the pattern",
		},
		Tags:    []string{"refactoring", "design-patterns", "architecture"},
		UseCount: 0,
	},
	{
		ID:          "tech-debt-reduction",
		Name:        "Technical Debt Reduction",
		Description: "Address technical debt, remove workarounds, update deprecated code",
		Category:    "refactoring",
		Framework:   "ape",
		Fields: map[string]string{
			"action":      "Address technical debt in {{CODE_AREA}}: {{DEBT_DESCRIPTION}}",
			"purpose":     "Reduce maintenance burden, improve code quality, and eliminate fragile workarounds before they cause production issues",
			"expectation": "Refactored code removing technical debt with explanation of what debt was removed, why it existed, and how refactored code is better. Include tests proving correctness.",
		},
		Tags:    []string{"refactoring", "tech-debt", "maintenance"},
		UseCount: 0,
	},

	// DOCUMENTATION TEMPLATES
	{
		ID:          "api-documentation",
		Name:        "API Documentation",
		Description: "Generate comprehensive API documentation with examples",
		Category:    "documentation",
		Framework:   "costar",
		Fields: map[string]string{
			"context":   "REST API with endpoints for {{DOMAIN}}. Need comprehensive documentation for external developers.",
			"objective": "Generate OpenAPI/Swagger documentation for all endpoints with descriptions, schemas, examples, and error responses",
			"style":     "Professional API documentation following OpenAPI 3.0 standard",
			"tone":      "Technical, clear, developer-friendly",
			"audience":  "External API consumers and integration developers",
			"response":  "OpenAPI 3.0 YAML specification including: API info, all endpoints, request/response schemas, example payloads, error codes, authentication, and common schemas",
		},
		Tags:    []string{"documentation", "api", "openapi", "swagger"},
		UseCount: 0,
	},
	{
		ID:          "readme-generator",
		Name:        "README File",
		Description: "Create comprehensive README with installation, usage, and examples",
		Category:    "documentation",
		Framework:   "costar",
		Fields: map[string]string{
			"context":   "Project: {{PROJECT_NAME}}, Purpose: {{PURPOSE}}, Tech Stack: {{TECH_STACK}}",
			"objective": "Create comprehensive README.md for GitHub repository",
			"style":     "Professional open-source project documentation with Markdown formatting",
			"tone":      "Welcoming, clear, encouraging contribution",
			"audience":  "Potential users, contributors, and developers",
			"response": `README.md including:
- Project description with badges
- Features list
- Prerequisites
- Installation instructions
- Configuration guide
- Usage examples
- API documentation link
- Contributing guidelines
- License
- Contact information`,
		},
		Tags:    []string{"documentation", "readme", "github"},
		UseCount: 0,
	},
	{
		ID:          "code-comments",
		Name:        "Code Comments & Docstrings",
		Description: "Add comprehensive inline comments and docstrings to existing code",
		Category:    "documentation",
		Framework:   "ape",
		Fields: map[string]string{
			"action":      "Add comprehensive comments and docstrings to: {{CODE}}",
			"purpose":     "Improve code understanding for team members and future maintainers",
			"expectation": "Code with docstrings for all public functions/classes, inline comments explaining complex logic, and examples where helpful. Follow {{LANGUAGE}} documentation conventions.",
		},
		Tags:    []string{"documentation", "comments", "docstrings"},
		UseCount: 0,
	},
	{
		ID:          "technical-spec",
		Name:        "Technical Specification",
		Description: "Write technical specification document for a feature or system",
		Category:    "documentation",
		Framework:   "risen",
		Fields: map[string]string{
			"role":         "technical writer with engineering background",
			"instructions": "Create technical specification for: {{FEATURE_OR_SYSTEM}}",
			"steps": `1. Overview and objectives
2. System architecture/components
3. Technical requirements
4. API/interface definitions
5. Data models and schemas
6. Error handling strategy
7. Security considerations
8. Performance requirements
9. Testing strategy`,
			"end_goal":  "Complete technical specification that engineers can implement from",
			"narrowing": "Audience: {{AUDIENCE}}, Detail level: {{DETAIL_LEVEL}}",
		},
		Tags:    []string{"documentation", "specification", "technical-writing"},
		UseCount: 0,
	},

	// ARCHITECTURE & DESIGN TEMPLATES
	{
		ID:          "system-design",
		Name:        "System Architecture Design",
		Description: "Design overall system architecture with components, data flow, and infrastructure",
		Category:    "architecture",
		Framework:   "risen",
		Fields: map[string]string{
			"role":         "system architect with experience in {{DOMAIN}}",
			"instructions": "Design system architecture for: {{SYSTEM_DESCRIPTION}}",
			"steps": `1. Identify system requirements and constraints
2. Define high-level components
3. Design data flow and interactions
4. Choose technology stack
5. Design data storage strategy
6. Plan for scalability and availability
7. Address security concerns
8. Consider monitoring and operations`,
			"end_goal": "Complete system architecture with component diagram, technology choices, data flow, scalability plan, and trade-off analysis",
			"narrowing": "Scale: {{SCALE}}, Budget: {{BUDGET}}, Constraints: {{CONSTRAINTS}}",
		},
		Tags:    []string{"architecture", "system-design", "infrastructure"},
		UseCount: 0,
	},
	{
		ID:          "database-design",
		Name:        "Database Schema Design",
		Description: "Design normalized database schema with tables, relationships, and indexes",
		Category:    "architecture",
		Framework:   "crispe",
		Fields: map[string]string{
			"capacity": "database architect with expertise in {{DATABASE_TYPE}}",
			"request":  "Design database schema for {{DOMAIN}}",
			"input":    "Requirements: {{REQUIREMENTS}}, Database: {{DATABASE}}, Expected Scale: {{SCALE}}",
			"steps": `1. Identify entities and attributes
2. Define relationships (1:1, 1:N, N:M)
3. Normalize to 3NF
4. Add indexes for common queries
5. Consider partitioning strategy for scale
6. Add constraints and validations
7. Plan migration strategy`,
			"performance": `Deliver:
- ER diagram
- Table definitions with all columns and types
- Relationships with foreign keys
- Indexes for optimization
- Constraints and validations
- Sample queries
- Migration scripts`,
		},
		Tags:    []string{"architecture", "database", "schema-design"},
		UseCount: 0,
	},
	{
		ID:          "api-design",
		Name:        "API Design",
		Description: "Design RESTful or GraphQL API with endpoints, resources, and versioning",
		Category:    "architecture",
		Framework:   "risen",
		Fields: map[string]string{
			"role":         "API architect specializing in {{API_TYPE}} design",
			"instructions": "Design {{API_TYPE}} API for {{DOMAIN}}",
			"steps": `1. Identify resources and operations
2. Define endpoint structure and naming
3. Design request/response formats
4. Plan authentication and authorization
5. Define error handling strategy
6. Add pagination, filtering, sorting
7. Version strategy
8. Rate limiting and caching`,
			"end_goal":  "Complete API design with all endpoints, schemas, examples, and OpenAPI specification",
			"narrowing": "API Type: {{API_TYPE}}, Consumers: {{CONSUMERS}}, Constraints: {{CONSTRAINTS}}",
		},
		Tags:    []string{"architecture", "api-design", "rest", "graphql"},
		UseCount: 0,
	},
	{
		ID:          "microservices-design",
		Name:        "Microservices Architecture",
		Description: "Design microservices architecture with service boundaries and communication",
		Category:    "architecture",
		Framework:   "risen",
		Fields: map[string]string{
			"role":         "microservices architect with domain-driven design expertise",
			"instructions": "Design microservices architecture for {{SYSTEM}}",
			"steps": `1. Identify bounded contexts using DDD
2. Define service responsibilities
3. Design inter-service communication
4. Plan data management per service
5. Design API gateway strategy
6. Address distributed system challenges
7. Plan deployment and orchestration
8. Design monitoring and observability`,
			"end_goal":  "Microservices architecture with service diagram, API contracts, data strategy, and deployment plan",
			"narrowing": "Services: max {{MAX_SERVICES}}, Communication: {{COMM_TYPE}}, Infrastructure: {{INFRA}}",
		},
		Tags:    []string{"architecture", "microservices", "distributed-systems"},
		UseCount: 0,
	},
}

// GetTemplatesByCategory returns templates filtered by category
func GetTemplatesByCategory(category string) []PromptTemplate {
	var filtered []PromptTemplate
	for _, template := range AllTemplates {
		if template.Category == category {
			filtered = append(filtered, template)
		}
	}
	return filtered
}

// GetTemplateByID returns a template by ID
func GetTemplateByID(id string) *PromptTemplate {
	for i := range AllTemplates {
		if AllTemplates[i].ID == id {
			return &AllTemplates[i]
		}
	}
	return nil
}

// SearchTemplates searches templates by name, description, or tags
func SearchTemplates(query string) []PromptTemplate {
	// Implement search logic
	var results []PromptTemplate
	// ... search implementation
	return results
}

// IncrementTemplateUseCount increments the use count for a template
func IncrementTemplateUseCount(templateID string) {
	for i := range AllTemplates {
		if AllTemplates[i].ID == templateID {
			AllTemplates[i].UseCount++
			break
		}
	}
}

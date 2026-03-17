package smartprompt

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/walterfan/lazy-ai-coder/internal/models"
)

// ParsePomXml parses Maven pom.xml content
func ParsePomXml(content string, ctx *models.ProjectContext) error {
	ctx.Language = "Java"
	ctx.BuildTool = "Maven"

	// Simple string matching for common frameworks
	if strings.Contains(content, "spring-boot-starter") {
		ctx.Framework = "Spring Boot"
		// Extract version
		if idx := strings.Index(content, "<spring-boot.version>"); idx != -1 {
			versionEnd := strings.Index(content[idx:], "</spring-boot.version>")
			if versionEnd != -1 {
				ctx.FrameworkVersion = strings.TrimSpace(content[idx+22 : idx+versionEnd])
			}
		}
	}

	// Check for database dependencies
	if strings.Contains(content, "mysql-connector") || strings.Contains(content, "mariadb") {
		ctx.Database = "MySQL"
	} else if strings.Contains(content, "postgresql") {
		ctx.Database = "PostgreSQL"
	} else if strings.Contains(content, "h2") {
		ctx.Database = "H2"
	}

	// Check for test framework
	if strings.Contains(content, "junit-jupiter") {
		ctx.TestFramework = "JUnit 5"
		ctx.HasTests = true
	} else if strings.Contains(content, "junit") {
		ctx.TestFramework = "JUnit 4"
		ctx.HasTests = true
	}

	// Extract some dependencies
	dependencies := []string{}
	if strings.Contains(content, "spring-boot-starter-web") {
		dependencies = append(dependencies, "Spring Web")
	}
	if strings.Contains(content, "spring-boot-starter-data-jpa") {
		dependencies = append(dependencies, "Spring Data JPA")
	}
	if strings.Contains(content, "lombok") {
		dependencies = append(dependencies, "Lombok")
	}
	ctx.Dependencies = dependencies

	return nil
}

// ParsePackageJson parses package.json content
func ParsePackageJson(content string, ctx *models.ProjectContext) error {
	ctx.Language = "JavaScript/TypeScript"
	ctx.BuildTool = "npm"

	// Parse JSON
	var pkg map[string]interface{}
	if err := json.Unmarshal([]byte(content), &pkg); err != nil {
		return err
	}

	// Check dependencies
	deps := make(map[string]string)
	if dependencies, ok := pkg["dependencies"].(map[string]interface{}); ok {
		for name, version := range dependencies {
			deps[name] = fmt.Sprintf("%v", version)
		}
	}
	if devDeps, ok := pkg["devDependencies"].(map[string]interface{}); ok {
		for name, version := range devDeps {
			deps[name] = fmt.Sprintf("%v", version)
		}
	}

	// Detect framework
	if _, ok := deps["react"]; ok {
		ctx.Framework = "React"
		if version, ok := deps["react"]; ok {
			ctx.FrameworkVersion = version
		}
	} else if _, ok := deps["vue"]; ok {
		ctx.Framework = "Vue"
		if version, ok := deps["vue"]; ok {
			ctx.FrameworkVersion = version
		}
	} else if _, ok := deps["@angular/core"]; ok {
		ctx.Framework = "Angular"
		if version, ok := deps["@angular/core"]; ok {
			ctx.FrameworkVersion = version
		}
	} else if _, ok := deps["fastapi"]; ok {
		ctx.Framework = "FastAPI"
		ctx.Language = "Python"
	}

	// Check for TypeScript
	if _, ok := deps["typescript"]; ok {
		ctx.Language = "TypeScript"
	}

	// Check test framework
	if _, ok := deps["jest"]; ok {
		ctx.TestFramework = "Jest"
		ctx.HasTests = true
	} else if _, ok := deps["mocha"]; ok {
		ctx.TestFramework = "Mocha"
		ctx.HasTests = true
	} else if _, ok := deps["vitest"]; ok {
		ctx.TestFramework = "Vitest"
		ctx.HasTests = true
	}

	// Collect key dependencies
	keyDeps := []string{}
	for dep := range deps {
		if strings.Contains(dep, "react") || strings.Contains(dep, "vue") ||
			strings.Contains(dep, "express") || strings.Contains(dep, "axios") {
			keyDeps = append(keyDeps, dep)
			if len(keyDeps) >= 5 {
				break
			}
		}
	}
	ctx.Dependencies = keyDeps

	return nil
}

// ParseRequirementsTxt parses requirements.txt content
func ParseRequirementsTxt(content string, ctx *models.ProjectContext) error {
	ctx.Language = "Python"
	ctx.BuildTool = "pip"

	lines := strings.Split(content, "\n")
	dependencies := []string{}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Extract package name (before == or >=)
		pkg := strings.Split(strings.Split(line, "==")[0], ">=")[0]
		pkg = strings.TrimSpace(pkg)

		// Detect framework
		if strings.Contains(strings.ToLower(pkg), "django") {
			ctx.Framework = "Django"
		} else if strings.Contains(strings.ToLower(pkg), "fastapi") {
			ctx.Framework = "FastAPI"
		} else if strings.Contains(strings.ToLower(pkg), "flask") {
			ctx.Framework = "Flask"
		}

		// Detect database
		if strings.Contains(strings.ToLower(pkg), "psycopg2") {
			ctx.Database = "PostgreSQL"
		} else if strings.Contains(strings.ToLower(pkg), "mysql") {
			ctx.Database = "MySQL"
		}

		// Detect test framework
		if strings.Contains(strings.ToLower(pkg), "pytest") {
			ctx.TestFramework = "pytest"
			ctx.HasTests = true
		}

		dependencies = append(dependencies, pkg)
		if len(dependencies) >= 10 {
			break
		}
	}

	ctx.Dependencies = dependencies
	return nil
}

// ParseGoMod parses go.mod content
func ParseGoMod(content string, ctx *models.ProjectContext) error {
	ctx.Language = "Go"
	ctx.BuildTool = "go mod"

	lines := strings.Split(content, "\n")
	dependencies := []string{}

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Detect framework
		if strings.Contains(line, "github.com/gin-gonic/gin") {
			ctx.Framework = "Gin"
		} else if strings.Contains(line, "github.com/labstack/echo") {
			ctx.Framework = "Echo"
		}

		// Detect database
		if strings.Contains(line, "gorm.io/gorm") {
			dependencies = append(dependencies, "GORM")
		}
		if strings.Contains(line, "github.com/go-sql-driver/mysql") {
			ctx.Database = "MySQL"
		} else if strings.Contains(line, "github.com/lib/pq") {
			ctx.Database = "PostgreSQL"
		}

		// Collect dependencies
		if strings.HasPrefix(line, "require ") || (strings.Contains(line, "github.com/") && !strings.HasPrefix(line, "module")) {
			parts := strings.Fields(line)
			if len(parts) >= 1 {
				dep := strings.TrimPrefix(parts[0], "require")
				dep = strings.TrimSpace(dep)
				if dep != "" && strings.Contains(dep, "/") {
					dependencies = append(dependencies, dep)
					if len(dependencies) >= 10 {
						break
					}
				}
			}
		}
	}

	// Check for test framework (Go has built-in testing)
	ctx.TestFramework = "testing (built-in)"
	ctx.HasTests = true // Assume tests exist for Go projects

	ctx.Dependencies = dependencies
	return nil
}

// ParseBuildGradle parses build.gradle content
func ParseBuildGradle(content string, ctx *models.ProjectContext) error {
	ctx.Language = "Java"
	ctx.BuildTool = "Gradle"

	// Similar to pom.xml parsing
	if strings.Contains(content, "spring-boot") {
		ctx.Framework = "Spring Boot"
	}

	if strings.Contains(content, "mysql") {
		ctx.Database = "MySQL"
	} else if strings.Contains(content, "postgresql") {
		ctx.Database = "PostgreSQL"
	}

	if strings.Contains(content, "junit") {
		ctx.TestFramework = "JUnit"
		ctx.HasTests = true
	}

	return nil
}

// ParseCargoToml parses Cargo.toml content
func ParseCargoToml(content string, ctx *models.ProjectContext) error {
	ctx.Language = "Rust"
	ctx.BuildTool = "Cargo"

	if strings.Contains(content, "actix-web") {
		ctx.Framework = "Actix Web"
	} else if strings.Contains(content, "rocket") {
		ctx.Framework = "Rocket"
	}

	ctx.TestFramework = "cargo test (built-in)"
	ctx.HasTests = true

	return nil
}


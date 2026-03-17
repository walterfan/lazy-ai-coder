package main

import (
	"github.com/walterfan/lazy-ai-coder/cmd"
	_ "github.com/walterfan/lazy-ai-coder/docs" // swagger docs
)

// @title Lazy AI Coder API
// @version 1.0
// @description API documentation for Lazy AI Coder - GitLab integration, LLM tools, and MCP server
// @description
// @description Features:
// @description - LLM chat and code analysis
// @description - GitLab integration (read files, review MRs)
// @description - MCP (Model Context Protocol) server
// @description - PlantUML diagram generation
// @description - Session management with memory

// @contact.name API Support
// @contact.email walter.fan@gmail.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8888
// @BasePath /api/v1

// @securityDefinitions.basic BasicAuth

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	cmd.Execute()
}

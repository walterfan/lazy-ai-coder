package rag

import (
	"context"
	"fmt"
	"os"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/cpp"
	"github.com/smacker/go-tree-sitter/golang"
	"github.com/smacker/go-tree-sitter/java"
	"github.com/smacker/go-tree-sitter/javascript"
	"github.com/smacker/go-tree-sitter/python"
)

// CodeMetadata contains extracted code information
type CodeMetadata struct {
	FilePath     string
	Language     LanguageType
	Content      string
	Functions    []FunctionInfo
	Classes      []ClassInfo
	Imports      []string
	Comments     []string
	Symbols      []string // All function/class names
}

// FunctionInfo represents a function/method
type FunctionInfo struct {
	Name      string
	Content   string
	StartLine int
	EndLine   int
	Comment   string
}

// ClassInfo represents a class/struct
type ClassInfo struct {
	Name      string
	Content   string
	StartLine int
	EndLine   int
	Methods   []FunctionInfo
	Comment   string
}

// CodeParser handles parsing of code files with tree-sitter
type CodeParser struct {
	parser *sitter.Parser
}

// NewCodeParser creates a new code parser
func NewCodeParser() *CodeParser {
	return &CodeParser{
		parser: sitter.NewParser(),
	}
}

// ParseCode extracts metadata from a code file
func (cp *CodeParser) ParseCode(filePath string) (*CodeMetadata, error) {
	info := DetectFileType(filePath)
	if info.Type != FileTypeCode && info.Type != FileTypeConfig {
		return nil, fmt.Errorf("not a code file: %s", filePath)
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return cp.ParseCodeContent(string(content), filePath, info.Language)
}

// ParseCodeContent parses code content and extracts metadata
func (cp *CodeParser) ParseCodeContent(content, filePath string, lang LanguageType) (*CodeMetadata, error) {
	metadata := &CodeMetadata{
		FilePath:  filePath,
		Language:  lang,
		Content:   content,
		Functions: []FunctionInfo{},
		Classes:   []ClassInfo{},
		Imports:   []string{},
		Comments:  []string{},
		Symbols:   []string{},
	}

	// Set language
	language := cp.getLanguage(lang)
	if language == nil {
		// If tree-sitter not available, return basic metadata
		return metadata, nil
	}

	cp.parser.SetLanguage(language)

	// Parse the source code
	tree, err := cp.parser.ParseCtx(context.Background(), nil, []byte(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse code: %w", err)
	}
	defer tree.Close()

	root := tree.RootNode()

	// Extract metadata based on language
	switch lang {
	case LangGo:
		cp.parseGo(root, []byte(content), metadata)
	case LangPython:
		cp.parsePython(root, []byte(content), metadata)
	case LangJava:
		cp.parseJava(root, []byte(content), metadata)
	case LangCpp:
		cp.parseCpp(root, []byte(content), metadata)
	case LangJavaScript:
		cp.parseJavaScript(root, []byte(content), metadata)
	default:
		// For unsupported languages, return basic metadata
		return metadata, nil
	}

	// Extract comments
	cp.extractComments(root, []byte(content), metadata)

	return metadata, nil
}

// getLanguage returns the tree-sitter language for a given language type
func (cp *CodeParser) getLanguage(lang LanguageType) *sitter.Language {
	switch lang {
	case LangGo:
		return golang.GetLanguage()
	case LangPython:
		return python.GetLanguage()
	case LangJava:
		return java.GetLanguage()
	case LangCpp:
		return cpp.GetLanguage()
	case LangJavaScript:
		return javascript.GetLanguage()
	default:
		return nil
	}
}

// parseGo extracts Go-specific metadata
func (cp *CodeParser) parseGo(node *sitter.Node, source []byte, metadata *CodeMetadata) {
	cp.traverse(node, source, func(n *sitter.Node, src []byte) {
		switch n.Type() {
		case "import_declaration":
			importPath := n.Content(src)
			metadata.Imports = append(metadata.Imports, strings.TrimSpace(importPath))

		case "function_declaration", "method_declaration":
			funcInfo := cp.extractFunction(n, src)
			if funcInfo.Name != "" {
				metadata.Functions = append(metadata.Functions, funcInfo)
				metadata.Symbols = append(metadata.Symbols, funcInfo.Name)
			}

		case "type_declaration":
			// Extract struct/interface names
			if n.ChildCount() > 0 {
				for i := 0; i < int(n.ChildCount()); i++ {
					child := n.Child(i)
					if child.Type() == "type_spec" {
						name := cp.getIdentifierName(child, src)
						if name != "" {
							classInfo := ClassInfo{
								Name:      name,
								Content:   child.Content(src),
								StartLine: int(child.StartPoint().Row) + 1,
								EndLine:   int(child.EndPoint().Row) + 1,
							}
							metadata.Classes = append(metadata.Classes, classInfo)
							metadata.Symbols = append(metadata.Symbols, name)
						}
					}
				}
			}
		}
	})
}

// parsePython extracts Python-specific metadata
func (cp *CodeParser) parsePython(node *sitter.Node, source []byte, metadata *CodeMetadata) {
	cp.traverse(node, source, func(n *sitter.Node, src []byte) {
		switch n.Type() {
		case "import_statement", "import_from_statement":
			importStmt := n.Content(src)
			metadata.Imports = append(metadata.Imports, strings.TrimSpace(importStmt))

		case "function_definition":
			funcInfo := cp.extractFunction(n, src)
			if funcInfo.Name != "" {
				metadata.Functions = append(metadata.Functions, funcInfo)
				metadata.Symbols = append(metadata.Symbols, funcInfo.Name)
			}

		case "class_definition":
			classInfo := cp.extractClass(n, src)
			if classInfo.Name != "" {
				metadata.Classes = append(metadata.Classes, classInfo)
				metadata.Symbols = append(metadata.Symbols, classInfo.Name)
			}
		}
	})
}

// parseJava extracts Java-specific metadata
func (cp *CodeParser) parseJava(node *sitter.Node, source []byte, metadata *CodeMetadata) {
	cp.traverse(node, source, func(n *sitter.Node, src []byte) {
		switch n.Type() {
		case "import_declaration":
			importStmt := n.Content(src)
			metadata.Imports = append(metadata.Imports, strings.TrimSpace(importStmt))

		case "method_declaration":
			funcInfo := cp.extractFunction(n, src)
			if funcInfo.Name != "" {
				metadata.Functions = append(metadata.Functions, funcInfo)
				metadata.Symbols = append(metadata.Symbols, funcInfo.Name)
			}

		case "class_declaration":
			classInfo := cp.extractClass(n, src)
			if classInfo.Name != "" {
				metadata.Classes = append(metadata.Classes, classInfo)
				metadata.Symbols = append(metadata.Symbols, classInfo.Name)
			}
		}
	})
}

// parseCpp extracts C++-specific metadata
func (cp *CodeParser) parseCpp(node *sitter.Node, source []byte, metadata *CodeMetadata) {
	cp.traverse(node, source, func(n *sitter.Node, src []byte) {
		switch n.Type() {
		case "preproc_include":
			includeStmt := n.Content(src)
			metadata.Imports = append(metadata.Imports, strings.TrimSpace(includeStmt))

		case "function_definition":
			funcInfo := cp.extractFunction(n, src)
			if funcInfo.Name != "" {
				metadata.Functions = append(metadata.Functions, funcInfo)
				metadata.Symbols = append(metadata.Symbols, funcInfo.Name)
			}

		case "class_specifier", "struct_specifier":
			classInfo := cp.extractClass(n, src)
			if classInfo.Name != "" {
				metadata.Classes = append(metadata.Classes, classInfo)
				metadata.Symbols = append(metadata.Symbols, classInfo.Name)
			}
		}
	})
}

// parseJavaScript extracts JavaScript-specific metadata
func (cp *CodeParser) parseJavaScript(node *sitter.Node, source []byte, metadata *CodeMetadata) {
	cp.traverse(node, source, func(n *sitter.Node, src []byte) {
		switch n.Type() {
		case "import_statement":
			importStmt := n.Content(src)
			metadata.Imports = append(metadata.Imports, strings.TrimSpace(importStmt))

		case "function_declaration", "method_definition":
			funcInfo := cp.extractFunction(n, src)
			if funcInfo.Name != "" {
				metadata.Functions = append(metadata.Functions, funcInfo)
				metadata.Symbols = append(metadata.Symbols, funcInfo.Name)
			}

		case "class_declaration":
			classInfo := cp.extractClass(n, src)
			if classInfo.Name != "" {
				metadata.Classes = append(metadata.Classes, classInfo)
				metadata.Symbols = append(metadata.Symbols, classInfo.Name)
			}
		}
	})
}

// extractFunction extracts function information from a node
func (cp *CodeParser) extractFunction(node *sitter.Node, source []byte) FunctionInfo {
	name := cp.getIdentifierName(node, source)
	return FunctionInfo{
		Name:      name,
		Content:   node.Content(source),
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
	}
}

// extractClass extracts class information from a node
func (cp *CodeParser) extractClass(node *sitter.Node, source []byte) ClassInfo {
	name := cp.getIdentifierName(node, source)
	return ClassInfo{
		Name:      name,
		Content:   node.Content(source),
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
	}
}

// getIdentifierName extracts identifier name from a node
func (cp *CodeParser) getIdentifierName(node *sitter.Node, source []byte) string {
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "identifier" || child.Type() == "type_identifier" {
			return child.Content(source)
		}
	}
	return ""
}

// extractComments extracts comments from the code
func (cp *CodeParser) extractComments(node *sitter.Node, source []byte, metadata *CodeMetadata) {
	cp.traverse(node, source, func(n *sitter.Node, src []byte) {
		if strings.Contains(n.Type(), "comment") {
			comment := strings.TrimSpace(n.Content(src))
			if len(comment) > 0 {
				metadata.Comments = append(metadata.Comments, comment)
			}
		}
	})
}

// traverse performs depth-first traversal of the syntax tree
func (cp *CodeParser) traverse(node *sitter.Node, source []byte, fn func(*sitter.Node, []byte)) {
	fn(node, source)
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		cp.traverse(child, source, fn)
	}
}

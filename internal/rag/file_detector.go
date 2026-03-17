package rag

import (
	"path/filepath"
	"strings"
)

// FileType represents the category of a file
type FileType int

const (
	FileTypeUnknown FileType = iota
	FileTypeDocument
	FileTypeCode
	FileTypeConfig
)

// LanguageType represents programming languages
type LanguageType string

const (
	LangGo         LanguageType = "go"
	LangPython     LanguageType = "python"
	LangJava       LanguageType = "java"
	LangCpp        LanguageType = "cpp"
	LangJavaScript LanguageType = "javascript"
	LangXML        LanguageType = "xml"
	LangYAML       LanguageType = "yaml"
	LangJSON       LanguageType = "json"
	LangUnknown    LanguageType = "unknown"
)

// FileInfo contains detected file information
type FileInfo struct {
	Path     string
	Type     FileType
	Language LanguageType
	Ext      string
}

// Document file extensions
var documentExts = map[string]bool{
	".md":   true,
	".rst":  true,
	".pdf":  true,
	".docx": true,
	".pptx": true,
	".txt":  true,
}

// Code file extensions to language mapping
var codeExts = map[string]LanguageType{
	".go":   LangGo,
	".py":   LangPython,
	".java": LangJava,
	".cpp":  LangCpp,
	".cc":   LangCpp,
	".cxx":  LangCpp,
	".c":    LangCpp,
	".h":    LangCpp,
	".hpp":  LangCpp,
	".js":   LangJavaScript,
	".jsx":  LangJavaScript,
	".ts":   LangJavaScript, // Treat TypeScript as JavaScript for now
	".tsx":  LangJavaScript,
}

// Config file extensions
var configExts = map[string]LanguageType{
	".xml":  LangXML,
	".yaml": LangYAML,
	".yml":  LangYAML,
	".json": LangJSON,
}

// DetectFileType analyzes a file path and returns its type and language
func DetectFileType(path string) FileInfo {
	ext := strings.ToLower(filepath.Ext(path))

	info := FileInfo{
		Path:     path,
		Type:     FileTypeUnknown,
		Language: LangUnknown,
		Ext:      ext,
	}

	// Check if it's a document
	if documentExts[ext] {
		info.Type = FileTypeDocument
		return info
	}

	// Check if it's code
	if lang, ok := codeExts[ext]; ok {
		info.Type = FileTypeCode
		info.Language = lang
		return info
	}

	// Check if it's config
	if lang, ok := configExts[ext]; ok {
		info.Type = FileTypeConfig
		info.Language = lang
		return info
	}

	return info
}

// ShouldProcessFile determines if a file should be processed based on filters
func ShouldProcessFile(path string, includeTypes []string, excludePatterns []string) bool {
	// Check exclude patterns first
	for _, pattern := range excludePatterns {
		if strings.Contains(path, pattern) {
			return false
		}
	}

	// If no include types specified, process all supported files
	if len(includeTypes) == 0 {
		info := DetectFileType(path)
		return info.Type != FileTypeUnknown
	}

	// Check if file matches any include types
	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(path)), ".")
	for _, t := range includeTypes {
		if ext == strings.ToLower(t) {
			return true
		}
	}

	return false
}

// GetSupportedExtensions returns all supported file extensions
func GetSupportedExtensions() []string {
	var exts []string

	for ext := range documentExts {
		exts = append(exts, ext)
	}
	for ext := range codeExts {
		exts = append(exts, ext)
	}
	for ext := range configExts {
		exts = append(exts, ext)
	}

	return exts
}

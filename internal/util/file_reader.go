package util

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ReadMultipleLocalFiles reads multiple local files and combines their content
func ReadMultipleLocalFiles(filePaths []string) (string, error) {
	var combinedContent strings.Builder

	for i, filePath := range filePaths {
		filePath = strings.TrimSpace(filePath)
		if filePath == "" {
			continue
		}

		content, err := os.ReadFile(filePath)
		if err != nil {
			return "", fmt.Errorf("failed to read local file '%s': %v", filePath, err)
		}

		// Add separator between files
		if i > 0 {
			combinedContent.WriteString("\n\n" + strings.Repeat("=", 80) + "\n\n")
		}

		// Add file header with filename
		fileName := filepath.Base(filePath)
		combinedContent.WriteString(fmt.Sprintf("// File: %s\n", fileName))
		combinedContent.WriteString(fmt.Sprintf("// Path: %s\n\n", filePath))
		combinedContent.WriteString(string(content))
	}

	return combinedContent.String(), nil
}

// ReadMultipleRemoteFiles reads multiple remote files and combines their content
func ReadMultipleRemoteFiles(fileUrls []string) (string, error) {
	var combinedContent strings.Builder

	for i, fileUrl := range fileUrls {
		fileUrl = strings.TrimSpace(fileUrl)
		if fileUrl == "" {
			continue
		}

		content, err := GetRemoteFileContent(fileUrl)
		if err != nil {
			return "", fmt.Errorf("failed to read remote file '%s': %v", fileUrl, err)
		}

		// Add separator between files
		if i > 0 {
			combinedContent.WriteString("\n\n" + strings.Repeat("=", 80) + "\n\n")
		}

		// Add file header with URL
		fileName := filepath.Base(fileUrl)
		combinedContent.WriteString(fmt.Sprintf("// File: %s\n", fileName))
		combinedContent.WriteString(fmt.Sprintf("// URL: %s\n\n", fileUrl))
		combinedContent.WriteString(content)
	}

	return combinedContent.String(), nil
}

// ReadMultipleGitLabFiles reads multiple GitLab files and combines their content
func ReadMultipleGitLabFiles(gitlabUrl, gitlabProject, gitlabBranch, privateToken string, filePaths []string) (string, error) {
	var combinedContent strings.Builder

	for i, filePath := range filePaths {
		filePath = strings.TrimSpace(filePath)
		if filePath == "" {
			continue
		}

		content, err := GetGitLabFileContent(gitlabUrl, gitlabProject, filePath, gitlabBranch, privateToken)
		if err != nil {
			return "", fmt.Errorf("failed to read GitLab file '%s': %v", filePath, err)
		}

		// Add separator between files
		if i > 0 {
			combinedContent.WriteString("\n\n" + strings.Repeat("=", 80) + "\n\n")
		}

		// Add file header with GitLab info
		fileName := filepath.Base(filePath)
		combinedContent.WriteString(fmt.Sprintf("// File: %s\n", fileName))
		combinedContent.WriteString(fmt.Sprintf("// GitLab Path: %s\n", filePath))
		combinedContent.WriteString(fmt.Sprintf("// Project: %s\n", gitlabProject))
		combinedContent.WriteString(fmt.Sprintf("// Branch: %s\n\n", gitlabBranch))
		combinedContent.WriteString(content)
	}

	return combinedContent.String(), nil
}


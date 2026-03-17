package util

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/walterfan/lazy-ai-coder/internal/log"
)

type ChangeItem struct {
	Diff          string `json:"diff"`
	NewPath       string `json:"new_path"`
	OldPath       string `json:"old_path"`
	AMode         string `json:"a_mode"`
	BMode         string `json:"b_mode"`
	NewFile       bool   `json:"new_file"`
	RenamedFile   bool   `json:"renamed_file"`
	DeletedFile   bool   `json:"deleted_file"`
	GeneratedFile bool   `json:"generated_file"`
}

type ChangeStatistics struct {
	ChangedLines int
	AddedLines   int
	DeletedLines int
}

type MergeRequestInfo struct {
	Title       string       `json:"title"`
	Description string       `json:"description"`
	Changes     []ChangeItem `json:"changes"`
}

type FileSummary struct {
	Index        int    `json:"index"`
	OldPath      string `json:"old_path"`
	NewPath      string `json:"new_path"`
	Added        bool   `json:"added"`
	Deleted      bool   `json:"deleted"`
	Renamed      bool   `json:"renamed"`
	Changed      bool   `json:"changed"`
	ChangedLines int    `json:"changed_lines"`
	AddedLines   int    `json:"added_lines"`
	DeletedLines int    `json:"deleted_lines"`
}

type MRSummary struct {
	Title       string        `json:"title"`
	Description string        `json:"description"`
	Files       []FileSummary `json:"files"`
	TotalFiles  int           `json:"total_files"`
}

// parseDiffStats parses git diff to extract added and deleted line counts
func parseDiffStats(diff string) ChangeStatistics {
	stats := ChangeStatistics{}

	// Split diff into lines
	lines := strings.Split(diff, "\n")

	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		// Lines starting with + are additions (but not +++ which is file header)
		if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			stats.AddedLines++
		}
		// Lines starting with - are deletions (but not --- which is file header)
		if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
			stats.DeletedLines++
		}
	}

	// Changed lines is the minimum of added and deleted (lines that were modified)
	stats.ChangedLines = stats.AddedLines + stats.DeletedLines

	return stats
}

func formatBoolAsEmoji(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}

// GetProjectIDByName gets the numeric project ID from GitLab using the project path (e.g., "namespace/project-name")
func GetProjectIDByName(gitlabURL, projectName, privateToken string) (string, error) {
	logger := log.GetLogger()

	// Construct the search API URL
	searchURL := fmt.Sprintf("%s/api/v4/projects?search=%s", gitlabURL, url.QueryEscape(projectName))

	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("PRIVATE-TOKEN", privateToken)

	httpClient, err := createHttpClient()
	if err != nil {
		return "", err
	}

	logger.Infof("Fetching project ID from: %s", searchURL)
	resp, err := httpClient.Do(req)
	if err != nil {
		logger.Errorf("Error fetching project ID: %v", err)
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("GitLab project search failed: %s - %s", resp.Status, string(body))
	}

	var projects []struct {
		ID                int    `json:"id"`
		Name              string `json:"name"`
		PathWithNamespace string `json:"path_with_namespace"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&projects); err != nil {
		return "", err
	}

	if len(projects) == 0 {
		return "", fmt.Errorf("no project found matching name: %s", projectName)
	}

	for _, p := range projects {
		if p.PathWithNamespace == projectName {
			return fmt.Sprintf("%d", p.ID), nil
		}
	}

	// If not exact match, return first result as fallback
	logger.Warnf("No exact match for project %s, returning first match: %s", projectName, projects[0].PathWithNamespace)
	return fmt.Sprintf("%d", projects[0].ID), nil
}

func ChangeItemsToMarkdown(mrInfo MergeRequestInfo) string {
	if len(mrInfo.Changes) == 0 {
		return "No changes found."
	}

	var markdown strings.Builder

	markdown.WriteString("## Merge Request\n\n")
	markdown.WriteString(fmt.Sprintf("* Title: `%s`\n", mrInfo.Title))
	markdown.WriteString(fmt.Sprintf("* Description: `%s`\n", mrInfo.Description))
	markdown.WriteString("\n### changes\n\n")
	for _, change := range mrInfo.Changes {

		markdown.WriteString(fmt.Sprintf("* Old Path: `%s`\n", change.OldPath))
		markdown.WriteString(fmt.Sprintf("* New Path: `%s`\n", change.NewPath))
		markdown.WriteString(fmt.Sprintf("* Added: %s\n", formatBoolAsEmoji(change.NewFile)))
		markdown.WriteString(fmt.Sprintf("* Renamed: %s\n", formatBoolAsEmoji(change.RenamedFile)))
		markdown.WriteString(fmt.Sprintf("* Deleted: %s\n", formatBoolAsEmoji(change.DeletedFile)))

		markdown.WriteString("\n* Code Diff:\n")
		markdown.WriteString("```\n")
		markdown.WriteString(change.Diff)
		markdown.WriteString("\n```\n")
		markdown.WriteString("\n------\n\n")
	}

	return markdown.String()
}

// ConvertToMRSummary converts MergeRequestInfo to MRSummary with statistics
func ConvertToMRSummary(mrInfo MergeRequestInfo) MRSummary {
	summary := MRSummary{
		Title:       mrInfo.Title,
		Description: mrInfo.Description,
		Files:       make([]FileSummary, 0, len(mrInfo.Changes)),
		TotalFiles:  len(mrInfo.Changes),
	}

	for i, change := range mrInfo.Changes {
		stats := parseDiffStats(change.Diff)

		// Determine if file was changed (not just added/deleted/renamed)
		changed := !change.NewFile && !change.DeletedFile && stats.ChangedLines > 0

		fileSummary := FileSummary{
			Index:        i + 1,
			OldPath:      change.OldPath,
			NewPath:      change.NewPath,
			Added:        change.NewFile,
			Deleted:      change.DeletedFile,
			Renamed:      change.RenamedFile,
			Changed:      changed,
			ChangedLines: stats.ChangedLines,
			AddedLines:   stats.AddedLines,
			DeletedLines: stats.DeletedLines,
		}

		summary.Files = append(summary.Files, fileSummary)
	}

	return summary
}

// MRSummaryToMarkdownTable converts MRSummary to markdown table format
func MRSummaryToMarkdownTable(summary MRSummary) string {
	var markdown strings.Builder

	markdown.WriteString("## Merge Request Summary\n\n")
	markdown.WriteString(fmt.Sprintf("**Title:** %s\n\n", summary.Title))
	if summary.Description != "" {
		markdown.WriteString(fmt.Sprintf("**Description:** %s\n\n", summary.Description))
	}
	markdown.WriteString(fmt.Sprintf("**Total Files Changed:** %d\n\n", summary.TotalFiles))

	// Table header
	markdown.WriteString("| # | Old Path | New Path | Added | Deleted | Renamed | Changed | Changed Lines | Added Lines | Deleted Lines |\n")
	markdown.WriteString("|---|----------|----------|-------|---------|---------|---------|---------------|-------------|---------------|\n")

	// Table rows
	for _, file := range summary.Files {
		markdown.WriteString(fmt.Sprintf("| %d | %s | %s | %s | %s | %s | %s | %d | %d | %d |\n",
			file.Index,
			file.OldPath,
			file.NewPath,
			formatBoolAsYN(file.Added),
			formatBoolAsYN(file.Deleted),
			formatBoolAsYN(file.Renamed),
			formatBoolAsYN(file.Changed),
			file.ChangedLines,
			file.AddedLines,
			file.DeletedLines,
		))
	}

	return markdown.String()
}

// formatBoolAsYN formats boolean as Y/N for table display
func formatBoolAsYN(b bool) string {
	if b {
		return "Y"
	}
	return "N"
}

func GetMergeRequestChange(gitlabURL, projectID, mergeRequestID, privateToken string) (string, error) {
	mrInfo, err := GetMergeRequestInfo(gitlabURL, projectID, mergeRequestID, privateToken)
	if err != nil {
		return "", err
	}
	return ChangeItemsToMarkdown(mrInfo), nil
}

// GetMergeRequestInfo fetches merge request information from GitLab API
func GetMergeRequestInfo(gitlabURL, projectID, mergeRequestID, privateToken string) (MergeRequestInfo, error) {
	logger := log.GetLogger()
	apiURL := fmt.Sprintf("%s/api/v4/projects/%s/merge_requests/%s/changes",
		gitlabURL,
		url.PathEscape(projectID), // Handles namespace or numeric ID
		url.PathEscape(mergeRequestID))
	// Create the request
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return MergeRequestInfo{}, err
	}
	req.Header.Set("PRIVATE-TOKEN", privateToken)
	// Send request
	httpClient, err := createHttpClient()
	if err != nil {
		return MergeRequestInfo{}, err
	}
	logger.Infof("sending request to %s", apiURL)
	resp, err := httpClient.Do(req)
	if err != nil {
		logger.Error("Error sending request to %s:", apiURL, err)
		return MergeRequestInfo{}, err
	}
	defer resp.Body.Close()
	logger.Infof("get response %d:", resp.StatusCode)
	// Check for successful response
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		// Log token info for debugging (only first/last 4 chars for security)
		tokenInfo := "empty"
		if len(privateToken) > 8 {
			tokenInfo = privateToken[:4] + "..." + privateToken[len(privateToken)-4:]
		} else if len(privateToken) > 0 {
			tokenInfo = "too_short"
		}
		logger.Warnf("GitLab API error: status=%s, token_hint=%s, response=%s", resp.Status, tokenInfo, string(body))
		return MergeRequestInfo{}, fmt.Errorf("GitLab API error: %s - %s", resp.Status, string(body))
	}

	// Parse response JSON
	var result MergeRequestInfo
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return MergeRequestInfo{}, err
	}

	return result, nil
}

func GetGitLabFileContent(gitlabURL, projectID, filePath, branch, privateToken string) (string, error) {
	// Construct the API URL
	apiURL := fmt.Sprintf("%s/api/v4/projects/%s/repository/files/%s?ref=%s",
		gitlabURL,
		url.PathEscape(projectID), // Handles namespace or numeric ID
		url.PathEscape(filePath),
		url.PathEscape(branch),
	)

	// Create the request
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("PRIVATE-TOKEN", privateToken)

	// Send request
	httpClient, err := createHttpClient()
	if err != nil {
		return "", err
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Check for successful response
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("GitLab API error: %s - %s", resp.Status, string(body))
	}

	// Parse response JSON
	var result struct {
		Content  string `json:"content"`
		Encoding string `json:"encoding"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	// Decode base64 content if needed
	if result.Encoding == "base64" {
		decoded, err := base64.StdEncoding.DecodeString(result.Content)
		if err != nil {
			return "", err
		}
		return string(decoded), nil
	}

	return result.Content, nil
}

// PostMergeRequestComment posts a comment to a GitLab merge request
func PostMergeRequestComment(gitlabURL, projectID, mergeRequestID, comment, privateToken string) error {
	logger := log.GetLogger()

	// Construct the API URL for posting a note (comment)
	apiURL := fmt.Sprintf("%s/api/v4/projects/%s/merge_requests/%s/notes",
		gitlabURL,
		url.PathEscape(projectID),
		url.PathEscape(mergeRequestID))

	// Prepare the request body
	requestBody := map[string]string{
		"body": comment,
	}
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create the request
	req, err := http.NewRequest("POST", apiURL, strings.NewReader(string(jsonData)))
	if err != nil {
		return err
	}
	req.Header.Set("PRIVATE-TOKEN", privateToken)
	req.Header.Set("Content-Type", "application/json")

	// Send request
	httpClient, err := createHttpClient()
	if err != nil {
		return err
	}

	logger.Infof("Posting comment to MR: %s", apiURL)
	resp, err := httpClient.Do(req)
	if err != nil {
		logger.Errorf("Error posting comment: %v", err)
		return err
	}
	defer resp.Body.Close()

	// Check for successful response
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GitLab API error: %s - %s", resp.Status, string(body))
	}

	logger.Infof("Comment posted successfully to MR %s", mergeRequestID)
	return nil
}

package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
)

var (
	serverURL string
	sessionID string
	format    string
)

var debugCmd = &cobra.Command{
	Use:   "debug",
	Short: "Debug and monitor the LLM agent",
	Long:  `Debug commands to monitor health, view metrics, and manage memory sessions`,
}

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Check system health",
	Long:  `Check the health status of the LLM agent and its components`,
	Run: func(cmd *cobra.Command, args []string) {
		detailed, _ := cmd.Flags().GetBool("detailed")

		endpoint := fmt.Sprintf("%s/health", serverURL)
		if detailed {
			endpoint = fmt.Sprintf("%s/health/detailed", serverURL)
		}

		resp, err := http.Get(endpoint)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)

		if format == "json" {
			fmt.Println(string(body))
		} else {
			var health map[string]interface{}
			json.Unmarshal(body, &health)

			fmt.Printf("Status: %s\n", health["status"])
			fmt.Printf("Timestamp: %s\n", health["timestamp"])

			if components, ok := health["components"].([]interface{}); ok {
				fmt.Println("\nComponents:")
				for _, comp := range components {
					c := comp.(map[string]interface{})
					fmt.Printf("  - %s: %s - %s\n", c["name"], c["status"], c["message"])
				}
			}
		}
	},
}

var sessionsCmd = &cobra.Command{
	Use:   "sessions",
	Short: "Manage memory sessions",
	Long:  `List, inspect, and manage memory sessions`,
}

var sessionsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all memory sessions",
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := http.Get(fmt.Sprintf("%s/api/v1/debug/memory/sessions", serverURL))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)

		if format == "json" {
			fmt.Println(string(body))
		} else {
			var result map[string]interface{}
			json.Unmarshal(body, &result)

			fmt.Printf("Total Sessions: %v\n\n", result["total_sessions"])

			if sessions, ok := result["sessions"].([]interface{}); ok && len(sessions) > 0 {
				w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
				fmt.Fprintln(w, "SESSION ID\tMESSAGES\tTOKENS\tAGE\tIDLE")
				for _, sess := range sessions {
					s := sess.(map[string]interface{})
					fmt.Fprintf(w, "%s\t%v\t%v\t%s\t%s\n",
						s["session_id"],
						s["message_count"],
						s["total_tokens"],
						s["age"],
						s["idle_time"],
					)
				}
				w.Flush()
			} else {
				fmt.Println("No active sessions")
			}
		}
	},
}

var sessionsGetCmd = &cobra.Command{
	Use:   "get <session_id>",
	Short: "Get detailed information about a session",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		sessionID := args[0]

		resp, err := http.Get(fmt.Sprintf("%s/api/v1/debug/memory/sessions/%s", serverURL, sessionID))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			fmt.Fprintf(os.Stderr, "Session not found: %s\n", sessionID)
			os.Exit(1)
		}

		body, _ := io.ReadAll(resp.Body)

		if format == "json" {
			fmt.Println(string(body))
		} else {
			var session map[string]interface{}
			json.Unmarshal(body, &session)

			fmt.Printf("Session ID: %s\n", session["session_id"])
			fmt.Printf("Messages: %v\n", session["message_count"])
			fmt.Printf("Tokens: %v\n", session["total_tokens"])
			fmt.Printf("Age: %s\n", session["age"])
			fmt.Printf("Idle Time: %s\n\n", session["idle_time"])

			if messages, ok := session["messages"].([]interface{}); ok {
				fmt.Println("Messages:")
				for i, msg := range messages {
					m := msg.(map[string]interface{})
					fmt.Printf("\n[%d] %s (tokens: %v):\n", i+1, m["role"], m["tokens"])
					fmt.Printf("%s\n", m["content"])
				}
			}
		}
	},
}

var sessionsDeleteCmd = &cobra.Command{
	Use:   "delete <session_id>",
	Short: "Delete a memory session",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		sessionID := args[0]

		req, _ := http.NewRequest("DELETE", fmt.Sprintf("%s/api/v1/debug/memory/sessions/%s", serverURL, sessionID), nil)
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			fmt.Fprintf(os.Stderr, "Failed to delete session: %s\n", sessionID)
			os.Exit(1)
		}

		fmt.Printf("Session deleted: %s\n", sessionID)
	},
}

var sessionsSummarizeCmd = &cobra.Command{
	Use:   "summarize <session_id>",
	Short: "Force summarization of a session",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		sessionID := args[0]

		resp, err := http.Post(fmt.Sprintf("%s/api/v1/debug/memory/sessions/%s/summarize", serverURL, sessionID), "application/json", nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)

		if format == "json" {
			fmt.Println(string(body))
		} else {
			var result map[string]interface{}
			json.Unmarshal(body, &result)

			fmt.Printf("Session: %s\n", result["session_id"])

			if before, ok := result["before"].(map[string]interface{}); ok {
				fmt.Printf("Before: %v messages, %v tokens\n", before["messages"], before["tokens"])
			}

			if after, ok := result["after"].(map[string]interface{}); ok {
				fmt.Printf("After: %v messages, %v tokens\n", after["messages"], after["tokens"])
			}

			if reduction, ok := result["reduction"].(map[string]interface{}); ok {
				fmt.Printf("Reduction: %v messages, %v tokens\n", reduction["messages"], reduction["tokens"])
			}
		}
	},
}

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Get memory statistics",
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := http.Get(fmt.Sprintf("%s/api/v1/debug/memory/stats", serverURL))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)

		if format == "json" {
			fmt.Println(string(body))
		} else {
			var stats map[string]interface{}
			json.Unmarshal(body, &stats)

			fmt.Printf("Total Sessions: %v\n", stats["total_sessions"])
			fmt.Printf("Total Messages: %v\n", stats["total_messages"])
			fmt.Printf("Total Tokens: %v\n\n", stats["total_tokens"])

			if config, ok := stats["configuration"].(map[string]interface{}); ok {
				fmt.Println("Configuration:")
				fmt.Printf("  Max Tokens: %v\n", config["max_tokens"])
				fmt.Printf("  Max Messages: %v\n", config["max_messages"])
				fmt.Printf("  Summary Tokens: %v\n", config["summary_tokens"])
				fmt.Printf("  Session Timeout: %v\n\n", config["session_timeout"])
			}

			if averages, ok := stats["averages"].(map[string]interface{}); ok {
				fmt.Println("Averages:")
				fmt.Printf("  Messages per Session: %.1f\n", averages["messages_per_session"])
				fmt.Printf("  Tokens per Session: %.0f\n\n", averages["tokens_per_session"])
			}

			if dist, ok := stats["distribution"].(map[string]interface{}); ok {
				if byAge, ok := dist["by_age"].(map[string]interface{}); ok && len(byAge) > 0 {
					fmt.Println("Sessions by Age:")
					for age, count := range byAge {
						fmt.Printf("  %s: %v\n", age, count)
					}
				}
			}
		}
	},
}

var cleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Cleanup expired sessions",
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := http.Post(fmt.Sprintf("%s/api/v1/debug/memory/cleanup", serverURL), "application/json", nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)

		if format == "json" {
			fmt.Println(string(body))
		} else {
			var result map[string]interface{}
			json.Unmarshal(body, &result)

			fmt.Printf("Sessions Before: %v\n", result["sessions_before"])
			fmt.Printf("Sessions After: %v\n", result["sessions_after"])
			fmt.Printf("Sessions Removed: %v\n", result["sessions_removed"])
			fmt.Printf("Timeout: %v\n", result["timeout"])
		}
	},
}

var metricsCmd = &cobra.Command{
	Use:   "metrics",
	Short: "View Prometheus metrics",
	Run: func(cmd *cobra.Command, args []string) {
		filter, _ := cmd.Flags().GetString("filter")

		resp, err := http.Get(fmt.Sprintf("%s/metrics", serverURL))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)

		if filter != "" {
			// Simple filter by substring
			lines := string(body)
			for _, line := range splitLines(lines) {
				if contains(line, filter) {
					fmt.Println(line)
				}
			}
		} else {
			fmt.Println(string(body))
		}
	},
}

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch memory statistics in real-time",
	Run: func(cmd *cobra.Command, args []string) {
		interval, _ := cmd.Flags().GetInt("interval")

		ticker := time.NewTicker(time.Duration(interval) * time.Second)
		defer ticker.Stop()

		// Print header
		fmt.Println("Watching memory statistics (Ctrl+C to stop)...")
		fmt.Println()

		printStats := func() {
			resp, err := http.Get(fmt.Sprintf("%s/api/v1/debug/memory/stats", serverURL))
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				return
			}
			defer resp.Body.Close()

			body, _ := io.ReadAll(resp.Body)
			var stats map[string]interface{}
			json.Unmarshal(body, &stats)

			fmt.Printf("\r[%s] Sessions: %v | Messages: %v | Tokens: %v",
				time.Now().Format("15:04:05"),
				stats["total_sessions"],
				stats["total_messages"],
				stats["total_tokens"],
			)
		}

		// Print initial stats
		printStats()

		// Watch for updates
		for range ticker.C {
			printStats()
		}
	},
}

// Helper functions
func splitLines(s string) []string {
	lines := []string{}
	current := ""
	for _, char := range s {
		if char == '\n' {
			lines = append(lines, current)
			current = ""
		} else {
			current += string(char)
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) && (s[:len(substr)] == substr ||
		contains(s[1:], substr))))
}

func init() {
	// Add debug command to root
	rootCmd.AddCommand(debugCmd)

	// Add subcommands
	debugCmd.AddCommand(healthCmd)
	debugCmd.AddCommand(sessionsCmd)
	debugCmd.AddCommand(statsCmd)
	debugCmd.AddCommand(cleanupCmd)
	debugCmd.AddCommand(metricsCmd)
	debugCmd.AddCommand(watchCmd)

	// Sessions subcommands
	sessionsCmd.AddCommand(sessionsListCmd)
	sessionsCmd.AddCommand(sessionsGetCmd)
	sessionsCmd.AddCommand(sessionsDeleteCmd)
	sessionsCmd.AddCommand(sessionsSummarizeCmd)

	// Global flags
	debugCmd.PersistentFlags().StringVarP(&serverURL, "server", "s", "http://localhost:8888", "Server URL")
	debugCmd.PersistentFlags().StringVarP(&format, "format", "f", "table", "Output format (table or json)")

	// Health command flags
	healthCmd.Flags().BoolP("detailed", "d", false, "Show detailed health information")

	// Metrics command flags
	metricsCmd.Flags().StringP("filter", "F", "", "Filter metrics by substring")

	// Watch command flags
	watchCmd.Flags().IntP("interval", "i", 5, "Refresh interval in seconds")
}

package hook

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const (
	appName        = "Claude Code"
	claudeSender   = "com.anthropic.claudefordesktop"
)

// SendNotification sends a macOS notification.
func SendNotification(title, subtitle, message string, sound string, group string) error {
	if hasTerminalNotifier() {
		return sendWithTerminalNotifier(title, subtitle, message, sound, group)
	}
	return sendWithOSAScript(title, subtitle, message, sound)
}

func hasTerminalNotifier() bool {
	_, err := exec.LookPath("terminal-notifier")
	return err == nil
}

func sendWithTerminalNotifier(title, subtitle, message, sound, group string) error {
	args := []string{
		"-title", title,
		"-subtitle", subtitle,
		"-message", message,
		"-sender", claudeSender,
	}
	if sound != "" {
		args = append(args, "-sound", sound)
	}
	if group != "" {
		args = append(args, "-group", group)
	}

	cmd := exec.Command("terminal-notifier", args...)
	return cmd.Run()
}

func sendWithOSAScript(title, subtitle, message, sound string) error {
	soundClause := ""
	if sound != "" {
		soundClause = fmt.Sprintf(` sound name "%s"`, sound)
	}
	script := fmt.Sprintf(`display notification "%s" with title "%s" subtitle "%s"%s`, message, title, subtitle, soundClause)
	cmd := exec.Command("osascript", "-e", script)
	return cmd.Run()
}

// NotifyInput sends notification when Claude needs input.
func NotifyInput(message string) error {
	projectName := getProjectName()
	if message == "" {
		message = "Claude needs your input to continue"
	}
	return SendNotification(
		fmt.Sprintf("🔔 %s", appName),
		projectName,
		message,
		"",
		fmt.Sprintf("claude-code-%s", projectName),
	)
}

// NotifyStop sends notification when Claude completes a task.
func NotifyStop(transcriptPath string) error {
	projectName := getProjectName()
	timeStr := getTime()

	message := fmt.Sprintf("Completed at %s", timeStr)
	if transcriptPath != "" {
		if finalResponse := extractFinalResponse(transcriptPath, 200); finalResponse != "" {
			message = finalResponse
		}
	}

	return SendNotification(
		fmt.Sprintf("✅ %s", appName),
		projectName,
		message,
		"",
		fmt.Sprintf("claude-code-%s", projectName),
	)
}

// NotifyError sends error notification.
func NotifyError(message string) error {
	projectName := getProjectName()
	if message == "" {
		message = "An error occurred during task execution"
	}
	return SendNotification(
		fmt.Sprintf("❌ %s", appName),
		projectName,
		message,
		"Basso",
		fmt.Sprintf("claude-code-%s", projectName),
	)
}

// NotifyTest sends a test notification.
func NotifyTest() error {
	projectName := getProjectName()
	hasNotifier := hasTerminalNotifier()

	if err := SendNotification(
		fmt.Sprintf("🧪 %s", appName),
		projectName,
		"Notifications are working!",
		"",
		fmt.Sprintf("claude-code-%s", projectName),
	); err != nil {
		return err
	}

	if hasNotifier {
		fmt.Println("✓ Test notification sent (using terminal-notifier)")
	} else {
		fmt.Println("✓ Test notification sent (using osascript fallback)")
		fmt.Println("\n  Tip: Install terminal-notifier for more reliable notifications:")
		fmt.Println("       brew install terminal-notifier")
	}

	return nil
}

func getProjectName() string {
	cwd, err := os.Getwd()
	if err != nil {
		return "unknown"
	}
	return filepath.Base(cwd)
}

func getTime() string {
	return time.Now().Format("3:04 PM")
}

func extractFinalResponse(transcriptPath string, maxLength int) string {
	path := filepath.Clean(transcriptPath)
	if !filepath.IsAbs(path) {
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		path = filepath.Join(home, path)
	}

	file, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer file.Close()

	var lastResponse string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var entry map[string]interface{}
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue
		}

		if entry["type"] == "assistant" {
			if message, ok := entry["message"].(map[string]interface{}); ok {
				if content, ok := message["content"].([]interface{}); ok {
					for _, block := range content {
						if blockMap, ok := block.(map[string]interface{}); ok {
							if blockMap["type"] == "text" {
								if text, ok := blockMap["text"].(string); ok {
									lastResponse = text
								}
							}
						} else if text, ok := block.(string); ok {
							lastResponse = text
						}
					}
				}
			}
		}
	}

	if lastResponse == "" {
		return ""
	}

	// Truncate and clean up for notification
	text := strings.TrimSpace(lastResponse)
	firstLine := strings.Split(text, "\n")[0]

	// Strip markdown formatting
	re := regexp.MustCompile(`\*\*(.+?)\*\*`)
	firstLine = re.ReplaceAllString(firstLine, "$1")
	re = regexp.MustCompile(`\*(.+?)\*`)
	firstLine = re.ReplaceAllString(firstLine, "$1")
	re = regexp.MustCompile("`(.+?)`")
	firstLine = re.ReplaceAllString(firstLine, "$1")
	re = regexp.MustCompile(`^#+\s*`)
	firstLine = re.ReplaceAllString(firstLine, "")

	if len(firstLine) > maxLength {
		return firstLine[:maxLength-3] + "..."
	}
	return firstLine
}

package hook

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// HookInput represents the JSON input from Claude Code hooks.
type HookInput struct {
	SessionID     string                 `json:"session_id"`
	ToolInput     map[string]interface{} `json:"tool_input"`
	TranscriptPath string                `json:"transcript_path"`
	Message       string                 `json:"message"`
	NotificationType string             `json:"notification_type"`
}

// GetStdinData reads stdin JSON data from hooks.
func GetStdinData() (*HookInput, error) {
	// Check if stdin is a TTY (interactive)
	if isTTY(os.Stdin) {
		return nil, nil
	}

	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, nil
	}

	var input HookInput
	if err := json.Unmarshal(data, &input); err != nil {
		return nil, err
	}

	return &input, nil
}

// GetFilePath extracts file_path from hook input (tool_input.file_path).
func GetFilePath(input *HookInput) string {
	if input == nil || input.ToolInput == nil {
		return ""
	}
	if path, ok := input.ToolInput["file_path"].(string); ok {
		return path
	}
	return ""
}

// GetTranscriptPath extracts transcript_path from hook input.
func GetTranscriptPath(input *HookInput) string {
	if input == nil {
		return ""
	}
	return input.TranscriptPath
}

// IsSubagent checks if this is a subagent based on transcript path.
func IsSubagent(transcriptPath string) bool {
	if transcriptPath == "" {
		return false
	}
	filename := filepath.Base(transcriptPath)
	return strings.HasPrefix(filename, "agent-")
}

func isTTY(f *os.File) bool {
	stat, err := f.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) != 0
}

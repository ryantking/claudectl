package hook

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ContextInfo generates context information for injection into prompts.
func ContextInfo() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	cwd, err = filepath.Abs(cwd)
	if err != nil {
		return "", err
	}

	var lines []string
	lines = append(lines, "<context-refresh>")
	lines = append(lines, fmt.Sprintf("Path: %s", cwd))

	// Current workspace (if in one)
	currentWS := getCurrentWorkspace(cwd)
	if currentWS != nil {
		wsLine := fmt.Sprintf("Current Workspace: %s", currentWS["branch"])
		if status, ok := currentWS["status"].(string); ok && status != "" {
			wsLine += fmt.Sprintf(" (%s)", status)
		}
		lines = append(lines, wsLine)
	}

	// Git info (current branch)
	branch := getGitBranch(cwd)
	if branch != "" {
		status := getGitStatusSummary(cwd)
		gitLine := fmt.Sprintf("Branch: %s", branch)
		if status != "" {
			gitLine += fmt.Sprintf(" (%s)", status)
		}
		lines = append(lines, gitLine)
	}

	// All git branches with cleanliness
	allBranches := getAllGitBranches(cwd)
	if len(allBranches) > 0 {
		lines = append(lines, "Git Branches:")
		for branchName, branchStatus := range allBranches {
			lines = append(lines, fmt.Sprintf("  %s: %s", branchName, branchStatus))
		}
	}

	// Workspace sessions
	workspaces := getWorkspaceSessions()
	if len(workspaces) > 0 {
		lines = append(lines, "Workspaces:")
		for _, ws := range workspaces {
			wsLine := fmt.Sprintf("  %v", ws["branch"])
			if status, ok := ws["status"].(string); ok && status != "" {
				wsLine += fmt.Sprintf(" (%s)", status)
			}
			lines = append(lines, wsLine)
		}
	}

	// PR status
	pr := getPRStatus(cwd)
	if pr != nil {
		prLine := fmt.Sprintf("PR #%v: %v", pr["number"], pr["title"])
		var details []string
		if review, ok := pr["review"].(string); ok && review != "" {
			details = append(details, strings.ToLower(strings.ReplaceAll(review, "_", " ")))
		}
		if checks, ok := pr["checks"].(string); ok && checks != "" {
			details = append(details, fmt.Sprintf("checks: %s", checks))
		}
		if len(details) > 0 {
			prLine += fmt.Sprintf(" (%s)", strings.Join(details, ", "))
		}
		lines = append(lines, prLine)
	}

	// Directory snapshot
	files := getDirectorySnapshot(cwd, 15)
	if len(files) > 0 {
		lines = append(lines, fmt.Sprintf("Directory: %s/", filepath.Base(cwd)))
		if len(files) <= 10 {
			lines = append(lines, fmt.Sprintf("  %s", strings.Join(files, ", ")))
		} else {
			lines = append(lines, fmt.Sprintf("  %s", strings.Join(files[:10], ", ")))
			lines = append(lines, fmt.Sprintf("  ... and %d more", len(files)-10))
		}
	}

	lines = append(lines, "</context-refresh>")
	return strings.Join(lines, "\n"), nil
}

func getGitBranch(repoRoot string) string {
	cmd := exec.Command("git", "-C", repoRoot, "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	branch := strings.TrimSpace(string(output))
	if branch == "HEAD" {
		return ""
	}
	return branch
}

func getGitStatusSummary(repoRoot string) string {
	cmd := exec.Command("git", "-C", repoRoot, "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return "clean"
	}

	var staged, modified, untracked int
	for _, line := range lines {
		if line == "" {
			continue
		}
		if len(line) < 2 {
			continue
		}
		if line[0] != ' ' && line[0] != '?' {
			staged++
		}
		if line[1] != ' ' && line[1] != '?' {
			modified++
		}
		if strings.HasPrefix(line, "??") {
			untracked++
		}
	}

	var parts []string
	if staged > 0 {
		parts = append(parts, fmt.Sprintf("%d staged", staged))
	}
	if modified > 0 {
		parts = append(parts, fmt.Sprintf("%d modified", modified))
	}
	if untracked > 0 {
		parts = append(parts, fmt.Sprintf("%d untracked", untracked))
	}

	if len(parts) == 0 {
		return "clean"
	}
	return strings.Join(parts, ", ")
}

func getAllGitBranches(repoRoot string) map[string]string {
	branches := make(map[string]string)
	cmd := exec.Command("git", "-C", repoRoot, "branch", "--list")
	output, err := cmd.Output()
	if err != nil {
		return branches
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "*") {
			line = strings.TrimPrefix(line, "*")
		}
		branch := strings.TrimSpace(line)
		if branch == "" {
			continue
		}
		// Simplified: mark as unknown (checking cleanliness is expensive)
		branches[branch] = "unknown"
	}
	return branches
}

func getWorkspaceSessions() []map[string]interface{} {
	cmd := exec.Command("agentctl", "workspace", "list", "--json")
	cmd.Env = os.Environ()
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	var workspaces []map[string]interface{}
	if err := json.Unmarshal(output, &workspaces); err != nil {
		return nil
	}
	return workspaces
}

func getCurrentWorkspace(cwd string) map[string]interface{} {
	workspaces := getWorkspaceSessions()
	if len(workspaces) == 0 {
		return nil
	}

	for _, ws := range workspaces {
		if path, ok := ws["path"].(string); ok {
			if strings.HasPrefix(cwd, path) {
				return ws
			}
		}
	}
	return nil
}

func getPRStatus(repoRoot string) map[string]interface{} {
	cmd := exec.Command("gh", "pr", "view", "--json", "number,title,state,url,reviewDecision,statusCheckRollup")
	cmd.Dir = repoRoot
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	var prData map[string]interface{}
	if err := json.Unmarshal(output, &prData); err != nil {
		return nil
	}

	if state, ok := prData["state"].(string); !ok || state != "OPEN" {
		return nil
	}

	result := map[string]interface{}{
		"number": prData["number"],
		"title":  prData["title"],
		"url":     prData["url"],
	}

	if review, ok := prData["reviewDecision"].(string); ok {
		result["review"] = review
	}

	// Summarize check status
	if checks, ok := prData["statusCheckRollup"].([]interface{}); ok {
		var passed, failed, pending int
		for _, check := range checks {
			if checkMap, ok := check.(map[string]interface{}); ok {
				if conclusion, ok := checkMap["conclusion"].(string); ok {
					if conclusion == "SUCCESS" {
						passed++
					} else if conclusion == "FAILURE" {
						failed++
					}
				} else if status, ok := checkMap["status"].(string); ok && status == "IN_PROGRESS" {
					pending++
				}
			}
		}
		if failed > 0 {
			result["checks"] = fmt.Sprintf("%d failing", failed)
		} else if pending > 0 {
			result["checks"] = fmt.Sprintf("%d pending", pending)
		} else if passed > 0 {
			result["checks"] = fmt.Sprintf("%d passed", passed)
		}
	}

	return result
}

func getDirectorySnapshot(cwd string, maxFiles int) []string {
	var files []string
	priorityPatterns := []string{
		"*.go", "*.py", "*.ts", "*.tsx", "*.js", "*.jsx", "*.rs", "*.rb", "*.java",
		"Makefile", "justfile", "package.json", "pyproject.toml", "Cargo.toml", "go.mod", "Gemfile",
	}

	seen := make(map[string]bool)
	for _, pattern := range priorityPatterns {
		matches, err := filepath.Glob(filepath.Join(cwd, pattern))
		if err != nil {
			continue
		}
		for _, match := range matches {
			if info, err := os.Stat(match); err != nil || info.IsDir() {
				continue
			}
			name := filepath.Base(match)
			if !seen[name] {
				seen[name] = true
				files = append(files, name)
				if len(files) >= maxFiles {
					return files
				}
			}
		}
	}

	// Add some subdirectories if space remains
	if len(files) < maxFiles {
		entries, err := os.ReadDir(cwd)
		if err == nil {
			for _, entry := range entries {
				if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
					dirName := entry.Name() + "/"
					if !seen[dirName] {
						seen[dirName] = true
						files = append(files, dirName)
						if len(files) >= maxFiles {
							break
						}
					}
				}
			}
		}
	}

	return files
}

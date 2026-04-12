package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type DependencyStatus struct {
	Name      string `json:"name"`
	Installed bool   `json:"installed"`
	Version   string `json:"version"`
	Error     string `json:"error"`
}

type DependencyCheckResult struct {
	AllInstalled bool               `json:"allInstalled"`
	Dependencies []DependencyStatus `json:"dependencies"`
	MissingTools []string           `json:"missingTools"`
	ErrorMessage string             `json:"errorMessage"`
}

var requiredDependencies = []string{"ffmpeg", "yt-dlp", "gallery-dl"}

// CheckDependencies checks if all required tools are installed
func (a *App) CheckDependencies() DependencyCheckResult {
	result := DependencyCheckResult{
		AllInstalled: true,
		Dependencies: []DependencyStatus{},
		MissingTools: []string{},
	}

	for _, tool := range requiredDependencies {
		status := checkTool(tool)
		result.Dependencies = append(result.Dependencies, status)

		if !status.Installed {
			result.AllInstalled = false
			result.MissingTools = append(result.MissingTools, tool)
		}
	}

	if !result.AllInstalled {
		result.ErrorMessage = fmt.Sprintf(
			"Missing dependencies: %s\nPlease install them to use YTDown.",
			strings.Join(result.MissingTools, ", "),
		)
	}

	return result
}

// InstallDependencies installs missing dependencies via Homebrew
func (a *App) InstallDependencies(tools []string) (bool, string) {
	// Tìm đường dẫn brew thực tế
	brewPath := ""
	if p, err := exec.LookPath("brew"); err == nil {
		brewPath = p
	} else {
		for _, p := range []string{"/opt/homebrew/bin/brew", "/usr/local/bin/brew"} {
			if _, err := os.Stat(p); err == nil {
				brewPath = p
				break
			}
		}
	}
	if brewPath == "" {
		return false, "Homebrew is not installed. Please install: https://brew.sh"
	}

	for _, tool := range tools {
		if status := checkTool(tool); status.Installed {
			continue
		}
		cmd := exec.Command(brewPath, "install", tool) // ← dùng path tuyệt đối
		if err := cmd.Run(); err != nil {
			return false, fmt.Sprintf("Failed to install %s: %v", tool, err)
		}
	}
	return true, ""
}

// checkTool checks if a tool is installed and returns its version
func checkTool(toolName string) DependencyStatus {
	status := DependencyStatus{Name: toolName}

	// Bước 1: Thử LookPath (hoạt động khi mở từ Terminal)
	if path, err := exec.LookPath(toolName); err == nil {
		status.Installed = true
		status.Version = getToolVersion(toolName, path)
		return status
	}

	// Bước 2: Fallback check các đường dẫn Homebrew thường gặp
	// (khi mở từ Finder, $PATH không có /opt/homebrew/bin)
	commonPaths := []string{
		"/opt/homebrew/bin/" + toolName, // Apple Silicon (M1/M2/M3)
		"/usr/local/bin/" + toolName,    // Intel Mac
		"/usr/bin/" + toolName,
	}

	for _, p := range commonPaths {
		if _, err := os.Stat(p); err == nil {
			status.Installed = true
			status.Version = getToolVersion(toolName, p)
			return status
		}
	}

	// Không tìm thấy ở đâu cả
	status.Installed = false
	status.Error = "not found in PATH or Homebrew directories"
	return status
}

// getToolVersion retrieves the version of a tool
func getToolVersion(toolName string, toolPath string) string {
	var cmd *exec.Cmd

	switch toolName {
	case "ffmpeg":
		cmd = exec.Command(toolPath, "-version")
	case "yt-dlp":
		cmd = exec.Command(toolPath, "--version")
	case "gallery-dl":
		cmd = exec.Command(toolPath, "--version")
	default:
		return "unknown"
	}

	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}

	// Extract first line which usually contains version
	lines := strings.Split(string(output), "\n")
	if len(lines) > 0 {
		version := strings.TrimSpace(lines[0])
		if version == "" {
			return "unknown"
		}
		fields := strings.Fields(version)

		switch toolName {
		case "ffmpeg":
			// Output: "ffmpeg version 7.0.2 Copyright..."
			// fields: ["ffmpeg", "version", "7.0.2", ...]
			if len(fields) >= 3 && fields[1] == "version" {
				return fields[2]
			}
		case "gallery-dl":
			// Output: "gallery-dl 1.28.1"
			// fields: ["gallery-dl", "1.28.1"]
			if len(fields) >= 2 {
				return fields[1]
			}
		default:
			// yt-dlp output: "2024.03.10" — trả về trực tiếp
			if len(fields) > 0 {
				return fields[0]
			}
		}
	}

	return "unknown"
}

// isBrewInstalled checks if Homebrew is installed
func isBrewInstalled() bool {
	if _, err := exec.LookPath("brew"); err == nil {
		return true
	}
	// Fallback khi mở từ Finder
	for _, p := range []string{"/opt/homebrew/bin/brew", "/usr/local/bin/brew"} {
		if _, err := os.Stat(p); err == nil {
			return true
		}
	}
	return false
}

// GetBrewInstallStatus returns information about Homebrew installation
func (a *App) GetBrewInstallStatus() map[string]interface{} {
	installed := isBrewInstalled()
	var path string
	var version string

	if installed {
		// Tìm path thực của brew (giống logic isBrewInstalled)
		if p, err := exec.LookPath("brew"); err == nil {
			path = p
		} else {
			for _, p := range []string{"/opt/homebrew/bin/brew", "/usr/local/bin/brew"} {
				if _, err := os.Stat(p); err == nil {
					path = p
					break
				}
			}
		}
		if path != "" {
			cmd := exec.Command(path, "--version") // ← Dùng path tuyệt đối
			if out, err := cmd.Output(); err == nil {
				version = strings.TrimSpace(string(out))
			}
		}
	}

	return map[string]interface{}{
		"installed": installed,
		"path":      path,
		"version":   version,
		"setupUrl":  "https://brew.sh",
	}
}

// PromptToInstallDependencies shows a dialog and installs if user agrees
func (a *App) PromptToInstallDependencies() (bool, string) {
	// Check what's missing
	check := a.CheckDependencies()

	if check.AllInstalled {
		return true, ""
	}

	// Show dialog to user
	confirmed, _ := runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
		Type:    runtime.QuestionDialog,
		Title:   "Install Dependencies",
		Message: fmt.Sprintf("YTDown requires the following tools:\n\n%s\n\nWould you like to install them now via Homebrew?", strings.Join(check.MissingTools, "\n")),
	})

	if confirmed != "Yes" {
		return false, "User declined to install dependencies"
	}

	// Install
	success, errMsg := a.InstallDependencies(check.MissingTools)
	return success, errMsg
}

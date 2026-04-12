package main

import (
	"fmt"
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
	// Check if Homebrew is installed
	if !isBrewInstalled() {
		return false, "Homebrew is not installed. Please install Homebrew first:\nhttps://brew.sh"
	}

	for _, tool := range tools {
		// Skip if already installed
		if status := checkTool(tool); status.Installed {
			fmt.Printf("✅ %s is already installed (v%s)\n", tool, status.Version)
			continue
		}

		fmt.Printf("📦 Installing %s via Homebrew...\n", tool)

		cmd := exec.Command("brew", "install", tool)
		if err := cmd.Run(); err != nil {
			return false, fmt.Sprintf("Failed to install %s: %v", tool, err)
		}

		fmt.Printf("✅ Successfully installed %s\n", tool)
	}

	return true, ""
}

// checkTool checks if a tool is installed and returns its version
func checkTool(toolName string) DependencyStatus {
	status := DependencyStatus{
		Name: toolName,
	}

	// Check if tool exists in PATH
	path, err := exec.LookPath(toolName)
	if err != nil {
		status.Installed = false
		status.Error = "not found in PATH"
		return status
	}

	status.Installed = true

	// Get version
	version := getToolVersion(toolName, path)
	status.Version = version

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
		// Clean up version string
		version = strings.Fields(version)[0]
		return version
	}

	return "unknown"
}

// isBrewInstalled checks if Homebrew is installed
func isBrewInstalled() bool {
	_, err := exec.LookPath("brew")
	return err == nil
}

// GetBrewInstallStatus returns information about Homebrew installation
func (a *App) GetBrewInstallStatus() map[string]interface{} {
	installed := isBrewInstalled()
	var path string
	var version string

	if installed {
		p, _ := exec.LookPath("brew")
		path = p

		cmd := exec.Command("brew", "--version")
		if out, err := cmd.Output(); err == nil {
			version = strings.TrimSpace(string(out))
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

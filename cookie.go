package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

// CookieMode defines how cookies are handled
type CookieMode string

const (
	CookieModeNone    CookieMode = "none"
	CookieModeManual  CookieMode = "manual"
	CookieModeBrowser CookieMode = "browser"
)

// CookieConfig represents the persistent cookie settings
type CookieConfig struct {
	Mode            CookieMode `json:"mode"`
	SelectedBrowser string     `json:"selected_browser"`
	ManualHeader    string     `json:"manual_header,omitempty"` // Used for persistent manual cookies if desired
}

type temporaryCookieState struct {
	mu       sync.RWMutex
	header   string
	tempFile string
}

type parsedCookie struct {
	Name  string
	Value string
}

// GetInstalledBrowsers returns IDs of supported browsers actually installed on the system
func GetInstalledBrowsers() []string {
	var available []string
	
	// macOS common browsers to check
	checkList := map[string]struct {
		bundleID string
		appPath  string
	}{
		"chrome":  {"com.google.Chrome", "/Applications/Google Chrome.app"},
		"firefox": {"org.mozilla.firefox", "/Applications/Firefox.app"},
		"safari":  {"com.apple.Safari", "/Applications/Safari.app"},
		"edge":    {"com.microsoft.edgemac", "/Applications/Microsoft Edge.app"},
		"brave":   {"com.brave.Browser", "/Applications/Brave Browser.app"},
		"opera":   {"com.operasoftware.Opera", "/Applications/Opera.app"},
		"vivaldi": {"com.vivaldi.Vivaldi", "/Applications/Vivaldi.app"},
	}

	for id, info := range checkList {
		if _, err := os.Stat(info.appPath); err == nil {
			available = append(available, id)
		}
	}
	
	return available
}

// GetUA returns a dynamic User-Agent based on the selected browser's actual version
func (m *CookieManager) GetUA() string {
	m.mu.RLock()
	browser := m.config.SelectedBrowser
	m.mu.RUnlock()

	osVersion := getMacOSVersion()
	osVersionUA := strings.ReplaceAll(osVersion, ".", "_")

	if browser == "" {
		// Default to Safari (pre-installed on every Mac) if none selected
		browser = "safari"
	}

	version := getBrowserVersionDynamic(browser)
	if version == "" {
		version = "17.0" // Generic fallback
	}

	switch strings.ToLower(browser) {
	case "chrome":
		return fmt.Sprintf("Mozilla/5.0 (Macintosh; Intel Mac OS X %s) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/%s Safari/537.36", osVersionUA, version)
	case "firefox":
		return fmt.Sprintf("Mozilla/5.0 (Macintosh; Intel Mac OS X %s; rv:%s) Gecko/20100101 Firefox/%s", osVersion, version, version)
	case "safari":
		return fmt.Sprintf("Mozilla/5.0 (Macintosh; Intel Mac OS X %s) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/%s Safari/605.1.15", osVersionUA, version)
	case "edge":
		return fmt.Sprintf("Mozilla/5.0 (Macintosh; Intel Mac OS X %s) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/%s Safari/537.36 Edg/%s", osVersionUA, version, version)
	default:
		// Generic WebKit based fallback
		return fmt.Sprintf("Mozilla/5.0 (Macintosh; Intel Mac OS X %s) AppleWebKit/605.1.15 (KHTML, like Gecko) Chrome/%s Safari/537.36", osVersionUA, version)
	}
}

func getBrowserVersionDynamic(id string) string {
	paths := map[string]string{
		"chrome":  "/Applications/Google Chrome.app/Contents/Info.plist",
		"firefox": "/Applications/Firefox.app/Contents/Info.plist",
		"safari":  "/Applications/Safari.app/Contents/Info.plist",
		"edge":    "/Applications/Microsoft Edge.app/Contents/Info.plist",
		"brave":   "/Applications/Brave Browser.app/Contents/Info.plist",
		"opera":   "/Applications/Opera.app/Contents/Info.plist",
		"vivaldi": "/Applications/Vivaldi.app/Contents/Info.plist",
	}

	plistPath, ok := paths[id]
	if !ok {
		return ""
	}

	cmd := exec.Command("defaults", "read", plistPath, "CFBundleShortVersionString")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func getMacOSVersion() string {
	cmd := exec.Command("sw_vers", "-productVersion")
	out, err := cmd.Output()
	if err != nil {
		return "10.15.7"
	}
	return strings.TrimSpace(string(out))
}

// Global cookie state manager
type CookieManager struct {
	mu     sync.RWMutex
	config CookieConfig
	state  temporaryCookieState
}

var manager = &CookieManager{
	config: CookieConfig{
		Mode: CookieModeNone,
	},
}

var cookieNamePattern = regexp.MustCompile(`^[A-Za-z0-9!#$%&'*+\-.^_` + "`" + `|~]+$`)

// GetConfigDir returns the path to the app's config directory
func GetConfigDir() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	dir := filepath.Join(configDir, "YTDown")
	_ = os.MkdirAll(dir, 0755)
	return dir
}

// LoadConfig loads the cookie configuration from disk
func (m *CookieManager) LoadConfig() {
	m.mu.Lock()
	defer m.mu.Unlock()

	dir := GetConfigDir()
	if dir == "" {
		return
	}

	path := filepath.Join(dir, "cookies_settings.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	_ = json.Unmarshal(data, &m.config)

	// If mode is manual, we don't restore the header from disk for security/freshness
	// unless we specifically want to. The user said manual should be cleared on exit,
	// so we keep it empty in memory on startup.
	if m.config.Mode == CookieModeManual {
		m.config.ManualHeader = ""
	}
}

// SaveConfig saves the cookie configuration to disk
func (m *CookieManager) SaveConfig() {
	m.mu.RLock()
	defer m.mu.RUnlock()

	dir := GetConfigDir()
	if dir == "" {
		return
	}

	path := filepath.Join(dir, "cookies_settings.json")
	
	// Prepare config for saving (clear manual header as it's session-only)
	saveCfg := m.config
	if saveCfg.Mode == CookieModeManual {
		saveCfg.ManualHeader = ""
	}

	data, _ := json.MarshalIndent(saveCfg, "", "  ")
	_ = os.WriteFile(path, data, 0644)
}

// GetCookieArgs returns the command line arguments for cookies
func (m *CookieManager) GetCookieArgs(tool string) []string {
	m.mu.RLock()
	cfg := m.config
	m.mu.RUnlock()

	switch cfg.Mode {
	case CookieModeBrowser:
		if cfg.SelectedBrowser != "" {
			return []string{"--cookies-from-browser", cfg.SelectedBrowser}
		}
	case CookieModeManual:
		m.state.mu.RLock()
		header := m.state.header
		tempFile := m.state.tempFile
		m.state.mu.RUnlock()

		if header != "" {
			if tool == "gallery-dl" {
				return []string{"-o", "http.headers.Cookie=" + header}
			}
			// For yt-dlp, use the temp file if available
			if tempFile != "" {
				return []string{"--cookies", tempFile}
			}
		}
	}
	return nil
}

// Compatibility functions for existing code

func setTemporaryYouTubeCookie(raw string) error {
	cookies, err := parseCookieInput(raw, true)
	if err != nil {
		return err
	}

	tempFile, err := writeTemporaryCookieFile(cookies, ".youtube.com")
	if err != nil {
		return err
	}

	manager.state.mu.Lock()
	if manager.state.tempFile != "" && manager.state.tempFile != tempFile {
		_ = os.RemoveAll(filepath.Dir(manager.state.tempFile))
	}

	pairs := make([]string, 0, len(cookies))
	for _, cookie := range cookies {
		pairs = append(pairs, cookie.Name+"="+cookie.Value)
	}

	manager.state.header = strings.Join(pairs, "; ")
	manager.state.tempFile = tempFile
	manager.state.mu.Unlock()

	manager.mu.Lock()
	manager.config.Mode = CookieModeManual
	manager.mu.Unlock()
	manager.SaveConfig()

	return nil
}

func setGalleryCookie(raw string) error {
	// Re-using the same logic as YouTube for unified cookies
	return setTemporaryYouTubeCookie(raw)
}

func getGalleryCookieHeader() string {
	manager.state.mu.RLock()
	defer manager.state.mu.RUnlock()
	return manager.state.header
}

func hasGalleryCookie() bool {
	manager.mu.RLock()
	mode := manager.config.Mode
	manager.mu.RUnlock()

	if mode == CookieModeBrowser {
		return true
	}

	manager.state.mu.RLock()
	defer manager.state.mu.RUnlock()
	return manager.state.header != ""
}

func hasTemporaryCookie() bool {
	manager.mu.RLock()
	mode := manager.config.Mode
	manager.mu.RUnlock()

	if mode == CookieModeBrowser {
		return true
	}

	manager.state.mu.RLock()
	defer manager.state.mu.RUnlock()
	return manager.state.tempFile != ""
}

func getTemporaryCookieFile() string {
	manager.state.mu.RLock()
	defer manager.state.mu.RUnlock()
	return manager.state.tempFile
}

func clearTemporaryYouTubeCookie() {
	manager.state.mu.Lock()
	if manager.state.tempFile != "" {
		_ = os.RemoveAll(filepath.Dir(manager.state.tempFile))
	}
	manager.state.header = ""
	manager.state.tempFile = ""
	manager.state.mu.Unlock()

	manager.mu.Lock()
	manager.config.Mode = CookieModeNone
	manager.mu.Unlock()
	manager.SaveConfig()
}

// Utility functions kept from original

func parseCookieInput(raw string, isYouTube bool) ([]parsedCookie, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, fmt.Errorf("cookie input is empty")
	}

	if extracted := extractCookieHeader(trimmed); extracted != "" {
		trimmed = extracted
	}

	parts := strings.Split(trimmed, ";")
	cookies := make([]parsedCookie, 0, len(parts))
	seen := make(map[string]bool)

	for _, part := range parts {
		token := strings.TrimSpace(strings.ReplaceAll(part, "\n", " "))
		token = strings.TrimSpace(strings.ReplaceAll(token, "\r", " "))
		if token == "" {
			continue
		}

		name, value, ok := strings.Cut(token, "=")
		if !ok {
			continue
		}

		name = strings.TrimSpace(name)
		value = strings.TrimSpace(value)
		if name == "" || value == "" {
			continue
		}
		if !cookieNamePattern.MatchString(name) {
			continue
		}

		if seen[name] {
			for i := range cookies {
				if cookies[i].Name == name {
					cookies[i].Value = value
					break
				}
			}
			continue
		}

		seen[name] = true
		cookies = append(cookies, parsedCookie{
			Name:  name,
			Value: value,
		})
	}

	if len(cookies) == 0 {
		return nil, fmt.Errorf("no valid cookie pairs found")
	}

	if isYouTube && !hasUsefulYouTubeAuthCookie(cookies) {
		return nil, fmt.Errorf("no usable YouTube authentication cookies found. Please copy a fresh YouTube Cookie header")
	}

	return cookies, nil
}

func extractCookieHeader(raw string) string {
	lines := strings.Split(raw, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(strings.ToLower(line), "cookie:") {
			return strings.TrimSpace(line[len("cookie:"):])
		}
	}
	return raw
}

func hasUsefulYouTubeAuthCookie(cookies []parsedCookie) bool {
	authNames := map[string]struct{}{
		"SAPISID":           {},
		"__Secure-1PSID":    {},
		"__Secure-3PSID":    {},
		"SID":               {},
		"SSID":              {},
		"HSID":              {},
		"LOGIN_INFO":        {},
		"__Secure-1PAPISID": {},
		"__Secure-3PAPISID": {},
	}

	for _, cookie := range cookies {
		if _, ok := authNames[cookie.Name]; ok {
			return true
		}
	}

	return false
}

func writeTemporaryCookieFile(cookies []parsedCookie, domain string) (string, error) {
	dir, err := os.MkdirTemp("", "ytdown-cookie-*")
	if err != nil {
		return "", fmt.Errorf("failed to prepare temporary cookie storage: %w", err)
	}

	path := filepath.Join(dir, "cookies.txt")
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		_ = os.RemoveAll(dir)
		return "", fmt.Errorf("failed to create temporary cookie file: %w", err)
	}
	defer file.Close()

	if _, err := file.WriteString("# Netscape HTTP Cookie File\n"); err != nil {
		_ = os.RemoveAll(dir)
		return "", fmt.Errorf("failed to write temporary cookie file: %w", err)
	}

	for _, cookie := range cookies {
		d := domain
		if d == "" {
			d = ".example.com"
		}
		line := fmt.Sprintf("%s\tTRUE\t/\tTRUE\t2147483647\t%s\t%s\n", d, cookie.Name, cookie.Value)
		if _, err := file.WriteString(line); err != nil {
			_ = os.RemoveAll(dir)
			return "", fmt.Errorf("failed to write temporary cookie file: %w", err)
		}
	}

	return path, nil
}

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

type temporaryCookieState struct {
	mu       sync.RWMutex
	header   string
	tempFile string
}

type parsedCookie struct {
	Name  string
	Value string
}

var youtubeCookieState temporaryCookieState

var cookieNamePattern = regexp.MustCompile(`^[A-Za-z0-9!#$%&'*+\-.^_` + "`" + `|~]+$`)

func setTemporaryYouTubeCookie(raw string) error {
	cookies, err := parseYouTubeCookieInput(raw)
	if err != nil {
		return err
	}

	tempFile, err := writeTemporaryCookieFile(cookies)
	if err != nil {
		return err
	}

	youtubeCookieState.mu.Lock()
	defer youtubeCookieState.mu.Unlock()

	if youtubeCookieState.tempFile != "" && youtubeCookieState.tempFile != tempFile {
		_ = os.RemoveAll(filepath.Dir(youtubeCookieState.tempFile))
	}

	pairs := make([]string, 0, len(cookies))
	for _, cookie := range cookies {
		pairs = append(pairs, cookie.Name+"="+cookie.Value)
	}

	youtubeCookieState.header = strings.Join(pairs, "; ")
	youtubeCookieState.tempFile = tempFile
	return nil
}

func hasTemporaryCookie() bool {
	youtubeCookieState.mu.RLock()
	defer youtubeCookieState.mu.RUnlock()
	return youtubeCookieState.tempFile != ""
}

func getTemporaryCookieFile() string {
	youtubeCookieState.mu.RLock()
	defer youtubeCookieState.mu.RUnlock()
	return youtubeCookieState.tempFile
}

func clearTemporaryYouTubeCookie() {
	youtubeCookieState.mu.Lock()
	defer youtubeCookieState.mu.Unlock()

	if youtubeCookieState.tempFile != "" {
		_ = os.RemoveAll(filepath.Dir(youtubeCookieState.tempFile))
	}

	youtubeCookieState.header = ""
	youtubeCookieState.tempFile = ""
}

func parseYouTubeCookieInput(raw string) ([]parsedCookie, error) {
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
		return nil, fmt.Errorf("no valid cookie pairs found. Please copy the full YouTube Cookie header again")
	}

	if !hasUsefulYouTubeAuthCookie(cookies) {
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

func writeTemporaryCookieFile(cookies []parsedCookie) (string, error) {
	dir, err := os.MkdirTemp("", "ytdown-cookie-*")
	if err != nil {
		return "", fmt.Errorf("failed to prepare temporary cookie storage: %w", err)
	}

	path := filepath.Join(dir, "youtube.cookies.txt")
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
		line := fmt.Sprintf(".youtube.com\tTRUE\t/\tTRUE\t2147483647\t%s\t%s\n", cookie.Name, cookie.Value)
		if _, err := file.WriteString(line); err != nil {
			_ = os.RemoveAll(dir)
			return "", fmt.Errorf("failed to write temporary cookie file: %w", err)
		}
	}

	return path, nil
}

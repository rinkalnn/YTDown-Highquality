package main

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type VideoInfo struct {
	Title     string `json:"title"`
	Thumbnail string `json:"thumbnail"`
	ID        string `json:"id"`
}

// DownloadVideo downloads a video using yt-dlp
func DownloadVideo(ctx context.Context, index int, url, format, quality, savePath string) error {
	ytdlpPath := getResourcePath("yt-dlp")

	// Force using system/brew ffmpeg and IGNORE bundled one
	ffmpegPath := ""
	for _, p := range []string{
		"/opt/homebrew/bin/ffmpeg",
		"/usr/local/bin/ffmpeg",
		"/usr/bin/ffmpeg",
	} {
		if info, err := os.Stat(p); err == nil && !info.IsDir() {
			ffmpegPath = p
			break
		}
	}
	// Fallback to searching in PATH if not found in common brew/system locations
	if ffmpegPath == "" {
		if path, err := exec.LookPath("ffmpeg"); err == nil {
			ffmpegPath = path
		}
	}

	if ytdlpPath == "" {
		return fmt.Errorf("yt-dlp not found. Please install it or use the Setup Dependencies button.")
	}

	if ffmpegPath == "" {
		LogError("[DL] ffmpeg not found in brew or system paths. Merging will likely fail.")
	} else {
		LogInfo("[DL] Using system ffmpeg from: %s", ffmpegPath)
	}

	// Fetch metadata first to get title, thumbnail, and ID
	info, _ := GetVideoMetadata(ctx, url)
	if info != nil {
		runtime.EventsEmit(ctx, "video-info", map[string]interface{}{
			"index":     index,
			"title":     info.Title,
			"thumbnail": info.Thumbnail,
			"id":        info.ID,
		})
	}

	// Build yt-dlp arguments based on format and quality
	args := buildDownloadArgs(format, quality, savePath, ffmpegPath)
	args = append(args, url)
	// Add verbose for better debugging in app.log
	// args = append(args, "--verbose")

	LogInfo("[DL] Running command: %s with %d args: %v", ytdlpPath, len(args), args)

	cmd := exec.CommandContext(ctx, ytdlpPath, args...)

	// Ensure /opt/homebrew/bin and other common paths are in PATH so yt-dlp can find ffmpeg, deno, etc.
	existingPath := os.Getenv("PATH")
	newPath := "/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin"
	if existingPath != "" {
		newPath = existingPath + ":" + newPath
	}
	cmd.Env = append(os.Environ(), "PATH="+newPath)

	// CombinedOutput is simpler but we need to stream progress,
	// so we'll pipe stderr to stdout to avoid buffer deadlocks.

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	cmd.Stderr = cmd.Stdout

	if err := cmd.Start(); err != nil {
		return err
	}

	// Read combined progress output
	scanner := bufio.NewScanner(stdout)

	// Set a larger buffer for the scanner (up to 1MB) to handle long output lines
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	// Custom split function to handle both \n and \r
	scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}
		for i, b := range data {
			if b == '\n' || b == '\r' {
				return i + 1, data[0:i], nil
			}
		}
		if atEOF {
			return len(data), data, nil
		}
		return 0, nil, nil
	})

	var finalFilePath string

	for scanner.Scan() {
		line := scanner.Text()

		// Log every line from yt-dlp for debugging, but prefix it to distinguish
		// We'll use LogDebug for regular progress to avoid cluttering INFO if needed,
		// but since we want to catch "hangs", INFO is safer for now.
		if strings.Contains(line, "ERROR:") || strings.Contains(line, "WARNING:") {
			LogError("[yt-dlp] %s", line)
		} else {
			// Only log important transitions or status updates to app.log to keep it readable
			// but enough to see where it "hangs"
			if strings.Contains(line, "[download]") || strings.Contains(line, "[ExtractAudio]") ||
				strings.Contains(line, "[Merger]") || strings.Contains(line, "[ffmpeg]") {
				LogInfo("[yt-dlp] %s", line)
			}
		}

		if strings.Contains(line, "[download]") {
			if strings.Contains(line, "Destination:") {
				// Extract filename from "[download] Destination: /path/to/Title.mp4"
				fullPath := strings.TrimSpace(strings.TrimPrefix(line, "[download] Destination: "))
				finalFilePath = fullPath
				title := filepath.Base(fullPath)
				// Remove extension
				if ext := filepath.Ext(title); ext != "" {
					title = title[:len(title)-len(ext)]
				}
				runtime.EventsEmit(ctx, "video-title", title)
			} else if strings.Contains(line, "has already been downloaded") {
				// Handle case where file exists: "[download] /path/to/Title.mp4 has already been downloaded"
				fullPath := strings.TrimSpace(strings.TrimPrefix(line, "[download] "))
				fullPath = strings.TrimSuffix(fullPath, " has already been downloaded")
				finalFilePath = fullPath
				title := filepath.Base(fullPath)
				if ext := filepath.Ext(title); ext != "" {
					title = title[:len(title)-len(ext)]
				}
				runtime.EventsEmit(ctx, "video-title", title)
				runtime.EventsEmit(ctx, "progress-update", map[string]interface{}{
					"index":      index,
					"percentage": 100.0,
					"speed":      "Done",
					"eta":        "00:00",
				})
			} else {
				progress := parseProgress(line)
				progress["index"] = index
				// Emit even if percentage is 0 to show starting state
				runtime.EventsEmit(ctx, "progress-update", progress)
			}
		}
		if strings.Contains(line, "[Merger]") || strings.Contains(line, "[ffmpeg]") || strings.Contains(line, "[VideoConvertor]") {
			// For merger, extract the final file path if available
			if strings.Contains(line, "Merging formats into \"") {
				re := regexp.MustCompile(`Merging formats into "([^"]+)"`)
				if match := re.FindStringSubmatch(line); len(match) > 1 {
					finalFilePath = match[1]
				}
			}
			runtime.EventsEmit(ctx, "progress-update", map[string]interface{}{
				"index":      index,
				"percentage": 100.0,
				"speed":      "Processing...",
				"eta":        "Almost done",
			})
		}
	}

	if err := scanner.Err(); err != nil {
		LogError("[DL] scanner error: %v", err)
	}

	if err := cmd.Wait(); err != nil {
		if ctx.Err() != nil {
			LogInfo("[DL] Download cancelled by user")
			return ctx.Err()
		}
		LogError("[DL] Command failed with error: %v", err)
		return fmt.Errorf("download failed: %v", err)
	}

	// If we have a final file path, emit it as complete
	if finalFilePath != "" {
		runtime.EventsEmit(ctx, "download-complete", map[string]interface{}{
			"index":    index,
			"filePath": finalFilePath,
		})
	}

	return nil
}

// buildDownloadArgs builds yt-dlp command arguments
func buildDownloadArgs(format, quality, savePath, ffmpegPath string) []string {
	args := []string{}

	switch format {
	case "MP4":
		if quality == "Best" || quality == "Best Quality" {
			// Matches the user's manual successful command: bestvideo+bestaudio/best
			args = append(args, "-f", "bestvideo+bestaudio/best")
		} else {
			// Map quality to height
			qualityHeight := qualityToHeight(quality)
			args = append(args, "-f", fmt.Sprintf("bestvideo[height<=%s]+bestaudio/best[height<=%s]", qualityHeight, qualityHeight))
		}
		args = append(args, "--merge-output-format", "mp4")

	case "MP3":
		args = append(args,
			"-f", "bestaudio",
			"--extract-audio",
			"--audio-format", "mp3",
			"--audio-quality", "0",
		)
	}

	// Common arguments
	args = append(args,
		"--no-playlist",
		"--no-continue",
		"--force-overwrites",
		"-o", filepath.Join(savePath, "%(title)s.%(ext)s"),
		"--referer", "https://www.youtube.com/",
	)

	if ffmpegPath != "" {
		args = append(args, "--ffmpeg-location", ffmpegPath)
	}

	// Get cookie and browser settings
	manager.mu.RLock()
	cookieMode := manager.config.Mode
	selectedBrowser := manager.config.SelectedBrowser
	manager.mu.RUnlock()

	// Apply browser-specific User-Agent if a browser is selected or manual cookie is used
	userAgent := "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36" // Default Chrome
	if cookieMode == CookieModeBrowser && selectedBrowser != "" {
		userAgent = getUserAgentForBrowser(selectedBrowser)
	}
	args = append(args, "--user-agent", userAgent)

	if cookieArgs := manager.GetCookieArgs("yt-dlp"); len(cookieArgs) > 0 {
		args = append(args, cookieArgs...)
	}

	return args
}

// getUserAgentForBrowser returns a realistic User-Agent for common browsers
func getUserAgentForBrowser(browser string) string {
	switch strings.ToLower(browser) {
	case "chrome", "google-chrome":
		return "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36"
	case "firefox":
		return "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:124.0) Gecko/20100101 Firefox/124.0"
	case "safari":
		return "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.4 Safari/605.1.15"
	case "edge":
		return "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36 Edg/123.0.0.0"
	case "brave":
		return "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36"
	default:
		return "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36"
	}
}

// qualityToHeight converts quality string to pixel height
func qualityToHeight(quality string) string {
	heightMap := map[string]string{
		"1080p": "1080",
		"720p":  "720",
		"480p":  "480",
		"360p":  "360",
	}
	if h, ok := heightMap[quality]; ok {
		return h
	}
	return "1080"
}

// parseProgress parses yt-dlp progress output
func parseProgress(line string) map[string]interface{} {
	progress := map[string]interface{}{
		"percentage": 0.0,
		"speed":      "0 MB/s",
		"eta":        "—",
		"raw":        line,
	}

	// Extract percentage: " 57.3%" or "100%" or "100.0%"
	rePct := regexp.MustCompile(`(\d+(?:\.\d+)?)%`)
	if match := rePct.FindStringSubmatch(line); len(match) > 1 {
		if pct, err := strconv.ParseFloat(match[1], 64); err == nil {
			progress["percentage"] = pct
		}
	}

	// Extract speed: "at 3.20MiB/s"
	reSpeed := regexp.MustCompile(`at\s+([^\s]+)`)
	if match := reSpeed.FindStringSubmatch(line); len(match) > 1 {
		progress["speed"] = match[1]
	}

	// Extract ETA: "ETA 00:43"
	reETA := regexp.MustCompile(`ETA\s+([^\s]+)`)
	if match := reETA.FindStringSubmatch(line); len(match) > 1 {
		progress["eta"] = match[1]
	}

	return progress
}

// GetVideoMetadata fetches video title, thumbnail and ID
func GetVideoMetadata(ctx context.Context, url string) (*VideoInfo, error) {
	ytdlpPath := getResourcePath("yt-dlp")
	if ytdlpPath == "" {
		return nil, fmt.Errorf("yt-dlp not found")
	}

	// Get title, thumbnail URL, and ID from yt-dlp
	args := []string{"--get-title", "--get-thumbnail", "--get-id", "--no-warnings", "--no-playlist"}
	if cookieArgs := manager.GetCookieArgs("yt-dlp"); len(cookieArgs) > 0 {
		args = append(args, cookieArgs...)
	}
	args = append(args, url)

	cmd := exec.CommandContext(ctx, ytdlpPath, args...)

	// Capture both stdout and stderr
	output, err := cmd.CombinedOutput()
	outputStr := strings.TrimSpace(string(output))

	if err != nil {
		LogError("[Metadata] yt-dlp failed: %v, output: %s", err, outputStr)
		return nil, err
	}

	lines := strings.Split(outputStr, "\n")

	if len(lines) < 3 {
		return nil, fmt.Errorf("could not extract title, thumbnail or ID")
	}

	title := strings.TrimSpace(lines[0])
	videoID := strings.TrimSpace(lines[1])
	thumbnailURL := strings.TrimSpace(lines[2])

	// Download thumbnail and convert to base64 data URL
	dataURL := downloadThumbnailAsBase64(thumbnailURL)

	return &VideoInfo{
		Title:     title,
		Thumbnail: dataURL,
		ID:        videoID,
	}, nil
}

// downloadThumbnailAsBase64 downloads thumbnail and returns as base64 data URL
func downloadThumbnailAsBase64(thumbnailURL string) string {
	if thumbnailURL == "" {
		return ""
	}

	// Download thumbnail with timeout
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(thumbnailURL)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ""
	}

	// Read thumbnail bytes
	thumbnailData, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}

	// Determine MIME type from URL
	mimeType := "image/jpeg"
	if strings.Contains(strings.ToLower(thumbnailURL), ".png") {
		mimeType = "image/png"
	} else if strings.Contains(strings.ToLower(thumbnailURL), ".webp") {
		mimeType = "image/webp"
	}

	// Convert to base64 data URL
	encoded := encodeBase64(thumbnailData)
	dataURL := fmt.Sprintf("data:%s;base64,%s", mimeType, encoded)

	return dataURL
}

// encodeBase64 encodes bytes to base64 string
func encodeBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// GetPlaylistVideos extracts all videos from a playlist
func GetPlaylistVideos(url string) ([]string, error) {
	ytdlpPath := getResourcePath("yt-dlp")
	if ytdlpPath == "" {
		return nil, fmt.Errorf("yt-dlp not found")
	}

	args := []string{"--flat-playlist", "-J"}
	if cookieArgs := manager.GetCookieArgs("yt-dlp"); len(cookieArgs) > 0 {
		args = append(args, cookieArgs...)
	}
	args = append(args, url)

	cmd := exec.Command(ytdlpPath, args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var data map[string]interface{}
	if err := json.Unmarshal(output, &data); err != nil {
		return nil, err
	}

	var videos []string
	if entries, ok := data["entries"].([]interface{}); ok {
		for _, entry := range entries {
			if e, ok := entry.(map[string]interface{}); ok {
				if id, ok := e["id"].(string); ok {
					videos = append(videos, "https://www.youtube.com/watch?v="+id)
				}
			}
		}
	}

	return videos, nil
}

// getResourcePath finds binary in system paths (brew) ONLY
func getResourcePath(name string) string {
	// 1. Try common system paths (Homebrew, etc.)
	for _, p := range []string{
		"/opt/homebrew/bin/" + name,
		"/usr/local/bin/" + name,
		"/usr/bin/" + name,
	} {
		if info, err := os.Stat(p); err == nil && !info.IsDir() {
			return p
		}
	}

	// 2. Last resort: try system PATH
	if path, err := exec.LookPath(name); err == nil {
		return path
	}

	return ""
}

type DownloadFailure struct {
	RequiresCookie bool
	DisplayMessage string
	Details        []string
}

func classifyDownloadFailure(err error, cookiePresent bool) DownloadFailure {
	if err == nil {
		return DownloadFailure{}
	}

	message := strings.TrimSpace(err.Error())
	lower := strings.ToLower(message)
	requiresCookie := looksLikeRestrictedAuthError(lower)

	details := []string{}
	if requiresCookie {
		details = append(details, "Restricted video or login required.")
		if cookiePresent {
			details = append(details, "Cookie invalid or insufficient.")
			details = append(details, "Please copy a fresh YouTube cookie.")
		} else {
			details = append(details, "Temporary cookie required.")
			details = append(details, "Paste a YouTube Cookie header to retry.")
		}
		return DownloadFailure{
			RequiresCookie: true,
			DisplayMessage: "Error",
			Details:        details,
		}
	}

	details = append(details, "Download failed.")
	details = append(details, summarizeErrorForUI(message))
	return DownloadFailure{
		RequiresCookie: false,
		DisplayMessage: "Error",
		Details:        details,
	}
}

func looksLikeRestrictedAuthError(message string) bool {
	patterns := []string{
		"sign in to confirm your age",
		"login required",
		"members-only",
		"private video",
		"private video. sign in",
		"confirm you're not a bot",
		"use --cookies",
		"this video is private",
		"authentication required",
		"sign in",
		"http error 403",
		"requested content is not available",
	}

	for _, pattern := range patterns {
		if strings.Contains(message, pattern) {
			return true
		}
	}

	return false
}

func summarizeErrorForUI(message string) string {
	message = strings.TrimSpace(message)
	if message == "" {
		return "Unknown yt-dlp error."
	}

	lines := strings.Split(message, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		line = strings.TrimPrefix(line, "ERROR: ")
		line = strings.TrimPrefix(line, "download failed: ")
		if len(line) > 120 {
			line = line[:117] + "..."
		}
		return line
	}

	return "Unknown yt-dlp error."
}

func IsRestrictedAuthError(err error) bool {
	if err == nil {
		return false
	}

	return looksLikeRestrictedAuthError(strings.ToLower(err.Error())) || errors.Is(err, os.ErrPermission)
}

// SanitizeFilename removes invalid characters from filename
func SanitizeFilename(filename string) string {
	invalidChars := []string{"/", ":", "*", "?", "\"", "<", ">", "|"}
	result := filename
	for _, char := range invalidChars {
		result = strings.ReplaceAll(result, char, "_")
	}
	return result
}

// OpenFileInFinder opens the file in Finder and highlights it
func OpenFileInFinder(filePath string) error {
	cmd := exec.Command("open", "-R", filePath)
	return cmd.Run()
}

// GetDefaultSavePath returns default download folder
func GetDefaultSavePath() string {
	usr, err := user.Current()
	if err != nil {
		return "~/Downloads"
	}
	return filepath.Join(usr.HomeDir, "Downloads")
}

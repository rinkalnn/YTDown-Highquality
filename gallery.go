package main

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// GalleryInfo stores basic info about a gallery
type GalleryInfo struct {
	URL   string `json:"url"`
	Title string `json:"title"`
}

// resolveShortURL follows redirects to find the final URL for short links
func resolveShortURL(url string) string {
	// Only resolve links that are known to be shorteners
	if !strings.Contains(url, "xhslink.com") && !strings.Contains(url, "vt.tiktok.com") {
		return url
	}

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
		Timeout: 10 * time.Second,
	}

	// Try HEAD first, then GET
	resp, err := client.Head(url)
	if err != nil {
		resp, err = client.Get(url)
		if err != nil {
			return url
		}
	}
	defer resp.Body.Close()

	if resp.Request != nil && resp.Request.URL != nil {
		finalURL := resp.Request.URL.String()
		LogInfo("[GDL] Resolved %s to %s", url, finalURL)
		return finalURL
	}

	return url
}

// DownloadGalleryWithOpts downloads images using gallery-dl with custom options
func DownloadGalleryWithOpts(ctx context.Context, index int, url string, options GalleryDownloadOptions) error {
	gallerydlPath := getResourcePath("gallery-dl")

	if gallerydlPath == "" {
		return fmt.Errorf("gallery-dl not found. Please install it using the Setup Dependencies button.")
	}

	// Resolve short URLs before passing to gallery-dl
	resolvedURL := resolveShortURL(url)

	args := []string{
		"--directory", options.SavePath,
	}

	if options.UgoiraToWebm {
		args = append(args, "--ugoira", "webm")
	}

	// Get dynamic User-Agent from system
	userAgent := manager.GetUA()
	if userAgent != "" {
		args = append(args, "-o", "http.user-agent="+userAgent)
	}

	// Decide whether to use cookies based on the domain
	// TikTok often fails with 403 Forbidden if cookies are present but not perfect.
	// Instagram REQUIRES cookies to work.
	useCookies := true
	if strings.Contains(resolvedURL, "tiktok.com") {
		useCookies = false
		LogInfo("[GDL] Skipping cookies for TikTok to avoid 403 Forbidden errors")
	}

	if useCookies {
		// Add cookie arguments from the unified manager
		if cookieArgs := manager.GetCookieArgs("gallery-dl"); len(cookieArgs) > 0 {
			args = append(args, cookieArgs...)
		} else if options.Browser != "" {
			// Fallback to legacy option only if global cookie is not set
			args = append(args, "--cookies-from-browser", options.Browser)
		}
	}

	// Force High Quality/Original for common sites
	args = append(args, "-o", "extractor.twitter.fullsize=True")
	args = append(args, "-o", "extractor.pixiv.ugoira=True")
	args = append(args, "-o", "extractor.tiktok.fullsize=True")

	if len(options.Formats) > 0 {
		// Create a filter for selected extensions
		// e.g. extension in ('jpg', 'jpeg', 'png')
		quotedFormats := make([]string, len(options.Formats))
		for i, fmt := range options.Formats {
			quotedFormats[i] = "'" + fmt + "'"
		}
		filter := fmt.Sprintf("extension in (%s)", strings.Join(quotedFormats, ", "))
		args = append(args, "--filter", filter)
	} else {
		// If no formats selected, we still want to skip videos by default 
		// if the user intended "Images" tab for images.
		args = append(args, "--filter", "extension not in ('mp4', 'm4v', 'webm', 'mov', 'avi', 'mkv', 'flv')")
	}

	if options.Archive {
		archivePath := filepath.Join(options.SavePath, "gallery-dl-archive.txt")
		args = append(args, "--download-archive", archivePath)
	}

	if options.ExtraArgs != "" {
		// Use a more robust way to split arguments that respects quotes
		// We'll implement a simple shell-style splitter here to avoid new dependencies
		extra, err := splitArguments(options.ExtraArgs)
		if err == nil {
			args = append(args, extra...)
		} else {
			LogError("[GDL] Error parsing extra args: %v", err)
		}
	}

	args = append(args, resolvedURL)

	LogInfo("[GDL] Running command: %s with args: %s", gallerydlPath, strings.Join(args, " "))

	cmd := exec.CommandContext(ctx, gallerydlPath, args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	// Read output to track progress
	scanner := bufio.NewScanner(stdout)
	count := 0
	
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		// If it's a file path (doesn't start with [), it's a downloaded file
		if !strings.HasPrefix(line, "[") {
			count++
			runtime.EventsEmit(ctx, "gallery-progress", map[string]interface{}{
				"index":      index,
				"percentage": 0.0, 
				"speed":      fmt.Sprintf("Downloaded %d files", count),
				"eta":        "Downloading...",
			})
			
			filename := filepath.Base(line)
			runtime.EventsEmit(ctx, "gallery-title", map[string]interface{}{
				"index": index,
				"title": "Gallery: " + filename,
			})
		} else {
			// Extract title or other info if possible
			// [pixiv][info] ...
		}
	}

	var stderrOutput strings.Builder
	errScanner := bufio.NewScanner(stderr)
	for errScanner.Scan() {
		line := errScanner.Text()
		LogInfo("[GDL] stderr: %s", line)
		stderrOutput.WriteString(line + "\n")
	}

	if err := cmd.Wait(); err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if stderrOutput.Len() > 0 {
			return fmt.Errorf("gallery download failed: %s", stderrOutput.String())
		}
		return fmt.Errorf("gallery download failed: %v", err)
	}

	runtime.EventsEmit(ctx, "gallery-complete", map[string]interface{}{
		"index":    index,
		"filePath": options.SavePath,
	})

	return nil
}

// DownloadGallery downloads images using gallery-dl (Legacy compatibility)
func DownloadGallery(ctx context.Context, index int, url, savePath string) error {
	return DownloadGalleryWithOpts(ctx, index, url, GalleryDownloadOptions{
		SavePath: savePath,
		Threads:  1,
	})
}

// splitArguments splits a command line string into separate arguments, 
// respecting single and double quotes.
func splitArguments(s string) ([]string, error) {
	var args []string
	var current strings.Builder
	var inQuotes rune

	for i := 0; i < len(s); i++ {
		r := rune(s[i])
		if inQuotes != 0 {
			if r == inQuotes {
				inQuotes = 0
			} else {
				current.WriteRune(r)
			}
		} else {
			if r == '\'' || r == '"' {
				inQuotes = r
			} else if r == ' ' || r == '\t' || r == '\n' || r == '\r' {
				if current.Len() > 0 {
					args = append(args, current.String())
					current.Reset()
				}
			} else {
				current.WriteRune(r)
			}
		}
	}

	if current.Len() > 0 {
		args = append(args, current.String())
	}

	if inQuotes != 0 {
		return nil, fmt.Errorf("unclosed quote")
	}

	return args, nil
}

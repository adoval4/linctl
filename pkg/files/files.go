package files

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ImageInfo represents information about an image found in markdown
type ImageInfo struct {
	URL         string
	AltText     string
	IsLinearURL bool
}

// ExtractImagesFromMarkdown finds all image URLs in markdown text
func ExtractImagesFromMarkdown(markdown string) []ImageInfo {
	// Match markdown image syntax: ![alt text](url)
	// Also match HTML img tags: <img src="url" />
	var images []ImageInfo

	// Markdown image regex: ![alt](url)
	markdownRegex := regexp.MustCompile(`!\[([^\]]*)\]\(([^)]+)\)`)
	matches := markdownRegex.FindAllStringSubmatch(markdown, -1)
	for _, match := range matches {
		if len(match) >= 3 {
			images = append(images, ImageInfo{
				URL:         match[2],
				AltText:     match[1],
				IsLinearURL: strings.Contains(match[2], "linear.app"),
			})
		}
	}

	// HTML img tag regex: <img src="url"
	htmlRegex := regexp.MustCompile(`<img[^>]+src="([^"]+)"`)
	matches = htmlRegex.FindAllStringSubmatch(markdown, -1)
	for _, match := range matches {
		if len(match) >= 2 {
			images = append(images, ImageInfo{
				URL:         match[1],
				AltText:     "",
				IsLinearURL: strings.Contains(match[1], "linear.app"),
			})
		}
	}

	return images
}

// DownloadImage downloads an image from a URL and saves it to the specified path
// authHeader is optional and will be used for authentication if provided (e.g., for Linear URLs)
func DownloadImage(ctx context.Context, url string, outputPath string, authHeader string) error {
	// Create HTTP request with context
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add authentication header if provided
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}

	// Execute request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download image: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %s", resp.Status)
	}

	// Create output directory if it doesn't exist
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create output file
	out, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer out.Close()

	// Copy image data to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write image data: %w", err)
	}

	return nil
}

// UploadFileInfo contains information needed to upload a file to Linear
type UploadFileInfo struct {
	UploadURL   string
	AssetURL    string
	Headers     map[string]string
	ContentType string
	Size        int64
}

// UploadToPresignedURL uploads file content to a pre-signed URL
func UploadToPresignedURL(ctx context.Context, info *UploadFileInfo, fileContent []byte) error {
	// Create PUT request
	req, err := http.NewRequestWithContext(ctx, "PUT", info.UploadURL, bytes.NewReader(fileContent))
	if err != nil {
		return fmt.Errorf("failed to create upload request: %w", err)
	}

	// Set required headers
	req.Header.Set("Content-Type", info.ContentType)
	req.Header.Set("Cache-Control", "public, max-age=31536000")

	// Add custom headers from Linear
	for key, value := range info.Headers {
		req.Header.Set(key, value)
	}

	// Execute upload
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}
	defer resp.Body.Close()

	// Check status
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload failed with status %s: %s", resp.Status, string(body))
	}

	return nil
}

// ReadFile reads a file from the filesystem
func ReadFile(filePath string) ([]byte, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}
	return data, nil
}

// GetFileInfo returns file metadata
func GetFileInfo(filePath string) (size int64, contentType string, err error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return 0, "", fmt.Errorf("failed to stat file: %w", err)
	}

	size = info.Size()

	// Determine content type from extension
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	case ".png":
		contentType = "image/png"
	case ".gif":
		contentType = "image/gif"
	case ".webp":
		contentType = "image/webp"
	case ".svg":
		contentType = "image/svg+xml"
	case ".bmp":
		contentType = "image/bmp"
	case ".mp4":
		contentType = "video/mp4"
	case ".webm":
		contentType = "video/webm"
	case ".mov":
		contentType = "video/quicktime"
	case ".pdf":
		contentType = "application/pdf"
	default:
		contentType = "application/octet-stream"
	}

	return size, contentType, nil
}

// SanitizeFilename creates a safe filename from a URL or alt text
func SanitizeFilename(name string) string {
	// Replace invalid characters with underscores
	reg := regexp.MustCompile(`[^a-zA-Z0-9._-]`)
	safe := reg.ReplaceAllString(name, "_")

	// Limit length
	if len(safe) > 200 {
		safe = safe[:200]
	}

	return safe
}

// InjectImageIntoMarkdown adds an image reference to markdown content
func InjectImageIntoMarkdown(markdown, imageURL, altText string) string {
	if altText == "" {
		altText = "image"
	}

	imageMarkdown := fmt.Sprintf("![%s](%s)", altText, imageURL)

	// If markdown is empty, just return the image
	if strings.TrimSpace(markdown) == "" {
		return imageMarkdown
	}

	// Append image to existing markdown with a newline
	return markdown + "\n\n" + imageMarkdown
}

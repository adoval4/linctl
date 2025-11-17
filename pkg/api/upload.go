package api

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/dorkitude/linctl/pkg/files"
)

// UploadFileToLinear uploads a file to Linear's cloud storage and returns the asset URL
func (c *Client) UploadFileToLinear(ctx context.Context, filePath string) (string, error) {
	// Get file metadata
	size, contentType, err := files.GetFileInfo(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to get file info: %w", err)
	}

	// Read file content
	fileContent, err := files.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	// Get filename
	filename := filepath.Base(filePath)

	// Request pre-signed upload URL from Linear
	uploadInfo, err := c.FileUpload(ctx, filename, int(size), contentType)
	if err != nil {
		return "", fmt.Errorf("failed to request upload URL: %w", err)
	}

	// Convert headers to map
	headers := make(map[string]string)
	for _, h := range uploadInfo.Headers {
		headers[h.Key] = h.Value
	}

	// Upload file to pre-signed URL
	uploadFileInfo := &files.UploadFileInfo{
		UploadURL:   uploadInfo.UploadURL,
		AssetURL:    uploadInfo.AssetURL,
		Headers:     headers,
		ContentType: uploadInfo.ContentType,
		Size:        size,
	}

	err = files.UploadToPresignedURL(ctx, uploadFileInfo, fileContent)
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	// Return the asset URL that can be used in markdown
	return uploadInfo.AssetURL, nil
}

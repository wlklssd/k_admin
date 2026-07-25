package vbenapi

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	avatarMaxSize   = 2 << 20
	avatarUploadDir = "avatars"
)

type validatedAvatar struct {
	Body        []byte
	ContentType string
	Name        string
	ObjectKey   string
}

func validateAvatar(file *multipart.FileHeader) (validatedAvatar, error) {
	if file == nil {
		return validatedAvatar{}, fmt.Errorf("avatar file is required")
	}
	if file.Size <= 0 || file.Size > avatarMaxSize {
		return validatedAvatar{}, fmt.Errorf("avatar image must be smaller than 2MB")
	}

	opened, err := file.Open()
	if err != nil {
		return validatedAvatar{}, fmt.Errorf("open avatar file failed")
	}
	defer opened.Close()

	body, err := io.ReadAll(io.LimitReader(opened, avatarMaxSize+1))
	if err != nil || len(body) == 0 || int64(len(body)) > avatarMaxSize {
		return validatedAvatar{}, fmt.Errorf("read avatar file failed")
	}

	contentType := http.DetectContentType(body)
	if !strings.HasPrefix(contentType, "image/") {
		return validatedAvatar{}, fmt.Errorf("avatar must be an image")
	}
	ext := safeImageExt(file.Filename, contentType)
	return validatedAvatar{
		Body:        body,
		ContentType: contentType,
		Name:        file.Filename,
		ObjectKey:   path.Join(avatarUploadDir, time.Now().Format("20060102"), uuid.NewString()+ext),
	}, nil
}

func safeImageExt(filename string, contentType string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp", ".svg":
		return ext
	}
	switch contentType {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	case "image/bmp":
		return ".bmp"
	default:
		return ".img"
	}
}

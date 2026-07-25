package vbenapi

import (
	"mime"
	"net/http"
	"path/filepath"

	"github.com/GoAdminGroup/go-admin/vbenapi/storage"
	"github.com/gin-gonic/gin"
)

func (s *Store) uploadUserAvatar(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		fail(c, http.StatusBadRequest, "avatar file is required")
		return
	}
	avatar, err := validateAvatar(file)
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}

	storageName, err := newFileServiceFromEnv().put(
		c.Request.Context(),
		avatar.ObjectKey,
		avatar.Body,
		avatar.ContentType,
	)
	if err != nil {
		fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	success(c, uploadedFileResponse{
		URL:     apiUploadURL(avatar.ObjectKey),
		Storage: storageName,
		Name:    avatar.Name,
	})
}

func (s *Store) serveUploadedFile(c *gin.Context) {
	objectKey, err := storage.CleanObjectKey(c.Param("path"))
	if err != nil {
		fail(c, http.StatusBadRequest, "invalid file path")
		return
	}

	service := newFileServiceFromEnv()
	if localPath, ok := service.localPath(objectKey); ok {
		c.File(localPath)
		return
	}

	body, info, err := service.openRemote(c.Request.Context(), objectKey)
	if err == nil {
		defer body.Close()
		contentType := info.ContentType
		if contentType == "" {
			contentType = mime.TypeByExtension(filepath.Ext(objectKey))
		}
		if contentType == "" {
			contentType = "application/octet-stream"
		}
		c.DataFromReader(http.StatusOK, -1, contentType, body, nil)
		return
	}

	fail(c, http.StatusNotFound, "file not found")
}

func apiUploadURL(objectKey string) string {
	return "/api/uploads/" + storage.EscapePath(objectKey)
}

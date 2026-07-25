package vbenapi

import (
	"errors"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/GoAdminGroup/go-admin/vbenapi/storage"
	"github.com/gin-gonic/gin"
)

func registerFileRoutes(api *gin.RouterGroup, s *Store) {
	files := api.Group("/files", s.requireAuth())
	files.POST("", s.requirePermission(fileUploadPermission), s.uploadManagedFile)
	files.GET("/:id", s.requirePermission(fileReadPermission), s.getManagedFileMetadata)
	files.GET("/:id/content", s.requirePermission(fileReadPermission), s.serveManagedFileContent)
	files.DELETE("/:id", s.requirePermission(fileDeletePermission), s.deleteManagedFile)
}

func (s *Store) uploadManagedFile(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxRequestSize)
	file, err := c.FormFile("file")
	if c.Request.MultipartForm != nil {
		defer c.Request.MultipartForm.RemoveAll()
	}
	if err != nil {
		if requestTooLarge(err) {
			fail(c, http.StatusRequestEntityTooLarge, "file is too large")
			return
		}
		fail(c, http.StatusBadRequest, "file is required")
		return
	}
	upload, err := validateManagedFile(file, c.PostForm("purpose"))
	if err != nil {
		var requestErr *fileRequestError
		if errors.As(err, &requestErr) {
			fail(c, requestErr.Status, requestErr.Message)
			return
		}
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	userID, ok := currentUserID(c)
	if !ok {
		fail(c, http.StatusUnauthorized, "invalid token")
		return
	}
	service := newManagedFileService(loadFileStorageSettings(), newDatabaseFileRepository(s.conn))
	record, err := service.createManagedFile(c.Request.Context(), upload, userID)
	if err != nil {
		failManagedFileService(c, err)
		return
	}
	success(c, fileResponse(record))
}

func (s *Store) getManagedFileMetadata(c *gin.Context) {
	id, ok := managedFileID(c)
	if !ok {
		return
	}
	record, err := newManagedFileService(loadFileStorageSettings(), newDatabaseFileRepository(s.conn)).getManagedFile(id)
	if err != nil {
		failManagedFileService(c, err)
		return
	}
	success(c, fileResponse(record))
}

func (s *Store) serveManagedFileContent(c *gin.Context) {
	id, ok := managedFileID(c)
	if !ok {
		return
	}
	body, info, record, err := newManagedFileService(loadFileStorageSettings(), newDatabaseFileRepository(s.conn)).openManagedFile(c.Request.Context(), id)
	if err != nil {
		failManagedFileService(c, err)
		return
	}
	defer body.Close()
	contentType := record.ContentType
	if contentType == "" {
		contentType = info.ContentType
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	c.Header("X-Content-Type-Options", "nosniff")
	if !strings.HasPrefix(contentType, "image/") {
		c.Header("Content-Disposition", contentDispositionAttachment(record.OriginalName))
	}
	size := record.Size
	if info.Size > 0 {
		size = info.Size
	}
	c.DataFromReader(http.StatusOK, size, contentType, body, nil)
}

func (s *Store) deleteManagedFile(c *gin.Context) {
	id, ok := managedFileID(c)
	if !ok {
		return
	}
	err := newManagedFileService(loadFileStorageSettings(), newDatabaseFileRepository(s.conn)).deleteManagedFile(c.Request.Context(), id)
	if err != nil {
		failManagedFileService(c, err)
		return
	}
	success(c, true)
}

func managedFileID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		fail(c, http.StatusBadRequest, "invalid file id")
		return 0, false
	}
	return id, true
}

func currentUserID(c *gin.Context) (int64, bool) {
	value, ok := c.Get("vben_user_id")
	if !ok {
		return 0, false
	}
	userID, ok := value.(int64)
	return userID, ok && userID > 0
}

func failManagedFileService(c *gin.Context, err error) {
	switch {
	case errors.Is(err, errFileNotFound), errors.Is(err, errFileContentNotFound):
		fail(c, http.StatusNotFound, "file not found")
	case errors.Is(err, errFileStorageUnavailable):
		fail(c, http.StatusServiceUnavailable, "file storage unavailable")
	default:
		fail(c, http.StatusInternalServerError, "file metadata unavailable")
	}
}

func requestTooLarge(err error) bool {
	return err != nil && (strings.Contains(strings.ToLower(err.Error()), "request body too large") || strings.Contains(strings.ToLower(err.Error()), "message too large"))
}

func managedFileContentURL(id int64) string {
	return "/api/files/" + strconv.FormatInt(id, 10) + "/content"
}

func contentDispositionAttachment(filename string) string {
	disposition := mime.FormatMediaType("attachment", map[string]string{"filename": filename})
	if disposition == "" {
		return "attachment"
	}
	return disposition
}

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

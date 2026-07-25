package vbenapi

import (
	"bytes"
	"context"
	"io"
	"os"

	"github.com/GoAdminGroup/go-admin/vbenapi/storage"
)

const (
	localStorageName = "local"
	minioStorageName = "minio"
)

type fileService struct {
	local  *storage.Local
	remote storage.ObjectStore
}

func newFileService(settings fileStorageSettings) *fileService {
	service := &fileService{local: storage.NewLocal(settings.LocalRoot)}
	if settings.MinioEnabled {
		service.remote = storage.NewMinio(settings.Minio)
	}
	return service
}

func newFileServiceFromEnv() *fileService {
	return newFileService(loadFileStorageSettings())
}

func (s *fileService) put(ctx context.Context, objectKey string, body []byte, contentType string) (string, error) {
	if s.remote != nil {
		if err := s.remote.Put(ctx, objectKey, bytes.NewReader(body), int64(len(body)), contentType); err == nil {
			return minioStorageName, nil
		}
	}
	if err := s.local.Put(ctx, objectKey, bytes.NewReader(body), int64(len(body)), contentType); err != nil {
		return "", err
	}
	return localStorageName, nil
}

func (s *fileService) localPath(objectKey string) (string, bool) {
	localPath, err := s.local.Path(objectKey)
	if err != nil {
		return "", false
	}
	if _, err := os.Stat(localPath); err != nil {
		return "", false
	}
	return localPath, true
}

func (s *fileService) openRemote(ctx context.Context, objectKey string) (io.ReadCloser, storage.ObjectInfo, error) {
	if s.remote == nil {
		return nil, storage.ObjectInfo{}, os.ErrNotExist
	}
	return s.remote.Open(ctx, objectKey)
}

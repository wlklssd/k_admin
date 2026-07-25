package storage

import (
	"context"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strings"
)

type Local struct {
	root string
}

func NewLocal(root string) *Local {
	return &Local{root: root}
}

func (l *Local) Put(_ context.Context, objectKey string, body io.Reader, _ int64, _ string) error {
	target, err := l.Path(objectKey)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return err
	}
	file, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(file, body)
	closeErr := file.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}

func (l *Local) Open(_ context.Context, objectKey string) (io.ReadCloser, ObjectInfo, error) {
	target, err := l.Path(objectKey)
	if err != nil {
		return nil, ObjectInfo{}, err
	}
	file, err := os.Open(target)
	if err != nil {
		return nil, ObjectInfo{}, err
	}
	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, ObjectInfo{}, err
	}
	return file, ObjectInfo{
		ContentType: mime.TypeByExtension(filepath.Ext(target)),
		Size:        info.Size(),
	}, nil
}

func (l *Local) Delete(_ context.Context, objectKey string) error {
	target, err := l.Path(objectKey)
	if err != nil {
		return err
	}
	return os.Remove(target)
}

func (l *Local) Path(objectKey string) (string, error) {
	cleaned, err := CleanObjectKey(objectKey)
	if err != nil {
		return "", err
	}
	root, err := filepath.Abs(l.root)
	if err != nil {
		return "", err
	}
	target, err := filepath.Abs(filepath.Join(root, filepath.FromSlash(cleaned)))
	if err != nil {
		return "", err
	}
	if target != root && !strings.HasPrefix(target, root+string(os.PathSeparator)) {
		return "", ErrInvalidObjectKey
	}
	return target, nil
}

package vbenapi

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/GoAdminGroup/go-admin/vbenapi/storage"
)

type fakeObjectStore struct {
	putBody  []byte
	putCalls int
	putErr   error
}

func (f *fakeObjectStore) Put(_ context.Context, _ string, body io.Reader, _ int64, _ string) error {
	f.putCalls++
	f.putBody, _ = io.ReadAll(body)
	return f.putErr
}

func (f *fakeObjectStore) Open(context.Context, string) (io.ReadCloser, storage.ObjectInfo, error) {
	return nil, storage.ObjectInfo{}, os.ErrNotExist
}

func (f *fakeObjectStore) Delete(context.Context, string) error {
	return nil
}

func TestFileServicePrefersRemoteStore(t *testing.T) {
	remote := &fakeObjectStore{}
	service := &fileService{
		local:  storage.NewLocal(t.TempDir()),
		remote: remote,
	}
	payload := []byte("avatar")

	storageName, err := service.put(context.Background(), "avatars/avatar.png", payload, "image/png")
	if err != nil {
		t.Fatalf("put remote object: %v", err)
	}
	if storageName != minioStorageName || remote.putCalls != 1 || !bytes.Equal(remote.putBody, payload) {
		t.Fatalf("unexpected remote upload result: storage=%q calls=%d body=%q", storageName, remote.putCalls, remote.putBody)
	}
	if _, ok := service.localPath("avatars/avatar.png"); ok {
		t.Fatal("did not expect a local fallback after remote success")
	}
}

func TestFileServiceFallsBackToLocalStore(t *testing.T) {
	remote := &fakeObjectStore{putErr: errors.New("remote unavailable")}
	service := &fileService{
		local:  storage.NewLocal(t.TempDir()),
		remote: remote,
	}
	payload := []byte("avatar")

	storageName, err := service.put(context.Background(), "avatars/avatar.png", payload, "image/png")
	if err != nil {
		t.Fatalf("put fallback object: %v", err)
	}
	if storageName != localStorageName || remote.putCalls != 1 {
		t.Fatalf("unexpected fallback result: storage=%q calls=%d", storageName, remote.putCalls)
	}
	localPath, ok := service.localPath("avatars/avatar.png")
	if !ok {
		t.Fatal("expected local fallback file")
	}
	actual, err := os.ReadFile(localPath)
	if err != nil {
		t.Fatalf("read fallback file: %v", err)
	}
	if !bytes.Equal(actual, payload) {
		t.Fatalf("unexpected fallback payload: %q", actual)
	}
}

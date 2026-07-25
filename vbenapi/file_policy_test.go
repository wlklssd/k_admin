package vbenapi

import (
	"bytes"
	"mime/multipart"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestValidateAvatarAcceptsDetectedImage(t *testing.T) {
	header := testMultipartFileHeader(t, "avatar.txt", []byte("\x89PNG\r\n\x1a\nimage"))

	avatar, err := validateAvatar(header)
	if err != nil {
		t.Fatalf("validate image avatar: %v", err)
	}
	if avatar.ContentType != "image/png" {
		t.Fatalf("unexpected content type: %q", avatar.ContentType)
	}
	if !strings.HasPrefix(avatar.ObjectKey, "avatars/") || !strings.HasSuffix(avatar.ObjectKey, ".png") {
		t.Fatalf("unexpected avatar object key: %q", avatar.ObjectKey)
	}
	if avatar.Name != "avatar.txt" {
		t.Fatalf("unexpected original name: %q", avatar.Name)
	}
}

func TestValidateAvatarRejectsNonImage(t *testing.T) {
	header := testMultipartFileHeader(t, "avatar.png", []byte("plain text"))

	_, err := validateAvatar(header)
	if err == nil || err.Error() != "avatar must be an image" {
		t.Fatalf("unexpected validation error: %v", err)
	}
}

func TestValidateAvatarRejectsOversizedFile(t *testing.T) {
	header := testMultipartFileHeader(t, "avatar.png", make([]byte, avatarMaxSize+1))

	_, err := validateAvatar(header)
	if err == nil || err.Error() != "avatar image must be smaller than 2MB" {
		t.Fatalf("unexpected validation error: %v", err)
	}
}

func TestSafeImageExt(t *testing.T) {
	tests := []struct {
		filename    string
		contentType string
		want        string
	}{
		{filename: "avatar.JPEG", contentType: "image/jpeg", want: ".jpeg"},
		{filename: "avatar.txt", contentType: "image/webp", want: ".webp"},
		{filename: "avatar", contentType: "image/unknown", want: ".img"},
	}
	for _, test := range tests {
		if actual := safeImageExt(test.filename, test.contentType); actual != test.want {
			t.Fatalf("safeImageExt(%q, %q) = %q, want %q", test.filename, test.contentType, actual, test.want)
		}
	}
}

func testMultipartFileHeader(t *testing.T, filename string, payload []byte) *multipart.FileHeader {
	t.Helper()
	body := bytes.NewBuffer(nil)
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("create multipart file: %v", err)
	}
	if _, err := part.Write(payload); err != nil {
		t.Fatalf("write multipart file: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	request := httptest.NewRequest("POST", "/api/users/avatar", body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	if err := request.ParseMultipartForm(int64(len(payload)) + 1024); err != nil {
		t.Fatalf("parse multipart form: %v", err)
	}
	t.Cleanup(func() {
		_ = request.MultipartForm.RemoveAll()
	})
	file, header, err := request.FormFile("file")
	if err != nil {
		t.Fatalf("read multipart file: %v", err)
	}
	file.Close()
	return header
}

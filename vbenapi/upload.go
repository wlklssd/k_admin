package vbenapi

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	avatarMaxSize   = 2 << 20
	uploadRoot      = "upload"
	avatarUploadDir = "avatars"
	minioRegion     = "us-east-1"
)

type uploadedFileResponse struct {
	URL     string `json:"url"`
	Storage string `json:"storage"`
	Name    string `json:"name"`
}

type minioSettings struct {
	Enabled    bool
	Endpoints  []string
	AccessKey  string
	SecretKey  string
	Bucket     string
	UseSSL     bool
	Region     string
	PublicBase string
}

func (s *Store) uploadUserAvatar(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		fail(c, http.StatusBadRequest, "avatar file is required")
		return
	}
	if file.Size <= 0 || file.Size > avatarMaxSize {
		fail(c, http.StatusBadRequest, "avatar image must be smaller than 2MB")
		return
	}

	opened, err := file.Open()
	if err != nil {
		fail(c, http.StatusBadRequest, "open avatar file failed")
		return
	}
	defer opened.Close()

	body, err := io.ReadAll(io.LimitReader(opened, avatarMaxSize+1))
	if err != nil || len(body) == 0 || int64(len(body)) > avatarMaxSize {
		fail(c, http.StatusBadRequest, "read avatar file failed")
		return
	}

	contentType := http.DetectContentType(body)
	if !strings.HasPrefix(contentType, "image/") {
		fail(c, http.StatusBadRequest, "avatar must be an image")
		return
	}
	ext := safeImageExt(file.Filename, contentType)
	objectKey := path.Join(avatarUploadDir, time.Now().Format("20060102"), uuid.NewString()+ext)

	settings := loadMinioSettings()
	if settings.Enabled {
		if err := uploadToFirstAvailableMinio(c.Request.Context(), settings, objectKey, body, contentType); err == nil {
			success(c, uploadedFileResponse{
				URL:     apiUploadURL(objectKey),
				Storage: "minio",
				Name:    file.Filename,
			})
			return
		}
	}

	if err := saveLocalUpload(objectKey, body); err != nil {
		fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	success(c, uploadedFileResponse{
		URL:     apiUploadURL(objectKey),
		Storage: "local",
		Name:    file.Filename,
	})
}

func (s *Store) serveUploadedFile(c *gin.Context) {
	objectKey, ok := cleanObjectKey(c.Param("path"))
	if !ok {
		fail(c, http.StatusBadRequest, "invalid file path")
		return
	}

	if localPath, ok := localUploadPath(objectKey); ok {
		if _, err := os.Stat(localPath); err == nil {
			c.File(localPath)
			return
		}
	}

	settings := loadMinioSettings()
	if settings.Enabled {
		body, contentType, err := getFromFirstAvailableMinio(c.Request.Context(), settings, objectKey)
		if err == nil {
			defer body.Close()
			if contentType == "" {
				contentType = mime.TypeByExtension(filepath.Ext(objectKey))
			}
			if contentType == "" {
				contentType = "application/octet-stream"
			}
			c.DataFromReader(http.StatusOK, -1, contentType, body, nil)
			return
		}
	}

	fail(c, http.StatusNotFound, "file not found")
}

func saveLocalUpload(objectKey string, body []byte) error {
	localPath, ok := localUploadPath(objectKey)
	if !ok {
		return fmt.Errorf("invalid file path")
	}
	if err := os.MkdirAll(filepath.Dir(localPath), 0755); err != nil {
		return err
	}
	return os.WriteFile(localPath, body, 0644)
}

func localUploadPath(objectKey string) (string, bool) {
	cleaned, ok := cleanObjectKey(objectKey)
	if !ok {
		return "", false
	}
	root, err := filepath.Abs(uploadRoot)
	if err != nil {
		return "", false
	}
	target, err := filepath.Abs(filepath.Join(root, filepath.FromSlash(cleaned)))
	if err != nil {
		return "", false
	}
	if target != root && !strings.HasPrefix(target, root+string(os.PathSeparator)) {
		return "", false
	}
	return target, true
}

func cleanObjectKey(value string) (string, bool) {
	value = strings.TrimSpace(strings.TrimPrefix(value, "/"))
	cleaned := path.Clean("/" + value)
	cleaned = strings.TrimPrefix(cleaned, "/")
	if cleaned == "." || cleaned == "" || strings.HasPrefix(cleaned, "../") || strings.Contains(cleaned, "\x00") {
		return "", false
	}
	return cleaned, true
}

func apiUploadURL(objectKey string) string {
	return "/api/uploads/" + escapePath(objectKey)
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

func loadMinioSettings() minioSettings {
	return minioSettings{
		Enabled:    envBool("KADMIN_MINIO_ENABLED", false),
		Endpoints:  uniqueStrings([]string{envString("KADMIN_MINIO_ENDPOINT", "127.0.0.1:19000"), envString("KADMIN_MINIO_INTERNAL_ENDPOINT", "minio:9000")}),
		AccessKey:  envString("KADMIN_MINIO_ACCESS_KEY", "kadmin_minio"),
		SecretKey:  envString("KADMIN_MINIO_SECRET_KEY", "kadmin_minio_pwd"),
		Bucket:     envString("KADMIN_MINIO_BUCKET", "kadmin"),
		UseSSL:     envBool("KADMIN_MINIO_USE_SSL", false),
		Region:     envString("KADMIN_MINIO_REGION", minioRegion),
		PublicBase: envString("KADMIN_MINIO_PUBLIC_BASE", ""),
	}
}

func uploadToFirstAvailableMinio(ctx context.Context, settings minioSettings, objectKey string, body []byte, contentType string) error {
	if settings.AccessKey == "" || settings.SecretKey == "" || settings.Bucket == "" {
		return fmt.Errorf("minio config is incomplete")
	}
	var lastErr error
	for _, endpoint := range settings.Endpoints {
		client := newMinioHTTPClient(settings, endpoint)
		if client == nil {
			continue
		}
		if err := client.health(ctx); err != nil {
			lastErr = err
			continue
		}
		if err := client.ensureBucket(ctx); err != nil {
			lastErr = err
			continue
		}
		if err := client.putObject(ctx, objectKey, body, contentType); err != nil {
			lastErr = err
			continue
		}
		return nil
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("minio endpoint is not configured")
	}
	return lastErr
}

func getFromFirstAvailableMinio(ctx context.Context, settings minioSettings, objectKey string) (io.ReadCloser, string, error) {
	if settings.AccessKey == "" || settings.SecretKey == "" || settings.Bucket == "" {
		return nil, "", fmt.Errorf("minio config is incomplete")
	}
	var lastErr error
	for _, endpoint := range settings.Endpoints {
		client := newMinioHTTPClient(settings, endpoint)
		if client == nil {
			continue
		}
		body, contentType, err := client.getObject(ctx, objectKey)
		if err != nil {
			lastErr = err
			continue
		}
		return body, contentType, nil
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("minio endpoint is not configured")
	}
	return nil, "", lastErr
}

type minioHTTPClient struct {
	settings minioSettings
	baseURL  string
	client   *http.Client
}

func newMinioHTTPClient(settings minioSettings, endpoint string) *minioHTTPClient {
	endpoint = strings.TrimSpace(endpoint)
	if endpoint == "" {
		return nil
	}
	if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		scheme := "http"
		if settings.UseSSL {
			scheme = "https"
		}
		endpoint = scheme + "://" + endpoint
	}
	endpoint = strings.TrimRight(endpoint, "/")
	if _, err := url.ParseRequestURI(endpoint); err != nil {
		return nil
	}
	return &minioHTTPClient{
		settings: settings,
		baseURL:  endpoint,
		client: &http.Client{
			Timeout: 3 * time.Second,
		},
	}
}

func (m *minioHTTPClient) health(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, m.baseURL+"/minio/health/live", nil)
	if err != nil {
		return err
	}
	resp, err := m.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("minio health check failed: %s", resp.Status)
	}
	return nil
}

func (m *minioHTTPClient) ensureBucket(ctx context.Context) error {
	resp, err := m.signedRequest(ctx, http.MethodHead, "/"+escapePath(m.settings.Bucket), nil, "")
	if err != nil {
		return err
	}
	resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	if resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("minio bucket check failed: %s", resp.Status)
	}

	resp, err = m.signedRequest(ctx, http.MethodPut, "/"+escapePath(m.settings.Bucket), nil, "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("minio bucket create failed: %s", resp.Status)
	}
	return nil
}

func (m *minioHTTPClient) putObject(ctx context.Context, objectKey string, body []byte, contentType string) error {
	resp, err := m.signedRequest(
		ctx,
		http.MethodPut,
		"/"+escapePath(m.settings.Bucket)+"/"+escapePath(objectKey),
		bytes.NewReader(body),
		contentType,
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("minio upload failed: %s", resp.Status)
	}
	return nil
}

func (m *minioHTTPClient) getObject(ctx context.Context, objectKey string) (io.ReadCloser, string, error) {
	resp, err := m.signedRequest(
		ctx,
		http.MethodGet,
		"/"+escapePath(m.settings.Bucket)+"/"+escapePath(objectKey),
		nil,
		"",
	)
	if err != nil {
		return nil, "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		resp.Body.Close()
		return nil, "", fmt.Errorf("minio get failed: %s", resp.Status)
	}
	return resp.Body, resp.Header.Get("Content-Type"), nil
}

func (m *minioHTTPClient) signedRequest(ctx context.Context, method string, requestPath string, body io.Reader, contentType string) (*http.Response, error) {
	var payload []byte
	var err error
	if body != nil {
		payload, err = io.ReadAll(body)
		if err != nil {
			return nil, err
		}
	}

	endpoint := m.baseURL + requestPath
	req, err := http.NewRequestWithContext(ctx, method, endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	m.sign(req, payload)
	return m.client.Do(req)
}

func (m *minioHTTPClient) sign(req *http.Request, payload []byte) {
	now := time.Now().UTC()
	date := now.Format("20060102")
	amzDate := now.Format("20060102T150405Z")
	payloadHash := sha256Hex(payload)

	req.Header.Set("X-Amz-Date", amzDate)
	req.Header.Set("X-Amz-Content-Sha256", payloadHash)

	signedHeaders := []string{"host", "x-amz-content-sha256", "x-amz-date"}
	canonicalHeaders := strings.Builder{}
	canonicalHeaders.WriteString("host:")
	canonicalHeaders.WriteString(req.URL.Host)
	canonicalHeaders.WriteByte('\n')
	canonicalHeaders.WriteString("x-amz-content-sha256:")
	canonicalHeaders.WriteString(payloadHash)
	canonicalHeaders.WriteByte('\n')
	canonicalHeaders.WriteString("x-amz-date:")
	canonicalHeaders.WriteString(amzDate)
	canonicalHeaders.WriteByte('\n')

	query := req.URL.Query()
	canonicalQuery := canonicalQueryString(query)
	scope := strings.Join([]string{date, m.settings.Region, "s3", "aws4_request"}, "/")
	canonicalRequest := strings.Join([]string{
		req.Method,
		req.URL.EscapedPath(),
		canonicalQuery,
		canonicalHeaders.String(),
		strings.Join(signedHeaders, ";"),
		payloadHash,
	}, "\n")

	stringToSign := strings.Join([]string{
		"AWS4-HMAC-SHA256",
		amzDate,
		scope,
		sha256Hex([]byte(canonicalRequest)),
	}, "\n")
	signingKey := awsSigningKey(m.settings.SecretKey, date, m.settings.Region, "s3")
	signature := hex.EncodeToString(hmacSHA256(signingKey, stringToSign))
	req.Header.Set("Authorization", fmt.Sprintf(
		"AWS4-HMAC-SHA256 Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		m.settings.AccessKey,
		scope,
		strings.Join(signedHeaders, ";"),
		signature,
	))
}

func canonicalQueryString(values url.Values) string {
	if len(values) == 0 {
		return ""
	}
	parts := make([]string, 0)
	for key, vals := range values {
		sort.Strings(vals)
		for _, value := range vals {
			parts = append(parts, url.QueryEscape(key)+"="+url.QueryEscape(value))
		}
	}
	sort.Strings(parts)
	return strings.Join(parts, "&")
}

func awsSigningKey(secret string, date string, region string, service string) []byte {
	kDate := hmacSHA256([]byte("AWS4"+secret), date)
	kRegion := hmacSHA256(kDate, region)
	kService := hmacSHA256(kRegion, service)
	return hmacSHA256(kService, "aws4_request")
}

func hmacSHA256(key []byte, value string) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(value))
	return mac.Sum(nil)
}

func sha256Hex(value []byte) string {
	sum := sha256.Sum256(value)
	return hex.EncodeToString(sum[:])
}

func escapePath(value string) string {
	segments := strings.Split(strings.Trim(value, "/"), "/")
	for i, segment := range segments {
		segments[i] = url.PathEscape(segment)
	}
	return strings.Join(segments, "/")
}

func envString(key string, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func envBool(key string, fallback bool) bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(key))) {
	case "1", "true", "t", "yes", "y", "on":
		return true
	case "0", "false", "f", "no", "n", "off":
		return false
	default:
		return fallback
	}
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]bool)
	res := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		res = append(res, value)
	}
	return res
}

package storage

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

type MinioConfig struct {
	Endpoints  []string
	AccessKey  string
	SecretKey  string
	Bucket     string
	UseSSL     bool
	Region     string
	PublicBase string
	Timeout    time.Duration
}

type Minio struct {
	config MinioConfig
}

func NewMinio(config MinioConfig) *Minio {
	return &Minio{config: config}
}

func (m *Minio) Put(ctx context.Context, objectKey string, body io.Reader, _ int64, contentType string) error {
	if err := m.validateConfig(); err != nil {
		return err
	}
	payload, err := io.ReadAll(body)
	if err != nil {
		return err
	}
	var lastErr error
	for _, endpoint := range m.config.Endpoints {
		client := newMinioHTTPClient(m.config, endpoint)
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
		if err := client.putObject(ctx, objectKey, payload, contentType); err != nil {
			lastErr = err
			continue
		}
		return nil
	}
	return minioEndpointError(lastErr)
}

func (m *Minio) Open(ctx context.Context, objectKey string) (io.ReadCloser, ObjectInfo, error) {
	if err := m.validateConfig(); err != nil {
		return nil, ObjectInfo{}, err
	}
	var lastErr error
	for _, endpoint := range m.config.Endpoints {
		client := newMinioHTTPClient(m.config, endpoint)
		if client == nil {
			continue
		}
		body, info, err := client.getObject(ctx, objectKey)
		if err != nil {
			lastErr = err
			continue
		}
		return body, info, nil
	}
	return nil, ObjectInfo{}, minioEndpointError(lastErr)
}

func (m *Minio) Delete(ctx context.Context, objectKey string) error {
	if err := m.validateConfig(); err != nil {
		return err
	}
	var lastErr error
	for _, endpoint := range m.config.Endpoints {
		client := newMinioHTTPClient(m.config, endpoint)
		if client == nil {
			continue
		}
		if err := client.deleteObject(ctx, objectKey); err != nil {
			lastErr = err
			continue
		}
		return nil
	}
	return minioEndpointError(lastErr)
}

func (m *Minio) validateConfig() error {
	if m.config.AccessKey == "" || m.config.SecretKey == "" || m.config.Bucket == "" {
		return fmt.Errorf("minio config is incomplete")
	}
	return nil
}

func minioEndpointError(lastErr error) error {
	if lastErr != nil {
		return lastErr
	}
	return fmt.Errorf("minio endpoint is not configured")
}

type minioHTTPClient struct {
	config  MinioConfig
	baseURL string
	client  *http.Client
}

func newMinioHTTPClient(config MinioConfig, endpoint string) *minioHTTPClient {
	endpoint = strings.TrimSpace(endpoint)
	if endpoint == "" {
		return nil
	}
	if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		scheme := "http"
		if config.UseSSL {
			scheme = "https"
		}
		endpoint = scheme + "://" + endpoint
	}
	endpoint = strings.TrimRight(endpoint, "/")
	if _, err := url.ParseRequestURI(endpoint); err != nil {
		return nil
	}
	timeout := config.Timeout
	if timeout <= 0 {
		timeout = 3 * time.Second
	}
	return &minioHTTPClient{
		config:  config,
		baseURL: endpoint,
		client:  &http.Client{Timeout: timeout},
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
	resp, err := m.signedRequest(ctx, http.MethodHead, "/"+EscapePath(m.config.Bucket), nil, "")
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

	resp, err = m.signedRequest(ctx, http.MethodPut, "/"+EscapePath(m.config.Bucket), nil, "")
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
		"/"+EscapePath(m.config.Bucket)+"/"+EscapePath(objectKey),
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

func (m *minioHTTPClient) getObject(ctx context.Context, objectKey string) (io.ReadCloser, ObjectInfo, error) {
	resp, err := m.signedRequest(
		ctx,
		http.MethodGet,
		"/"+EscapePath(m.config.Bucket)+"/"+EscapePath(objectKey),
		nil,
		"",
	)
	if err != nil {
		return nil, ObjectInfo{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		resp.Body.Close()
		return nil, ObjectInfo{}, fmt.Errorf("minio get failed: %s", resp.Status)
	}
	return resp.Body, ObjectInfo{
		ContentType: resp.Header.Get("Content-Type"),
		Size:        resp.ContentLength,
	}, nil
}

func (m *minioHTTPClient) deleteObject(ctx context.Context, objectKey string) error {
	resp, err := m.signedRequest(
		ctx,
		http.MethodDelete,
		"/"+EscapePath(m.config.Bucket)+"/"+EscapePath(objectKey),
		nil,
		"",
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("minio delete failed: %s", resp.Status)
	}
	return nil
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

	canonicalQuery := canonicalQueryString(req.URL.Query())
	scope := strings.Join([]string{date, m.config.Region, "s3", "aws4_request"}, "/")
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
	signingKey := awsSigningKey(m.config.SecretKey, date, m.config.Region, "s3")
	signature := hex.EncodeToString(hmacSHA256(signingKey, stringToSign))
	req.Header.Set("Authorization", fmt.Sprintf(
		"AWS4-HMAC-SHA256 Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		m.config.AccessKey,
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

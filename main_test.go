package main

import (
	"bytes"
	"net"
	"strings"
	"testing"
)

func TestListenHTTPReportsAddressInUse(t *testing.T) {
	occupied, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("reserve test address: %v", err)
	}
	defer occupied.Close()

	listener, err := listenHTTP(occupied.Addr().String())
	if listener != nil {
		_ = listener.Close()
	}
	if err == nil {
		t.Fatal("expected an address-in-use error")
	}
	if !strings.Contains(err.Error(), "HTTP 服务监听") {
		t.Fatalf("unexpected startup error: %v", err)
	}
}

func TestSetupBackendConvertsInitializationPanicToError(t *testing.T) {
	_, _, err := setupBackend(nil, nil, nil)
	if err == nil {
		t.Fatal("expected initialization error")
	}
	if !strings.Contains(err.Error(), "初始化 GoAdmin 公共组件失败") {
		t.Fatalf("unexpected initialization error: %v", err)
	}
}

func TestGinDebugErrorWriterOnlyFiltersNonErrorDebugMessages(t *testing.T) {
	var output bytes.Buffer
	writer := ginDebugErrorWriter{destination: &output}

	messages := []string{
		"[GIN-debug] [WARNING] debug mode\n",
		"[GIN-debug] GET /api\n",
		"[GIN-debug] [ERROR] listen failed\n",
		"KAdmin 后端服务启动成功\n",
	}
	for _, message := range messages {
		written, err := writer.Write([]byte(message))
		if err != nil {
			t.Fatalf("write message: %v", err)
		}
		if written != len(message) {
			t.Fatalf("written bytes = %d, want %d", written, len(message))
		}
	}

	want := "[GIN-debug] [ERROR] listen failed\nKAdmin 后端服务启动成功\n"
	if output.String() != want {
		t.Fatalf("filtered output = %q, want %q", output.String(), want)
	}
}

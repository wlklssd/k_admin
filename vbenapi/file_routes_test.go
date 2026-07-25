package vbenapi

import (
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRegisterExistingUploadRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	registerUserManagementRoutes(engine.Group("/api"), &Store{})

	wanted := map[string]bool{
		"GET /api/uploads/*path": false,
		"POST /api/users/avatar": false,
	}
	for _, route := range engine.Routes() {
		key := route.Method + " " + route.Path
		if _, exists := wanted[key]; exists {
			wanted[key] = true
		}
	}
	for route, registered := range wanted {
		if !registered {
			t.Fatalf("route %s was not registered", route)
		}
	}
}

func TestAPIUploadURLEscapesObjectKeySegments(t *testing.T) {
	actual := apiUploadURL("avatars/20260725/avatar name.png")
	want := "/api/uploads/avatars/20260725/avatar%20name.png"
	if actual != want {
		t.Fatalf("apiUploadURL() = %q, want %q", actual, want)
	}
}

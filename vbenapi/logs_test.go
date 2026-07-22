package vbenapi

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/GoAdminGroup/go-admin/plugins/admin/models"
	"github.com/gin-gonic/gin"
)

func TestRegisterLogRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	registerLogRoutes(engine.Group("/api"), &Store{})

	wanted := map[string]bool{
		"DELETE /api/logs":     false,
		"DELETE /api/logs/:id": false,
		"GET /api/logs":        false,
		"GET /api/logs/:id":    false,
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

func TestParseManagedLogFilter(t *testing.T) {
	context, _ := gin.CreateTestContext(httptest.NewRecorder())
	context.Request = httptest.NewRequest(
		"GET",
		"/api/logs?page=2&pageSize=500&eventType=request&level=warn&method=post&success=false&statusCode=401&startedAt=2026-07-22T00:00:00Z&endedAt=2026-07-22T12:00:00Z",
		nil,
	)

	filter, err := parseManagedLogFilter(context)
	if err != nil {
		t.Fatalf("parse valid filter: %v", err)
	}
	if filter.Page != 2 || filter.PageSize != maxLogPageSize || filter.Method != "POST" {
		t.Fatalf("unexpected pagination or method: %#v", filter)
	}
	if filter.Success == nil || *filter.Success || filter.StatusCode == nil || *filter.StatusCode != 401 {
		t.Fatalf("unexpected result filters: %#v", filter)
	}

	where, args := managedLogWhere(filter)
	for _, condition := range []string{"l.event_type = ?", "l.level = ?", "l.method = ?", "l.success = ?", "l.status_code = ?", "l.occurred_at >= ?", "l.occurred_at <= ?"} {
		if !strings.Contains(where, condition) {
			t.Fatalf("missing condition %q in %q", condition, where)
		}
	}
	if len(args) != 7 {
		t.Fatalf("expected 7 query arguments, got %d", len(args))
	}
}

func TestParseManagedLogFilterRejectsInvalidValues(t *testing.T) {
	for _, query := range []string{
		"eventType=unknown",
		"level=notice",
		"success=maybe",
		"statusCode=99",
		"startedAt=not-a-date",
		"startedAt=2026-07-23T00:00:00Z&endedAt=2026-07-22T00:00:00Z",
	} {
		context, _ := gin.CreateTestContext(httptest.NewRecorder())
		context.Request = httptest.NewRequest("GET", "/api/logs?"+query, nil)
		if _, err := parseManagedLogFilter(context); err == nil {
			t.Fatalf("expected query %q to fail validation", query)
		}
	}
}

func TestUserHasLogPermission(t *testing.T) {
	user := models.UserModel{Permissions: []models.PermissionModel{{Slug: logListPermission}}}
	if !userHasPermission(user, logListPermission) {
		t.Fatal("expected matching log permission to be accepted")
	}
	if userHasPermission(user, logDeletePermission) {
		t.Fatal("did not expect an unrelated permission to be accepted")
	}
	user.Permissions = []models.PermissionModel{{Slug: "*"}}
	if !userHasPermission(user, logDeletePermission) {
		t.Fatal("expected wildcard permission to be accepted")
	}
}

func TestDefaultLogMenuBinding(t *testing.T) {
	binding, ok := vbenMenuRouteBindings["/kadmin/logs"]
	if !ok || binding.Component != "/kadmin/components/LogManagementView" {
		t.Fatalf("unexpected log menu binding: %#v", binding)
	}
	for _, root := range defaultMenuSeeds {
		for _, child := range root.Children {
			if child.URI == "/kadmin/logs" && child.Order == 7 {
				return
			}
		}
	}
	t.Fatal("default log menu seed was not found")
}

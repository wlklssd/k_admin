package kadmin

import (
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRegisterApplicationRoutesIncludesEveryModule(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	registerApplicationRoutes(engine.Group("/api"), &Store{})

	wanted := map[string]bool{
		"POST /api/auth/login":             false,
		"GET /api/user/info":               false,
		"GET /api/menu/all":                false,
		"GET /api/rbac/overview":           false,
		"GET /api/admin-menus":             false,
		"GET /api/users":                   false,
		"GET /api/dictionaries/overview":   false,
		"POST /api/files":                  false,
		"POST /api/users/avatar":           false,
		"GET /api/system/config/login":     false,
		"GET /api/logs":                    false,
		"GET /api/jobs":                    false,
		"GET /api/job-logs":                false,
		"GET /api/system-monitor":          false,
		"PATCH /api/system-monitor/status": false,
	}
	for _, route := range engine.Routes() {
		key := route.Method + " " + route.Path
		if _, ok := wanted[key]; ok {
			wanted[key] = true
		}
	}
	for route, registered := range wanted {
		if !registered {
			t.Errorf("application route %s was not registered", route)
		}
	}
}

package kadmin

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/GoAdminGroup/go-admin/internal/kadmin/modules/files"
	"github.com/GoAdminGroup/go-admin/internal/kadmin/modules/jobs"
	"github.com/GoAdminGroup/go-admin/internal/kadmin/modules/loginlogs"
	"github.com/GoAdminGroup/go-admin/internal/kadmin/modules/monitor"
	"github.com/GoAdminGroup/go-admin/modules/db"
	"github.com/GoAdminGroup/go-admin/plugins/admin/models"
	"github.com/gin-gonic/gin"
)

// fakeMenuPermissionConnection 返回角色菜单绑定查询所需的菜单 URI，
// 模拟 goadmin_role_menu JOIN goadmin_menu 的结果。
type fakeMenuPermissionConnection struct {
	db.Connection
	uris []string
}

func (f *fakeMenuPermissionConnection) Name() string { return "sqlite" }

func (f *fakeMenuPermissionConnection) QueryWith(_ *sql.Tx, _ string, query string, _ ...interface{}) ([]map[string]interface{}, error) {
	if !strings.Contains(query, "goadmin_role_menu") {
		return nil, nil
	}
	rows := make([]map[string]interface{}, 0, len(f.uris))
	for _, uri := range f.uris {
		rows = append(rows, map[string]interface{}{"uri": uri})
	}
	return rows, nil
}

func TestMenuPermissionSlugsMatchSeedsAndBindings(t *testing.T) {
	seededURIs := make(map[string]bool)
	for _, root := range defaultMenuSeeds {
		seededURIs[root.URI] = true
		for _, child := range root.Children {
			seededURIs[child.URI] = true
		}
	}
	seededSlugs := make(map[string]bool)
	for _, seed := range defaultPermissionSeeds {
		seededSlugs[seed.Slug] = true
	}

	for uri, slugs := range menuPermissionSlugs {
		if _, ok := vbenMenuRouteBindings[uri]; !ok {
			t.Errorf("menu URI %q has no vben route binding", uri)
		}
		if !seededURIs[uri] {
			t.Errorf("menu URI %q has no menu seed", uri)
		}
		for _, slug := range slugs {
			if !seededSlugs[slug] {
				t.Errorf("menu URI %q maps to unseeded permission %q", uri, slug)
			}
		}
	}
}

func TestMenuPermissionSlugsForUser(t *testing.T) {
	store := &Store{conn: &fakeMenuPermissionConnection{uris: []string{"/kadmin", "/kadmin/jobs", "/unbound-page"}}}
	user := models.UserModel{Roles: []models.RoleModel{{Id: 2}}}

	slugs, err := store.menuPermissionSlugsForUser(user)
	if err != nil {
		t.Fatal(err)
	}
	for _, wanted := range []string{
		jobs.ListPermission,
		jobs.CreatePermission,
		jobs.UpdatePermission,
		jobs.DeletePermission,
		jobs.RunPermission,
		jobs.LogListPermission,
	} {
		if !slugs[wanted] {
			t.Errorf("expected menu /kadmin/jobs to grant %q", wanted)
		}
	}
	if slugs[files.UploadPermission] {
		t.Fatal("unrelated resource permission must not be granted")
	}

	empty, err := (&Store{conn: &fakeMenuPermissionConnection{}}).menuPermissionSlugsForUser(models.UserModel{})
	if err != nil {
		t.Fatal(err)
	}
	if len(empty) != 0 {
		t.Fatalf("user without roles must have no menu permissions, got %v", empty)
	}
}

func TestPermissionRequiredAcceptsMenuGrantedUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &Store{conn: &fakeMenuPermissionConnection{uris: []string{"/kadmin/jobs"}}}
	engine := gin.New()
	engine.GET("/api/jobs", store.permissionRequired(func(*gin.Context) (models.UserModel, bool) {
		return models.UserModel{Roles: []models.RoleModel{{Id: 2}}}, true
	}, jobs.ListPermission), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/jobs", nil))

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d; body=%s", recorder.Code, http.StatusNoContent, recorder.Body.String())
	}
}

func TestPermissionRequiredDeniesUserWithoutMenuOrPermission(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &Store{conn: &fakeMenuPermissionConnection{uris: []string{"/kadmin/jobs"}}}
	engine := gin.New()
	engine.GET("/api/login-audits", store.permissionRequired(func(*gin.Context) (models.UserModel, bool) {
		return models.UserModel{Roles: []models.RoleModel{{Id: 2}}}, true
	}, loginlogs.ListPermission), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/login-audits", nil))

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body=%s", recorder.Code, http.StatusForbidden, recorder.Body.String())
	}
}

func TestUserAccessCodesIncludeMenuPermissions(t *testing.T) {
	store := &Store{conn: &fakeMenuPermissionConnection{uris: []string{"/kadmin/logs"}}}
	user := models.UserModel{Roles: []models.RoleModel{{Id: 2, Slug: "operator"}}}

	codes, err := store.userAccessCodes(user)
	if err != nil {
		t.Fatal(err)
	}
	wanted := map[string]bool{
		"operator":       false,
		logListPermission:   false,
		logDeletePermission: false,
	}
	for _, code := range codes {
		if _, ok := wanted[code]; ok {
			wanted[code] = true
		}
	}
	for code, found := range wanted {
		if !found {
			t.Errorf("access code %q was not returned", code)
		}
	}

	if codes, err := store.userAccessCodes(models.UserModel{}); err != nil || len(codes) != 0 {
		t.Fatalf("user without roles should have no access codes, got %v err=%v", codes, err)
	}
}

func TestMenuPermissionSlugsCoverPermissionGatedMenus(t *testing.T) {
	// 每个带权限保护的页面菜单都必须有映射，否则“菜单可见即接口可用”会再次失效。
	for _, uri := range []string{
		"/kadmin/users",
		"/kadmin/rbac",
		"/kadmin/menus",
		"/kadmin/dictionary",
		"/kadmin/settings",
		"/kadmin/logs",
		"/kadmin/login-audits",
		"/kadmin/jobs",
		"/kadmin/monitor",
		"/kadmin/load-ranking",
		"/kadmin/resources",
	} {
		if _, ok := menuPermissionSlugs[uri]; !ok {
			t.Errorf("permission-gated menu %q has no permission mapping", uri)
		}
	}
}

func TestMenuPermissionSlugsGrantAllModulePermissions(t *testing.T) {
	// 菜单授权应覆盖对应模块的全部接口权限，保证页面操作按钮与后端一致。
	moduleSlugs := map[string]map[string]bool{
		"/kadmin/logs": {
			logListPermission:   false,
			logDeletePermission: false,
		},
		"/kadmin/login-audits": {
			loginlogs.ListPermission:      false,
			loginlogs.DeletePermission:    false,
			loginlogs.RetentionPermission: false,
		},
		"/kadmin/jobs": {
			jobs.ListPermission:    false,
			jobs.CreatePermission:  false,
			jobs.UpdatePermission:  false,
			jobs.DeletePermission:  false,
			jobs.RunPermission:     false,
			jobs.LogListPermission: false,
		},
		"/kadmin/monitor": {
			monitor.ViewPermission:   false,
			monitor.UpdatePermission: false,
		},
	}
	for uri, wanted := range moduleSlugs {
		for _, slug := range menuPermissionSlugs[uri] {
			if _, ok := wanted[slug]; ok {
				wanted[slug] = true
			}
		}
		for slug, found := range wanted {
			if !found {
				t.Errorf("menu %q does not grant %q", uri, slug)
			}
		}
	}
}

package vbenapi

import (
	"net/http"

	"github.com/GoAdminGroup/go-admin/modules/db"
	"github.com/GoAdminGroup/go-admin/modules/db/dialect"
	"github.com/GoAdminGroup/go-admin/plugins/admin/models"
	"github.com/gin-gonic/gin"
)

const (
	logListPermission    = "system:log:list"
	logDeletePermission  = "system:log:delete"
	fileUploadPermission = "system:file:upload"
	fileReadPermission   = "system:file:read"
	fileDeletePermission = "system:file:delete"
)

type permissionSeed struct {
	Name       string
	Slug       string
	HTTPMethod string
	HTTPPath   string
}

var defaultPermissionSeeds = []permissionSeed{
	{Name: "查看请求日志", Slug: logListPermission, HTTPMethod: "GET", HTTPPath: "/api/logs*"},
	{Name: "删除请求日志", Slug: logDeletePermission, HTTPMethod: "DELETE", HTTPPath: "/api/logs*"},
	{Name: "上传文件", Slug: fileUploadPermission, HTTPMethod: "POST", HTTPPath: "/api/files"},
	{Name: "读取文件", Slug: fileReadPermission, HTTPMethod: "GET", HTTPPath: "/api/files*"},
	{Name: "删除文件", Slug: fileDeletePermission, HTTPMethod: "DELETE", HTTPPath: "/api/files/*"},
}

func (s *Store) syncDefaultPermissions() error {
	rows, err := db.WithDriver(s.conn).Table("goadmin_permissions").Select("slug").All()
	if err != nil {
		return err
	}
	existing := make(map[string]bool, len(rows))
	for _, row := range rows {
		existing[toString(row["slug"])] = true
	}

	for _, seed := range defaultPermissionSeeds {
		if existing[seed.Slug] {
			continue
		}
		if _, err := db.WithDriver(s.conn).Table("goadmin_permissions").Insert(dialect.H{
			"name":        seed.Name,
			"slug":        seed.Slug,
			"http_method": seed.HTTPMethod,
			"http_path":   seed.HTTPPath,
			"created_at":  nowString(),
			"updated_at":  nowString(),
		}); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) requirePermission(codes ...string) gin.HandlerFunc {
	return permissionRequired(s.currentUser, codes...)
}

func permissionRequired(resolveUser func(*gin.Context) (models.UserModel, bool), codes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := resolveUser(c)
		if !ok {
			fail(c, http.StatusUnauthorized, "invalid token")
			c.Abort()
			return
		}
		if user.IsSuperAdmin() || isAdminUser(user.UserName) || userHasPermission(user, codes...) {
			c.Set("vben_user", user)
			c.Next()
			return
		}
		fail(c, http.StatusForbidden, "permission denied")
		c.Abort()
	}
}

func userHasPermission(user models.UserModel, codes ...string) bool {
	wanted := make(map[string]bool, len(codes))
	for _, code := range codes {
		wanted[code] = true
	}
	for _, permission := range user.Permissions {
		if permission.Slug == "*" || wanted[permission.Slug] {
			return true
		}
	}
	return false
}

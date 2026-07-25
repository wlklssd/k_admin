package bootstrap

import "github.com/GoAdminGroup/go-admin/internal/kadmin/modules/files"

const (
	LogListPermission   = "system:log:list"
	LogDeletePermission = "system:log:delete"
)

type PermissionSeed struct {
	Name       string
	Slug       string
	HTTPMethod string
	HTTPPath   string
}

func DefaultPermissions() []PermissionSeed {
	return []PermissionSeed{
		{Name: "查看请求日志", Slug: LogListPermission, HTTPMethod: "GET", HTTPPath: "/api/logs*"},
		{Name: "删除请求日志", Slug: LogDeletePermission, HTTPMethod: "DELETE", HTTPPath: "/api/logs*"},
		{Name: "上传文件", Slug: files.UploadPermission, HTTPMethod: "POST", HTTPPath: "/api/files"},
		{Name: "读取文件", Slug: files.ReadPermission, HTTPMethod: "GET", HTTPPath: "/api/files*"},
		{Name: "删除文件", Slug: files.DeletePermission, HTTPMethod: "DELETE", HTTPPath: "/api/files/*"},
	}
}

package bootstrap

import (
	"github.com/GoAdminGroup/go-admin/internal/kadmin/modules/files"
	"github.com/GoAdminGroup/go-admin/internal/kadmin/modules/jobs"
)

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
		{Name: "查看定时任务", Slug: jobs.ListPermission, HTTPMethod: "GET", HTTPPath: "/api/jobs*"},
		{Name: "创建定时任务", Slug: jobs.CreatePermission, HTTPMethod: "POST", HTTPPath: "/api/jobs"},
		{Name: "修改定时任务", Slug: jobs.UpdatePermission, HTTPMethod: "PUT,PATCH", HTTPPath: "/api/jobs/*"},
		{Name: "删除定时任务", Slug: jobs.DeletePermission, HTTPMethod: "DELETE", HTTPPath: "/api/jobs/*"},
		{Name: "立即执行任务", Slug: jobs.RunPermission, HTTPMethod: "POST", HTTPPath: "/api/jobs/*/run"},
		{Name: "查看任务日志", Slug: jobs.LogListPermission, HTTPMethod: "GET", HTTPPath: "/api/job-logs*"},
	}
}

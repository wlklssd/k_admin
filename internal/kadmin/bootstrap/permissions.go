package bootstrap

import (
	"github.com/GoAdminGroup/go-admin/internal/kadmin/modules/files"
	"github.com/GoAdminGroup/go-admin/internal/kadmin/modules/jobs"
	"github.com/GoAdminGroup/go-admin/internal/kadmin/modules/loadrank"
	"github.com/GoAdminGroup/go-admin/internal/kadmin/modules/loginlogs"
	"github.com/GoAdminGroup/go-admin/internal/kadmin/modules/monitor"
)

const (
	LogListPermission            = "system:log:list"
	LogDeletePermission          = "system:log:delete"
	UserManagePermission         = "system:user:manage"
	RbacManagePermission         = "system:rbac:manage"
	MenuManagePermission         = "system:menu:manage"
	DictionaryManagePermission   = "system:dict:manage"
	SystemConfigManagePermission = "system:config:manage"
)

type PermissionSeed struct {
	Name       string
	Slug       string
	HTTPMethod string
	HTTPPath   string
}

func DefaultPermissions() []PermissionSeed {
	return []PermissionSeed{
		{Name: "用户管理", Slug: UserManagePermission, HTTPMethod: "GET,POST,PUT,DELETE", HTTPPath: "/api/users*"},
		{Name: "权限管理", Slug: RbacManagePermission, HTTPMethod: "GET,POST,PUT,DELETE", HTTPPath: "/api/rbac*"},
		{Name: "菜单管理", Slug: MenuManagePermission, HTTPMethod: "GET,POST,PUT,DELETE", HTTPPath: "/api/admin-menus*"},
		{Name: "字典管理", Slug: DictionaryManagePermission, HTTPMethod: "GET,POST,PUT,DELETE", HTTPPath: "/api/dictionaries*"},
		{Name: "参数配置", Slug: SystemConfigManagePermission, HTTPMethod: "GET,PUT", HTTPPath: "/api/system/config*"},
		{Name: "查看请求日志", Slug: LogListPermission, HTTPMethod: "GET", HTTPPath: "/api/logs*"},
		{Name: "删除请求日志", Slug: LogDeletePermission, HTTPMethod: "DELETE", HTTPPath: "/api/logs*"},
		{Name: "查看登录审计", Slug: loginlogs.ListPermission, HTTPMethod: "GET", HTTPPath: "/api/login-audits*"},
		{Name: "清理登录审计", Slug: loginlogs.DeletePermission, HTTPMethod: "DELETE,POST", HTTPPath: "/api/login-audits*"},
		{Name: "设置登录审计保留周期", Slug: loginlogs.RetentionPermission, HTTPMethod: "PATCH", HTTPPath: "/api/login-audits/retention"},
		{Name: "上传文件", Slug: files.UploadPermission, HTTPMethod: "POST", HTTPPath: "/api/files"},
		{Name: "读取文件", Slug: files.ReadPermission, HTTPMethod: "GET", HTTPPath: "/api/files*"},
		{Name: "删除文件", Slug: files.DeletePermission, HTTPMethod: "DELETE", HTTPPath: "/api/files/*"},
		{Name: "查看定时任务", Slug: jobs.ListPermission, HTTPMethod: "GET", HTTPPath: "/api/jobs*"},
		{Name: "创建定时任务", Slug: jobs.CreatePermission, HTTPMethod: "POST", HTTPPath: "/api/jobs"},
		{Name: "修改定时任务", Slug: jobs.UpdatePermission, HTTPMethod: "PUT,PATCH", HTTPPath: "/api/jobs/*"},
		{Name: "删除定时任务", Slug: jobs.DeletePermission, HTTPMethod: "DELETE", HTTPPath: "/api/jobs/*"},
		{Name: "立即执行任务", Slug: jobs.RunPermission, HTTPMethod: "POST", HTTPPath: "/api/jobs/*/run"},
		{Name: "查看任务日志", Slug: jobs.LogListPermission, HTTPMethod: "GET", HTTPPath: "/api/job-logs*"},
		{Name: "查看系统监控", Slug: monitor.ViewPermission, HTTPMethod: "GET", HTTPPath: "/api/system-monitor"},
		{Name: "启停系统监控", Slug: monitor.UpdatePermission, HTTPMethod: "PATCH", HTTPPath: "/api/system-monitor/status"},
		{Name: "查看接口负载排行", Slug: loadrank.ViewPermission, HTTPMethod: "GET", HTTPPath: "/api/load-ranking*"},
		{Name: "启停接口采样", Slug: loadrank.UpdatePermission, HTTPMethod: "PATCH", HTTPPath: "/api/load-ranking/status"},
	}
}

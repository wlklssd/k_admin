package bootstrap

import (
	"github.com/GoAdminGroup/go-admin/internal/kadmin/modules/codegen"
	"github.com/GoAdminGroup/go-admin/internal/kadmin/modules/files"
	"github.com/GoAdminGroup/go-admin/internal/kadmin/modules/jobs"
	"github.com/GoAdminGroup/go-admin/internal/kadmin/modules/loadrank"
	"github.com/GoAdminGroup/go-admin/internal/kadmin/modules/loginlogs"
	"github.com/GoAdminGroup/go-admin/internal/kadmin/modules/monitor"
	"github.com/GoAdminGroup/go-admin/internal/kadmin/modules/notifications"
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

type PermissionScope string

const (
	PermissionScopePage   PermissionScope = "page"
	PermissionScopeButton PermissionScope = "button"
)

type PermissionSeed struct {
	Name       string
	Slug       string
	HTTPMethod string
	HTTPPath   string
	PageURI    string
	PageTitle  string
	Scope      PermissionScope
	Button     string
}

func DefaultPermissions() []PermissionSeed {
	return []PermissionSeed{
		{
			Name: "用户管理", Slug: UserManagePermission,
			HTTPMethod: "GET,POST,PUT,DELETE", HTTPPath: "/api/users*",
			PageURI: "/kadmin/users", PageTitle: "用户管理", Scope: PermissionScopePage,
		},
		{
			Name: "权限管理", Slug: RbacManagePermission,
			HTTPMethod: "GET,POST,PUT,DELETE", HTTPPath: "/api/rbac*",
			PageURI: "/kadmin/rbac", PageTitle: "权限管理", Scope: PermissionScopePage,
		},
		{
			Name: "菜单管理", Slug: MenuManagePermission,
			HTTPMethod: "GET,POST,PUT,DELETE", HTTPPath: "/api/admin-menus*",
			PageURI: "/kadmin/menus", PageTitle: "菜单管理", Scope: PermissionScopePage,
		},
		{
			Name: "字典管理", Slug: DictionaryManagePermission,
			HTTPMethod: "GET,POST,PUT,DELETE", HTTPPath: "/api/dictionaries*",
			PageURI: "/kadmin/dictionary", PageTitle: "字典管理", Scope: PermissionScopePage,
		},
		{
			Name: "参数配置", Slug: SystemConfigManagePermission,
			HTTPMethod: "GET,PUT", HTTPPath: "/api/system/config*",
			PageURI: "/kadmin/settings", PageTitle: "参数配置", Scope: PermissionScopePage,
		},
		{
			Name: "查看请求日志", Slug: LogListPermission, HTTPMethod: "GET", HTTPPath: "/api/logs*",
			PageURI: "/kadmin/logs", PageTitle: "日志管理", Scope: PermissionScopePage,
		},
		{
			Name: "删除请求日志", Slug: LogDeletePermission, HTTPMethod: "DELETE", HTTPPath: "/api/logs*",
			PageURI: "/kadmin/logs", PageTitle: "日志管理", Scope: PermissionScopeButton, Button: "logs.delete",
		},
		{
			Name: "查看登录审计", Slug: loginlogs.ListPermission, HTTPMethod: "GET", HTTPPath: "/api/login-audits*",
			PageURI: "/kadmin/login-audits", PageTitle: "登录审计", Scope: PermissionScopePage,
		},
		{
			Name: "清理登录审计", Slug: loginlogs.DeletePermission, HTTPMethod: "DELETE,POST", HTTPPath: "/api/login-audits*",
			PageURI: "/kadmin/login-audits", PageTitle: "登录审计", Scope: PermissionScopeButton, Button: "login-audits.delete",
		},
		{
			Name: "设置登录审计保留周期", Slug: loginlogs.RetentionPermission, HTTPMethod: "PATCH", HTTPPath: "/api/login-audits/retention",
			PageURI: "/kadmin/login-audits", PageTitle: "登录审计", Scope: PermissionScopeButton, Button: "login-audits.retention",
		},
		{
			Name: "上传文件", Slug: files.UploadPermission, HTTPMethod: "POST", HTTPPath: "/api/files",
			PageURI: "/kadmin/resources", PageTitle: "资源工作台", Scope: PermissionScopeButton, Button: "resources.upload",
		},
		{
			Name: "读取文件", Slug: files.ReadPermission, HTTPMethod: "GET", HTTPPath: "/api/files*",
			PageURI: "/kadmin/resources", PageTitle: "资源工作台", Scope: PermissionScopePage,
		},
		{
			Name: "删除文件", Slug: files.DeletePermission, HTTPMethod: "DELETE", HTTPPath: "/api/files/*",
			PageURI: "/kadmin/resources", PageTitle: "资源工作台", Scope: PermissionScopeButton, Button: "resources.delete",
		},
		{
			Name: "查看定时任务", Slug: jobs.ListPermission, HTTPMethod: "GET", HTTPPath: "/api/jobs*",
			PageURI: "/kadmin/jobs", PageTitle: "定时任务", Scope: PermissionScopePage,
		},
		{
			Name: "创建定时任务", Slug: jobs.CreatePermission, HTTPMethod: "POST", HTTPPath: "/api/jobs",
			PageURI: "/kadmin/jobs", PageTitle: "定时任务", Scope: PermissionScopeButton, Button: "jobs.create",
		},
		{
			Name: "修改定时任务", Slug: jobs.UpdatePermission, HTTPMethod: "PUT,PATCH", HTTPPath: "/api/jobs/*",
			PageURI: "/kadmin/jobs", PageTitle: "定时任务", Scope: PermissionScopeButton, Button: "jobs.update",
		},
		{
			Name: "删除定时任务", Slug: jobs.DeletePermission, HTTPMethod: "DELETE", HTTPPath: "/api/jobs/*",
			PageURI: "/kadmin/jobs", PageTitle: "定时任务", Scope: PermissionScopeButton, Button: "jobs.delete",
		},
		{
			Name: "立即执行任务", Slug: jobs.RunPermission, HTTPMethod: "POST", HTTPPath: "/api/jobs/*/run",
			PageURI: "/kadmin/jobs", PageTitle: "定时任务", Scope: PermissionScopeButton, Button: "jobs.run",
		},
		{
			Name: "查看任务日志", Slug: jobs.LogListPermission, HTTPMethod: "GET", HTTPPath: "/api/job-logs*",
			PageURI: "/kadmin/jobs", PageTitle: "定时任务", Scope: PermissionScopeButton, Button: "jobs.logs",
		},
		{
			Name: "查看系统监控", Slug: monitor.ViewPermission, HTTPMethod: "GET", HTTPPath: "/api/system-monitor",
			PageURI: "/kadmin/monitor", PageTitle: "系统监控", Scope: PermissionScopePage,
		},
		{
			Name: "启停系统监控", Slug: monitor.UpdatePermission, HTTPMethod: "PATCH", HTTPPath: "/api/system-monitor/status",
			PageURI: "/kadmin/monitor", PageTitle: "系统监控", Scope: PermissionScopeButton, Button: "monitor.update",
		},
		{
			Name: "查看接口负载排行", Slug: loadrank.ViewPermission, HTTPMethod: "GET", HTTPPath: "/api/load-ranking*",
			PageURI: "/kadmin/load-ranking", PageTitle: "接口负载排行", Scope: PermissionScopePage,
		},
		{
			Name: "启停接口采样", Slug: loadrank.UpdatePermission, HTTPMethod: "PATCH", HTTPPath: "/api/load-ranking/status",
			PageURI: "/kadmin/load-ranking", PageTitle: "接口负载排行", Scope: PermissionScopeButton, Button: "load-ranking.update",
		},
		{
			Name: "查看代码生成", Slug: codegen.ListPermission, HTTPMethod: "GET", HTTPPath: "/api/codegen*",
			PageURI: "/kadmin/codegen", PageTitle: "代码生成", Scope: PermissionScopePage,
		},
		{
			Name: "导入与配置代码生成", Slug: codegen.ImportPermission, HTTPMethod: "POST,PUT,DELETE", HTTPPath: "/api/codegen*",
			PageURI: "/kadmin/codegen", PageTitle: "代码生成", Scope: PermissionScopeButton, Button: "codegen.import",
		},
		{
			Name: "预览与生成代码", Slug: codegen.GeneratePermission, HTTPMethod: "POST", HTTPPath: "/api/codegen*",
			PageURI: "/kadmin/codegen", PageTitle: "代码生成", Scope: PermissionScopeButton, Button: "codegen.generate",
		},
		{
			Name: "查看站内通知", Slug: notifications.ListPermission, HTTPMethod: "GET,PATCH,DELETE", HTTPPath: "/api/notifications*",
			PageURI: "/kadmin/notifications", PageTitle: "站内通知", Scope: PermissionScopePage,
		},
		{
			Name: "发送站内通知", Slug: notifications.CreatePermission, HTTPMethod: "POST", HTTPPath: "/api/notifications*",
			PageURI: "/kadmin/notifications", PageTitle: "站内通知", Scope: PermissionScopeButton, Button: "notifications.create",
		},
	}
}

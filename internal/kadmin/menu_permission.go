package kadmin

import (
	"github.com/GoAdminGroup/go-admin/internal/kadmin/modules/files"
	"github.com/GoAdminGroup/go-admin/internal/kadmin/modules/jobs"
	"github.com/GoAdminGroup/go-admin/internal/kadmin/modules/loadrank"
	"github.com/GoAdminGroup/go-admin/internal/kadmin/modules/loginlogs"
	"github.com/GoAdminGroup/go-admin/internal/kadmin/modules/monitor"
	"github.com/GoAdminGroup/go-admin/modules/db"
	"github.com/GoAdminGroup/go-admin/plugins/admin/models"
)

// menuPermissionSlugs 将菜单 URI 映射到该页面需要调用的接口权限。
// 在权限工作台为角色勾选菜单即授予对应页面的权限；取消勾选即收回。
// 与 goadmin_role_permissions 直接配置的权限是叠加关系，互不影响。
var menuPermissionSlugs = map[string][]string{
	"/kadmin/users": {
		userManagePermission,
	},
	"/kadmin/rbac": {
		rbacManagePermission,
	},
	"/kadmin/menus": {
		menuManagePermission,
	},
	"/kadmin/dictionary": {
		dictionaryManagePermission,
	},
	"/kadmin/settings": {
		systemConfigManagePermission,
	},
	"/kadmin/logs": {
		logListPermission,
		logDeletePermission,
	},
	"/kadmin/login-audits": {
		loginlogs.ListPermission,
		loginlogs.DeletePermission,
		loginlogs.RetentionPermission,
	},
	"/kadmin/jobs": {
		jobs.ListPermission,
		jobs.CreatePermission,
		jobs.UpdatePermission,
		jobs.DeletePermission,
		jobs.RunPermission,
		jobs.LogListPermission,
	},
	"/kadmin/monitor": {
		monitor.ViewPermission,
		monitor.UpdatePermission,
	},
	"/kadmin/load-ranking": {
		loadrank.ViewPermission,
		loadrank.UpdatePermission,
	},
	"/kadmin/resources": {
		files.UploadPermission,
		files.ReadPermission,
		files.DeletePermission,
	},
}

// menuPermissionSlugsForUser 返回用户角色已授权菜单对应的页面权限 slug 集合。
// 与菜单可见性使用同一数据源（goadmin_role_menu），保证“菜单可见即接口可用”。
func (s *Store) menuPermissionSlugsForUser(user models.UserModel) (map[string]bool, error) {
	roleIDs := user.GetAllRoleId()
	if len(roleIDs) == 0 {
		return nil, nil
	}
	rows, err := db.WithDriver(s.conn).
		Table("goadmin_role_menu").
		LeftJoin("goadmin_menu", "goadmin_menu.id", "=", "goadmin_role_menu.menu_id").
		WhereIn("goadmin_role_menu.role_id", roleIDs).
		Select("goadmin_menu.uri").
		All()
	if err != nil {
		return nil, err
	}

	slugs := make(map[string]bool)
	for _, row := range rows {
		for _, slug := range menuPermissionSlugs[normalizeMenuURI(toString(row["uri"]))] {
			slugs[slug] = true
		}
	}
	return slugs, nil
}

// userHasMenuPermission 判断用户是否通过菜单授权获得指定接口权限。
func (s *Store) userHasMenuPermission(user models.UserModel, codes ...string) (bool, error) {
	slugs, err := s.menuPermissionSlugsForUser(user)
	if err != nil {
		return false, err
	}
	for _, code := range codes {
		if slugs[code] {
			return true, nil
		}
	}
	return false, nil
}

// userAccessCodes 返回用户全部访问标识：角色标识、直接配置的权限以及菜单隐含的页面权限。
func (s *Store) userAccessCodes(user models.UserModel) ([]string, error) {
	codes := accessCodes(user)
	slugs, err := s.menuPermissionSlugsForUser(user)
	if err != nil {
		return nil, err
	}
	if len(slugs) == 0 {
		return codes, nil
	}

	set := make(map[string]bool, len(codes)+len(slugs))
	for _, code := range codes {
		set[code] = true
	}
	for slug := range slugs {
		set[slug] = true
	}
	merged := make([]string, 0, len(set))
	for code := range set {
		merged = append(merged, code)
	}
	return merged, nil
}

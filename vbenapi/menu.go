package vbenapi

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/GoAdminGroup/go-admin/modules/db"
	"github.com/gin-gonic/gin"
)

type menuItem struct {
	ID       int64
	ParentID int64
	Order    int64
	Title    string
	Icon     string
	URI      string
}

type vbenMenu struct {
	ID        int64                  `json:"id,omitempty"`
	ParentID  int64                  `json:"parentId,omitempty"`
	Name      string                 `json:"name"`
	Path      string                 `json:"path"`
	Component string                 `json:"component,omitempty"`
	Redirect  string                 `json:"redirect,omitempty"`
	Meta      map[string]interface{} `json:"meta"`
	Children  []vbenMenu             `json:"children,omitempty"`
}

type menuRouteBinding struct {
	Name      string
	Path      string
	Component string
	Icon      string
}

var vbenMenuRouteBindings = map[string]menuRouteBinding{
	"/": {
		Name: "Dashboard",
		Path: "/dashboard",
		Icon: "lucide:layout-dashboard",
	},
	"/dashboard": {
		Name: "Dashboard",
		Path: "/dashboard",
		Icon: "lucide:layout-dashboard",
	},
	"/dashboard/analytics": {
		Name:      "Analytics",
		Path:      "/dashboard/analytics",
		Component: "/dashboard/analytics/index",
		Icon:      "lucide:area-chart",
	},
	"/dashboard/workspace": {
		Name:      "Workspace",
		Path:      "/dashboard/workspace",
		Component: "/dashboard/workspace/index",
		Icon:      "carbon:workspace",
	},
	"/kadmin": {
		Name: "KAdmin",
		Path: "/kadmin",
		Icon: "lucide:settings-2",
	},
	"/kadmin/users": {
		Name:      "KAdminUsers",
		Path:      "/kadmin/users",
		Component: "/kadmin/components/UserManagementView",
		Icon:      "lucide:users",
	},
	"/kadmin/rbac": {
		Name:      "KAdminRbac",
		Path:      "/kadmin/rbac",
		Component: "/kadmin/components/RbacWorkbench",
		Icon:      "lucide:shield-check",
	},
	"/kadmin/dictionary": {
		Name:      "KAdminDictionary",
		Path:      "/kadmin/dictionary",
		Component: "/kadmin/components/DictionaryManagementView",
		Icon:      "lucide:book-open",
	},
	"/kadmin/settings": {
		Name:      "KAdminSettings",
		Path:      "/kadmin/settings",
		Component: "/kadmin/components/SettingsView",
		Icon:      "lucide:sliders-horizontal",
	},
	"/kadmin/resources": {
		Name:      "KAdminResources",
		Path:      "/kadmin/resources",
		Component: "/kadmin/components/ResourceWorkbench",
		Icon:      "lucide:folder-kanban",
	},
}

func registerMenuRoutes(api *gin.RouterGroup, s *Store) {
	menuGroup := api.Group("/menu", s.requireAuth())
	menuGroup.GET("/all", s.menus)
	menuGroup.GET("/list", s.menus)
}

func (s *Store) menus(c *gin.Context) {
	user, ok := s.currentUser(c)
	if !ok {
		fail(c, http.StatusUnauthorized, "invalid token")
		return
	}

	rows, err := db.WithDriver(s.conn).
		Table("goadmin_menu").
		OrderBy("order", "asc").
		All()
	if err != nil {
		fail(c, http.StatusInternalServerError, err.Error())
		return
	}

	allowed := make(map[int64]bool)
	if user.IsSuperAdmin() || isAdminUser(user.UserName) {
		for _, row := range rows {
			allowed[toInt64(row["id"])] = true
		}
	} else {
		for _, id := range user.MenuIds {
			allowed[id] = true
		}
	}

	items := make([]menuItem, 0, len(rows))
	for _, row := range rows {
		id := toInt64(row["id"])
		if !allowed[id] {
			continue
		}
		items = append(items, menuItem{
			ID:       id,
			ParentID: toInt64(row["parent_id"]),
			Order:    toInt64(row["order"]),
			Title:    toString(row["title"]),
			Icon:     toString(row["icon"]),
			URI:      toString(row["uri"]),
		})
	}

	success(c, buildMenuTree(items, 0))
}

func buildMenuTree(items []menuItem, parentID int64) []vbenMenu {
	res := make([]vbenMenu, 0)
	for _, item := range items {
		if item.ParentID != parentID {
			continue
		}

		children := buildMenuTree(items, item.ID)
		menu := item.toVbenMenu(children)
		res = append(res, menu)
	}
	return res
}

func (m menuItem) toVbenMenu(children []vbenMenu) vbenMenu {
	path, iframeSrc := menuPath(m)
	binding, hasBinding := vbenMenuBinding(m.URI)
	if hasBinding {
		path = binding.Path
	}

	meta := map[string]interface{}{
		"title": m.Title,
		"order": m.Order,
	}
	icon := normalizeMenuIcon(m.Icon)
	if icon == "" && hasBinding {
		icon = binding.Icon
	}
	if icon != "" {
		meta["icon"] = icon
	}
	if iframeSrc != "" && !hasBinding {
		meta["iframeSrc"] = iframeSrc
	}

	name := "GoAdminMenu" + strconv.FormatInt(m.ID, 10)
	if hasBinding {
		name = binding.Name
	}

	menu := vbenMenu{
		ID:       m.ID,
		ParentID: m.ParentID,
		Name:     name,
		Path:     path,
		Meta:     meta,
		Children: children,
	}

	if len(children) > 0 {
		menu.Redirect = children[0].Path
		return menu
	}

	if hasBinding {
		menu.Component = binding.Component
		return menu
	}

	if iframeSrc != "" {
		menu.Component = "IFrameView"
	}
	return menu
}

func menuPath(m menuItem) (path string, iframeSrc string) {
	uri := normalizeMenuURI(m.URI)
	if uri == "" || uri == "#" {
		return "/goadmin/menu-" + strconv.FormatInt(m.ID, 10), ""
	}

	if strings.HasPrefix(uri, "http://") || strings.HasPrefix(uri, "https://") {
		return "/goadmin/external-" + strconv.FormatInt(m.ID, 10), uri
	}

	return "/goadmin" + strings.TrimSuffix(uri, "/"), "/admin" + uri
}

func vbenMenuBinding(uri string) (menuRouteBinding, bool) {
	binding, ok := vbenMenuRouteBindings[normalizeMenuURI(uri)]
	return binding, ok
}

func normalizeMenuURI(uri string) string {
	uri = strings.TrimSpace(uri)
	if uri == "" || uri == "#" {
		return uri
	}
	if strings.HasPrefix(uri, "http://") || strings.HasPrefix(uri, "https://") {
		return uri
	}
	if !strings.HasPrefix(uri, "/") {
		uri = "/" + uri
	}
	if uri != "/" {
		uri = strings.TrimSuffix(uri, "/")
	}
	return uri
}

func normalizeMenuIcon(icon string) string {
	icon = strings.TrimSpace(icon)
	if icon == "" || strings.Contains(icon, ":") {
		return icon
	}

	switch icon {
	case "fa-bar-chart", "fa-dashboard":
		return "lucide:layout-dashboard"
	case "fa-tasks":
		return "lucide:settings-2"
	case "fa-users":
		return "lucide:users"
	case "fa-user":
		return "lucide:user"
	case "fa-ban":
		return "lucide:ban"
	case "fa-bars":
		return "lucide:menu"
	case "fa-history":
		return "lucide:history"
	default:
		return icon
	}
}

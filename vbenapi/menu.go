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
	if user.IsSuperAdmin() {
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
	path, link := menuPath(m)
	meta := map[string]interface{}{
		"title": m.Title,
		"order": m.Order,
	}
	if m.Icon != "" {
		meta["icon"] = m.Icon
	}
	if link != "" {
		meta["link"] = link
	}

	menu := vbenMenu{
		ID:       m.ID,
		ParentID: m.ParentID,
		Name:     "GoAdminMenu" + strconv.FormatInt(m.ID, 10),
		Path:     path,
		Meta:     meta,
		Children: children,
	}

	if len(children) > 0 {
		menu.Redirect = children[0].Path
		return menu
	}

	// Add a Vben view at apps/web-*/src/views/legacy/iframe/index.vue
	// to render meta.link during the migration from GoAdmin pages.
	menu.Component = "/legacy/iframe/index"
	return menu
}

func menuPath(m menuItem) (path string, link string) {
	uri := strings.TrimSpace(m.URI)
	if uri == "" || uri == "#" {
		return "/goadmin/menu-" + strconv.FormatInt(m.ID, 10), ""
	}

	if strings.HasPrefix(uri, "http://") || strings.HasPrefix(uri, "https://") {
		return "/goadmin/external-" + strconv.FormatInt(m.ID, 10), uri
	}

	if !strings.HasPrefix(uri, "/") {
		uri = "/" + uri
	}

	return "/goadmin" + strings.TrimSuffix(uri, "/"), "/admin" + uri
}

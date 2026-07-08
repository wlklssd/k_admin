package vbenapi

import (
	"net/http"
	"sort"
	"strings"

	"github.com/GoAdminGroup/go-admin/modules/db"
	"github.com/GoAdminGroup/go-admin/modules/db/dialect"
	"github.com/gin-gonic/gin"
)

type managedMenu struct {
	ID         int64         `json:"id"`
	ParentID   int64         `json:"parentId"`
	Type       int64         `json:"type"`
	Order      int64         `json:"order"`
	Title      string        `json:"title"`
	Icon       string        `json:"icon"`
	URI        string        `json:"uri"`
	Header     string        `json:"header"`
	PluginName string        `json:"pluginName"`
	UUID       string        `json:"uuid"`
	CreatedAt  string        `json:"createdAt"`
	UpdatedAt  string        `json:"updatedAt"`
	Children   []managedMenu `json:"children,omitempty"`
}

type menuPayload struct {
	ParentID   int64  `json:"parentId"`
	Type       *int64 `json:"type"`
	Order      int64  `json:"order"`
	Title      string `json:"title"`
	Icon       string `json:"icon"`
	URI        string `json:"uri"`
	Header     string `json:"header"`
	PluginName string `json:"pluginName"`
	UUID       string `json:"uuid"`
}

func registerMenuManagementRoutes(api *gin.RouterGroup, s *Store) {
	menuGroup := api.Group("/admin-menus", s.requireAuth(), s.requireAdmin())
	menuGroup.GET("", s.adminMenus)
	menuGroup.GET("/tree", s.adminMenuTree)
	menuGroup.POST("", s.createAdminMenu)
	menuGroup.PUT("/:id", s.updateAdminMenu)
	menuGroup.DELETE("/:id", s.deleteAdminMenu)
}

func (s *Store) adminMenus(c *gin.Context) {
	menus, err := s.loadManagedMenus()
	if err != nil {
		fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	success(c, menus)
}

func (s *Store) adminMenuTree(c *gin.Context) {
	menus, err := s.loadManagedMenus()
	if err != nil {
		fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	success(c, buildManagedMenuTree(menus, 0))
}

func (s *Store) createAdminMenu(c *gin.Context) {
	var req menuPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "invalid menu payload")
		return
	}
	normalizeMenuPayload(&req)
	if req.Title == "" {
		fail(c, http.StatusBadRequest, "menu title is required")
		return
	}
	if req.ParentID > 0 {
		exists, err := s.menuExists(req.ParentID)
		if err != nil {
			fail(c, http.StatusInternalServerError, err.Error())
			return
		}
		if !exists {
			fail(c, http.StatusBadRequest, "parent menu does not exist")
			return
		}
	}

	id, err := db.WithDriver(s.conn).Table("goadmin_menu").Insert(dialect.H{
		"parent_id":   req.ParentID,
		"type":        *req.Type,
		"order":       req.Order,
		"title":       req.Title,
		"icon":        req.Icon,
		"uri":         req.URI,
		"header":      req.Header,
		"plugin_name": req.PluginName,
		"uuid":        req.UUID,
		"created_at":  nowString(),
		"updated_at":  nowString(),
	})
	if err != nil {
		fail(c, http.StatusInternalServerError, err.Error())
		return
	}

	menu, err := s.loadManagedMenu(id)
	if err != nil {
		fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	success(c, menu)
}

func (s *Store) updateAdminMenu(c *gin.Context) {
	menuID, ok := pathID(c)
	if !ok {
		return
	}

	var req menuPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "invalid menu payload")
		return
	}
	normalizeMenuPayload(&req)
	if req.Title == "" {
		fail(c, http.StatusBadRequest, "menu title is required")
		return
	}
	if req.ParentID == menuID {
		fail(c, http.StatusBadRequest, "parent menu cannot be itself")
		return
	}
	if req.ParentID > 0 {
		exists, err := s.menuExists(req.ParentID)
		if err != nil {
			fail(c, http.StatusInternalServerError, err.Error())
			return
		}
		if !exists {
			fail(c, http.StatusBadRequest, "parent menu does not exist")
			return
		}
		descendant, err := s.isMenuDescendant(req.ParentID, menuID)
		if err != nil {
			fail(c, http.StatusInternalServerError, err.Error())
			return
		}
		if descendant {
			fail(c, http.StatusBadRequest, "parent menu cannot be a descendant")
			return
		}
	}

	_, err := db.WithDriver(s.conn).
		Table("goadmin_menu").
		Where("id", "=", menuID).
		Update(dialect.H{
			"parent_id":   req.ParentID,
			"type":        *req.Type,
			"order":       req.Order,
			"title":       req.Title,
			"icon":        req.Icon,
			"uri":         req.URI,
			"header":      req.Header,
			"plugin_name": req.PluginName,
			"uuid":        req.UUID,
			"updated_at":  nowString(),
		})
	if err != nil {
		fail(c, http.StatusInternalServerError, err.Error())
		return
	}

	menu, err := s.loadManagedMenu(menuID)
	if err != nil {
		fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	success(c, menu)
}

func (s *Store) deleteAdminMenu(c *gin.Context) {
	menuID, ok := pathID(c)
	if !ok {
		return
	}
	hasChildren, err := s.menuHasChildren(menuID)
	if err != nil {
		fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	if hasChildren {
		fail(c, http.StatusBadRequest, "menu has children")
		return
	}

	_ = db.WithDriver(s.conn).Table("goadmin_role_menu").Where("menu_id", "=", menuID).Delete()
	if err := db.WithDriver(s.conn).Table("goadmin_menu").Where("id", "=", menuID).Delete(); err != nil {
		fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	success(c, true)
}

func (s *Store) loadManagedMenus() ([]managedMenu, error) {
	rows, err := db.WithDriver(s.conn).
		Table("goadmin_menu").
		OrderBy("order", "asc").
		All()
	if err != nil {
		return nil, err
	}

	menus := make([]managedMenu, 0, len(rows))
	for _, row := range rows {
		menus = append(menus, managedMenuFromRow(row))
	}
	sort.SliceStable(menus, func(i, j int) bool {
		if menus[i].Order == menus[j].Order {
			return menus[i].ID < menus[j].ID
		}
		return menus[i].Order < menus[j].Order
	})
	return menus, nil
}

func (s *Store) loadManagedMenu(id int64) (managedMenu, error) {
	rows, err := db.WithDriver(s.conn).
		Table("goadmin_menu").
		Where("id", "=", id).
		All()
	if err != nil {
		return managedMenu{}, err
	}
	if len(rows) == 0 {
		return managedMenu{}, nil
	}
	return managedMenuFromRow(rows[0]), nil
}

func managedMenuFromRow(row map[string]interface{}) managedMenu {
	return managedMenu{
		ID:         toInt64(row["id"]),
		ParentID:   toInt64(row["parent_id"]),
		Type:       toInt64(row["type"]),
		Order:      toInt64(row["order"]),
		Title:      toString(row["title"]),
		Icon:       toString(row["icon"]),
		URI:        toString(row["uri"]),
		Header:     toString(row["header"]),
		PluginName: toString(row["plugin_name"]),
		UUID:       toString(row["uuid"]),
		CreatedAt:  toDateTimeString(row["created_at"]),
		UpdatedAt:  toDateTimeString(row["updated_at"]),
	}
}

func buildManagedMenuTree(items []managedMenu, parentID int64) []managedMenu {
	res := make([]managedMenu, 0)
	for _, item := range items {
		if item.ParentID != parentID {
			continue
		}
		item.Children = buildManagedMenuTree(items, item.ID)
		res = append(res, item)
	}
	return res
}

func normalizeMenuPayload(req *menuPayload) {
	req.Title = strings.TrimSpace(req.Title)
	req.Icon = strings.TrimSpace(req.Icon)
	req.URI = normalizeMenuURI(req.URI)
	req.Header = strings.TrimSpace(req.Header)
	req.PluginName = strings.TrimSpace(req.PluginName)
	req.UUID = strings.TrimSpace(req.UUID)
	// type 未传时默认为菜单(1)；显式传 0 表示目录/分组，需原样保留。
	if req.Type == nil {
		t := int64(1)
		req.Type = &t
	}
}

func (s *Store) menuExists(id int64) (bool, error) {
	rows, err := db.WithDriver(s.conn).Table("goadmin_menu").Where("id", "=", id).All()
	if err != nil {
		return false, err
	}
	return len(rows) > 0, nil
}

func (s *Store) menuHasChildren(id int64) (bool, error) {
	rows, err := db.WithDriver(s.conn).Table("goadmin_menu").Where("parent_id", "=", id).All()
	if err != nil {
		return false, err
	}
	return len(rows) > 0, nil
}

func (s *Store) isMenuDescendant(candidateID, parentID int64) (bool, error) {
	menus, err := s.loadManagedMenus()
	if err != nil {
		return false, err
	}
	parentByID := make(map[int64]int64, len(menus))
	for _, menu := range menus {
		parentByID[menu.ID] = menu.ParentID
	}
	for current := candidateID; current != 0; current = parentByID[current] {
		if current == parentID {
			return true, nil
		}
		next, ok := parentByID[current]
		if !ok || next == current {
			break
		}
	}
	return false, nil
}

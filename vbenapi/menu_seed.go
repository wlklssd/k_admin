package vbenapi

import (
	"github.com/GoAdminGroup/go-admin/modules/db"
	"github.com/GoAdminGroup/go-admin/modules/db/dialect"
)

type menuSeed struct {
	Title    string
	Icon     string
	URI      string
	Order    int64
	Children []menuSeed
}

var defaultMenuSeeds = []menuSeed{
	{
		Title: "Dashboard",
		Icon:  "lucide:layout-dashboard",
		URI:   "/dashboard",
		Order: 1,
		Children: []menuSeed{
			{
				Title: "分析页",
				Icon:  "lucide:area-chart",
				URI:   "/dashboard/analytics",
				Order: 1,
			},
			{
				Title: "工作台",
				Icon:  "carbon:workspace",
				URI:   "/dashboard/workspace",
				Order: 2,
			},
		},
	},
	{
		Title: "KAdmin 管理",
		Icon:  "lucide:settings-2",
		URI:   "/kadmin",
		Order: 10,
		Children: []menuSeed{
			{
				Title: "用户管理",
				Icon:  "lucide:users",
				URI:   "/kadmin/users",
				Order: 1,
			},
			{
				Title: "权限管理",
				Icon:  "lucide:shield-check",
				URI:   "/kadmin/rbac",
				Order: 2,
			},
			{
				Title: "菜单管理",
				Icon:  "lucide:menu",
				URI:   "/kadmin/menus",
				Order: 3,
			},
			{
				Title: "字典管理",
				Icon:  "lucide:book-open",
				URI:   "/kadmin/dictionary",
				Order: 4,
			},
			{
				Title: "参数配置",
				Icon:  "lucide:sliders-horizontal",
				URI:   "/kadmin/settings",
				Order: 5,
			},
			{
				Title: "资源工作台",
				Icon:  "lucide:folder-kanban",
				URI:   "/kadmin/resources",
				Order: 6,
			},
		},
	},
}

func (s *Store) syncDefaultMenus() error {
	rows, err := db.WithDriver(s.conn).Table("goadmin_menu").All()
	if err != nil {
		return err
	}

	idsByURI := make(map[string]int64, len(rows))
	for _, row := range rows {
		uri := normalizeMenuURI(toString(row["uri"]))
		if uri != "" && idsByURI[uri] == 0 {
			idsByURI[uri] = toInt64(row["id"])
		}
	}

	for _, seed := range defaultMenuSeeds {
		if err := s.ensureMenuSeed(seed, 0, idsByURI); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) ensureMenuSeed(seed menuSeed, parentID int64, idsByURI map[string]int64) error {
	uri := normalizeMenuURI(seed.URI)
	id := idsByURI[uri]
	if id == 0 && uri == "/dashboard" {
		id = idsByURI["/"]
	}
	if id == 0 {
		insertID, err := db.WithDriver(s.conn).Table("goadmin_menu").Insert(dialect.H{
			"parent_id":  parentID,
			"type":       1,
			"order":      seed.Order,
			"title":      seed.Title,
			"icon":       seed.Icon,
			"uri":        seed.URI,
			"created_at": nowString(),
			"updated_at": nowString(),
		})
		if err != nil {
			return err
		}
		id = insertID
		idsByURI[uri] = id
	}

	for _, child := range seed.Children {
		if err := s.ensureMenuSeed(child, id, idsByURI); err != nil {
			return err
		}
	}
	return nil
}

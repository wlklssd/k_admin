package kadmin

import (
	"sort"

	"github.com/GoAdminGroup/go-admin/internal/kadmin/bootstrap"
	"github.com/GoAdminGroup/go-admin/modules/db"
	"github.com/GoAdminGroup/go-admin/modules/db/dialect"
)

type menuSeed = bootstrap.MenuSeed

var defaultMenuSeeds = bootstrap.DefaultMenus()

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
	return s.normalizeMenuParentTypes()
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
			"type":       menuSeedType(seed),
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

func menuSeedType(seed menuSeed) int64 {
	if len(seed.Children) > 0 {
		return menuTypeDirectory
	}
	return menuTypeItem
}

func (s *Store) normalizeMenuParentTypes() error {
	rows, err := db.WithDriver(s.conn).Table("goadmin_menu").All()
	if err != nil {
		return err
	}

	typeByID := make(map[int64]int64, len(rows))
	parentIDs := make(map[int64]struct{})
	for _, row := range rows {
		id := toInt64(row["id"])
		typeByID[id] = toInt64(row["type"])
		if parentID := toInt64(row["parent_id"]); parentID > 0 {
			parentIDs[parentID] = struct{}{}
		}
	}

	ids := make([]int64, 0, len(parentIDs))
	for id := range parentIDs {
		if typeByID[id] != menuTypeDirectory {
			ids = append(ids, id)
		}
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	for _, id := range ids {
		if _, err := db.WithDriver(s.conn).
			Table("goadmin_menu").
			Where("id", "=", id).
			Update(dialect.H{
				"type":       menuTypeDirectory,
				"updated_at": nowString(),
			}); err != nil {
			return err
		}
	}
	return nil
}

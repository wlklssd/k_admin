package vbenapi

import "testing"

func TestExpandAllowedMenuAncestors(t *testing.T) {
	items := []menuItem{
		{ID: 1, ParentID: 0},
		{ID: 2, ParentID: 1},
		{ID: 3, ParentID: 2},
		{ID: 4, ParentID: 1},
	}

	allowed := expandAllowedMenuAncestors(items, map[int64]bool{3: true})
	for _, id := range []int64{1, 2, 3} {
		if !allowed[id] {
			t.Fatalf("expected menu %d to be included", id)
		}
	}
	if allowed[4] {
		t.Fatal("did not expect an ungranted sibling to be included")
	}
}

func TestSortMenuItemsUsesStableHierarchyOrder(t *testing.T) {
	items := []menuItem{
		{ID: 4, ParentID: 1, Order: 0},
		{ID: 3, ParentID: 0, Order: 1},
		{ID: 2, ParentID: 0, Order: 0},
		{ID: 1, ParentID: 0, Order: 0},
	}

	sortMenuItems(items)
	want := []int64{1, 2, 3, 4}
	for index, id := range want {
		if items[index].ID != id {
			t.Fatalf("expected id %d at index %d, got %d", id, index, items[index].ID)
		}
	}
}

func TestDirectoryWithoutChildrenDoesNotBecomePage(t *testing.T) {
	menu := menuItem{
		ID:    1,
		Type:  menuTypeDirectory,
		Title: "Empty directory",
		URI:   "/custom-directory",
	}.toVbenMenu(nil)

	if menu.Component != "" {
		t.Fatalf("expected directory component to be empty, got %q", menu.Component)
	}
}

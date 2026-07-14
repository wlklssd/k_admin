package vbenapi

import (
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRegisterMenuManagementRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	registerMenuManagementRoutes(engine.Group("/api"), &Store{})

	wanted := map[string]bool{
		"PUT /api/admin-menus":     false,
		"PUT /api/admin-menus/:id": false,
	}
	for _, route := range engine.Routes() {
		key := route.Method + " " + route.Path
		if _, exists := wanted[key]; exists {
			wanted[key] = true
		}
	}
	for route, registered := range wanted {
		if !registered {
			t.Fatalf("route %s was not registered", route)
		}
	}
}

func TestValidateMenuPositions(t *testing.T) {
	menus := []managedMenu{
		{ID: 1, ParentID: 0, Type: menuTypeDirectory, Order: 0},
		{ID: 2, ParentID: 1, Type: menuTypeItem, Order: 0},
		{ID: 3, ParentID: 0, Type: menuTypeDirectory, Order: 1},
	}

	tests := []struct {
		name      string
		positions []menuPosition
		wantError bool
	}{
		{
			name: "valid reparent and reorder",
			positions: []menuPosition{
				{ID: 3, ParentID: 0, Order: 0},
				{ID: 1, ParentID: 0, Order: 1},
				{ID: 2, ParentID: 3, Order: 0},
			},
		},
		{
			name: "missing menu",
			positions: []menuPosition{
				{ID: 1, ParentID: 0, Order: 0},
				{ID: 2, ParentID: 1, Order: 0},
			},
			wantError: true,
		},
		{
			name: "duplicate sibling order",
			positions: []menuPosition{
				{ID: 1, ParentID: 0, Order: 0},
				{ID: 2, ParentID: 1, Order: 0},
				{ID: 3, ParentID: 0, Order: 0},
			},
			wantError: true,
		},
		{
			name: "cyclic hierarchy",
			positions: []menuPosition{
				{ID: 1, ParentID: 2, Order: 0},
				{ID: 2, ParentID: 1, Order: 0},
				{ID: 3, ParentID: 0, Order: 0},
			},
			wantError: true,
		},
		{
			name: "duplicate menu",
			positions: []menuPosition{
				{ID: 1, ParentID: 0, Order: 0},
				{ID: 1, ParentID: 0, Order: 1},
				{ID: 3, ParentID: 0, Order: 2},
			},
			wantError: true,
		},
		{
			name: "unknown parent",
			positions: []menuPosition{
				{ID: 1, ParentID: 0, Order: 0},
				{ID: 2, ParentID: 99, Order: 0},
				{ID: 3, ParentID: 0, Order: 1},
			},
			wantError: true,
		},
		{
			name: "menu item as parent",
			positions: []menuPosition{
				{ID: 1, ParentID: 0, Order: 0},
				{ID: 2, ParentID: 0, Order: 1},
				{ID: 3, ParentID: 2, Order: 0},
			},
			wantError: true,
		},
		{
			name: "negative order",
			positions: []menuPosition{
				{ID: 1, ParentID: 0, Order: 0},
				{ID: 2, ParentID: 1, Order: -1},
				{ID: 3, ParentID: 0, Order: 1},
			},
			wantError: true,
		},
		{
			name: "non-continuous sibling orders",
			positions: []menuPosition{
				{ID: 1, ParentID: 0, Order: 0},
				{ID: 2, ParentID: 1, Order: 0},
				{ID: 3, ParentID: 0, Order: 2},
			},
			wantError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateMenuPositions(menus, test.positions)
			if test.wantError && err == nil {
				t.Fatal("expected validation error")
			}
			if !test.wantError && err != nil {
				t.Fatalf("unexpected validation error: %v", err)
			}
		})
	}
}

func TestValidateMenuType(t *testing.T) {
	for _, menuType := range []int64{menuTypeDirectory, menuTypeItem} {
		if err := validateMenuType(menuType); err != nil {
			t.Fatalf("expected menu type %d to be valid: %v", menuType, err)
		}
	}
	if err := validateMenuType(2); err == nil {
		t.Fatal("expected unsupported menu type to be rejected")
	}
}

func TestMenuSeedType(t *testing.T) {
	leaf := menuSeed{Title: "leaf"}
	directory := menuSeed{Title: "directory", Children: []menuSeed{leaf}}

	if got := leaf.menuType(); got != menuTypeItem {
		t.Fatalf("expected leaf type %d, got %d", menuTypeItem, got)
	}
	if got := directory.menuType(); got != menuTypeDirectory {
		t.Fatalf("expected directory type %d, got %d", menuTypeDirectory, got)
	}
}

package vbenapi

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/GoAdminGroup/go-admin/modules/db"
	"github.com/gin-gonic/gin"
)

const (
	defaultHomePath = "/dashboard/workspace"
)

type Store struct {
	conn           db.Connection
	configMu       sync.Mutex
	menuMutationMu sync.Mutex
	auth           *authService
}

func Register(r *gin.Engine, conn db.Connection) error {
	s := &Store{
		conn: conn,
		auth: newAuthServiceFromEnv(),
	}
	if err := s.syncDefaultPermissions(); err != nil {
		return fmt.Errorf("同步默认权限失败: %w", err)
	}
	if err := s.syncDefaultMenus(); err != nil {
		return fmt.Errorf("同步默认菜单失败: %w", err)
	}

	api := r.Group("/api", cors())
	api.OPTIONS("/*path", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	registerAuthRoutes(api, s)
	registerUserRoutes(api, s)
	registerMenuRoutes(api, s)
	registerRBACRoutes(api, s)
	registerMenuManagementRoutes(api, s)
	registerUserManagementRoutes(api, s)
	registerDictionaryRoutes(api, s)
	registerModuleRoutes(api, s)
	registerSystemConfigRoutes(api, s)
	registerLogRoutes(api, s)
	return nil
}

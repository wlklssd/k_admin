package vbenapi

import (
	"log"
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

func Register(r *gin.Engine, conn db.Connection) {
	s := &Store{
		conn: conn,
		auth: newAuthServiceFromEnv(),
	}
	if err := s.syncDefaultMenus(); err != nil {
		log.Printf("sync vben menus failed: %v", err)
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
}

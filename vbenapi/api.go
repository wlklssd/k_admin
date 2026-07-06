package vbenapi

import (
	"net/http"
	"sync"
	"time"

	"github.com/GoAdminGroup/go-admin/modules/db"
	"github.com/gin-gonic/gin"
)

const (
	defaultHomePath = "/dashboard"
	tokenTTL        = 2 * time.Hour
)

type tokenInfo struct {
	UserID    int64
	ExpiresAt time.Time
}

type Store struct {
	conn   db.Connection
	mu     sync.RWMutex
	tokens map[string]tokenInfo
}

func Register(r *gin.Engine, conn db.Connection) {
	s := &Store{
		conn:   conn,
		tokens: make(map[string]tokenInfo),
	}

	api := r.Group("/api", cors())
	api.OPTIONS("/*path", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	registerAuthRoutes(api, s)
	registerUserRoutes(api, s)
	registerMenuRoutes(api, s)
	registerRBACRoutes(api, s)
	registerUserManagementRoutes(api, s)
	registerDictionaryRoutes(api, s)
	registerModuleRoutes(api, s)
}

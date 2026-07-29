package kadmin

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/GoAdminGroup/go-admin/internal/kadmin/modules/files"
	"github.com/GoAdminGroup/go-admin/internal/kadmin/modules/jobs"
	"github.com/GoAdminGroup/go-admin/internal/kadmin/modules/loginlogs"
	"github.com/GoAdminGroup/go-admin/internal/kadmin/modules/monitor"
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
	security       *securityService
	audit          *businessAuditRecorder
	jobs           *jobs.Manager
	loginLogs      *loginlogs.Manager
	monitor        *monitor.Manager
}

type Runtime struct {
	audit     *businessAuditRecorder
	jobs      *jobs.Manager
	loginLogs *loginlogs.Manager
	monitor   *monitor.Manager
}

func (r *Runtime) Close() {
	if r == nil {
		return
	}
	if r.audit != nil {
		r.audit.Close()
	}
	if r.monitor != nil {
		r.monitor.Close()
	}
	if r.jobs != nil {
		r.jobs.Close()
	}
	if r.loginLogs != nil {
		r.loginLogs.Close()
	}
}

func Register(r *gin.Engine, conn db.Connection) (*Runtime, error) {
	s := &Store{
		conn: conn,
		auth: newAuthServiceFromEnv(),
	}
	s.security = newSecurityService(s.auth)
	if err := s.syncDefaultPermissions(); err != nil {
		return nil, fmt.Errorf("同步默认权限失败: %w", err)
	}
	if err := s.syncDefaultMenus(); err != nil {
		return nil, fmt.Errorf("同步默认菜单失败: %w", err)
	}
	s.audit = newBusinessAuditRecorder(conn)

	api := r.Group("/api", cors(), s.idempotencyMiddleware(), s.businessAuditMiddleware())
	api.OPTIONS("/*path", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	loginLogManager, err := loginlogs.Register(api, loginlogs.Dependencies{
		Connection:        s.conn,
		RequireAuth:       s.requireAuth(),
		RequirePermission: s.requirePermission,
	})
	if err != nil {
		s.audit.Close()
		return nil, fmt.Errorf("初始化登录审计失败: %w", err)
	}
	s.loginLogs = loginLogManager

	manager, err := jobs.Register(api, jobs.Dependencies{
		Connection:        s.conn,
		RequireAuth:       s.requireAuth(),
		RequirePermission: s.requirePermission,
		RefreshCache:      s.refreshBuiltInCache,
	})
	if err != nil {
		loginLogManager.Close()
		s.audit.Close()
		return nil, fmt.Errorf("初始化定时任务失败: %w", err)
	}
	s.jobs = manager
	monitorManager, err := monitor.Register(api, monitor.Dependencies{
		Connection:        s.conn,
		RequireAuth:       s.requireAuth(),
		RequirePermission: s.requirePermission,
	})
	if err != nil {
		manager.Close()
		loginLogManager.Close()
		s.audit.Close()
		return nil, fmt.Errorf("初始化系统监控失败: %w", err)
	}
	s.monitor = monitorManager
	registerApplicationRoutes(api, s)
	return &Runtime{audit: s.audit, jobs: manager, loginLogs: loginLogManager, monitor: monitorManager}, nil
}

func registerApplicationRoutes(api *gin.RouterGroup, s *Store) {
	registerAuthRoutes(api, s)
	registerUserRoutes(api, s)
	registerMenuRoutes(api, s)
	registerRBACRoutes(api, s)
	registerMenuManagementRoutes(api, s)
	registerUserManagementRoutes(api, s)
	registerDictionaryRoutes(api, s)
	files.Register(api, files.Dependencies{
		Connection:        s.conn,
		RequireAuth:       s.requireAuth(),
		RequireAdmin:      s.requireAdmin(),
		RequirePermission: s.requirePermission,
	})
	registerSystemConfigRoutes(api, s)
	registerLogRoutes(api, s)
	if s.loginLogs == nil {
		loginlogs.RegisterRoutes(api, nil, loginlogs.Dependencies{
			RequireAuth:       s.requireAuth(),
			RequirePermission: s.requirePermission,
		})
	}
	if s.jobs == nil {
		jobs.RegisterRoutes(api, nil, jobs.Dependencies{
			RequireAuth:       s.requireAuth(),
			RequirePermission: s.requirePermission,
		})
	}
	if s.monitor == nil {
		monitor.RegisterRoutes(api, nil, monitor.Dependencies{
			RequireAuth:       s.requireAuth(),
			RequirePermission: s.requirePermission,
		})
	}
}

func (s *Store) refreshBuiltInCache() error {
	if err := s.syncDefaultPermissions(); err != nil {
		return err
	}
	s.menuMutationMu.Lock()
	defer s.menuMutationMu.Unlock()
	return s.syncDefaultMenus()
}

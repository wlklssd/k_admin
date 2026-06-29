package vbenapi

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/GoAdminGroup/go-admin/modules/auth"
	"github.com/GoAdminGroup/go-admin/modules/db"
	"github.com/GoAdminGroup/go-admin/plugins/admin/models"
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

	authGroup := api.Group("/auth")
	authGroup.POST("/login", s.login)
	authGroup.POST("/logout", s.requireAuth(), s.logout)
	authGroup.GET("/codes", s.requireAuth(), s.accessCodes)

	userGroup := api.Group("/user", s.requireAuth())
	userGroup.GET("/info", s.userInfo)
	userGroup.GET("/menu", s.menus)
	userGroup.GET("/menu/list", s.menus)
	userGroup.GET("/permissions", s.accessCodes)

	menuGroup := api.Group("/menu", s.requireAuth())
	menuGroup.GET("/all", s.menus)
	menuGroup.GET("/list", s.menus)
}

type loginRequest struct {
	Username string `json:"username" form:"username"`
	Password string `json:"password" form:"password"`
}

type loginResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	Token        string `json:"token"`
	TokenType    string `json:"tokenType"`
	ExpiresAt    int64  `json:"expiresAt"`
}

func (s *Store) login(c *gin.Context) {
	var req loginRequest
	_ = c.ShouldBind(&req)
	if req.Username == "" {
		req.Username = c.PostForm("username")
	}
	if req.Password == "" {
		req.Password = c.PostForm("password")
	}

	user, ok := auth.Check(req.Password, req.Username, s.conn)
	if !ok {
		fail(c, http.StatusUnauthorized, "wrong username or password")
		return
	}

	token, expiresAt, err := s.issueToken(user.Id)
	if err != nil {
		fail(c, http.StatusInternalServerError, "create token failed")
		return
	}

	success(c, loginResponse{
		AccessToken:  token,
		RefreshToken: token,
		Token:        token,
		TokenType:    "Bearer",
		ExpiresAt:    expiresAt.UnixMilli(),
	})
}

func (s *Store) logout(c *gin.Context) {
	token := tokenFromRequest(c)
	if token != "" {
		s.mu.Lock()
		delete(s.tokens, token)
		s.mu.Unlock()
	}
	success(c, true)
}

func (s *Store) userInfo(c *gin.Context) {
	user, ok := s.currentUser(c)
	if !ok {
		fail(c, http.StatusUnauthorized, "invalid token")
		return
	}

	success(c, gin.H{
		"userId":      user.Id,
		"username":    user.UserName,
		"realName":    user.Name,
		"avatar":      user.Avatar,
		"desc":        "",
		"roles":       roleSlugs(user),
		"accessCodes": accessCodes(user),
		"homePath":    defaultHomePath,
	})
}

func (s *Store) accessCodes(c *gin.Context) {
	user, ok := s.currentUser(c)
	if !ok {
		fail(c, http.StatusUnauthorized, "invalid token")
		return
	}
	success(c, accessCodes(user))
}

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

func (s *Store) requireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := tokenFromRequest(c)
		if token == "" {
			fail(c, http.StatusUnauthorized, "missing token")
			c.Abort()
			return
		}

		s.mu.RLock()
		info, ok := s.tokens[token]
		s.mu.RUnlock()
		if !ok || time.Now().After(info.ExpiresAt) {
			s.mu.Lock()
			delete(s.tokens, token)
			s.mu.Unlock()
			fail(c, http.StatusUnauthorized, "invalid token")
			c.Abort()
			return
		}

		c.Set("vben_user_id", info.UserID)
		c.Set("vben_token", token)
		c.Next()
	}
}

func (s *Store) currentUser(c *gin.Context) (models.UserModel, bool) {
	idVal, ok := c.Get("vben_user_id")
	if !ok {
		return models.User(), false
	}

	id, ok := idVal.(int64)
	if !ok {
		return models.User(), false
	}

	user := models.User().SetConn(s.conn).Find(id)
	if user.IsEmpty() {
		return user, false
	}

	user = user.WithRoles().WithPermissions().WithMenus()
	return user, user.HasMenu()
}

func (s *Store) issueToken(userID int64) (string, time.Time, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", time.Time{}, err
	}

	token := hex.EncodeToString(raw)
	expiresAt := time.Now().Add(tokenTTL)

	s.mu.Lock()
	s.tokens[token] = tokenInfo{UserID: userID, ExpiresAt: expiresAt}
	s.mu.Unlock()

	return token, expiresAt, nil
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

func roleSlugs(user models.UserModel) []string {
	roles := make([]string, 0, len(user.Roles))
	for _, role := range user.Roles {
		if role.Slug != "" {
			roles = append(roles, role.Slug)
		}
	}
	if len(roles) == 0 && user.IsSuperAdmin() {
		roles = append(roles, "super")
	}
	return roles
}

func accessCodes(user models.UserModel) []string {
	set := make(map[string]bool)
	for _, role := range roleSlugs(user) {
		set[role] = true
	}
	for _, permission := range user.Permissions {
		if permission.Slug != "" {
			set[permission.Slug] = true
		}
	}

	codes := make([]string, 0, len(set))
	for code := range set {
		codes = append(codes, code)
	}
	return codes
}

func tokenFromRequest(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
			return strings.TrimSpace(parts[1])
		}
		return strings.TrimSpace(authHeader)
	}

	if token := c.GetHeader("Access-Token"); token != "" {
		return token
	}
	if token := c.GetHeader("X-Access-Token"); token != "" {
		return token
	}
	if token := c.Query("token"); token != "" {
		return token
	}
	return ""
}

func success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "ok",
		"msg":     "ok",
		"data":    data,
	})
}

func fail(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{
		"code":    status,
		"message": message,
		"msg":     message,
		"data":    nil,
	})
}

func cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Headers", "Authorization, Access-Token, X-Access-Token, Content-Type, X-Requested-With")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		}
		c.Next()
	}
}

func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	switch value := v.(type) {
	case string:
		return value
	case []byte:
		return string(value)
	default:
		return ""
	}
}

func toInt64(v interface{}) int64 {
	switch value := v.(type) {
	case int:
		return int64(value)
	case int8:
		return int64(value)
	case int16:
		return int64(value)
	case int32:
		return int64(value)
	case int64:
		return value
	case uint:
		return int64(value)
	case uint8:
		return int64(value)
	case uint16:
		return int64(value)
	case uint32:
		return int64(value)
	case uint64:
		return int64(value)
	case []byte:
		i, _ := strconv.ParseInt(string(value), 10, 64)
		return i
	case string:
		i, _ := strconv.ParseInt(value, 10, 64)
		return i
	default:
		return 0
	}
}

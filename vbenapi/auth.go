package vbenapi

import (
	"net/http"

	"github.com/GoAdminGroup/go-admin/modules/auth"
	"github.com/gin-gonic/gin"
)

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

func registerAuthRoutes(api *gin.RouterGroup, s *Store) {
	authGroup := api.Group("/auth")
	authGroup.POST("/login", s.login)
	authGroup.POST("/logout", s.requireAuth(), s.logout)
	authGroup.GET("/codes", s.requireAuth(), s.accessCodes)
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

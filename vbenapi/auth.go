package vbenapi

import (
	"net/http"
	"strings"

	"github.com/GoAdminGroup/go-admin/modules/auth"
	"github.com/gin-gonic/gin"
)

type loginRequest struct {
	Username string `json:"username" form:"username"`
	Password string `json:"password" form:"password"`
}

type refreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" form:"refreshToken"`
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
	authGroup.POST("/refresh", s.refreshToken)
	authGroup.POST("/logout", s.logout)
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
		user, ok = s.checkConfiguredDefaultAdmin(req.Username, req.Password)
		if !ok {
			fail(c, http.StatusUnauthorized, "wrong username or password")
			return
		}
	}
	enabled, err := s.userAccountEnabled(user.Id)
	if err != nil {
		fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	if !enabled {
		fail(c, http.StatusForbidden, "account disabled")
		return
	}

	tokens, err := s.auth.issueTokenPair(user.Id)
	if err != nil {
		fail(c, http.StatusInternalServerError, "create token failed")
		return
	}

	success(c, loginResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		Token:        tokens.AccessToken,
		TokenType:    "Bearer",
		ExpiresAt:    tokens.AccessExpiresAt.UnixMilli(),
	})
}

func (s *Store) refreshToken(c *gin.Context) {
	refreshToken := refreshTokenFromRequest(c)
	if refreshToken == "" {
		fail(c, http.StatusUnauthorized, "missing refresh token")
		return
	}

	userID, err := s.auth.consumeRefreshToken(refreshToken)
	if err != nil {
		if err == errInvalidRefreshToken {
			fail(c, http.StatusUnauthorized, "invalid refresh token")
			return
		}
		fail(c, http.StatusInternalServerError, "auth storage unavailable")
		return
	}

	enabled, err := s.userAccountEnabled(userID)
	if err != nil {
		fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	if !enabled {
		fail(c, http.StatusForbidden, "account disabled")
		return
	}

	tokens, err := s.auth.issueTokenPair(userID)
	if err != nil {
		fail(c, http.StatusInternalServerError, "create token failed")
		return
	}

	success(c, loginResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		Token:        tokens.AccessToken,
		TokenType:    "Bearer",
		ExpiresAt:    tokens.AccessExpiresAt.UnixMilli(),
	})
}

func (s *Store) logout(c *gin.Context) {
	if token := tokenFromRequest(c); token != "" {
		if claims, err := s.auth.parseAccessToken(token); err == nil {
			if err := s.auth.blacklistAccessToken(claims); err != nil {
				fail(c, http.StatusInternalServerError, "auth storage unavailable")
				return
			}
		}
	}

	if err := s.auth.revokeRefreshToken(refreshTokenFromRequest(c)); err != nil {
		fail(c, http.StatusInternalServerError, "auth storage unavailable")
		return
	}

	success(c, true)
}

func refreshTokenFromRequest(c *gin.Context) string {
	var req refreshTokenRequest
	_ = c.ShouldBind(&req)
	if token := strings.TrimSpace(req.RefreshToken); token != "" {
		return token
	}
	if token := strings.TrimSpace(c.GetHeader("X-Refresh-Token")); token != "" {
		return token
	}
	if token := strings.TrimSpace(c.Query("refreshToken")); token != "" {
		return token
	}
	return ""
}

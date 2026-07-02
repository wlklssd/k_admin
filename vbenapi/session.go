package vbenapi

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"github.com/GoAdminGroup/go-admin/plugins/admin/models"
	"github.com/gin-gonic/gin"
)

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
	return user, user.HasMenu() || isAdminUser(user.UserName)
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

package vbenapi

import (
	"net/http"

	"github.com/GoAdminGroup/go-admin/plugins/admin/models"
	"github.com/gin-gonic/gin"
)

func registerUserRoutes(api *gin.RouterGroup, s *Store) {
	userGroup := api.Group("/user", s.requireAuth())
	userGroup.GET("/info", s.userInfo)
	userGroup.GET("/menu", s.menus)
	userGroup.GET("/menu/list", s.menus)
	userGroup.GET("/permissions", s.accessCodes)
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

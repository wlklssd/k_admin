package vbenapi

import "github.com/gin-gonic/gin"

func registerModuleRoutes(api *gin.RouterGroup, s *Store) {
	registerGeneratedRoutes(api, s)
	registerCustomModuleRoutes(api, s)
}

func registerGeneratedRoutes(api *gin.RouterGroup, s *Store) {
	// Generated business API routes should be registered here.
}

func registerCustomModuleRoutes(api *gin.RouterGroup, s *Store) {
	registerRBACRoutes(api, s)
	// Hand-written business API routes should be registered here.
}

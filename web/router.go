package web

import (
	"kubernetes-update-manager/updater/manager"

	"github.com/getsentry/raven-go"
	"github.com/gin-contrib/sentry"
	"github.com/gin-gonic/gin"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
)

// @title Kubernetes Update Manager API
// @version 1.0
// @description API to update configurations in a kubernetes cluster without needing to have full access to the kubernetes API.

// @contact.name XCNT DEV Team
// @contact.url https://xcnt.io
// @contact.email dev@xcnt.io

// @license.name MIT
// @license.url https://tldrlegal.com/license/mit-license

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

// GetWeb returns the initialized web engine.
func GetWeb(config *Config) *gin.Engine {
	engine, _ := getWeb(config)
	return engine
}

func getWeb(config *Config) (*gin.Engine, *manager.Manager) {
	router := gin.New()
	router.Use(sentry.Recovery(raven.DefaultClient, false))

	mgr := registerRoutes(router, config)
	return router, mgr
}

func registerRoutes(router *gin.Engine, config *Config) *manager.Manager {
	router.GET("/health", CheckHealth(config))
	router.GET("/swagger/*any", ginSwagger.DisablingWrapHandler(swaggerFiles.Handler, "NAME_OF_ENV_VARIABLE"))
	return registerUpdaterRoutes(router, config)
}

func registerUpdaterRoutes(router *gin.Engine, config *Config) *manager.Manager {
	updater := NewUpdaterHandler(config)
	router.GET("/updates/:uuid", updater.GetItem)
	router.DELETE("/updates/:uuid", updater.Delete)
	router.POST("/updates", updater.Post)
	return updater.manager
}

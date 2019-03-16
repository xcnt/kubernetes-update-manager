package web

import (
	"crypto/subtle"
	"kubernetes-update-manager/updater/manager"
	"net/http"
	"strings"

	"github.com/getsentry/raven-go"
	"github.com/gin-contrib/sentry"
	"github.com/gin-gonic/gin"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"

	// Side loading docs
	_ "kubernetes-update-manager/web/docs"
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
	engine, _ := getWeb(config, true)
	return engine
}

func getWeb(config *Config, useMiddlewares bool) (*gin.Engine, *manager.Manager) {
	router := gin.New()
	if useMiddlewares {
		router.Use(gin.Logger(), gin.Recovery(), sentry.Recovery(raven.DefaultClient, false))
	}

	mgr := registerRoutes(router, config)
	return router, mgr
}

func registerRoutes(router *gin.Engine, config *Config) *manager.Manager {
	router.GET("/health", CheckHealth(config))
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	return registerUpdaterRoutes(router, config)
}

func registerUpdaterRoutes(router *gin.Engine, config *Config) *manager.Manager {
	updater := NewUpdaterHandler(config)
	authCheck := RequireAuth(config.APIKey)
	router.GET("/updates/:uuid", authCheck, updater.GetItem)
	router.DELETE("/updates/:uuid", authCheck, updater.Delete)
	router.POST("/updates", authCheck, updater.Post)
	return updater.manager
}

// RequireAuth returns a usable middleware who includes authorization checks in the given endpoint.
func RequireAuth(apiKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authorization := c.GetHeader("Authorization")
		authorizationData := strings.SplitN(authorization, " ", 2)
		authorizationString := authorizationData[len(authorizationData)-1]

		if SecureCompare(authorizationString, apiKey) {
			c.Next()
		} else {
			c.AbortWithStatus(http.StatusUnauthorized)
		}
	}
}

// SecureCompare compares two strings in a time constant way to avoid possible timing attacks on the password check.
func SecureCompare(left, right string) bool {
	for len(left) < len(right) {
		left += " "
	}
	left = left[:len(right)]
	return subtle.ConstantTimeCompare([]byte(left), []byte(right)) == 1
}

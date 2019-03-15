package web

import (
	"github.com/gin-gonic/gin"
	"github.com/gookit/color"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CheckHealth can be used as a health endpoint to verify that the API is accessible.
// @Summary Check Health of API
// @Description Tries to reach the kubernetes API and returns if it can be reached.
// @Success 204
// @Failure 500
// @Router /health [get]
func CheckHealth(config *Config) gin.HandlerFunc {
	return func(context *gin.Context) {
		clientSet := config.Clientset
		_, err := clientSet.Core().Nodes().List(metaV1.ListOptions{})
		if err != nil {
			color.Error.Println(err.Error())
			context.Status(500)
		} else {
			context.Status(204)
		}
	}
}

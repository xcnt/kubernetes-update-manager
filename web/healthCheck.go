package web

import (
	"github.com/gin-gonic/gin"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CheckHealth can be used as a health endpoint to verify that the API is accessible.
func CheckHealth(config *Config) gin.HandlerFunc {
	return func(context *gin.Context) {
		clientSet := config.Clientset
		_, err := clientSet.Core().Nodes().List(metaV1.ListOptions{})
		if err != nil {
			context.Status(500)
		} else {
			context.Status(204)
		}
	}
}

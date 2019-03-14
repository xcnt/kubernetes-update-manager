package web

import (
	"kubernetes-update-manager/updater"
	"kubernetes-update-manager/updater/manager"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

const (
	// UUIDParam is the parameter which the API expects to have
	UUIDParam = "uuid"
	// ImageParam is the parameter for the specified image
	ImageParam = "image"
	// UpdateClassifierParam returns the parameter name for the update classification configuration
	UpdateClassifierParam = "update_classifier"
)

// NewUpdaterHandler configuration configures an updaterhandler which can be used to register endpoints for gin requests.
func NewUpdaterHandler(config *Config) *UpdaterHandler {
	return &UpdaterHandler{
		config:  config,
		manager: manager.NewManager(config.Clientset),
	}
}

// UpdaterHandler represents the state necessary in a web interface context to handle update requests.
type UpdaterHandler struct {
	config  *Config
	manager *manager.Manager
}

// GetItem represents the get method for the specified get item.
func (updateHandler *UpdaterHandler) GetItem(context *gin.Context) {
	manager := updateHandler.manager
	defer manager.Cleanup()
	uuidString := context.Param(UUIDParam)
	updateProgress, err := manager.GetByString(uuidString)
	if os.IsNotExist(err) {
		context.Status(http.StatusNotFound)
	} else if err != nil {
		context.Status(http.StatusBadRequest)
	} else {
		context.JSON(200, serializeUpdateProgress(updateProgress))
	}
}

// Post represents the POST method to create an update request.
func (updateHandler *UpdaterHandler) Post(context *gin.Context) {
	manager := updateHandler.manager
	defer manager.Cleanup()
	config := updateHandler.config
	imageString := context.Param(ImageParam)
	updateClassifier := context.Param(UpdateClassifierParam)
	namespaces := config.Namespaces
	if config.AutoloadNamespaces {
		var err error
		namespaces, err = updater.ListNamespaces(updater.NewClientsetWrapper(config.Clientset))
		if err != nil {
			context.AbortWithError(500, err)
			return
		}
	}
	updateConfig := updater.NewConfig(config.Clientset, updater.NewImage(imageString), updateClassifier)
	updateConfig.SetNamespaces(namespaces)
	updateProgress, err := manager.Create(updateConfig)
	if err != nil {
		context.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	context.JSON(200, serializeUpdateProgress(updateProgress))
}

// Delete represents the DELETE method to remove an update request from the manager.
func (updateHandler *UpdaterHandler) Delete(context *gin.Context) {
	manager := updateHandler.manager
	defer manager.Cleanup()
	uuid := context.Param(UUIDParam)
	manager.DeleteByString(uuid)
	context.Status(http.StatusNoContent)
}

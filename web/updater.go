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
// @Summary Retrieves update information
// @Description retrieves via an uuid the current information of an update progress.
// @Tags updates
// @Produce json
// @Param uuid path string true "The uuid of the update progress which information should be requested"
// @Security ApiKeyAuth
// @Success 200 {object} web.UpdateProgressSerialized
// @Failure 404
// @Failure 400
// @Failure 401
// @Router /updates/{uuid} [get]
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
		context.JSON(http.StatusOK, serializeUpdateProgress(updateProgress))
	}
}

// Post represents the POST method to create an update request.
// @Summary Creates an update
// @Description retrieves via an uuid the current information of an update progress.
// @Tags updates
// @Produce json
// @Security ApiKeyAuth
// @Param image body string true "The image included in the update request"
// @Param update_classifier body string true "The update classifier which should be used for searching for the update status"
// @Success 200 {object} web.UpdateProgressSerialized
// @Failure 400
// @Failure 500
// @Failure 401
// @Router /updates [post]
func (updateHandler *UpdaterHandler) Post(context *gin.Context) {
	manager := updateHandler.manager
	defer manager.Cleanup()
	config := updateHandler.config
	imageString, _ := context.GetPostForm(ImageParam)
	updateClassifier, _ := context.GetPostForm(UpdateClassifierParam)
	if len(imageString) == 0 {
		context.AbortWithStatus(http.StatusBadRequest)
		return
	}
	if len(updateClassifier) == 0 {
		context.AbortWithStatus(http.StatusBadRequest)
		return
	}
	namespaces := config.Namespaces
	if config.AutoloadNamespaces {
		var err error
		namespaces, err = updater.ListNamespaces(updater.NewClientsetWrapper(config.Clientset))
		if err != nil {
			context.AbortWithError(http.StatusInternalServerError, err)
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
	context.JSON(http.StatusCreated, serializeUpdateProgress(updateProgress))
}

// Delete represents the DELETE method to remove an update request from the manager.
// @Summary Deletes a status information of an update
// @Description deletes a status update for the provided uuid
// @Param uuid path string true "The uuid of the update progress which information should be requested"
// @Security ApiKeyAuth
// @Success 204
// @Failure 401
// @Router /updates/{uuid} [delete]
func (updateHandler *UpdaterHandler) Delete(context *gin.Context) {
	manager := updateHandler.manager
	defer manager.Cleanup()
	uuid := context.Param(UUIDParam)
	manager.DeleteByString(uuid)
	context.Status(http.StatusNoContent)
}

package client

import (
	"errors"
	"fmt"
	"kubernetes-update-manager/web"
	"net/http"
	"net/url"
	"os"
	"path"

	"github.com/google/uuid"
	"github.com/levigross/grequests"
)

const (
	// ImageParam is the parameter name of the specific image
	ImageParam = web.ImageParam
	// UpdateClassifierParam is the parameter used to be sent to the client
	UpdateClassifierParam = web.UpdateClassifierParam
)

var (
	// ErrUnauthorized is returned when an API can not be requested.
	ErrUnauthorized = errors.New("unauthorized")
)

// NewUpdateExecution returns a newly created, not yet executed instance of the update execution configuration.
func NewUpdateExecution(command *UpdateCommand) *UpdateExecution {
	return &UpdateExecution{
		updateCommand: command,
	}
}

// UpdateExecution holds informations about a specific update and allows to retrieve the current information from a remote update manager.
type UpdateExecution struct {
	updateProgressUUID string
	updateCommand      *UpdateCommand
}

// UUID returns the uuid assigned to the update progress which is represented by the execution status.
func (updateExecution *UpdateExecution) UUID() uuid.UUID {
	u, err := uuid.Parse(updateExecution.updateProgressUUID)
	if err != nil {
		return uuid.Nil
	}
	return u
}

// authenticatedRequestOptions returns pre authenticated request options.
func (updateExecution *UpdateExecution) authenticatedRequestOptions() *grequests.RequestOptions {
	updateCommand := updateExecution.updateCommand
	return &grequests.RequestOptions{
		Headers: map[string]string{
			"Authorization": fmt.Sprintf("APIKey %s", updateCommand.APIKey),
		},
	}
}

// Start starts the request pipeline for the command configuration. It returns ErrUnauthorized if the authentication with the remote server fails.
func (updateExecution *UpdateExecution) Start() error {
	updateCommand := updateExecution.updateCommand
	request := updateExecution.authenticatedRequestOptions()
	request.Data = map[string]string{
		ImageParam:            updateCommand.Image,
		UpdateClassifierParam: updateCommand.UpdateClassifier,
	}
	response, err := grequests.Post(updateCommand.TargetEndpoint, request)
	if err != nil {
		return err
	}
	err = verifyRemoteStatusCode(response.StatusCode)
	if err != nil {
		return err
	}

	updateProgressSerialized := &web.UpdateProgressSerialized{}
	err = response.JSON(updateProgressSerialized)
	if err != nil {
		return err
	}

	updateExecution.updateProgressUUID = updateProgressSerialized.UUID
	return nil
}

// Get retrieves the current information for the update progress to be returned. It returns os.ErrNotExist, if the update progress with the specified uuid does not exist. It returns ErrUnauthorized if the authentication with the remote server fails.
func (updateExecution *UpdateExecution) Get() (*web.UpdateProgressSerialized, error) {
	options := updateExecution.authenticatedRequestOptions()

	objectURL := updateExecution.objectURL()
	response, err := grequests.Get(objectURL.String(), options)
	if err != nil {
		return nil, err
	}
	err = verifyRemoteStatusCode(response.StatusCode)
	if err != nil {
		return nil, err
	}

	updateProgressSerialized := &web.UpdateProgressSerialized{}
	err = response.JSON(updateProgressSerialized)
	if err != nil {
		return nil, err
	}

	return updateProgressSerialized, nil
}

// Finish deletes the update progress on the update manager. It should be called when no more information needs to be returned. It returns ErrUnauthorized if the authentication with the remote server fails.
func (updateExecution *UpdateExecution) Finish() error {
	options := updateExecution.authenticatedRequestOptions()
	objectURL := updateExecution.objectURL()
	response, err := grequests.Delete(objectURL.String(), options)
	if err != nil {
		return err
	}
	err = verifyRemoteStatusCode(response.StatusCode)
	if err != nil {
		return err
	}
	return nil
}

func (updateExecution *UpdateExecution) objectURL() *url.URL {
	parsedURL, _ := url.Parse(updateExecution.updateCommand.TargetEndpoint)
	parsedURL.Path = path.Join(parsedURL.Path, updateExecution.UUID().String())
	return parsedURL
}

func verifyRemoteStatusCode(statusCode int) error {
	switch statusCode {
	case http.StatusNotFound:
		return os.ErrNotExist
	case http.StatusUnauthorized:
		return ErrUnauthorized
	case http.StatusOK, http.StatusNoContent, http.StatusCreated:
		break
		// Ok
	default:
		return fmt.Errorf("Unexpected status code %d", statusCode)
	}
	return nil
}

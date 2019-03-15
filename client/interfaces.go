package client

import (
	"kubernetes-update-manager/web"

	"github.com/google/uuid"
)

// ExecutionStatus is the interface to retrieve from a remote server informationen about a current update progress
type ExecutionStatus interface {
	// UUID returns the uuid assigned to the update progress which is represented by the execution status.
	UUID() uuid.UUID
	// Get retrieves the current information for the update progress to be returned. It returns os.ErrNotExist, if the update progress with the specified uuid does not exist. It returns ErrUnauthorized if the authentication with the remote server fails.
	Get() (*web.UpdateProgressSerialized, error)
	// Finish deletes the update progress on the update manager. It should be called when no more information needs to be returned.
	Finish() error
}

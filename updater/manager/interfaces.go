package manager

import (
	"kubernetes-update-manager/updater"

	"github.com/google/uuid"
)

// UpdateProgress implements an identifiable update manager and identifies by a specified uuid
type UpdateProgress interface {
	// UUID returns the unique identifier for the specified update progress
	UUID() uuid.UUID
	updater.UpdateProgress
}

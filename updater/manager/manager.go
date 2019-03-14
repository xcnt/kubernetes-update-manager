package manager

import (
	"kubernetes-update-manager/updater"
	"os"
	"time"

	"github.com/google/uuid"
	"k8s.io/client-go/kubernetes"
)

// NewManager returns a manager initialized with the provided configuration.
func NewManager(clientset kubernetes.Interface) *Manager {
	return &Manager{
		Update:        updater.Update,
		Plan:          updater.Plan,
		clientset:     clientset,
		updates:       map[uuid.UUID]UpdateProgress{},
		thresholdTime: 10 * time.Minute,
	}
}

// Manager is the main entry point for providing status updates for updates as well as storing them for retrieval.
type Manager struct {
	Update        func(updater.UpdatePlan, updater.KubernetesWrapper) updater.UpdateProgress
	Plan          func(*updater.Config) (updater.UpdatePlan, error)
	clientset     kubernetes.Interface
	updates       map[uuid.UUID]UpdateProgress
	thresholdTime time.Duration
}

// Cleanup removes updates which are finished and passed a specific time threshold after completion
func (manager *Manager) Cleanup() {
	for updateProgressKey, updateProgress := range manager.updates {
		if updateProgress.Finished() && updateProgress.FinishTime().Add(manager.thresholdTime).Before(time.Now()) {
			delete(manager.updates, updateProgressKey)
		}
	}
}

// GetByString returns the element being present in the given string which has to be convertable to an uuid.
// It delegates the actual lookup to the get function and returns an additional error if the provided string is not
// a parseable uuid.
func (manager *Manager) GetByString(uuidString string) (UpdateProgress, error) {
	toGetUUID, err := uuid.Parse(uuidString)
	if err != nil {
		return nil, err
	}
	return manager.Get(toGetUUID)
}

// Get returns the status of the process with the provided uuid. Returns os.ErrNotExist if no update could be found with the provied uuid.
func (manager *Manager) Get(toGetUUID uuid.UUID) (UpdateProgress, error) {
	update, ok := manager.updates[toGetUUID]
	if !ok {
		return nil, os.ErrNotExist
	}
	return update, nil
}

// Schedule takes the specified update plan, starts it and stores the result in the manager.
func (manager *Manager) Schedule(updatePlan updater.UpdatePlan, config *updater.Config) (UpdateProgress, error) {
	updateProgress := WrapUpdateProgress(manager.Update(updatePlan, config))
	manager.updates[updateProgress.UUID()] = updateProgress
	return updateProgress, nil
}

// Create creates and schedules an update plan adn returns the update progress
func (manager *Manager) Create(config *updater.Config) (UpdateProgress, error) {
	updateProgress, err := manager.Plan(config)
	if err != nil {
		return nil, err
	}
	return manager.Schedule(updateProgress, config)
}

// DeleteByString deletes the specific uuid string representation from the update manager. Does nothing
// if the string is not convertable to a uuid or the element does not exist.
func (manager *Manager) DeleteByString(uuidStringToDelete string) {
	toGetUUID, err := uuid.Parse(uuidStringToDelete)
	if err == nil {
		manager.Delete(toGetUUID)
	}
}

// Delete removes the specified uuid from the update manager if it is present
func (manager *Manager) Delete(uuidToDelete uuid.UUID) {
	delete(manager.updates, uuidToDelete)
}

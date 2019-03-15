package client

// UpdateCommand holds the configuration to run an update to the client
type UpdateCommand struct {
	// TargetEndpoint is used to specify the URL which should be used to communicate with the update manager
	TargetEndpoint string
	// Image is the name of the image which should be updated
	Image string
	// UpdateClassifier specifies the classifier which should be communicated to the update manager to run the specified update configuration
	UpdateClassifier string
	// APIKey specifies the api key used for authentication against the kubernetes update manager
	APIKey string
}

// Run executes the update command.
func (updateCommand *UpdateCommand) Run() (ExecutionStatus, error) {
	updateExecution := NewUpdateExecution(updateCommand)
	err := updateExecution.Start()
	if err != nil {
		return nil, err
	}
	return updateExecution, err
}

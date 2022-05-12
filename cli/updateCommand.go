package cli

import (
	"errors"
	"fmt"
	"kubernetes-update-manager/client"
	"kubernetes-update-manager/web"
	"os"
	"time"

	"github.com/gookit/color"
	"github.com/gosuri/uiprogress"

	cli "github.com/urfave/cli/v2"
)

var (
	// FlagURL is the target URL to the remote server.
	FlagURL = &cli.StringFlag{
		Name:    "url",
		Usage:   "The url where the update manager resides in. This must be the complete path! Use http://xcnt.io/updates instead of https://xcnt.io/",
		EnvVars: []string{"UPDATE_MANGER_URL"},
	}
	// FlagImage represents the image which should be executed
	FlagImage = &cli.StringFlag{
		Name:    "image",
		Aliases: []string{"u"},
		Usage:   "The docker image which should be updated.",
		EnvVars: []string{"UPDATE_MANGER_IMAGE"},
	}
	// FlagUpdateClassifier is the update classifier flag which should be sent to the server
	FlagUpdateClassifier = &cli.StringFlag{
		Name:    "update-classifier",
		Aliases: []string{"c", "classifier"},
		Usage:   "The update classifier which should be sent to the server for update.",
		EnvVars: []string{"UPDATE_MANGER_UPDATE_CLASSIFIER", "UPDATE_MANGER_CLASSIFIER"},
	}

	// ErrNoTargetEndpoint is returned if no target endpoint is provided
	ErrNoTargetEndpoint = errors.New("The target endpoint for the remote update manager is not specified")
	// ErrNoImage is returned if no image has been provided ot the update command
	ErrNoImage = errors.New("The image for the update was not provided")
	// ErrNoUpdateClassifier is returned if no update classifier was provided to the update command
	ErrNoUpdateClassifier = errors.New("The update classifier was not provided to the update command")
)

// UpdateCommand can be used to notify a remove server about an update
func UpdateCommand() *cli.Command {
	return &cli.Command{
		Name:    "update",
		Aliases: []string{"u"},
		Usage:   "Runs an update command on a remote server",
		Flags:   UpdateFlags(),
		Action:  UpdateAction,
	}
}

// UpdateFlags return the flags which are available in the update command.
func UpdateFlags() []cli.Flag {
	return []cli.Flag{
		FlagURL,
		FlagImage,
		FlagUpdateClassifier,
		FlagAPIKey,
	}
}

// UpdateAction is the action which is executed when the update command is picked.
func UpdateAction(c *cli.Context) error {
	updateCommand := updateCommandFromContext(c)
	if len(updateCommand.TargetEndpoint) == 0 {
		return ErrNoTargetEndpoint
	}
	if len(updateCommand.Image) == 0 {
		return ErrNoImage
	}
	if len(updateCommand.UpdateClassifier) == 0 {
		return ErrNoUpdateClassifier
	}
	if len(updateCommand.APIKey) == 0 {
		return ErrNoAPIKey
	}

	color.Info.Println(
		fmt.Sprintf("Updating %s with image %s and update classifier %s",
			updateCommand.TargetEndpoint,
			updateCommand.Image,
			updateCommand.UpdateClassifier))
	status, err := updateCommand.Run()
	if err != nil {
		return err
	}

	return monitorUpdate(status)
}

func monitorUpdate(status client.ExecutionStatus) error {
	finished := false

	var err error
	var currentStatus *web.UpdateProgressSerialized
	var jobsProgress *uiprogress.Bar
	var deploymentsProgress *uiprogress.Bar
	uiprogress.Start()

	for !finished {
		currentStatus, err = status.Get()
		if os.IsNotExist(err) {
			color.Warn.Println("Update not found, expect it to be already done and deleted.")
			return nil
		}
		jobsCount := currentStatus.Counts.Jobs
		deploymentsCount := currentStatus.Counts.Deployments

		if jobsProgress == nil && jobsCount.Total > 0 {
			jobsProgress = addJobsBar(jobsCount.Total)
		}
		if deploymentsProgress == nil && deploymentsCount.Total > 0 {
			deploymentsProgress = addDeploymentsBar(deploymentsCount.Total)
		}

		if jobsProgress != nil {
			jobsProgress.Set(jobsCount.Updated)
		}

		if deploymentsProgress != nil {
			deploymentsProgress.Set(deploymentsCount.Updated)
		}
		finished = currentStatus.Status.Finished
		time.Sleep(time.Second * 1)
	}

	if currentStatus.Status.Failed {
		err = errors.New("Update failed")
		color.Error.Println(err.Error())
		os.Exit(1)
		return err
	}
	color.FgGreen.Println("Finished")
	return nil
}

func updateCommandFromContext(c *cli.Context) *client.UpdateCommand {
	return &client.UpdateCommand{
		TargetEndpoint:   c.String(FlagURL.Name),
		Image:            c.String(FlagImage.Name),
		UpdateClassifier: c.String(FlagUpdateClassifier.Name),
		APIKey:           c.String(FlagAPIKey.Name),
	}
}

func addJobsBar(totalJobs int) *uiprogress.Bar {
	bar := uiprogress.AddBar(totalJobs).
		AppendCompleted().
		PrependFunc(func(b *uiprogress.Bar) string { return fmt.Sprintf("%d jobs: ", totalJobs) }).
		PrependElapsed()
	return bar
}

func addDeploymentsBar(totalDeployments int) *uiprogress.Bar {
	bar := uiprogress.AddBar(totalDeployments).
		AppendCompleted().
		PrependFunc(func(b *uiprogress.Bar) string { return fmt.Sprintf("%d deployments: ", totalDeployments) }).
		PrependElapsed()
	return bar
}

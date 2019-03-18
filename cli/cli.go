package cli

import (
	"os"
	"strings"

	"github.com/sirupsen/logrus"

	"gopkg.in/urfave/cli.v2"
)

var (
	// FlagLogLevel specifies the log level of the application.
	FlagLogLevel = &cli.StringFlag{
		Name:        "log-level",
		Usage:       "The log level which shoudl be configured. This can be debug, info, warn, error or critical.",
		DefaultText: "info",
		EnvVars:     []string{"LOG_LEVEL"},
		Value:       "info",
	}
)

// New returns a new CLI runtime for execution.
func New() *cli.App {
	app := &cli.App{
		Name:    "kubernetes-update-manager",
		Version: Version,
		Flags:   []cli.Flag{FlagLogLevel},
		Before: func(context *cli.Context) {
			logLevel := context.String(FlagLogLevel.Name)
			toApplyLogLevel := logrus.Info
			switch strings.ToLower(logLevel) {
			case "debug":
				toApplyLogLevel = logrus.DebugLevel
			case "warn":
				toApplyLogLevel = logrus.WarnLevel
			case "error":
				toApplyLogLevel = logrus.ErrorLevel
			case "critical":
				toApplyLogLevel = logrus.FatalLevel
			default:
				toApplyLogLevel = logrus.InfoLevel
			}
			log.SetFormatter(&log.JSONFormatter{})
			log.SetOutput(os.Stdout)
			log.SetLevel(toApplyLogLevel)
		},
		Commands: []*cli.Command{
			ServerCommand(),
			UpdateCommand(),
		},
	}
	return app
}

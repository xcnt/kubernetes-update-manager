package cli

import (
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	cli "github.com/urfave/cli/v2"
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
		Before: func(context *cli.Context) error {
			logLevel := context.String(FlagLogLevel.Name)
			toApplyLogLevel := log.InfoLevel
			switch strings.ToLower(logLevel) {
			case "debug":
				toApplyLogLevel = log.DebugLevel
			case "warn":
				toApplyLogLevel = log.WarnLevel
			case "error":
				toApplyLogLevel = log.ErrorLevel
			case "critical":
				toApplyLogLevel = log.FatalLevel
			default:
				toApplyLogLevel = log.InfoLevel
			}
			log.SetFormatter(&log.JSONFormatter{})
			log.SetOutput(os.Stdout)
			log.SetLevel(toApplyLogLevel)
			return nil
		},
		Commands: []*cli.Command{
			ServerCommand(),
			UpdateCommand(),
		},
	}
	return app
}

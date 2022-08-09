package cli

import (
	"errors"
	"fmt"
	"kubernetes-update-manager/web"
	"strings"

	"github.com/getsentry/raven-go"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	cli "github.com/urfave/cli/v2"
)

var (
	// FlagHost is used to configure the listening host when running a server
	FlagHost = &cli.StringFlag{
		Name:        "host",
		Value:       "0.0.0.0",
		DefaultText: "Listens to all hosts in default",
		Usage:       "The host where the application should listen to",
		EnvVars:     []string{"UPDATE_MANAGER_LISTENING_HOST"},
	}
	// FlagPort is used to configure the listening port when running a server
	FlagPort = &cli.IntFlag{
		Name:        "port",
		Aliases:     []string{"p"},
		Value:       9000,
		DefaultText: "Listens to port 9000",
		Usage:       "The port the server uses to host it's API on",
		EnvVars:     []string{"UPDATE_MANAGER_LISTENING_PORT"},
	}
	// FlagAutoloadNamespaces specifies if namespaces should automatically be loaded from the cluster
	FlagAutoloadNamespaces = &cli.BoolFlag{
		Name:        "autload-namespaces",
		Aliases:     []string{"an"},
		Value:       true,
		DefaultText: "Autoloads namespaces",
		Usage:       "Specification if all namespaces should be scanned when applying an update. This is per default true and can be turned off if only specific namespaces should be scanned for update information.",
		EnvVars:     []string{"UPDATE_MANAGER_AUTOLOAD_NAMESPACES"},
	}
	// FlagNamespaces is used to specify the namespaces which should be scanned if the autoload functionality has been disabled.
	FlagNamespaces = &cli.StringSliceFlag{
		Name:        "namespaces",
		Aliases:     []string{"n"},
		Value:       cli.NewStringSlice(),
		DefaultText: "empty",
		Usage:       "A list of namespaces which should be scanned for an update. This is only used if autload of namespaces has been switched off.",
		EnvVars:     []string{"UPDATE_MANAGER_NAMESPACES"},
	}
	// FlagAPIKey specifies the pre-shared API key to use the update manager instance.
	FlagAPIKey = &cli.StringFlag{
		Name:    "api-key",
		Usage:   "The pre-shared API key used to authenticate API calls. This is a required field and must be set.",
		EnvVars: []string{"UPDATE_MANAGER_API_KEY"},
	}
	// FlagSentryDSN is used to configure the endpoint where sentry error messages should be sent to if there is an error in the process.
	FlagSentryDSN = &cli.StringFlag{
		Name:    "sentry-dsn",
		Usage:   "The sentry dsn which should be used when reporting errors from the server",
		EnvVars: []string{"SENTRY_DSN"},
	}

	// ErrNoAPIKey is returned if no API Key has been provided for authentication purposes.
	ErrNoAPIKey = errors.New("No API key provided for authenticating the server")
)

// ServerCommand returns the command which shoudl be added to the CLI to run the server.
func ServerCommand() *cli.Command {
	return &cli.Command{
		Name:    "server",
		Aliases: []string{"s"},
		Usage:   "starts the server inside of the cluster",
		Flags:   ServerFlags(),
		Action:  ServerAction,
	}
}

// ServerAction is the action executed if the server command is chosen.
func ServerAction(c *cli.Context) error {
	fmt.Println("Starting server")
	config, err := webConfigFromContext(c)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	sentryDSN := c.String(FlagSentryDSN.Name)
	if len(sentryDSN) > 0 {
		raven.SetDSN(sentryDSN)
	}

	server := web.GetWeb(config)
	host := c.String(FlagHost.Name)
	if host != "0.0.0.0" {
		host = fmt.Sprintf(":%d", c.Int(FlagPort.Name))
	} else {
		host = fmt.Sprintf("%s:%d", host, c.Int(FlagPort.Name))
	}
	err = server.Run(host)
	if err != nil {
		fmt.Println(err.Error())
	}
	return err
}

// webConfigFromContext takes the cli context for the web server, checks for necessary options and returns an initialized web configuration option.
func webConfigFromContext(c *cli.Context) (*web.Config, error) {
	config := web.Config{}
	config.APIKey = strings.TrimSpace(c.String(FlagAPIKey.Name))
	if len(config.APIKey) == 0 {
		return nil, ErrNoAPIKey
	}

	config.AutoloadNamespaces = c.Bool(FlagAutoloadNamespaces.Name)
	config.Namespaces = c.StringSlice(FlagNamespaces.Name)

	kuberneteConfig, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(kuberneteConfig)
	if err != nil {
		return nil, err
	}
	config.Clientset = clientset
	return &config, nil
}

// ServerFlags returns the cli Flags for the server configuration.
func ServerFlags() []cli.Flag {
	return []cli.Flag{
		FlagHost,
		FlagPort,
		FlagAutoloadNamespaces,
		FlagNamespaces,
		FlagAPIKey,
		FlagSentryDSN,
	}
}

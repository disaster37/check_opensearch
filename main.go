package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/disaster37/check_opensearch/v2/check"
	nagiosPlugin "github.com/disaster37/go-nagios"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"github.com/urfave/cli/v2/altsrc"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

var (
	version string
	commit  string
)

func run(args []string) error {

	// Logger setting
	formatter := new(prefixed.TextFormatter)
	formatter.FullTimestamp = true
	formatter.ForceFormatting = true
	log.SetFormatter(formatter)
	log.SetOutput(os.Stdout)

	// CLI settings
	app := cli.NewApp()
	app.Usage = "Check Opensearch"
	app.Version = fmt.Sprintf("%s-%s", version, commit)
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:  "config",
			Usage: "Load configuration from `FILE`",
		},
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:     "url",
			Usage:    "The Opensearch URL",
			EnvVars:  []string{"OPENSEARCH_URL"},
			Required: true,
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:     "user",
			Usage:    "The Opensearch user",
			EnvVars:  []string{"OPENSEARCH_USER"},
			Required: true,
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:     "password",
			Usage:    "The Opensearch password",
			EnvVars:  []string{"OPENSEARCH_PASSWORD"},
			Required: true,
		}),
		&cli.BoolFlag{
			Name:  "self-signed-certificate",
			Usage: "Disable the TLS certificate check",
		},
		&cli.BoolFlag{
			Name:  "debug",
			Usage: "Display debug output",
		},
	}
	app.Commands = []*cli.Command{
		{
			Name:     "check-ism-indice",
			Usage:    "Check the ISM on specific indice. Set indice _all to check all ISM policies",
			Category: "ISM",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "indice",
					Usage:    "The indice name",
					Required: true,
				},
				&cli.StringSliceFlag{
					Name:  "exclude",
					Usage: "The indice name to exclude",
				},
			},
			Action: check.CheckISMError,
		},
		{
			Name:     "check-repository-snapshot",
			Usage:    "Check snapshots state on repository",
			Category: "SM",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "repository",
					Usage:    "The repisitory name",
					Required: true,
				},
			},
			Action: check.CheckSMError,
		},
		{
			Name:     "check-sm-policy",
			Usage:    "Check SM policy",
			Category: "SM",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "name",
					Usage: "The policy name",
				},
			},
			Action: check.CheckSLMPolicy,
		},
		{
			Name:     "check-indice-locked",
			Usage:    "Check if there are indice locked. You can use _all as indice name to check all indices",
			Category: "Indice",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "indice",
					Usage:    "The indice name",
					Required: true,
				},
			},
			Action: check.CheckIndiceLocked,
		},
		{
			Name:     "check-transform",
			Usage:    "Check that Transform have not in error state",
			Category: "Transform",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "name",
					Usage: "The transform id or empty for check all transform",
				},
				&cli.StringSliceFlag{
					Name:  "exclude",
					Usage: "The transform id to exclude",
				},
			},
			Action: check.CheckTransformError,
		},
	}

	app.Before = func(c *cli.Context) error {

		if c.Bool("debug") {
			log.SetLevel(log.DebugLevel)
		}

		if c.String("config") != "" {
			before := altsrc.InitInputSourceWithContext(app.Flags, altsrc.NewYamlSourceFromFlagFunc("config"))
			return before(c)
		}
		return nil
	}

	sort.Sort(cli.CommandsByName(app.Commands))

	err := app.Run(args)
	return err
}

func main() {
	err := run(os.Args)
	if err != nil {
		monitoringData := nagiosPlugin.NewMonitoring()
		monitoringData.SetStatus(nagiosPlugin.STATUS_UNKNOWN)
		monitoringData.AddMessage("Error appear during check: %s", err)
		monitoringData.ToSdtOut()
	}
}

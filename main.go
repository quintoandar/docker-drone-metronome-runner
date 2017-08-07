package main

import (
	"os"
	"strconv"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/urfave/cli"
)

var version string

func main() {
	app := cli.NewApp()
	app.Name = "Metronome job runner plugin"
	app.Usage = "metronome job runner plugin"
	app.Action = run
	app.Version = version
	app.Flags = []cli.Flag{

		cli.StringFlag{
			Name:   "url",
			Usage:  "dcos url",
			EnvVar: "PLUGIN_URL,DCOS_URL",
		},
		cli.StringFlag{
			Name:   "token",
			Usage:  "dcos access token",
			EnvVar: "PLUGIN_TOKEN,DCOS_ACS_TOKEN",
		},
		cli.StringFlag{
			Name:   "job",
			Usage:  "job name for metronome",
			EnvVar: "PLUGIN_JOB",
		},
		cli.StringFlag{
			Name:   "timeout",
			Usage:  "job timeout in minutes",
			Value:  "30",
			EnvVar: "PLUGIN_TIMEOUT",
		},
		cli.BoolFlag{
			Name:   "debug",
			Usage:  "set to true for debug log",
			EnvVar: "PLUGIN_DEBUG",
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func run(c *cli.Context) error {
	if c.Bool("debug") {
		log.SetLevel(log.DebugLevel)
	}

	timeout, err := strconv.Atoi(c.String("timeout"))

	if err != nil {
		log.WithFields(log.Fields{
			"timeout": c.String("timeout"),
			"error":   err,
		}).Error("invalid timeout configuration")
		return err
	}

	plugin := Plugin{
		URL:     c.String("url"),
		Token:   c.String("token"),
		Job:     c.String("job"),
		Timeout: time.Duration(timeout) * time.Minute,
	}

	return plugin.Exec()
}

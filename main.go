package main

import (
	"errors"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

var (
	logger = logrus.New()

	// 2020-07-30
	// https://docs.aws.amazon.com/general/latest/gr/rande.html#ec2_region
	allEC2Regions = []string{
		"us-east-2",
		"us-east-1",
		"us-west-1",
		"us-west-2",
		// "af-south-1", Africa (Cape Town)
		// "ap-east-1", Asia Pacific (Hong Kong)
		"ap-south-1",
		// "ap-northeast-3", Asia Pacific (Osaka-Local)
		"ap-northeast-2",
		"ap-southeast-1",
		"ap-southeast-2",
		"ap-northeast-1",
		"ca-central-1",
		// "cn-north-1", China (Beijing)
		// "cn-northwest-1", China (Ningxia)
		"eu-central-1",
		"eu-west-1",
		"eu-west-2",
		// "eu-south-1", Europe (Milan)
		"eu-west-3",
		"eu-north-1",
		"sa-east-1",
		// "us-gov-east-1",
		// "us-gov-west-1",
	}
)

type globalArguments struct {
	logLevel string
	regions  string
}

func (x *globalArguments) Regions() []string {
	if x.regions == "all" {
		return allEC2Regions
	}
	return strings.Split(x.regions, ",")
}

func (x *globalArguments) Configure() error {
	switch x.logLevel {
	case "trace":
		logger.SetLevel(logrus.TraceLevel)
	case "debug":
		logger.SetLevel(logrus.DebugLevel)
	case "info":
		logger.SetLevel(logrus.InfoLevel)
	case "warn":
		logger.SetLevel(logrus.WarnLevel)
	case "error":
		logger.SetLevel(logrus.ErrorLevel)
	case "":
		break // ignore
	default:
		logger.WithField("loglevel", x.logLevel).Error("invalid log level")
		return errors.New("LogLevel must be in [trace|debug|info|warn|error]")
	}

	return nil
}

func main() {
	args := globalArguments{}

	app := &cli.App{
		Name:    "flowlogconf",
		Version: "v0.2.0",
		Usage:   "AWS VPC Flow Logs batch config tool!",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "loglevel",
				Aliases:     []string{"l"},
				Usage:       "Set log level [trace|debug|info|warn|error]",
				Value:       "info",
				Destination: &args.logLevel,
			},
			&cli.StringFlag{
				Name:        "regions",
				Aliases:     []string{"r"},
				Usage:       "Target regions (comma separated)",
				Value:       "all",
				Destination: &args.regions,
			},
		},
		Commands: []*cli.Command{
			showCommand(&args),
			addCommand(&args),
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		logger.WithError(err).Error("Exit with error.")
		os.Exit(1)
	}
}

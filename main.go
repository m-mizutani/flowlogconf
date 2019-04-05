package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var (
	logger = logrus.New()

	// 2019-04-05
	// https://docs.aws.amazon.com/ja_jp/general/latest/gr/rande.html#ec2_region
	allEC2Regions = []string{
		"us-east-2",
		"us-east-1",
		"us-west-1",
		"us-west-2",
		"ap-south-1",
		// "ap-northeast-3",
		"ap-northeast-2",
		"ap-southeast-1",
		"ap-southeast-2",
		"ap-northeast-1",
		"ca-central-1",
		// "cn-north-1",
		// "cn-northwest-1",
		"eu-central-1",
		"eu-west-1",
		"eu-west-2",
		"eu-west-3",
		"eu-north-1",
		"sa-east-1",
		// "us-gov-east-1",
		// "us-gov-west-1",
	}
)

func isSupportedRegion(target string) bool {
	for _, region := range allEC2Regions {
		if region == target {
			return true
		}
	}

	return false
}

type globalConfig struct {
	regions []string
}

func globalSetup(c *cli.Context, g *globalConfig) error {
	switch c.String("loglevel") {
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
		logger.WithField("loglevel", c.String("loglevel")).Error("invalid log level")
		return errors.New("LogLevel must be in [trace|debug|info|warn|error]")
	}

	if regions := c.String("regions"); regions != "" {
		for _, region := range strings.Split(regions, ",") {
			if !isSupportedRegion(region) {
				return fmt.Errorf("%s is not supported region", region)
			}

			g.regions = append(g.regions, region)
		}
	}

	if parent := c.Parent(); parent != nil {
		return globalSetup(parent, g)
	}
	return nil
}

func main() {
	config := globalConfig{}
	var trafficType string
	var dryrun bool

	app := cli.NewApp()
	app.Name = "flowlogconf"
	app.Version = "v0.1.0"
	app.Usage = "AWS VPC Flow Logs batch config tool"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "loglevel, l",
			Usage: "Set log level [trace|debug|info|warn|error]",
			Value: "info",
		},
		cli.StringFlag{
			Name:  "regions, r",
			Usage: "Target regions (comma separated)",
			Value: strings.Join(allEC2Regions, ","),
		},
	}

	app.Commands = []cli.Command{
		cli.Command{
			Name: "show",
			Action: func(c *cli.Context) error {
				logger.Info("Show config")

				if err := globalSetup(c, &config); err != nil {
					return err
				}
				if err := showConfigs(config.regions); err != nil {
					return err
				}

				return nil
			},
		},
		cli.Command{
			Name: "add",
			Action: func(c *cli.Context) error {
				logger.Info("Add config")

				if err := globalSetup(c, &config); err != nil {
					return err
				}

				if c.NArg() != 1 {
					return errors.New("Invalid arguments: add [s3bucket] [traffic type]")
				}

				s3bucket := c.Args().Get(0)

				if err := addS3Configs(s3bucket, trafficType, config.regions, dryrun); err != nil {
					return err
				}

				return nil
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:        "traffic-type, t",
					Usage:       "Set traffic type [ACCEPT | REJECT | ALL]",
					Value:       "ALL",
					Destination: &trafficType,
				},
				cli.BoolFlag{
					Name:        "dryrun, d",
					Usage:       "Enable dryrun mode",
					Destination: &dryrun,
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		logger.WithError(err).Fatal("Exit with error.")
	}
}

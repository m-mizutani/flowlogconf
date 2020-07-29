package main

import (
	"github.com/urfave/cli/v2"
)

type addArguments struct {
	trafficType string
	dryrun      bool
	s3Bucket    string
}

func addCommand(args *globalArguments) *cli.Command {
	var addArgs addArguments
	return &cli.Command{
		Name: "add",
		Action: func(c *cli.Context) error {
			if err := args.Configure(); err != nil {
				return err
			}

			if err := addS3Configs(addArgs.s3Bucket, addArgs.trafficType, args.Regions(), addArgs.dryrun); err != nil {
				return err
			}

			return nil
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "s3bucket",
				Aliases:     []string{"b"},
				Destination: &addArgs.s3Bucket,
				Required:    true,
			},
			&cli.StringFlag{
				Name:        "traffic-type",
				Aliases:     []string{"t"},
				Usage:       "Set traffic type [ACCEPT | REJECT | ALL]",
				Value:       "ALL",
				Destination: &addArgs.trafficType,
			},
			&cli.BoolFlag{
				Name:        "dryrun",
				Aliases:     []string{"d"},
				Usage:       "Enable dryrun mode",
				Destination: &addArgs.dryrun,
			},
		},
	}
}

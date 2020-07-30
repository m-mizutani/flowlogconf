package main

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

type addArguments struct {
	trafficType     string
	dryrun          bool
	s3Bucket        string
	logFormat       string
	customLogFormat string
}

func addCommand(args *globalArguments) *cli.Command {
	var addArgs addArguments
	return &cli.Command{
		Name: "add",
		Action: func(c *cli.Context) error {
			if err := args.Configure(); err != nil {
				return err
			}

			if addArgs.dryrun {
				logger.Warn("Dryrun mode")
			}

			if err := addS3Configs(args.Regions(), addArgs); err != nil {
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
			&cli.StringFlag{
				Name:        "log-format",
				Aliases:     []string{"f"},
				Usage:       "Log format [default | full]",
				Value:       "default",
				Destination: &addArgs.logFormat,
			},
			&cli.StringFlag{
				Name:        "custom-log-format",
				Usage:       "Costome Log format",
				Destination: &addArgs.customLogFormat,
			},
		},
	}
}

var flowlogFieldNames = []string{
	// v2
	"version",
	"account-id",
	"interface-id",
	"srcaddr",
	"dstaddr",
	"srcport",
	"dstport",
	"protocol",
	"packets",
	"bytes",
	"start",
	"end",
	"action",
	"log-status",
	// v3
	"vpc-id",
	"subnet-id",
	"instance-id",
	"tcp-flags",
	"type",
	"pkt-srcaddr",
	"pkt-dstaddr",
	// v4
	"region",
	"az-id",
	"sublocation-type",
	"sublocation-id",
}

func logFormatFull() string {
	var fields []string

	for _, fname := range flowlogFieldNames {
		fields = append(fields, fmt.Sprintf("${%s}", fname))
	}

	return strings.Join(fields, " ")
}

func addS3Configs(regions []string, args addArguments) error {
	resourceType := "VPC"

	vpcs, err := getVpcList(regions)
	if err != nil {
		return err
	}

	configs, err := getFlowLogConfigs(regions)
	if err != nil {
		return err
	}

	configMap := toConfigMap(configs)
	regionMap := toRegionMap(vpcs)

	for region, vpcList := range regionMap {
		var vpcIDs []string
		for _, vpc := range vpcList {
			if !hasS3Config(args.s3Bucket, args.trafficType, configMap[vpc.vpcID]) {
				vpcIDs = append(vpcIDs, vpc.vpcID)
			}
		}

		if len(vpcIDs) > 0 {
			logger.Infof("Add: %s (%s) @%s => %s (%s)",
				vpcIDs, resourceType, region, args.s3Bucket, args.trafficType)

			if err := addFlowLogS3(args, resourceType, region, vpcIDs); err != nil {
				logger.WithError(err).Errorf("Fail to add S3 flowlog: %s %s %s %s", args.s3Bucket, args.trafficType, region, vpcIDs)
			}
		}
	}

	return nil
}

func addFlowLogS3(args addArguments, resource, region string, vpcIDs []string) error {

	dstBucket := fmt.Sprintf("arn:aws:s3:::%s", args.s3Bucket)

	ssn := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(region),
	}))
	svc := ec2.New(ssn)

	for _, vpcID := range vpcIDs {
		resourceIds := []*string{aws.String(vpcID)}

		input := &ec2.CreateFlowLogsInput{
			LogDestination:     aws.String(dstBucket),
			LogDestinationType: aws.String("s3"),
			ResourceIds:        resourceIds,
			ResourceType:       aws.String(resource),
			TrafficType:        aws.String(args.trafficType),
		}

		if args.logFormat == "full" {
			format := logFormatFull()
			logger.WithField("format", format).Trace("Set log format")
			input.LogFormat = aws.String(format)
		}

		if !args.dryrun {
			result, err := svc.CreateFlowLogs(input)
			if err != nil {
				logger.WithError(err).WithFields(logrus.Fields{
					"region":    region,
					"dstBucket": dstBucket,
					"vpcID":     vpcID,
				}).Error("Fail to create FlowLog")
			} else {
				logger.WithField("result", result).Debug("Create FlowLog")
			}
		}
	}

	return nil
}

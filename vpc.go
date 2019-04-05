package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type vpcInfo struct {
	vpcID     string
	cidrBlock string
	region    string
}

func getVpcList(regions []string) ([]vpcInfo, error) {
	var vpcList []vpcInfo

	for _, region := range regions {
		ssn := session.Must(session.NewSession(&aws.Config{
			Region: aws.String(region),
		}))

		svc := ec2.New(ssn)

		var nextToken string
		for {
			input := &ec2.DescribeVpcsInput{}
			if nextToken != "" {
				input.NextToken = &nextToken
			}

			result, err := svc.DescribeVpcs(input)
			if err != nil {
				return nil, errors.Wrapf(err, "Fail to fetch VPC descriptions (%s)", region)
			}
			logger.WithFields(logrus.Fields{
				"region":    region,
				"vpc count": len(result.Vpcs),
			}).Debug("Got VPC descriptions")

			for _, vpc := range result.Vpcs {
				vpcList = append(vpcList, vpcInfo{
					vpcID:     *vpc.VpcId,
					cidrBlock: *vpc.CidrBlock,
					region:    region,
				})
			}

			if result.NextToken == nil {
				break
			}
			nextToken = *result.NextToken
		}
	}

	return vpcList, nil
}

type flowLogConfig struct {
	flowlog *ec2.FlowLog
	region  string
}

func getFlowLogConfigs(regions []string) ([]flowLogConfig, error) {
	var configs []flowLogConfig

	for _, region := range regions {
		ssn := session.Must(session.NewSession(&aws.Config{
			Region: aws.String(region),
		}))

		var nextToken string
		for {
			svc := ec2.New(ssn)
			input := &ec2.DescribeFlowLogsInput{}
			if nextToken != "" {
				input.NextToken = &nextToken
			}

			result, err := svc.DescribeFlowLogs(input)
			if err != nil {
				return nil, errors.Wrapf(err, "Fail to fetch FlowLog descriptions (%s)", region)
			}

			logger.WithFields(logrus.Fields{
				"region":        region,
				"flowlog count": len(result.FlowLogs),
			}).Debug("Got FlowLogs")

			for _, flowlog := range result.FlowLogs {
				configs = append(configs, flowLogConfig{flowlog, region})
			}

			if result.NextToken == nil {
				break
			}
			nextToken = *result.NextToken
		}
	}

	return configs, nil
}

func addFlowLogS3(dstBucket, traffic, resource, region string, vpcIDs []string) error {
	var resourceIds []*string
	for _, vpcID := range vpcIDs {
		resourceIds = append(resourceIds, &vpcID)
	}

	ssn := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(region),
	}))
	svc := ec2.New(ssn)
	input := &ec2.CreateFlowLogsInput{
		LogDestination:     aws.String(dstBucket),
		LogDestinationType: aws.String("s3"),
		ResourceIds:        resourceIds,
		ResourceType:       aws.String(resource),
		TrafficType:        aws.String(traffic),
	}

	result, err := svc.CreateFlowLogs(input)
	if err != nil {
		return errors.Wrapf(err, "Fail to create FlowLog (%s, %s, %s)", region, dstBucket, vpcIDs)
	}

	logger.WithField("result", result).Debug("Create FlowLog")

	return nil
}

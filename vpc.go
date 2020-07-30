package main

import (
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/sirupsen/logrus"
)

type vpcInfo struct {
	vpcID     string
	cidrBlock string
	region    string
}

func getVpcList(regions []string) ([]*vpcInfo, error) {
	var vpcList []*vpcInfo
	ch := make(chan *vpcInfo, 4096)
	done := make(chan bool)

	go func() {
		var wg sync.WaitGroup
		for _, r := range regions {
			wg.Add(1)

			go func(region string) {
				defer wg.Done()
				logger.WithField("region", region).Debug("Retrieving VPC info")

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
						logger.WithField("region", region).WithError(err).Error("Fail to fetch VPC descriptions")
						return
					}

					logger.WithFields(logrus.Fields{
						"region":    region,
						"vpc count": len(result.Vpcs),
					}).Debug("Got VPC descriptions")

					for _, vpc := range result.Vpcs {
						ch <- &vpcInfo{
							vpcID:     *vpc.VpcId,
							cidrBlock: *vpc.CidrBlock,
							region:    region,
						}
					}

					if result.NextToken == nil {
						break
					}
					nextToken = *result.NextToken
				}
			}(r)
		}
		wg.Wait()
		close(ch)
		done <- true
	}()

	<-done
	for vpc := range ch {
		vpcList = append(vpcList, vpc)
	}

	return vpcList, nil
}

type flowLogConfig struct {
	flowlog *ec2.FlowLog
	region  string
}

func getFlowLogConfigs(regions []string) ([]*flowLogConfig, error) {
	var configs []*flowLogConfig
	ch := make(chan *flowLogConfig, 4096)
	done := make(chan bool)

	go func() {
		var wg sync.WaitGroup
		for _, r := range regions {
			wg.Add(1)

			go func(region string) {
				defer wg.Done()
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
						logger.WithField("region", region).WithError(err).Error("Fail to fetch FlowLog descriptions")
						return
					}

					logger.WithFields(logrus.Fields{
						"region":        region,
						"flowlog count": len(result.FlowLogs),
					}).Debug("Got FlowLogs")

					for _, flowlog := range result.FlowLogs {
						ch <- &flowLogConfig{flowlog, region}
					}

					if result.NextToken == nil {
						break
					}
					nextToken = *result.NextToken
				}
			}(r)
		}

		wg.Wait()
		close(ch)
		done <- true
	}()

	<-done
	for flowLog := range ch {
		configs = append(configs, flowLog)
	}

	return configs, nil
}

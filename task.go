package main

import (
	"fmt"
)

func toConfigMap(configs []flowLogConfig) map[string][]flowLogConfig {
	configMap := map[string][]flowLogConfig{}
	for _, config := range configs {
		if config.flowlog.ResourceId == nil {
			logger.WithField("flowlog", config).Warn("ResourceID is not set")
		}
		vpcID := *config.flowlog.ResourceId

		if _, ok := configMap[vpcID]; !ok {
			configMap[vpcID] = []flowLogConfig{}
		}

		configMap[vpcID] = append(configMap[vpcID], config)
	}

	return configMap
}

func toRegionMap(vpcs []vpcInfo) map[string][]vpcInfo {
	regionMap := map[string][]vpcInfo{}
	for _, vpc := range vpcs {
		if _, ok := regionMap[vpc.region]; !ok {
			regionMap[vpc.region] = []vpcInfo{}
		}

		regionMap[vpc.region] = append(regionMap[vpc.region], vpc)
	}

	return regionMap
}

func showConfigs(regions []string) error {
	vpcs, err := getVpcList(regions)
	if err != nil {
		return err
	}

	configs, err := getFlowLogConfigs(regions)
	if err != nil {
		return err
	}

	configMap := toConfigMap(configs)

	for _, vpc := range vpcs {
		fmt.Printf("%25s %20s %15s ", vpc.vpcID, vpc.region, vpc.cidrBlock)

		if len(configMap[vpc.vpcID]) > 0 {
			for _, config := range configMap[vpc.vpcID] {
				fmt.Printf("%10s %s", *config.flowlog.LogDestinationType, *config.flowlog.LogDestination)
			}
		} else {
			fmt.Printf("N/A")
		}

		fmt.Printf("\n")
	}

	return nil
}

func hasS3Config(dstBucket, traffic string, configs []flowLogConfig) bool {
	for _, config := range configs {
		if config.flowlog.LogDestinationType == nil ||
			config.flowlog.LogDestination == nil ||
			config.flowlog.TrafficType == nil {
			continue
		}

		if *config.flowlog.LogDestinationType == "s3" &&
			*config.flowlog.LogDestination == dstBucket &&
			*config.flowlog.TrafficType == traffic {
			return true
		}
	}

	return false
}

func addS3Configs(dstBucket, traffic string, regions []string, dryrun bool) error {
	dstArn := fmt.Sprintf("arn:aws:s3:::%s", dstBucket)
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
			if !hasS3Config(dstArn, traffic, configMap[vpc.vpcID]) {
				vpcIDs = append(vpcIDs, vpc.vpcID)
			}
		}

		if len(vpcIDs) > 0 {
			logger.Infof("Add: %s (%s) @%s => %s (%s)",
				vpcIDs, resourceType, region, dstArn, traffic)

			if !dryrun {
				if err := addFlowLogS3(dstArn, traffic, resourceType, region, vpcIDs); err != nil {
					// return errors.Wrapf(err, "Fail to add S3 flowlog: %s %s %s %s", dstArn, traffic, region, vpcIDs)
					logger.WithError(err).Errorf("Fail to add S3 flowlog: %s %s %s %s", dstArn, traffic, region, vpcIDs)
				}
			}
		}
	}

	return nil
}

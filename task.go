package main

import "fmt"

func toConfigMap(configs []*flowLogConfig) map[string][]*flowLogConfig {
	configMap := map[string][]*flowLogConfig{}
	for _, config := range configs {
		if config.flowlog.ResourceId == nil {
			logger.WithField("flowlog", config).Warn("ResourceID is not set")
		}
		vpcID := *config.flowlog.ResourceId

		if _, ok := configMap[vpcID]; !ok {
			configMap[vpcID] = []*flowLogConfig{}
		}

		configMap[vpcID] = append(configMap[vpcID], config)
	}

	return configMap
}

func toRegionMap(vpcs []*vpcInfo) map[string][]*vpcInfo {
	regionMap := map[string][]*vpcInfo{}
	for _, vpc := range vpcs {
		if _, ok := regionMap[vpc.region]; !ok {
			regionMap[vpc.region] = []*vpcInfo{}
		}

		regionMap[vpc.region] = append(regionMap[vpc.region], vpc)
	}

	return regionMap
}

func hasS3Config(s3Bucket, traffic string, configs []*flowLogConfig) bool {
	dstBucket := fmt.Sprintf("arn:aws:s3:::%s", s3Bucket)

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

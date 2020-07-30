package main

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

func showCommand(args *globalArguments) *cli.Command {
	return &cli.Command{
		Name: "show",
		Action: func(c *cli.Context) error {
			if err := args.Configure(); err != nil {
				return err
			}
			if err := showConfigs(args.Regions()); err != nil {
				return err
			}

			return nil
		},
	}
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

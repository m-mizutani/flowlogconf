package main

import "github.com/urfave/cli/v2"

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

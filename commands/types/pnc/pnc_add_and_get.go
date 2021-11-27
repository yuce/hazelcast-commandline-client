package commands

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/hazelcast/hazelcast-commandline-client/internal"
)

var pncValue int64

var pncAddAndGetCmd = &cobra.Command{
	Use:   "add-and-get [--name pncname] [--value num]",
	Short: "set PN Counter",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.TODO()
		config, err := internal.MakeConfig(cmd)
		if err != nil { //TODO error look like unhandled although it is handled in MakeConfig.Find a better approach
			return
		}
		p, err := getPNCounter(config, pncName)
		if err != nil {
			return
		}
		value, err := p.AddAndGet(ctx, pncValue)
		if err != nil {
			fmt.Printf("Error: Cannot add and get value to PN Counter %s\n", pncName)
			return
		}
		fmt.Println(value)
	},
}

func init() {
	pncAddAndGetCmd.Flags().StringVarP(&pncName, "name", "m", "", "PN Counter name")
	pncAddAndGetCmd.Flags().Int64VarP(&pncValue, "value", "v", 0, "value to add")
	pncAddAndGetCmd.MarkFlagRequired("name")
	pncAddAndGetCmd.MarkFlagRequired("value")
}

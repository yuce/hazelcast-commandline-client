package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/logger"
	"github.com/hazelcast/hazelcast-go-client/types"
	"github.com/spf13/cobra"

	"github.com/hazelcast/hazelcast-commandline-client/internal"
)

var pncName string

var PNCCmd = &cobra.Command{
	Use: "pnc {get} --name pncname",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	PNCCmd.AddCommand(pncGetCmd)
	PNCCmd.AddCommand(pncAddAndGetCmd)
}

func getPNCounter(clientConfig *hazelcast.Config, name string) (result *hazelcast.PNCounter, err error) {
	defer func() {
		obj := recover()
		if panicErr, ok := obj.(error); ok {
			err = panicErr
			if msg, handled := internal.TranslateError(err, clientConfig.Cluster.Cloud.Enabled); handled {
				fmt.Println("Error:", msg)
				return
			}
			fmt.Println("Error:", err)
		}
	}()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if clientConfig == nil {
		clientConfig = &hazelcast.Config{}
	}
	clientConfig.Cluster.ConnectionStrategy.Retry.InitialBackoff = types.Duration(1 * time.Second)
	configCopy := clientConfig.Clone()
	configCopy.Logger.Level = logger.OffLevel // internal event loop prints error logs
	hzcClient, err := hazelcast.StartNewClientWithConfig(ctx, configCopy)
	if err != nil {
		if msg, handled := internal.TranslateError(err, clientConfig.Cluster.Cloud.Enabled); handled {
			fmt.Println("Error:", msg)
			return
		}
		fmt.Println("Error:", err)
		return nil, fmt.Errorf("error creating the client: %w", err)
	}
	if result, err = hzcClient.GetPNCounter(ctx, name); err != nil {
		fmt.Println("Error:", err)
	}
	return
}

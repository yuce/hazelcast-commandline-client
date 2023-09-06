//go:build std || viridian

package viridian

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/viridian"
)

type ClusterStopCmd struct{}

func (cm ClusterStopCmd) Init(cc plug.InitContext) error {
	var long, short string
	if viridian.InternalOpsEnabled() {
		cc.SetCommandUsage("stop-cluster [flags]")
		long = `Stops the cluster.

Make sure you login before running this command.
`
		short = "Stops the cluster"
		cc.SetPositionalArgCount(0, 0)
		cc.SetCommandGroup("viridian")
	} else {
		cc.SetCommandUsage("stop-cluster [cluster-ID/name] [flags]")
		long = `Stops the given Viridian cluster.

Make sure you login before running this command.
`
		short = "Stops the given Viridian cluster"
		cc.SetPositionalArgCount(1, 1)
	}
	cc.SetCommandHelp(long, short)
	cc.AddStringFlag(propAPIKey, "", "", false, "Viridian API Key")
	return nil
}

func (cm ClusterStopCmd) Exec(ctx context.Context, ec plug.ExecContext) error {
	api, err := getAPI(ec)
	if err != nil {
		return err
	}
	var clusterNameOrID string
	if viridian.InternalOpsEnabled() {
		vc, err := loadVRDConfig()
		if err != nil {
			return fmt.Errorf("loading vrd config: %w", err)
		}
		clusterNameOrID = vc.ClusterID
	} else {
		clusterNameOrID = ec.Args()[0]
	}
	_, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText("Pausing the cluster")
		err := api.StopCluster(ctx, clusterNameOrID)
		if err != nil {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		return handleErrorResponse(ec, err)
	}
	stop()
	ec.PrintlnUnnecessary(fmt.Sprintf("Cluster %s was stopped.", clusterNameOrID))
	return nil
}

func init() {
	if viridian.InternalOpsEnabled() {
		check.Must(plug.Registry.RegisterCommand("stop-cluster", &ClusterStopCmd{}))
	} else {
		check.Must(plug.Registry.RegisterCommand("viridian:stop-cluster", &ClusterStopCmd{}))
	}
}

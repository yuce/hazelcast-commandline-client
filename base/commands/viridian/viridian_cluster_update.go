//go:build std || viridian

package viridian

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type ClusterUpdateCmd struct{}

func (ClusterUpdateCmd) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("update-cluster")
	long := `Updates the given Viridian cluster.

Make sure you login before running this command.
`
	short := "Updates the given Viridian cluster"
	cc.SetCommandHelp(long, short)
	cc.SetPositionalArgCount(0, 0)
	cc.AddStringFlag(propAPIKey, "", "", false, "Viridian API Key")
	if enableInternalOps {
		cc.SetCommandGroup("viridian")
	}
	return nil
}

func (ClusterUpdateCmd) Exec(ctx context.Context, ec plug.ExecContext) error {
	vc, err := loadVRDConfig()
	if err != nil {
		return err
	}
	_, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		api, err := getAPI(ec)
		sp.SetText(fmt.Sprintf("Stopping cluster: %s", vc.ClusterID))
		if err := api.StopCluster(ctx, vc.ClusterID); err != nil {
			return nil, err
		}
		if err := waitClusterState(ctx, ec, api, vc.ClusterID, stateStopped); err != nil {
			return nil, err
		}
		sp.SetText(fmt.Sprintf("Resuming cluster: %s", vc.ClusterID))
		if err := api.ResumeCluster(ctx, vc.ClusterID); err != nil {
			return nil, err
		}
		if err := waitClusterState(ctx, ec, api, vc.ClusterID, stateRunning); err != nil {
			return nil, err
		}
		return nil, err
	})
	if err != nil {
		return err
	}
	stop()
	ec.PrintlnUnnecessary(fmt.Sprintf("Cluster %s was updated.", vc.ClusterID))
	return nil
}

func init() {
	if enableInternalOps {
		check.Must(plug.Registry.RegisterCommand("update-cluster", &ClusterUpdateCmd{}))
	}
}

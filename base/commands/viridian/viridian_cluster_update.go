//go:build std || viridian

package viridian

import (
	"context"

	"github.com/hazelcast/hazelcast-commandline-client/clc/ux/stage"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
	"github.com/hazelcast/hazelcast-commandline-client/internal/viridian"
)

type ClusterUpdateCommand struct{}

func (ClusterUpdateCommand) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("update-cluster")
	long := `Updates the given Viridian cluster.

Make sure you login before running this command.
`
	short := "Updates the given Viridian cluster"
	cc.SetCommandHelp(long, short)
	cc.AddStringFlag(propAPIKey, "", "", false, "Viridian API Key")
	if viridian.InternalOpsEnabled() {
		cc.SetCommandGroup("viridian")
	}
	return nil
}

func (ClusterUpdateCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	vc, err := loadVRDConfig()
	if err != nil {
		return err
	}
	api, err := getAPI(ec)
	if err != nil {
		return err
	}
	ec.PrintlnUnnecessary("")
	sp := stage.NewFixedProvider[string](
		stage.Stage[string]{
			ProgressMsg: "Stopping the cluster",
			SuccessMsg:  "Stopped the cluster",
			FailureMsg:  "Failed stopping the cluster",
			Func: func(ctx context.Context, status stage.Statuser[string]) (string, error) {
				clusterID := status.Value()
				if _, err := api.StopCluster(ctx, clusterID); err != nil {
					return clusterID, err
				}
				if err := waitClusterState(ctx, ec, api, clusterID, stateStopped); err != nil {
					return clusterID, err
				}
				return clusterID, nil
			},
		},
		stage.Stage[string]{
			ProgressMsg: "Resuming the cluster",
			SuccessMsg:  "Resumed the cluster",
			FailureMsg:  "Failed resuming the cluster",
			Func: func(ctx context.Context, status stage.Statuser[string]) (string, error) {
				clusterID := status.Value()
				if _, err := api.ResumeCluster(ctx, clusterID); err != nil {
					return clusterID, err
				}
				if err := waitClusterState(ctx, ec, api, clusterID, stateRunning); err != nil {
					return clusterID, err
				}
				return clusterID, nil
			},
		},
	)
	if _, err := stage.Execute(ctx, ec, vc.ClusterID, sp); err != nil {
		return err
	}
	ec.PrintlnUnnecessary("OK Cluster update completed successfully.\n")
	return ec.AddOutputRows(ctx, output.Row{
		output.Column{
			Name:  "Cluster ID",
			Type:  serialization.TypeString,
			Value: vc.ClusterID,
		},
	})
}

func (ClusterUpdateCommand) Unwrappable() {}

func init() {
	if viridian.InternalOpsEnabled() {
		check.Must(plug.Registry.RegisterCommand("update-cluster", &ClusterUpdateCommand{}))
	}
}

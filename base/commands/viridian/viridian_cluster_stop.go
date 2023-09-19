//go:build std || viridian

package viridian

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-commandline-client/clc/ux/stage"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
	"github.com/hazelcast/hazelcast-commandline-client/internal/viridian"
)

type ClusterStopCommand struct{}

func (ClusterStopCommand) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("stop-cluster")
	long := `Stops the cluster.

Make sure you login before running this command.
`
	var short string
	if viridian.InternalOpsEnabled() {
		cc.SetCommandGroup("viridian")
		short = "Stops the cluster"
	} else {
		cc.SetCommandUsage("stop-cluster [cluster-ID/name] [flags]")
		short = "Stops the given Viridian cluster"
		cc.AddStringFlag(propAPIKey, "", "", false, "Viridian API Key")
		cc.AddStringArg(argClusterID, argTitleClusterID)
	}
	cc.SetCommandHelp(long, short)
	return nil
}

func (ClusterStopCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	api, err := getAPI(ec)
	if err != nil {
		return err
	}
	var nameOrID string
	if viridian.InternalOpsEnabled() {
		vc, err := loadVRDConfig()
		if err != nil {
			return fmt.Errorf("loading vrd config: %w", err)
		}
		nameOrID = vc.ClusterID
	} else {
		nameOrID = ec.GetStringArg(argClusterID)
	}
	st := stage.Stage[viridian.Cluster]{
		ProgressMsg: "Initiating cluster stop",
		SuccessMsg:  "Initiated cluster stop",
		FailureMsg:  "Failed to initiate cluster stop",
		Func: func(ctx context.Context, status stage.Statuser[viridian.Cluster]) (viridian.Cluster, error) {
			return api.StopCluster(ctx, nameOrID)
		},
	}
	cluster, err := stage.Execute(ctx, ec, viridian.Cluster{}, stage.NewFixedProvider(st))
	if err != nil {
		return handleErrorResponse(ec, err)
	}
	ec.PrintlnUnnecessary("")
	return ec.AddOutputRows(ctx, output.Row{
		output.Column{
			Name:  "ID",
			Type:  serialization.TypeString,
			Value: cluster.ID,
		},
	})
}

func init() {
	if viridian.InternalOpsEnabled() {
		check.Must(plug.Registry.RegisterCommand("stop-cluster", &ClusterStopCommand{}))
	} else {
		check.Must(plug.Registry.RegisterCommand("viridian:stop-cluster", &ClusterStopCommand{}))
	}
}

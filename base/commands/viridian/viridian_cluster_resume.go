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

type ClusterResumeCommand struct{}

func (ClusterResumeCommand) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("resume-cluster")
	var long, short string
	if viridian.InternalOpsEnabled() {
		long = `Resumes the cluster.

Make sure you login before running this command.
`
		short = "Resumes the cluster"
		cc.SetCommandGroup("viridian")
	} else {
		long = `Resumes the given Viridian cluster.

Make sure you login before running this command.
`
		short = "Resumes the given Viridian cluster"
		cc.AddStringArg(argClusterID, argTitleClusterID)
	}
	cc.SetCommandHelp(long, short)
	cc.AddStringFlag(propAPIKey, "", "", false, "Viridian API Key")
	return nil
}

func (ClusterResumeCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
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
		ProgressMsg: "Starting to resume the cluster",
		SuccessMsg:  "Started to resume the cluster",
		FailureMsg:  "Failed to start resuming the cluster",
		Func: func(ctx context.Context, status stage.Statuser[viridian.Cluster]) (viridian.Cluster, error) {
			return api.ResumeCluster(ctx, nameOrID)
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
		check.Must(plug.Registry.RegisterCommand("resume-cluster", &ClusterResumeCommand{}))
	} else {
		check.Must(plug.Registry.RegisterCommand("viridian:resume-cluster", &ClusterResumeCommand{}))
	}
}

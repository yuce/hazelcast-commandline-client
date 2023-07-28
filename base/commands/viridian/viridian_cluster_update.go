//go:build std || viridian

package viridian

import (
	"context"

	"github.com/hazelcast/hazelcast-commandline-client/clc/ux/stage"
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
	api, err := getAPI(ec)
	if err != nil {
		return err
	}
	ec.PrintlnUnnecessary("")
	sp := stage.NewFixedProvider(
		stopStage(ctx, ec, api, vc.ClusterID),
		resumeStage(ctx, ec, api, vc.ClusterID),
	)
	if err := stage.Execute(ctx, ec, sp); err != nil {
		return err
	}
	ec.PrintlnUnnecessary("")
	ec.PrintlnUnnecessary("OK Cluster update completed successfully.")
	return nil
}

func (ClusterUpdateCmd) Unwrappable() {}

func init() {
	if enableInternalOps {
		check.Must(plug.Registry.RegisterCommand("update-cluster", &ClusterUpdateCmd{}))
	}
}

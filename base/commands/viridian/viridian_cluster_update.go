//go:build std || viridian

package viridian

import (
	"context"

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
	cc.SetPositionalArgCount(1, 1)
	cc.AddStringFlag(propAPIKey, "", "", false, "Viridian API Key")
	if enableInternalOps {
		cc.SetCommandGroup("viridian")
	}
	return nil
}

func (ClusterUpdateCmd) Exec(ctx context.Context, ec plug.ExecContext) error {
	return nil
}

func init() {
	if enableInternalOps {
		check.Must(plug.Registry.RegisterCommand("update-cluster", &ClusterUpdateCmd{}))
	}
}

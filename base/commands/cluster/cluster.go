package cluster

import (
	"context"

	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type Cmd struct{}

func (cm Cmd) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("cluster [command]")
	cc.AddCommandGroup("cluster", "Cluster Operations")
	cc.SetCommandGroup("cluster")
	cc.SetTopLevel(true)
	help := "Cluster operations"
	cc.SetCommandHelp(help, help)
	return nil
}

func (cm Cmd) Exec(context.Context, plug.ExecContext) error {
	return nil
}

func init() {
	check.Must(plug.Registry.RegisterCommand("cluster", &Cmd{}))
}

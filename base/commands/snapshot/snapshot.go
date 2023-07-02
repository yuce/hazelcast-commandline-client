package snapshot

import (
	"context"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/prv/check"
	"github.com/hazelcast/hazelcast-commandline-client/prv/plug"
)

type Cmd struct{}

func (cm Cmd) Init(cc plug.InitContext) error {
	cc.SetCommandGroup(clc.GroupJetID)
	cc.SetTopLevel(true)
	help := "Jet snapshot operations"
	cc.SetCommandUsage("snapshot [command]")
	cc.SetCommandHelp(help, help)
	return nil
}

func (cm Cmd) Exec(context.Context, plug.ExecContext) error {
	return nil
}

func init() {
	check.Must(plug.Registry.RegisterCommand("snapshot", &Cmd{}))
}

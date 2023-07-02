package job

import (
	"context"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	. "github.com/hazelcast/hazelcast-commandline-client/prv/check"
	"github.com/hazelcast/hazelcast-commandline-client/prv/plug"
)

type Cmd struct{}

func (cm Cmd) Init(cc plug.InitContext) error {
	cc.SetCommandGroup(clc.GroupJetID)
	cc.SetTopLevel(true)
	help := "Jet job operations"
	cc.SetCommandUsage("job [command]")
	cc.SetCommandHelp(help, help)
	return nil
}

func (cm Cmd) Exec(context.Context, plug.ExecContext) error {
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("job", &Cmd{}))
}

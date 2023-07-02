//go:build base || objects

package object

import (
	"context"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	. "github.com/hazelcast/hazelcast-commandline-client/prv/check"
	"github.com/hazelcast/hazelcast-commandline-client/prv/plug"
)

type ObjectCommand struct{}

func (cm ObjectCommand) Init(cc plug.InitContext) error {
	cc.SetCommandGroup(clc.GroupDDSID)
	cc.SetTopLevel(true)
	help := "Generic distributed data structure operations"
	cc.SetCommandUsage("object [command]")
	cc.SetCommandHelp(help, help)
	return nil
}

func (cm ObjectCommand) Exec(context.Context, plug.ExecContext) error {
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("object", &ObjectCommand{}))
}

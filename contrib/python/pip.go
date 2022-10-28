package python

import (
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type PipCommand struct{}

func (cm PipCommand) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("pip [flags]")
	help := "Install Python packages"
	cc.SetCommandHelp(help, help)
	cc.SetCommandGroup("python")
	return nil
}

func (cm PipCommand) Exec(ec plug.ExecContext) error {
	//TODO implement me
	panic("implement me")
}

func init() {
	Must(plug.Registry.RegisterCommand("pip", &PipCommand{}))
}

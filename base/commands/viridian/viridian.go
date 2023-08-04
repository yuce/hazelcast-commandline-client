//go:build std || viridian

package viridian

import (
	"context"

	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/viridian"
)

type Cmd struct{}

func (cm Cmd) Init(cc plug.InitContext) error {
	if !viridian.InternalOpsEnabled() {
		cc.SetTopLevel(true)
		cc.SetCommandUsage("viridian [command]")
		help := "Various Viridian operations"
		cc.SetCommandHelp(help, help)
	}
	cc.AddCommandGroup("viridian", "Viridian")
	cc.SetCommandGroup("viridian")
	return nil
}

func (cm Cmd) Exec(ctx context.Context, ec plug.ExecContext) error {
	return nil
}

func init() {
	if viridian.InternalOpsEnabled() {
		plug.Registry.RegisterGlobalInitializer("10-viridian-ops", &Cmd{})
	} else {
		check.Must(plug.Registry.RegisterCommand("viridian", &Cmd{}))
	}
}

//go:build std || viridian

package viridian

import (
	"context"

	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/viridian"
)

type Command struct{}

func (Command) Init(cc plug.InitContext) error {
	if !viridian.InternalOpsEnabled() {
		cc.SetCommandUsage("viridian")
		cc.SetTopLevel(true)
		help := "Various Viridian operations"
		cc.SetCommandHelp(help, help)
	}
	cc.AddCommandGroup("viridian", "Viridian")
	cc.SetCommandGroup("viridian")
	return nil
}

func (Command) Exec(ctx context.Context, ec plug.ExecContext) error {
	return nil
}

func init() {
	if viridian.InternalOpsEnabled() {
		plug.Registry.RegisterGlobalInitializer("10-viridian-ops", &Command{})
	} else {
		check.Must(plug.Registry.RegisterCommand("viridian", &Command{}))
	}
}

//go:build std || set

package set

import (
	"github.com/hazelcast/hazelcast-commandline-client/base/commands"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

func init() {
	c := commands.NewDestroyCommand("Set", "set", getSet)
	check.Must(plug.Registry.RegisterCommand("set:destroy", c))
}

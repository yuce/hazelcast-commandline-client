//go:build std || map

package _map

import (
	"github.com/hazelcast/hazelcast-commandline-client/base/commands"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/proto/codec"
)

func init() {
	c := commands.NewMapEntrySetCommand("Map", "map", codec.EncodeMapEntrySetRequest, codec.DecodeMapEntrySetResponse, getMap)
	check.Must(plug.Registry.RegisterCommand("map:entry-set", c))
}

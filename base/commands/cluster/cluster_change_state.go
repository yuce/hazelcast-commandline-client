package cluster

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/proto/codec"
)

type ChangeStateCmd struct{}

func (cm ChangeStateCmd) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("change-state [new-state]")
	long := `Sets cluster state

The new-state may be one of the following:

* ACTIVE
* NO_MIGRATION
* FROZEN
* PASSIVE
* IN_TRANSITION
	
The states are case insensitive.
`
	short := "Sets the cluster state"
	cc.SetCommandHelp(long, short)
	cc.SetPositionalArgCount(1, 1)
	return nil
}

func (cm ChangeStateCmd) Exec(ctx context.Context, ec plug.ExecContext) error {
	ci, err := ec.ClientInternal(ctx)
	if err != nil {
		return err
	}
	state, err := stringToState(ec.Args()[0])
	if err != nil {
		return err
	}
	_, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText("Changing cluster state")
		req := codec.EncodeMCChangeClusterStateRequest(state)
		_, err := ci.InvokeOnRandomTarget(ctx, req, nil)
		if err != nil {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		return fmt.Errorf("changing cluster state: %w", err)
	}
	stop()
	return nil
}

func init() {
	check.Must(plug.Registry.RegisterCommand("cluster:change-state", &ChangeStateCmd{}))
}

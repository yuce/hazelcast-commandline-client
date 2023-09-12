package cluster

import (
	"context"
	"strings"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/cmd"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/proto/codec"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
)

const (
	argNewState      = "newState"
	argTitleNewState = "new state"
)

type ChangeStateCmd struct{}

func (cm ChangeStateCmd) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("change-state")
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
	cc.AddStringArg(argNewState, argTitleNewState)
	return nil
}

func (cm ChangeStateCmd) Exec(ctx context.Context, ec plug.ExecContext) error {
	stateStr := ec.GetStringArg(argNewState)
	state, err := stringToState(stateStr)
	if err != nil {
		return err
	}
	row, stop, err := cmd.ExecuteBlocking(ctx, ec, func(ctx context.Context, sp clc.Spinner) (output.Row, error) {
		ci, err := cmd.ClientInternal(ctx, ec, sp)
		if err != nil {
			return nil, err
		}
		sp.SetText("Changing cluster state")
		req := codec.EncodeMCChangeClusterStateRequest(state)
		_, err = ci.InvokeOnRandomTarget(ctx, req, nil)
		if err != nil {
			return nil, err
		}
		row := output.Row{
			output.Column{
				Name:  "New State",
				Type:  serialization.TypeString,
				Value: strings.ToUpper(stateStr),
			},
		}
		return row, nil
	})
	if err != nil {
		return err
	}
	stop()
	ec.PrintlnUnnecessary("OK Changed cluster state.\n")
	return ec.AddOutputRows(ctx, row)
}

func init() {
	check.Must(plug.Registry.RegisterCommand("cluster:change-state", &ChangeStateCmd{}))
}

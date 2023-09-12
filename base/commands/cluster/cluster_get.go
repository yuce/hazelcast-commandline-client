package cluster

import (
	"context"
	"fmt"
	"time"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/cmd"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/proto/codec"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
)

type GetCmd struct{}

func (cm GetCmd) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("get")
	help := "Gets cluster information"
	cc.SetCommandHelp(help, help)
	return nil
}

func (cm GetCmd) Exec(ctx context.Context, ec plug.ExecContext) error {
	row, stop, err := cmd.ExecuteBlocking(ctx, ec, func(ctx context.Context, sp clc.Spinner) (output.Row, error) {
		ci, err := cmd.ClientInternal(ctx, ec, sp)
		if err != nil {
			return nil, err
		}
		sp.SetText("Getting cluster information")
		req := codec.EncodeMCGetClusterMetadataRequest()
		resp, err := ci.InvokeOnRandomTarget(ctx, req, nil)
		if err != nil {
			return nil, err
		}
		currentState, memberVersion, _, clusterTime, clusterID, clusterIdExists := codec.DecodeMCGetClusterMetadataResponse(resp)
		cid := "N/A"
		if clusterIdExists {
			cid = clusterID.String()
		}
		row := output.Row{
			output.Column{
				Name:  "State",
				Type:  serialization.TypeString,
				Value: stateToString(currentState),
			},
			output.Column{
				Name:  "Member Version",
				Type:  serialization.TypeString,
				Value: memberVersion,
			},
			/*
				// disabling for now
				output.Column{
					Name:  "Jet Version",
					Type:  serialization.TypeString,
					Value: jetVersion,
				},`
			*/
			output.Column{
				Name:  "Cluster Time",
				Type:  serialization.TypeString,
				Value: time.UnixMilli(clusterTime).Format(time.RFC3339),
			},
			output.Column{
				Name:  "Cluster ID",
				Type:  serialization.TypeString,
				Value: cid,
			},
		}
		return row, nil
	})
	if err != nil {
		return fmt.Errorf("retrieving cluster information: %w", err)
	}
	stop()
	return ec.AddOutputRows(ctx, row)
}

func init() {
	check.Must(plug.Registry.RegisterCommand("cluster:get", &GetCmd{}))
}

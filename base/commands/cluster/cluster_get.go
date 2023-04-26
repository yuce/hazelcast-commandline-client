package cluster

import (
	"context"
	"fmt"
	"time"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
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
	cc.SetPositionalArgCount(0, 0)
	return nil
}

func (cm GetCmd) Exec(ctx context.Context, ec plug.ExecContext) error {
	ci, err := ec.ClientInternal(ctx)
	if err != nil {
		return err
	}
	_, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
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
		rows := output.Row{
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
		err = ec.AddOutputRows(ctx, rows)
		if err != nil {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		return fmt.Errorf("retrieving cluster information: %w", err)
	}
	stop()
	return nil
}

func init() {
	check.Must(plug.Registry.RegisterCommand("cluster:get", &GetCmd{}))
}

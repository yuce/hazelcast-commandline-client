//go:build std || viridian

package viridian

import (
	"context"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/ux/stage"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
	"github.com/hazelcast/hazelcast-commandline-client/internal/viridian"
)

type ClusterCreateCmd struct{}

func (ClusterCreateCmd) Init(cc plug.InitContext) error {
	long := `Creates a Viridian cluster.

Make sure you login before running this command.
`
	short := "Creates a Viridian cluster"
	cc.SetCommandUsage("create-cluster")
	cc.SetCommandHelp(long, short)
	cc.SetPositionalArgCount(0, 0)
	cc.AddStringFlag(propAPIKey, "", "", false, "Viridian API Key")
	cc.AddStringFlag(flagName, "", "", false, "specify the cluster name; if not given an auto-generated name is used.")
	if viridian.InternalOpsEnabled() {
		cc.SetCommandGroup("viridian")
		cc.AddStringFlag(flagImageTag, "", "", true, "Image name in the format: name.surname")
		cc.AddStringFlag(flagHazelcastVersion, "", "", true, "Hazelcast version")
	} else {
		cc.AddStringFlag(flagClusterType, "", viridian.ClusterTypeServerless, false, "type for the cluster")
	}
	return nil
}

func (ClusterCreateCmd) Exec(ctx context.Context, ec plug.ExecContext) error {
	ec.PrintlnUnnecessary("")
	api, err := getAPI(ec)
	if err != nil {
		return err
	}
	clusterType := ec.Props().GetString(flagClusterType)
	if clusterType == "" {
		clusterType = viridian.ClusterTypeServerless
	}
	imageTag := ec.Props().GetString(flagImageTag)
	hzVersion := ec.Props().GetString(flagHazelcastVersion)
	name := ec.Props().GetString(flagName)
	if name == "" {
		if imageTag != "" {
			name = imageTag
		} else {
			name = makeClusterName()
		}
	}
	stageState := map[string]any{}
	sp := stage.NewFixedProvider(
		createStage(ctx, ec, api, name, clusterType, imageTag, hzVersion, stageState),
		importConfigStage(ctx, ec, api, stageState, ""),
	)
	if err := stage.Execute(ctx, ec, sp); err != nil {
		return err
	}
	cluster := stageState["cluster"].(viridian.Cluster)
	verbose := ec.Props().GetBool(clc.PropertyVerbose)
	ec.PrintlnUnnecessary("")
	if verbose {
		row := output.Row{
			output.Column{
				Name:  "ID",
				Type:  serialization.TypeString,
				Value: cluster.ID,
			},
			output.Column{
				Name:  "Name",
				Type:  serialization.TypeString,
				Value: cluster.Name,
			},
		}
		return ec.AddOutputRows(ctx, row)
	}
	ec.PrintlnUnnecessary("OK Cluster creation completed successfully.")
	return nil
}

func (ClusterCreateCmd) Unwrappable() {}

func init() {
	if viridian.InternalOpsEnabled() {
		check.Must(plug.Registry.RegisterCommand("create-cluster", &ClusterCreateCmd{}))
	} else {
		check.Must(plug.Registry.RegisterCommand("viridian:create-cluster", &ClusterCreateCmd{}))
	}
}

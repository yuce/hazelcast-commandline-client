//go:build std || viridian

package viridian

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/config"
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
	if enableInternalOps {
		cc.SetCommandGroup("viridian")
		cc.AddStringFlag(flagImage, "", "", true, "Image name in the NAME:HZ_VERSION format")
	} else {
		cc.AddStringFlag(flagClusterType, "", viridian.ClusterTypeServerless, false, "type for the cluster")
	}
	return nil
}

func (ClusterCreateCmd) Exec(ctx context.Context, ec plug.ExecContext) error {
	api, err := getAPI(ec)
	if err != nil {
		return err
	}
	name := ec.Props().GetString(flagName)
	if name == "" {
		name = clusterName()
	}
	clusterType := ec.Props().GetString(flagClusterType)
	if clusterType == "" {
		clusterType = viridian.ClusterTypeServerless
	}
	image := ec.Props().GetString(flagImage)
	if enableInternalOps {
		_, _, err = splitImageName(image)
		if err != nil {
			return err
		}
	}
	//hzVersion := ec.Props().GetString(flagHazelcastVersion)
	ec.PrintlnUnnecessary("")
	var cluster viridian.Cluster
	sp := stage.NewFixedProvider(
		createStage(ctx, ec, api, name, image, clusterType, func(cs viridian.Cluster) {
			cluster = cs
		}),
		importConfigStage(ctx, ec, api, cluster, cluster.Name),
	)
	if err := stage.Execute(ctx, ec, sp); err != nil {
		return err
	}
	verbose := ec.Props().GetBool(clc.PropertyVerbose)
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
	return nil
}

func (ClusterCreateCmd) Unwrappable() {}

func tryImportConfig(ctx context.Context, ec plug.ExecContext, api *viridian.API, cluster viridian.Cluster) {
	cfgName := cluster.Name
	cp, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText("Waiting for the cluster to get ready")
		if err := waitClusterState(ctx, ec, api, cluster.ID, stateRunning); err != nil {
			// do not import the config and exit early
			return nil, err
		}
		sp.SetText("Importing configuration")
		zipPath, stop, err := api.DownloadConfig(ctx, cluster.ID)
		if err != nil {
			return nil, err
		}
		defer stop()
		cfgPath, err := config.CreateFromZip(ctx, ec, cfgName, zipPath)
		if err != nil {
			return nil, err
		}
		return cfgPath, nil
	})
	if err != nil {
		ec.Logger().Error(err)
		return
	}
	stop()
	ec.PrintlnUnnecessary(fmt.Sprintf("Cluster %s was created.", cluster.Name))
	ec.Logger().Info("Imported configuration %s and saved to: %s", cfgName, cp)
	ec.PrintlnUnnecessary(fmt.Sprintf("Imported configuration: %s", cfgName))
}

func init() {
	if enableInternalOps {
		check.Must(plug.Registry.RegisterCommand("create-cluster", &ClusterCreateCmd{}))
	} else {
		check.Must(plug.Registry.RegisterCommand("viridian:create-cluster", &ClusterCreateCmd{}))
	}
}

package viridian

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-commandline-client/clc/config"
	"github.com/hazelcast/hazelcast-commandline-client/clc/ux/stage"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/viridian"
)

func createStage(ctx context.Context, ec plug.ExecContext, api *viridian.API, name, clusterType, image string, stageStage map[string]any) stage.Stage {
	return stage.Stage{
		ProgressMsg: fmt.Sprintf("Creating cluster %s", name),
		SuccessMsg:  fmt.Sprintf("Created cluster %s", name),
		FailureMsg:  fmt.Sprintf("Failed creating cluster"),
		Func: func(status stage.Statuser) error {
			var imageTag, hzVersion string
			var err error
			if image != "" {
				imageTag, hzVersion, err = splitImageName(image)
				if err != nil {
					return err
				}
			}
			c, err := getFirstAvailableK8sCluster(ctx, api)
			if err != nil {
				return err
			}
			cs, err := api.CreateCluster(ctx, name, clusterType, c.ID, imageTag, hzVersion)
			if err != nil {
				return handleErrorResponse(ec, err)
			}
			if err := waitClusterState(ctx, ec, api, cs.ID, stateRunning); err != nil {
				return handleErrorResponse(ec, err)
			}
			if viridian.InternalOpsEnabled() {
				vc := vrdConfig{
					ClusterID: cs.ID,
					ImageName: image,
				}
				if err := saveVRDConfig(vc); err != nil {
					return err
				}
			}
			stageStage["cluster"] = cs
			return nil
		},
	}
}

func importConfigStage(ctx context.Context, ec plug.ExecContext, api *viridian.API, stageStage map[string]any, cfgName string) stage.Stage {
	return stage.Stage{
		ProgressMsg: "Importing configuration",
		SuccessMsg:  "Imported configuration",
		FailureMsg:  "Failed importing the configuration",
		Func: func(status stage.Statuser) error {
			cluster := stageStage["cluster"].(viridian.Cluster)
			zip, stop, err := api.DownloadConfig(ctx, cluster.ID)
			if err != nil {
				return handleErrorResponse(ec, err)
			}
			defer stop()
			if cfgName == "" {
				cfgName = cluster.Name
			}
			path, err := config.CreateFromZip(ctx, ec, cfgName, zip)
			if err != nil {
				return handleErrorResponse(ec, err)
			}
			ec.Logger().Info("Imported configuration %s and saved to: %s", cfgName, path)
			return nil
		},
	}
}

func stopStage(ctx context.Context, ec plug.ExecContext, api *viridian.API, clusterID string) stage.Stage {
	return stage.Stage{
		ProgressMsg: fmt.Sprintf("Stopping cluster: %s", clusterID),
		SuccessMsg:  fmt.Sprintf("Stopped cluster: %s", clusterID),
		FailureMsg:  fmt.Sprintf("Could not stop cluster: %s", clusterID),
		Func: func(status stage.Statuser) error {
			if err := api.StopCluster(ctx, clusterID); err != nil {
				return handleErrorResponse(ec, err)
			}
			if err := waitClusterState(ctx, ec, api, clusterID, stateStopped); err != nil {
				return handleErrorResponse(ec, err)
			}
			return nil
		},
	}
}

func resumeStage(ctx context.Context, ec plug.ExecContext, api *viridian.API, clusterID string) stage.Stage {
	return stage.Stage{
		ProgressMsg: fmt.Sprintf("Resuming cluster: %s", clusterID),
		SuccessMsg:  fmt.Sprintf("Resumed cluster: %s", clusterID),
		FailureMsg:  fmt.Sprintf("Could not resume cluster: %s", clusterID),
		Func: func(status stage.Statuser) error {
			if err := api.ResumeCluster(ctx, clusterID); err != nil {
				return handleErrorResponse(ec, err)
			}
			if err := waitClusterState(ctx, ec, api, clusterID, stateRunning); err != nil {
				return handleErrorResponse(ec, err)
			}
			return nil
		},
	}
}

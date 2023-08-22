//go:build std || viridian

package viridian

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/viridian"
)

type DownloadLogsCmd struct{}

func (cm DownloadLogsCmd) Init(cc plug.InitContext) error {
	if viridian.InternalOpsEnabled() {
		cc.SetCommandUsage("download-logs")
	} else {
		cc.SetCommandUsage("download-logs [cluster-ID/name] [flags]")
	}
	long := `Downloads the logs of the given Viridian cluster for the logged in API key.

Make sure you login before running this command.
`
	short := "Downloads the logs of the given Viridian cluster"
	cc.SetCommandHelp(long, short)
	if viridian.InternalOpsEnabled() {
		cc.SetCommandGroup("viridian")
		cc.SetPositionalArgCount(0, 0)
	} else {
		cc.SetPositionalArgCount(1, 1)
	}
	cc.AddStringFlag(propAPIKey, "", "", false, "Viridian API Key")
	cc.AddStringFlag(flagOutputDir, "o", "", false, "output directory for the log files, if not given current directory is used")
	return nil
}

func (cm DownloadLogsCmd) Exec(ctx context.Context, ec plug.ExecContext) error {
	api, err := getAPI(ec)
	if err != nil {
		return err
	}
	outDir := ec.Props().GetString(flagOutputDir)
	// extract target info
	if err := validateOutputDir(outDir); err != nil {
		return err
	}
	var clusterNameOrID string
	if viridian.InternalOpsEnabled() {
		vc, err := loadVRDConfig()
		if err != nil {
			return fmt.Errorf("loading vrd config: %w", err)
		}
		clusterNameOrID = vc.ClusterID
	} else {
		clusterNameOrID = ec.Args()[0]
	}
	_, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText("Downloading cluster logs")
		err := api.DownloadClusterLogs(ctx, outDir, clusterNameOrID)
		if err != nil {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		return handleErrorResponse(ec, err)
	}
	stop()
	return nil
}

func validateOutputDir(dir string) error {
	info, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if info.IsDir() {
		return nil
	}
	return errors.New("output-dir is not a directory")
}

func init() {
	if viridian.InternalOpsEnabled() {
		check.Must(plug.Registry.RegisterCommand("download-logs", &DownloadLogsCmd{}))
	} else {
		check.Must(plug.Registry.RegisterCommand("viridian:download-logs", &DownloadLogsCmd{}))
	}
}

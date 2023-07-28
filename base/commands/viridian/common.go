//go:build std || viridian

package viridian

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/viridian"
)

const (
	stateRunning      = "RUNNING"
	stateFailed       = "FAILED"
	stateStopped      = "STOPPED"
	vrdConfigFilename = "vrd_config.yaml"
)

var (
	ErrClusterFailed  = errors.New("cluster failed")
	EnableInternalOps = "no"
	enableInternalOps = false
)

func getAPI(ec plug.ExecContext) (*viridian.API, error) {
	t, err := FindTokens(ec)
	if err != nil {
		return nil, err
	}
	return viridian.NewAPI(secretPrefix, t.Key, t.AccessToken, t.RefreshToken, t.ExpiresIn), nil
}

func waitClusterState(ctx context.Context, ec plug.ExecContext, api *viridian.API, clusterIDOrName, state string) error {
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		cs, err := api.ListClusters(ctx)
		if err != nil {
			return err
		}
		for _, c := range cs {
			if c.ID != clusterIDOrName && c.Name != clusterIDOrName {
				continue
			}
			ok, err := matchClusterState(c, state)
			if err != nil {
				return err
			}
			if ok {
				return nil
			}
			ec.Logger().Info("Waiting for cluster %s with state %s to transition to %s", clusterIDOrName, c.State, state)
			time.Sleep(2 * time.Second)
		}
	}
}

func matchClusterState(cluster viridian.Cluster, state string) (bool, error) {
	if cluster.State == state {
		return true, nil
	}
	if cluster.State == stateFailed {
		return false, ErrClusterFailed
	}
	return false, nil
}

func handleErrorResponse(ec plug.ExecContext, err error) error {
	ec.Logger().Error(err)
	// if it is a http client error, return the simplified error to user
	var ce viridian.HTTPClientError
	if errors.As(err, &ce) {
		if ce.Code() == http.StatusUnauthorized {
			return fmt.Errorf("authentication error, did you login?")
		}
		return fmt.Errorf(ce.Text())
	}
	// if it is not a http client error, return it directly as always
	return err
}

func fixClusterState(state string) string {
	// this is a temporary workaround until this is changed in the API
	state = strings.Replace(state, "STOPPED", "PAUSED", 1)
	state = strings.Replace(state, "STOP", "PAUSE", 1)
	return state
}

type vrdConfig struct {
	ClusterID string
	ImageName string
}

func vrdConfigPath() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return paths.Join(dir, vrdConfigFilename), nil
}

func saveVRDConfig(cfg vrdConfig) error {
	path, err := vrdConfigPath()
	if err != nil {
		return err
	}
	b, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0600)
}

func loadVRDConfig() (vrdConfig, error) {
	var vc vrdConfig
	path, err := vrdConfigPath()
	if err != nil {
		return vc, err
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return vc, err
	}
	if err := yaml.Unmarshal(b, &vc); err != nil {
		return vc, err
	}
	return vc, nil
}

func splitImageName(image string) (name, hzVersion string, err error) {
	ps := strings.SplitN(image, ":", 2)
	if len(ps) != 2 {
		return "", "", fmt.Errorf("invalid image name: %s", image)
	}
	return ps[0], ps[1], nil
}

func init() {
	if EnableInternalOps == "yes" {
		enableInternalOps = true
	}
}

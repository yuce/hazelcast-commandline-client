package viridian

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

const ClusterTypeDevMode = "DEVMODE"
const ClusterTypeServerless = "SERVERLESS"

type platformCustomization struct {
	HzImageTag string `json:"hzImageTag"`
	HzVersion  string `json:"hzVersion"`
}

type createClusterRequest struct {
	KubernetesClusterID   int                    `json:"kubernetesClusterId"`
	Name                  string                 `json:"name"`
	ClusterTypeID         int64                  `json:"clusterTypeId"`
	PlanName              string                 `json:"planName"`
	PlatformCustomization *platformCustomization `json:"platformCustomization,omitempty"`
	Prerelease            bool                   `json:"preRelease"`
}

type createClusterResponse Cluster

func (a *API) CreateCluster(ctx context.Context, name string, clusterType string, k8sClusterID int, prerelease bool, imageTag, imageVersion string) (Cluster, error) {
	if name == "" {
		return Cluster{}, fmt.Errorf("cluster name cannot be blank")
	}
	cType, err := a.FindClusterType(ctx, clusterType)
	if err != nil {
		return Cluster{}, err
	}
	clusterTypeID := cType.ID
	planName := ClusterTypeServerless
	c := createClusterRequest{
		KubernetesClusterID: k8sClusterID,
		Name:                name,
		ClusterTypeID:       clusterTypeID,
		PlanName:            planName,
		Prerelease:          prerelease,
	}
	if imageTag != "" {
		c.PlatformCustomization = &platformCustomization{
			HzImageTag: imageTag,
			HzVersion:  imageVersion,
		}
	}
	cluster, err := RetryOnAuthFail(ctx, a, func(ctx context.Context, token string) (Cluster, error) {
		u := a.makeURL("/cluster")
		c, err := doPost[createClusterRequest, createClusterResponse](ctx, u, a.Token, c)
		return Cluster(c), err
	})
	if err != nil {
		return Cluster{}, fmt.Errorf("creating cluster: %w", err)
	}
	return cluster, nil
}

func (a *API) StopCluster(ctx context.Context, idOrName string) (Cluster, error) {
	c, err := a.FindCluster(ctx, idOrName)
	if err != nil {
		return c, err
	}
	ok, err := RetryOnAuthFail(ctx, a, func(ctx context.Context, token string) (bool, error) {
		u := a.makeURL("/cluster/%s/stop", c.ID)
		return doPost[[]byte, bool](ctx, u, a.Token, nil)
	})
	if err != nil {
		return c, fmt.Errorf("stopping cluster: %w", err)
	}
	if !ok {
		return c, errors.New("could not stop the cluster")
	}
	return c, nil
}

func (a *API) ListClusters(ctx context.Context) ([]Cluster, error) {
	csw, err := RetryOnAuthFail(ctx, a, func(ctx context.Context, token string) (Wrapper[[]Cluster], error) {
		u := a.makeURL("/cluster?size=500")
		return doGet[Wrapper[[]Cluster]](ctx, u, a.Token)
	})
	if err != nil {
		return nil, fmt.Errorf("listing clusters: %w", err)
	}
	return csw.Content, nil
}

func (a *API) ResumeCluster(ctx context.Context, idOrName string) (Cluster, error) {
	c, err := a.FindCluster(ctx, idOrName)
	if err != nil {
		return c, err
	}
	ok, err := RetryOnAuthFail(ctx, a, func(ctx context.Context, token string) (bool, error) {
		u := a.makeURL("/cluster/%s/resume", c.ID)
		return doPost[[]byte, bool](ctx, u, a.Token, nil)
	})
	if err != nil {
		return c, fmt.Errorf("resuming cluster: %w", err)
	}
	if !ok {
		return c, errors.New("could not resume the cluster")
	}
	return c, nil
}

func (a *API) DeleteCluster(ctx context.Context, idOrName string, force bool) (Cluster, error) {
	c, err := a.FindCluster(ctx, idOrName)
	if err != nil {
		return c, err
	}
	qs := ""
	if force {
		qs = "force=true"
	}
	u := a.makeURL("/cluster/%s?%s", c.ID, qs)
	_, err = RetryOnAuthFail(ctx, a, func(ctx context.Context, token string) (any, error) {
		err = doDelete(ctx, u, a.Token)
		if err != nil {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		return c, fmt.Errorf("deleting cluster: %w", err)
	}
	return c, nil
}

func (a *API) GetCluster(ctx context.Context, idOrName string) (Cluster, error) {
	cluster, err := a.FindCluster(ctx, idOrName)
	if err != nil {
		return Cluster{}, err
	}
	c, err := RetryOnAuthFail(ctx, a, func(ctx context.Context, token string) (Cluster, error) {
		u := a.makeURL("/cluster/%s", cluster.ID)
		return doGet[Cluster](ctx, u, a.Token)
	})
	if err != nil {
		return Cluster{}, fmt.Errorf("retrieving cluster: %w", err)
	}
	return c, nil
}

func (a *API) ListClusterTypes(ctx context.Context) ([]ClusterType, error) {
	csw, err := RetryOnAuthFail(ctx, a, func(ctx context.Context, token string) (Wrapper[[]ClusterType], error) {
		u := a.makeURL("/cluster_types")
		return doGet[Wrapper[[]ClusterType]](ctx, u, a.Token)
	})
	if err != nil {
		return nil, fmt.Errorf("listing cluster types: %w", err)
	}
	return csw.Content, nil
}

func (a *API) makeURL(format string, args ...any) string {
	var sb strings.Builder
	sb.WriteString(a.APIBaseURL)
	sb.WriteString(fmt.Sprintf(format, args...))
	return sb.String()
}

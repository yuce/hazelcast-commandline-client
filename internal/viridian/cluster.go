package viridian

import (
	"context"
	"errors"
	"fmt"
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
}

type createClusterResponse Cluster

func (a *API) CreateCluster(ctx context.Context, name, clusterType string, k8sClusterID int, imageTag, imageVersion string) (Cluster, error) {
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
	}
	if imageTag != "" {
		c.PlatformCustomization = &platformCustomization{
			HzImageTag: imageTag,
			HzVersion:  imageVersion,
		}
	}
	cluster, err := RetryOnAuthFail(ctx, a, func(ctx context.Context, token string) (Cluster, error) {
		c, err := doPost[createClusterRequest, createClusterResponse](ctx, "/cluster", a.Token, c)
		return Cluster(c), err
	})
	if err != nil {
		return Cluster{}, fmt.Errorf("creating cluster: %w", err)
	}
	return cluster, nil
}

func (a *API) StopCluster(ctx context.Context, idOrName string) error {
	c, err := a.FindCluster(ctx, idOrName)
	if err != nil {
		return err
	}
	ok, err := RetryOnAuthFail(ctx, a, func(ctx context.Context, token string) (bool, error) {
		return doPost[[]byte, bool](ctx, fmt.Sprintf("/cluster/%s/stop", c.ID), a.Token, nil)
	})
	if err != nil {
		return fmt.Errorf("stopping cluster: %w", err)
	}
	if !ok {
		return errors.New("could not stop the cluster")
	}
	return nil
}

func (a *API) ListClusters(ctx context.Context) ([]Cluster, error) {
	csw, err := RetryOnAuthFail(ctx, a, func(ctx context.Context, token string) (Wrapper[[]Cluster], error) {
		return doGet[Wrapper[[]Cluster]](ctx, "/cluster", a.Token)
	})
	if err != nil {
		return nil, fmt.Errorf("listing clusters: %w", err)
	}
	return csw.Content, nil
}

func (a *API) ResumeCluster(ctx context.Context, idOrName string) error {
	c, err := a.FindCluster(ctx, idOrName)
	if err != nil {
		return err
	}
	ok, err := RetryOnAuthFail(ctx, a, func(ctx context.Context, token string) (bool, error) {
		return doPost[[]byte, bool](ctx, fmt.Sprintf("/cluster/%s/resume", c.ID), a.Token, nil)
	})
	if err != nil {
		return fmt.Errorf("resuming cluster: %w", err)
	}
	if !ok {
		return errors.New("could not resume the cluster")
	}
	return nil
}

func (a *API) DeleteCluster(ctx context.Context, idOrName string, force bool) error {
	c, err := a.FindCluster(ctx, idOrName)
	if err != nil {
		return err
	}
	qs := ""
	if force {
		qs = "force=true"
	}
	u := fmt.Sprintf("/cluster/%s?%s", c.ID, qs)
	_, err = RetryOnAuthFail(ctx, a, func(ctx context.Context, token string) (any, error) {
		err = doDelete(ctx, u, a.Token)
		if err != nil {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		return fmt.Errorf("deleting cluster: %w", err)
	}
	return nil
}

func (a *API) GetCluster(ctx context.Context, idOrName string) (Cluster, error) {
	cluster, err := a.FindCluster(ctx, idOrName)
	if err != nil {
		return Cluster{}, err
	}
	c, err := RetryOnAuthFail(ctx, a, func(ctx context.Context, token string) (Cluster, error) {
		return doGet[Cluster](ctx, fmt.Sprintf("/cluster/%s", cluster.ID), a.Token)
	})
	if err != nil {
		return Cluster{}, fmt.Errorf("retrieving cluster: %w", err)
	}
	return c, nil
}

func (a *API) ListClusterTypes(ctx context.Context) ([]ClusterType, error) {
	csw, err := RetryOnAuthFail(ctx, a, func(ctx context.Context, token string) (Wrapper[[]ClusterType], error) {
		return doGet[Wrapper[[]ClusterType]](ctx, "/cluster_types", a.Token)
	})
	if err != nil {
		return nil, fmt.Errorf("listing cluster types: %w", err)
	}
	return csw.Content, nil
}

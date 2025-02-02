package it

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"

	hz "github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/logger"
)

const (
	envAPIBaseURL = "HZ_CLOUD_COORDINATOR_BASE_URL"
	envAPIKey     = "CLC_VIRIDIAN_API_KEY"
	envAPISecret  = "CLC_VIRIDIAN_API_SECRET"
)

type TestCluster interface {
	DefaultConfig() hz.Config
}

type DedicatedTestCluster struct {
	RC          *RemoteControllerClientWrapper
	ClusterID   string
	MemberUUIDs []string
	Port        int
}

func (c DedicatedTestCluster) Shutdown() {
	// TODO: add Terminate method.
	for _, memberUUID := range c.MemberUUIDs {
		c.RC.ShutdownMember(context.Background(), c.ClusterID, memberUUID)
	}
}

func (c DedicatedTestCluster) Terminate() {
	for _, memberUUID := range c.MemberUUIDs {
		c.RC.TerminateMember(context.Background(), c.ClusterID, memberUUID)
	}

}

func (c DedicatedTestCluster) DefaultConfig() hz.Config {
	config := hz.Config{}
	config.Cluster.Name = c.ClusterID
	config.Cluster.Network.SetAddresses(fmt.Sprintf("localhost:%d", c.Port))
	if SSLEnabled() {
		config.Cluster.Network.SSL.Enabled = true
		config.Cluster.Network.SSL.SetTLSConfig(&tls.Config{InsecureSkipVerify: true})
	}
	if TraceLoggingEnabled() {
		config.Logger.Level = logger.TraceLevel
	}
	return config
}

func (c DedicatedTestCluster) DefaultConfigWithNoSSL() hz.Config {
	config := hz.Config{}
	config.Cluster.Name = c.ClusterID
	config.Cluster.Network.SetAddresses(fmt.Sprintf("localhost:%d", c.Port))
	if TraceLoggingEnabled() {
		config.Logger.Level = logger.TraceLevel
	}
	return config
}

func (c DedicatedTestCluster) StartMember(ctx context.Context) (*Member, error) {
	return c.RC.StartMember(ctx, c.ClusterID)
}

type testLogger interface {
	Logf(format string, args ...interface{})
}

type SingletonTestCluster struct {
	mu       *sync.Mutex
	cls      TestCluster
	launcher func() TestCluster
	name     string
}

func NewSingletonTestCluster(name string, launcher func() TestCluster) *SingletonTestCluster {
	return &SingletonTestCluster{
		name:     name,
		mu:       &sync.Mutex{},
		launcher: launcher,
	}
}

func (c *SingletonTestCluster) Launch(t testLogger) TestCluster {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cls != nil {
		return c.cls
	}
	t.Logf("Launching the auto-shutdown test cluster: %s", c.name)
	c.cls = c.launcher()
	return c.cls
}

func StartNewClusterWithOptions(clusterName string, port, memberCount int) *DedicatedTestCluster {
	ensureRemoteController(false)
	config := XMLConfig(clusterName, port)
	return rc.startNewCluster(memberCount, config, port)
}

func StartNewClusterWithConfig(memberCount int, config string, port int) *DedicatedTestCluster {
	ensureRemoteController(false)
	return rc.startNewCluster(memberCount, config, port)
}

func XMLConfig(clusterName string, port int) string {
	return fmt.Sprintf(`
        <hazelcast xmlns="http://www.hazelcast.com/schema/config"
            xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
            xsi:schemaLocation="http://www.hazelcast.com/schema/config
            http://www.hazelcast.com/schema/config/hazelcast-config-4.0.xsd">
            <cluster-name>%s</cluster-name>
            <network>
               <port>%d</port>
            </network>
			<jet enabled="true" resource-upload-enabled="true" />
			<map name="test-mapstore">
				<map-store enabled="true">
					<class-name>com.hazelcast.client.test.SampleMapStore</class-name>
				</map-store>
			</map>
        </hazelcast>
	`, clusterName, port)
}

type ViridianClusterInfo struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	State string `json:"state"`
}

type keyValue map[string]any

type Wrapper[T any] struct {
	Content T
}

type ViridianAPI struct {
	token string
}

func loginNewViridianAPI(ctx context.Context, apiKey, apiSecret string) *ViridianAPI {
	a := &ViridianAPI{}
	a.login(ctx, apiKey, apiSecret)
	return a
}

func (a *ViridianAPI) login(ctx context.Context, apiKey, apiSecret string) {
	req := map[string]any{
		"apiKey":    apiKey,
		"apiSecret": apiSecret,
	}
	res, err := doPost[keyValue, keyValue](ctx, "/customers/api/login", "", req)
	if err != nil {
		panic(err)
	}
	a.token = res["token"].(string)
}

func (a *ViridianAPI) CreateCluster(ctx context.Context, name string) (ViridianClusterInfo, error) {
	req := map[string]any{
		"kubernetesClusterId": 1,
		"clusterTypeId":       6,
		"name":                name,
		"planName":            "SERVERLESS",
	}
	res, err := doPost[keyValue, ViridianClusterInfo](ctx, "/cluster", a.token, req)
	if err != nil {
		return ViridianClusterInfo{}, err
	}
	return res, nil
}

func (a *ViridianAPI) DeleteCluster(ctx context.Context, clusterID string) error {
	return doDelete(ctx, "/cluster/"+clusterID, a.token)
}

func (a *ViridianAPI) GetCluster(ctx context.Context, clusterID string) (ViridianClusterInfo, error) {
	res, err := doGet[ViridianClusterInfo](ctx, "/cluster/"+clusterID, a.token)
	if err != nil {
		return ViridianClusterInfo{}, err
	}
	return res, nil
}

func (a *ViridianAPI) GetClusterWithName(ctx context.Context, name string) (ViridianClusterInfo, error) {
	cs, err := a.ListClusters(ctx)
	if err != nil {
		return ViridianClusterInfo{}, err
	}
	for _, c := range cs {
		if c.Name == name {
			return c, nil
		}
	}
	return ViridianClusterInfo{}, errors.New("cluster does not exist")
}

func (a *ViridianAPI) ListClusters(ctx context.Context) ([]ViridianClusterInfo, error) {
	res, err := doGet[Wrapper[[]ViridianClusterInfo]](ctx, "/cluster", a.token)
	if err != nil {
		return nil, err
	}
	return res.Content, nil
}

func (a *ViridianAPI) StopCluster(ctx context.Context, id string) error {
	ok, err := doPost[[]byte, bool](ctx, fmt.Sprintf("/cluster/%s/stop", id), a.token, nil)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("could not stop the cluster")
	}
	return nil
}

func (a *ViridianAPI) ResumeCluster(ctx context.Context, id string) error {
	ok, err := doPost[[]byte, bool](ctx, fmt.Sprintf("/cluster/%s/resume", id), a.token, nil)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("could not resume the cluster")
	}
	return nil
}

func makeUrl(path string) string {
	path = strings.TrimLeft(path, "/")
	path = "/" + path
	return os.Getenv(envAPIBaseURL) + path
}

func doPost[Req, Res any](ctx context.Context, path, token string, request Req) (res Res, err error) {
	m, err := json.Marshal(request)
	if err != nil {
		return res, fmt.Errorf("creating request payload: %w", err)
	}
	b, err := doPostBytes(ctx, makeUrl(path), token, m)
	if err != nil {
		return res, err
	}
	if err = json.Unmarshal(b, &res); err != nil {
		return res, err
	}
	return res, nil
}

func doPostBytes(ctx context.Context, url, token string, body []byte) ([]byte, error) {
	reader := bytes.NewBuffer(body)
	req, err := http.NewRequest(http.MethodPost, url, reader)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req = req.WithContext(ctx)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}
	rb, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}
	if res.StatusCode >= 200 && res.StatusCode < 300 {
		return rb, nil
	}
	return nil, fmt.Errorf("%d: %s", res.StatusCode, string(rb))
}

func doDelete(ctx context.Context, path, token string) error {
	req, err := http.NewRequest(http.MethodDelete, makeUrl(path), nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req = req.WithContext(ctx)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("sending request: %w", err)
	}
	rb, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("reading response: %w", err)
	}
	if res.StatusCode >= 200 && res.StatusCode < 300 {
		return nil
	}
	return fmt.Errorf("%d: %s", res.StatusCode, string(rb))
}

func doGet[Res any](ctx context.Context, path, token string) (res Res, err error) {
	req, err := http.NewRequest(http.MethodGet, makeUrl(path), nil)
	if err != nil {
		return res, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req = req.WithContext(ctx)
	rawRes, err := http.DefaultClient.Do(req)
	if err != nil {
		return res, fmt.Errorf("sending request: %w", err)
	}
	rb, err := io.ReadAll(rawRes.Body)
	if err != nil {
		return res, fmt.Errorf("reading response: %w", err)
	}
	if rawRes.StatusCode == 200 {
		if err = json.Unmarshal(rb, &res); err != nil {
			return res, err
		}
		return res, nil
	}
	return res, fmt.Errorf("%d: %s", rawRes.StatusCode, string(rb))
}

type viridianTestCluster struct {
	api *ViridianAPI
}

func newViridianTestCluster() *viridianTestCluster {
	api := loginNewViridianAPI(context.Background(), ViridianAPIKey(), ViridianAPISecret())
	return &viridianTestCluster{api: api}
}

func (c viridianTestCluster) DefaultConfig() hz.Config {
	return hz.Config{}
}

func ViridianAPIKey() string {
	return os.Getenv(envAPIKey)
}

func ViridianAPISecret() string {
	return os.Getenv(envAPISecret)
}

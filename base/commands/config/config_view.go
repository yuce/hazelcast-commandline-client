//go:build base

package config

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type ViewCmd struct{}

func (cm ViewCmd) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("view [config]")
	long := fmt.Sprintf(`Displays a known configuration
	
A known configuration is a directory at %s that contains config.yaml.
Directory names which start with . or _ are ignored.
	
If the configuration name is not given, default configuration is displayed.	
`, paths.Configs())
	short := "Display a known configuration"
	cc.SetCommandHelp(long, short)
	cc.SetPositionalArgCount(0, 1)
	return nil
}

func (cm ViewCmd) Exec(ctx context.Context, ec plug.ExecContext) error {
	cn := ""
	if len(ec.Args()) > 0 {
		cn = ec.Args()[0]
	}
	path := paths.ResolveConfigPath(cn)
	if !paths.Exists(path) {
		return fmt.Errorf("configuration does not exist: %s", path)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	fmt.Println(string(b))
	return nil
}

func (cm ViewCmd) findConfigs(cd string) ([]string, error) {
	var cs []string
	es, err := os.ReadDir(cd)
	if err != nil {
		return nil, err
	}
	for _, e := range es {
		if !e.IsDir() {
			continue
		}
		if strings.HasPrefix(e.Name(), ".") || strings.HasPrefix(e.Name(), "_") {
			continue
		}
		if paths.Exists(paths.Join(cd, e.Name(), "config.yaml")) {
			cs = append(cs, e.Name())
		}
	}
	return cs, nil
}

func init() {
	Must(plug.Registry.RegisterCommand("config:view", &ViewCmd{}))
}

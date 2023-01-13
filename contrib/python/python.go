package python

import (
	"context"
	"errors"
	"os"

	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type Command struct{}

func (cm Command) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("python [flags] [file.py]")
	long := "Run Python with the CLC module enabled"
	short := "Run Python"
	cc.SetCommandHelp(long, short)
	cc.SetCommandGroup("python")
	return nil
}

func (cm Command) Exec(ctx context.Context, ec plug.ExecContext) error {
	ec.(InteractiveSetter).SetInteractive(true)
	vev, cancel, err := ec.ExecuteBlocking(ctx, "Creating the virtual environment", func(ctx context.Context) (any, error) {
		ve, err := NewVirtualEnv(ec, ec.Logger())
		if err != nil {
			return nil, err
		}
		exists, err := ve.Exists()
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
		if !exists {
			if err := ve.Create(); err != nil {
				return nil, err
			}
		}
		err = ve.InstallRequirements(
			"hazelcast-python-client==5.1",
			"psutil==5.9.3",
			"PyYAML==6.0",
		)
		if err != nil {
			return nil, err
		}
		return ve, nil
	})
	if err != nil {
		return err
	}
	defer cancel()
	ve := vev.(VirtualEnv)
	if err := ve.Exec("python", ec.Args()...); err != nil {
		return err
	}
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("python", &Command{}))
}

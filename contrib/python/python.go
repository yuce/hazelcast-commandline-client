package python

import (
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

func (cm Command) Exec(ec plug.ExecContext) error {
	ve, err := NewVirtualEnv(ec)
	if err != nil {
		return err
	}
	exists, err := ve.Exists()
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if !exists {
		if err := ve.Create(); err != nil {
			return err
		}
		err := ve.InstallRequirements(
			"hazelcast-python-client==5.1",
			"psutil==5.9.3",
			"PyYAML==6.0",
		)
		if err != nil {
			return err
		}
	}
	if err := ve.Exec("python", ec.Args()...); err != nil {
		return err
	}
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("python", &Command{}))
}

package python

import (
	"errors"
	"os"

	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type NotebookCommand struct{}

func (cm NotebookCommand) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("notebook")
	short := "Run a Jupyter Notebook"
	long := "Run a Jupyter Notebook with CLC module enabled"
	cc.SetCommandHelp(long, short)
	cc.SetCommandGroup("python")
	return nil
}

func (cm NotebookCommand) Exec(ec plug.ExecContext) error {
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
	}
	err = ve.InstallRequirements(
		"hazelcast-python-client==5.1",
		"psutil==5.9.3",
		"PyYAML==6.0",
		"notebook==6.5.1",
	)
	if err != nil {
		return err
	}
	if err := runJupyterNotebook(ve); err != nil {
		return err
	}
	return nil
}

func runJupyterNotebook(ve VirtualEnv) error {
	// cd to the notebooks dir first
	dir := paths.Join(paths.Home(), "notebooks")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	// ignore the error
	_ = os.Chdir(dir)
	return ve.Exec("jupyter", "notebook")
}

func init() {
	Must(plug.Registry.RegisterCommand("notebook", &NotebookCommand{}))
}

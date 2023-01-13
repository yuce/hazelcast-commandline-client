package python

import (
	"context"
	"errors"
	"fmt"
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
	cc.SetPositionalArgCount(0, 0)
	return nil
}

func (cm NotebookCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
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
			"https://github.com/yuce/hazelcast-python-client/archive/dbapi-support.zip#hazelcast-python-client==5.1",
			"pandas==1.5.2",
			"matplotlib==3.6.3",
			"psutil==5.9.3",
			"PyYAML==6.0",
			"notebook==6.5.1",
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
	path := paths.Join(dir, fmt.Sprintf("%s.ipynb", sampleNotebookName))
	if err := createDefaultNotebook(path); err != nil {
		panic(err)
	}
	// ignore the error
	_ = os.Chdir(dir)
	return ve.Exec("jupyter", "notebook", "--notebook-dir", dir)
}

func createDefaultNotebook(path string) error {
	if !paths.Exists(path) {
		return os.WriteFile(path, []byte(sampleNotebook), 0660)
	}
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("notebook", &NotebookCommand{}))
}

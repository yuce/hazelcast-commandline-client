package notebook

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type Command struct{}

const (
	python3Path = "/bin/python3"
)

func (cm Command) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("notebook")
	return nil
}

func (cm Command) Exec(ec plug.ExecContext) error {
	cn := ec.Props().GetString(clc.PropertyConfigPath)
	if cn == "" {
		return fmt.Errorf("config name is required")
	}
	I2(fmt.Fprintln(ec.Stdout(), "config:", cn))
	// TODO: make this more robust
	baseDir := filepath.Dir(cn)
	_, venvName := filepath.Split(baseDir)
	if err := os.MkdirAll(paths.Venvs(), 0700); err != nil {
		return fmt.Errorf("creating venvs direcory: %w", err)
	}
	venvDir := paths.Join(paths.Venvs(), venvName)
	/*
		fmt.Println("created VENV at:", venvDir)
		if err := createVenv(ec, venvDir); err != nil {
			return err
		}
		if err := installRequirements(venvDir); err != nil {
			return err
		}
	*/
	if err := writePythonModule(venvDir); err != nil {
		return err
	}
	if err := runJupyterNotebook(venvDir, cn); err != nil {
		return err
	}
	return nil
}

func writePythonModule(venvDir string) error {
	path := paths.Join(venvDir, "lib", "python3.10", "site-packages", "viridian.py")
	return os.WriteFile(path, []byte(PythonModule), 0600)
}

func runJupyterNotebook(venvDir string, configPath string) error {
	cmdPath := paths.Join(venvDir, "bin", "jupyter")
	c := exec.Command(cmdPath, "notebook")
	c.Env = append(c.Env, fmt.Sprintf("CLC_CONFIG=%s", configPath))
	return c.Run()
}

func createVenv(ec plug.ExecContext, path string) error {
	c := exec.Command(python3Path, "-m", "venv", path)
	c.Stdout = ec.Stdout()
	c.Stderr = ec.Stderr()
	/*
		if err := c.Start(); err != nil {
			return fmt.Errorf("creating venv (start): %w", err)
		}
		if err := c.Wait(); err != nil {
			return fmt.Errorf("creating venv (wait): %w", err)
		}
		op, err := c.StdoutPipe()
		if err != nil {
			return fmt.Errorf("creating venv (getting a pipe): %w", err)
		}
		_, err = io.Copy(c.Stderr, op)
		if err != nil {
			return fmt.Errorf("creating venv (copying stderr): %w", err)
		}
	*/
	return c.Run()
	//return nil
}

func installRequirements(venvPath string) error {
	reqPath, err := createRequirementsFile()
	if err != nil {
		return err
	}
	defer os.Remove(reqPath)
	// TODO: Windows
	pipPath := paths.Join(venvPath, "bin", "pip")
	c := exec.Command(pipPath, "install", "-r", reqPath)
	return c.Run()
}

func requirements() string {
	return `
hazelcast-python-client==5.1
psutil==5.9.3
PyYAML==6.0
notebook==6.5.1	
	`
}

func createRequirementsFile() (string, error) {
	rs := requirements()
	f, err := os.CreateTemp("", "requirements")
	if err != nil {
		return "", err
	}
	if _, err := f.Write([]byte(rs)); err != nil {
		return "", err
	}
	return f.Name(), nil
}

func init() {
	Must(plug.Registry.RegisterCommand("notebook", &Command{}))
}

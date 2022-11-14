package python

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	"github.com/hazelcast/hazelcast-commandline-client/internal/log"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

const (
	python3LibPath = "lib/python3.10"
)

type VirtualEnv struct {
	path    string
	cfgPath string
	lg      log.Logger
	ec      plug.ExecContext
}

func NewVirtualEnv(ec plug.ExecContext, lg log.Logger) (VirtualEnv, error) {
	ve := VirtualEnv{
		ec: ec,
		lg: lg,
	}
	cn := ec.Props().GetString(clc.PropertyConfig)
	if cn == "" {
		return ve, fmt.Errorf("config name is required")
	}
	ve.cfgPath = cn
	// TODO: make this more robust
	baseDir := filepath.Dir(cn)
	_, venvName := filepath.Split(baseDir)
	if err := os.MkdirAll(paths.Venvs(), 0700); err != nil {
		return ve, fmt.Errorf("creating venvs direcory: %w", err)
	}
	ve.path = paths.Join(paths.Venvs(), venvName)
	return ve, nil
}

func (ve VirtualEnv) Path() string {
	return ve.path
}

func (ve VirtualEnv) Exists() (bool, error) {
	_, err := os.Stat(ve.path)
	if err != nil {
		if err == os.ErrNotExist {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (ve VirtualEnv) Create() error {
	ve.ec.Logger().Info("Creating virtual env at: %s", ve.path)
	pyPath, err := python3Path()
	if err != nil {
		return errors.New("Python3 not found")
	}
	c := exec.Command(pyPath, "-m", "venv", ve.path)
	err = c.Run()
	if err != nil {
		return err
	}
	return ve.writePythonModule()
}

func (ve VirtualEnv) InstallRequirements(reqs ...string) error {
	ve.ec.Logger().Info("Installing requirements")
	reqPath, err := ve.createRequirementsFile(reqs...)
	if err != nil {
		return err
	}
	defer os.Remove(reqPath)
	// TODO: Windows
	c := exec.Command(ve.binPath("pip"), "install", "-r", reqPath)
	return c.Run()
}

func (ve VirtualEnv) Exec(cmd string, args ...string) error {
	cmdPath := paths.Join(ve.path, "bin", cmd)
	env := os.Environ()
	env = append(env,
		fmt.Sprintf("CLC_CONFIG=%s", ve.cfgPath),
		fmt.Sprintf("CLC_HOME=%s", paths.Home()),
	)
	args = append([]string{cmdPath}, args...)
	return syscall.Exec(cmdPath, args, env)
}

func (ve VirtualEnv) createRequirementsFile(reqs ...string) (string, error) {
	rs := strings.Join(reqs, "\n")
	f, err := os.CreateTemp("", "requirements.txt")
	if err != nil {
		return "", err
	}
	if _, err := f.Write([]byte(rs)); err != nil {
		return "", err
	}
	return f.Name(), nil
}

func (ve VirtualEnv) writePythonModule() error {
	spPath, err := ve.python3SitePackages()
	if err != nil {
		return err
	}
	path := paths.Join(ve.path, spPath, "clc.py")
	ve.ec.Logger().Debugf("Writing the Python module to: %s", path)
	return os.WriteFile(path, []byte(PythonModule), 0600)
}

func (ve VirtualEnv) binPath(cmd string) string {
	bp := "bin"
	if runtime.GOOS == "windows" {
		bp = "Scripts"
	}
	return paths.Join(ve.path, bp, cmd)
}

func python3Path() (string, error) {
	path, err := exec.LookPath("python3")
	if err != nil {
		path, err = exec.LookPath("python")
	}
	return path, err
}

func (ve VirtualEnv) python3SitePackages() (string, error) {
	c := exec.Command(ve.binPath("python"), "-c", `import site; print(site.getsitepackages()[0])`)
	b, err := c.Output()
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			ne := errors.New(string(ee.Stderr))
			ve.lg.Error(ne)
			return "", ne
		}
		return "", err
	}
	path := strings.TrimSpace(string(b))
	return path, nil
}

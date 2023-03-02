package ctx

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/zeromicro/go-zero/tools/goctl/rpc/execx"
	"github.com/zeromicro/go-zero/tools/goctl/util/pathx"
)

const goModuleWithoutGoFiles = "command-line-arguments"

var errInvalidGoMod = errors.New("invalid go module")

// Module contains the relative data of go module,
// which is the result of the command go list
type Module struct {
	Path      string
	Main      bool
	Dir       string
	GoMod     string
	GoVersion string
}

func (m *Module) validate() error {
	if m.Path == goModuleWithoutGoFiles || m.Dir == "" {
		return errInvalidGoMod
	}
	return nil
}

// projectFromGoMod is used to find the go module and project file path
// the workDir flag specifies which folder we need to detect based on
// only valid for go mod project
func projectFromGoMod(workDir string) (*ProjectContext, error) {
	fmt.Println("projectFrom workDir", workDir)
	if len(workDir) == 0 {
		return nil, errors.New("the work directory is not found")
	}
	if _, err := os.Stat(workDir); err != nil {
		return nil, err
	}

	workDir, err := pathx.ReadLink(workDir)
	if err != nil {
		return nil, err
	}

	fmt.Println("workDir", workDir)

	m, err := getRealModule(workDir, execx.Run)
	if err != nil {
		return nil, err
	}
	if err := m.validate(); err != nil {
		return nil, err
	}

	var ret ProjectContext
	ret.WorkDir = workDir
	ret.Name = filepath.Base(m.Dir)
	dir, err := pathx.ReadLink(m.Dir)
	if err != nil {
		return nil, err
	}

	ret.Dir = dir
	ret.Path = m.Path
	return &ret, nil
}

func getRealModule(workDir string, execRun execx.RunFunc) (*Module, error) {
	var execDir string = workDir
	// 返回上一级再执行 go list
	s := strings.Split(workDir, "\\")
	var command string
	if runtime.GOOS == "windows" {
		if len(s) >= 2 {
			execDir = strings.Join(s[:len(s)-1], "\\")
			name := s[len(s)-2] + "-service"
			command = "go mod init " + name

		}
	} else {
		execDir = strings.Join(s[:len(s)-2], "\\")
		name := s[len(s)-2] + "-service"
		command = "go mod init " + name
	}
	fmt.Println("getRealModule", command, execDir)
	execRun(command, execDir)

	data, err := execRun("go list -json -m", execDir)
	if err != nil {
		return nil, err
	}
	modules, err := decodePackages(strings.NewReader(data))
	if err != nil {
		return nil, err
	}
	for _, m := range modules {
		if strings.HasPrefix(workDir, m.Dir) {
			return &m, nil
		}
	}
	return nil, errors.New("no matched module")
}

func decodePackages(rc io.Reader) ([]Module, error) {
	var modules []Module
	decoder := json.NewDecoder(rc)
	for decoder.More() {
		var m Module
		if err := decoder.Decode(&m); err != nil {
			return nil, fmt.Errorf("invalid module: %v", err)
		}
		modules = append(modules, m)
	}

	return modules, nil
}

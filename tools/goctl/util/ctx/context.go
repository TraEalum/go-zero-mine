package ctx

import (
	"errors"
	"runtime"
	"strings"

	"github.com/zeromicro/go-zero/tools/goctl/rpc/execx"
)

var errModuleCheck = errors.New("the work directory must be found in the go mod or the $GOPATH")

// ProjectContext is a structure for the project,
// which contains WorkDir, Name, Path and Dir
type ProjectContext struct {
	WorkDir string
	// Name is the root name of the project
	// eg: go-zero、greet
	Name string
	// Path identifies which module a project belongs to, which is module value if it's a go mod project,
	// or else it is the root name of the project, eg: github.com/zeromicro/go-zero、greet
	Path string
	// Dir is the path of the project, eg: /Users/keson/goland/go/go-zero、/Users/keson/go/src/greet
	Dir string
}

// Prepare checks the project which module belongs to,and returns the path and module.
// workDir parameter is the directory of the source of generating code,
// where can be found the project path and the project module,
func Prepare(workDir string) (*ProjectContext, error) {
	var s []string
	var dir string
	var goModDir, serviceName string

	if runtime.GOOS == "windows" {
		s = strings.Split(workDir, "\\")
		goModDir = strings.Join(s[:len(s)-1], "\\")
	} else {
		s = strings.Split(workDir, "/") // 兼容linux
		goModDir = strings.Join(s[:len(s)-1], "/")
	}

	if len(s) >= 2 {
		serviceName = "go mod init " + s[len(s)-2] + "-service"
	}

	// 先移除 go.work go.work.sum 这两个文件会导致 go list 命令检测不了 go.mod
	// 执行 rm  go.work go.work.sum
	if len(s) > 2 { // 这个问题主要是在windows存在
		dir = strings.Join(s[:len(s)-2], "\\") // 回退到 app这个路径执行命令
		execx.Run("rm go.work", dir)
		execx.Run("rm go.work.sum", dir)
	}

	execx.Run(serviceName, goModDir)

	return background(workDir)
}

func background(workDir string) (*ProjectContext, error) {
	isGoMod, err := IsGoMod(workDir)
	if err != nil {
		return nil, err
	}

	if isGoMod {
		return projectFromGoMod(workDir)
	}
	return projectFromGoPath(workDir)
}

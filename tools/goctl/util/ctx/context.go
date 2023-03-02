package ctx

import (
	"errors"
	"fmt"
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
	fmt.Println("Prepare,Prepare", workDir)
	ctx, err := background(workDir)
	if err == nil {
		return ctx, nil
	}

	s := strings.Split(workDir, "\\")
	var dir string
	// 先移除 go.work go.work.sum 这两个文件会导致 go list 命令检测不了 go.mod
	// 执行 rm  go.work go.work.sum
	if len(s) > 3 {
		dir = strings.Join(s[:len(s)-3], "\\")
		fmt.Println(fmt.Printf("s:[%+v],dir:%s", s, dir))
		execx.Run("rm go.work", dir)
		execx.Run("rm go.work.sum", dir)

	}

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

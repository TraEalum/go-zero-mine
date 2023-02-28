package ctx

import (
	"errors"
	"path/filepath"
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
	ctx, err := background(workDir)
	if err == nil {
		return ctx, nil
	}

	//获取此路径下面的最低的路径名称
	name := filepath.Base(workDir)

	s := strings.Split(workDir, "\\")
	if len(s) > 2 {
		//获取倒数第二个路径
		name = s[len(s)-2] + "-service"
		workDir = strings.Join(s[:len(s)-1], "\\")
	}

	_, err = execx.Run("go mod init "+name, workDir)
	if err != nil {
		return nil, err
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

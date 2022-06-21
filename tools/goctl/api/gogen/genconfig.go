package gogen

import (
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/tools/goctl/api/spec"
	"github.com/zeromicro/go-zero/tools/goctl/config"
	"github.com/zeromicro/go-zero/tools/goctl/util"
	"github.com/zeromicro/go-zero/tools/goctl/util/format"
	"github.com/zeromicro/go-zero/tools/goctl/vars"
)

const (
	configFile     = "config"
	configTemplate = `package config

import (
	{{.authImport}}
	"github.com/zeromicro/go-zero/zrpc"
)
type Config struct {
	rest.RestConf
	{{.auth}}
	{{.jwtTrans}}
	{{.rpc}}
}
`

	jwtTemplate = ` struct {
		AccessSecret string
		AccessExpire int64
	}
`
	jwtTransTemplate = ` struct {
		Secret     string
		PrevSecret string
	}
`
)

func genConfig(dir string, cfg *config.Config, api *spec.ApiSpec) error {
	filename, err := format.FileNamingFormat(cfg.NamingFormat, configFile)
	if err != nil {
		return err
	}

	authNames := getAuths(api)
	var auths []string
	for _, item := range authNames {
		auths = append(auths, fmt.Sprintf("%s %s", item, jwtTemplate))
	}
	// generate about rpc
	types := api.Types
	need2gen := []spec.Type{}
	//Filter out unnecessary generation types
	for _, tp := range types {
		name := tp.Name()
		if !isStartWith([]string{"Update", "Query", "Create"}, name) {
			need2gen = append(need2gen, tp)
		}
	}
	var builder strings.Builder
	if len(need2gen) != 0 {
		for _, tp := range need2gen {
			name := util.Title(tp.Name())
			str := fmt.Sprintf("%s zrpc.RpcClientConf\n", name)
			builder.WriteString(str)
		}
	}
	jwtTransNames := getJwtTrans(api)
	var jwtTransList []string
	for _, item := range jwtTransNames {
		jwtTransList = append(jwtTransList, fmt.Sprintf("%s %s", item, jwtTransTemplate))
	}
	authImportStr := fmt.Sprintf("\"%s/rest\"", vars.ProjectOpenSourceURL)

	return genFile(fileGenConfig{
		dir:             dir,
		subdir:          configDir,
		filename:        filename + ".go",
		templateName:    "configTemplate",
		category:        category,
		templateFile:    configTemplateFile,
		builtinTemplate: configTemplate,
		data: map[string]string{
			"authImport": authImportStr,
			"auth":       strings.Join(auths, "\n"),
			"jwtTrans":   strings.Join(jwtTransList, "\n"),
			"rpc":        builder.String(),
		},
	})
}

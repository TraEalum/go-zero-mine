package gogen

import (
	_ "embed"
	"github.com/zeromicro/go-zero/tools/goctl/api/spec"
	"github.com/zeromicro/go-zero/tools/goctl/config"
	"strings"
)

//go:embed config-extend.tpl
var configExtendTemplate string

func genConfigExtend(dir string, cfg *config.Config, api *spec.ApiSpec) error {
	name := strings.ToLower(api.Service.Name)
	filename := "config_extend"

	return genFile(fileGenConfig{
		dir:             dir,
		subdir:          configDir,
		filename:        filename + ".go",
		templateName:    "configExtendTemplate",
		category:        category,
		templateFile:    configExtendTemplateFileFile,
		builtinTemplate: configExtendTemplate,
		data: map[string]string{
			"serviceKey": name,
		},
	})
}

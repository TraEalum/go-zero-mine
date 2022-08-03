package generator

import (
	_ "embed"
	conf "github.com/zeromicro/go-zero/tools/goctl/config"
	"github.com/zeromicro/go-zero/tools/goctl/rpc/parser"
	"github.com/zeromicro/go-zero/tools/goctl/util"
	"github.com/zeromicro/go-zero/tools/goctl/util/pathx"
	"path/filepath"
)

//go:embed config-extend.tpl
var configExtendTemplate string
var specialStr = `{{
		IpAddr: cfg.Nacos.Host,
		Port:   cfg.Nacos.Port,
	}}`

// GenConfigExtend generates the configuration structure definition file of the rpc service
func (g *Generator) GenConfigExtend(ctx DirContext, proto parser.Proto, cfg *conf.Config) error {
	dir := ctx.GetConfigExtend()
	fileName := filepath.Join(dir.Filename, "config_extend.go")
	if pathx.FileExists(fileName) {
		return nil
	}

	text, err := pathx.LoadTemplate(category, configExtendTemplateFileFile, configExtendTemplate)
	if err != nil {
		return err
	}

	return util.With("config-extend").GoFmt(true).Parse(text).SaveTo(map[string]interface{}{
		"serviceKey":   proto.Service.Name,
		"serverConfig": specialStr,
	}, fileName, true)
}

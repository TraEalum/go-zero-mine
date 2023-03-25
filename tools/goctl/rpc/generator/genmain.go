package generator

import (
	_ "embed"
	"fmt"
	"path/filepath"
	"strings"

	conf "github.com/zeromicro/go-zero/tools/goctl/config"
	"github.com/zeromicro/go-zero/tools/goctl/rpc/parser"
	"github.com/zeromicro/go-zero/tools/goctl/util"
	"github.com/zeromicro/go-zero/tools/goctl/util/format"
	"github.com/zeromicro/go-zero/tools/goctl/util/pathx"
	"github.com/zeromicro/go-zero/tools/goctl/util/stringx"
)

//go:embed main.tpl
var mainTemplate string

type MainServiceTemplateData struct {
	Service   string
	ServerPkg string
	Pkg       string
}

// GenMain generates the main file of the rpc service, which is an rpc service program call entry
func (g *Generator) GenMain(ctx DirContext, proto parser.Proto, cfg *conf.Config,
	c *ZRpcContext) error {
	mainFilename, err := format.FileNamingFormat(cfg.NamingFormat, ctx.GetServiceName().Source())
	if err != nil {
		return err
	}

	fileName := filepath.Join(ctx.GetMain().Filename, fmt.Sprintf("%v.go", mainFilename))
	imports := make([]string, 0)
	//pbImport := fmt.Sprintf(`proto "proto/%s"`, proto.Service[0].Name)
	pbImport := fmt.Sprintf(`proto "proto/%s"`, proto.PbPackage)
	svcImport := fmt.Sprintf(`"%v"`, ctx.GetSvc().Package)
	configImport := fmt.Sprintf(`"%v"`, ctx.GetConfig().Package)
	imports = append(imports, configImport, pbImport, svcImport)

	var serviceNames []MainServiceTemplateData
	for _, e := range proto.Service {
		var (
			remoteImport string
			serverPkg    string
		)
		if !c.Multiple {
			serverPkg = "server"
			remoteImport = fmt.Sprintf(`"%v"`, ctx.GetServer().Package)
		} else {
			childPkg, err := ctx.GetServer().GetChildPackage(e.Name)
			if err != nil {
				return err
			}

			serverPkg = filepath.Base(childPkg + "Server")
			remoteImport = fmt.Sprintf(`%s "%v"`, serverPkg, childPkg)
		}
		imports = append(imports, remoteImport)
		serviceNames = append(serviceNames, MainServiceTemplateData{
			Service:   parser.CamelCase(e.Name),
			ServerPkg: serverPkg,
			Pkg:       "proto",
		})
	}

	text, err := pathx.LoadTemplate(category, mainTemplateFile, mainTemplate)
	if err != nil {
		return err
	}

	etcFileName, err := format.FileNamingFormat(cfg.NamingFormat, ctx.GetServiceName().Source())
	if err != nil {
		return err
	}

	return util.With("main").GoFmt(true).Parse(text).SaveTo(map[string]interface{}{
		"serviceName":  etcFileName,
		"serviceNames": serviceNames,
		"imports":      strings.Join(imports, pathx.NL),
		"pkg":          "proto",
		"serviceNew":   stringx.From(proto.Service[0].Name).ToCamel(),
		"service":      parser.CamelCase(proto.Service[0].Name),
		"serviceKey":   proto.Service[0].Name,
	}, fileName, false)
}

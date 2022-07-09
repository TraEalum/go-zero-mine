package generator

import (
	_ "embed"
	"fmt"
	"path/filepath"

	conf "github.com/zeromicro/go-zero/tools/goctl/config"
	"github.com/zeromicro/go-zero/tools/goctl/rpc/parser"
	"github.com/zeromicro/go-zero/tools/goctl/util"
	"github.com/zeromicro/go-zero/tools/goctl/util/format"
	"github.com/zeromicro/go-zero/tools/goctl/util/pathx"
)

const modelsTemplate = `{{if .hasComment}}{{.comment}}{{end}}
func (l *{{.logicName}}) {{.method}} ({{if .hasReq}}in {{.request}}{{if .stream}},stream {{.streamBody}}{{end}}{{else}}stream {{.streamBody}}{{end}}) ({{if .hasReply}}{{.response}},{{end}} error) {
	return {{if .hasReply}}&{{.responseType}}{},{{end}} nil
}
`

//go:embed svc.tpl
var svcTemplate string

// GenSvc generates the servicecontext.go file, which is the resource dependency of a service,
// such as rpc dependency, model dependency, etc.
func (g *Generator) GenSvc(ctx DirContext, proto parser.Proto, cfg *conf.Config) error {
	dir := ctx.GetSvc()
	svcFilename, err := format.FileNamingFormat(cfg.NamingFormat, "service_context")
	if err != nil {
		return err
	}

	modelDefine, modelInit := genModels(proto.Tables)

	fileName := filepath.Join(dir.Filename, svcFilename+".go")
	text, err := pathx.LoadTemplate(category, svcTemplateFile, svcTemplate)
	if err != nil {
		return err
	}

	return util.With("svc").GoFmt(true).Parse(text).SaveTo(map[string]interface{}{
		"imports":     fmt.Sprintf(`"%v"`, ctx.GetConfig().Package),
		"modelDefine": modelDefine,
		"modelInit":   modelInit,
		"serviceName": proto.Service.Name,
	}, fileName, true)
}

func genModels(tables []string) (string, string) {
	modelDefine := ""
	modelInit := ""

	for _, tbl := range tables {
		modelDefine += fmt.Sprintf("%sModel model.%sModel\n", tbl, tbl)
		modelInit += fmt.Sprintf("%sModel: model.New%sModel(sqlConn),\n", tbl, tbl)
	}

	return modelDefine, modelInit
}

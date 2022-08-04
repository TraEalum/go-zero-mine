package generator

import (
	_ "embed"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/zeromicro/go-zero/core/collection"
	conf "github.com/zeromicro/go-zero/tools/goctl/config"
	"github.com/zeromicro/go-zero/tools/goctl/rpc/parser"
	"github.com/zeromicro/go-zero/tools/goctl/util"
	"github.com/zeromicro/go-zero/tools/goctl/util/format"
	"github.com/zeromicro/go-zero/tools/goctl/util/pathx"
	"github.com/zeromicro/go-zero/tools/goctl/util/stringx"
)

const logicFunctionTemplate = `{{if .hasComment}}{{.comment}}{{end}}
func (l *{{.logicName}}) {{.method}} ({{if .hasReq}}in {{.request}}{{if .stream}},stream {{.streamBody}}{{end}}{{else}}stream {{.streamBody}}{{end}}) ({{if .hasReply}}{{.response}},{{end}} error) {
	// todo: add your logic here and delete this line
	
	return {{if .hasReply}}&{{.responseType}}{},{{end}} nil
}
`

//go:embed logic.tpl
var logicTemplate string

// GenLogic generates the logic file of the rpc service, which corresponds to the RPC definition items in proto.
func (g *Generator) GenLogic(ctx DirContext, proto parser.Proto, cfg *conf.Config) error {
	dir := ctx.GetLogic()
	service := proto.Service.Service.Name
	for _, rpc := range proto.Service.RPC {
		logicFilename, err := format.FileNamingFormat(cfg.NamingFormat, rpc.Name+"_logic")
		if err != nil {
			return err
		}

		filename := filepath.Join(dir.Filename, logicFilename+".go")
		functions, err := g.genLogicFunction(service, proto.PbPackage, rpc)
		if err != nil {
			return err
		}

		imports := collection.NewSet()
		imports.AddStr(fmt.Sprintf(`"%v"`, ctx.GetSvc().Package))
		imports.AddStr(fmt.Sprintf(`"%v"`, ctx.GetPb().Package))
		imports.AddStr(fmt.Sprintf(`"go-service/comm/errorm"`))
		if functions.HasSqlc {
			imports.AddStr(fmt.Sprintf(`"github.com/zeromicro/go-zero/core/stores/sqlc"`))
		}
		if functions.HasUtil {
			imports.AddStr(fmt.Sprintf(`"go-service/comm/util"`))
		}
		if functions.HasModel {
			imports.AddStr(fmt.Sprintf(`"go-service/app/%s/model"`, proto.Service.Name))
		}
		text, err := pathx.LoadTemplate(category, logicTemplateFileFile, logicTemplate)
		if err != nil {
			return err
		}
		err = util.With("logic").GoFmt(true).Parse(text).SaveTo(map[string]interface{}{
			"logicName": fmt.Sprintf("%sLogic", stringx.From(rpc.Name).ToCamel()),
			"functions": functions.Fn,
			"imports":   strings.Join(imports.KeysStr(), pathx.NL),
		}, filename, false)
		if err != nil {
			return err
		}
	}
	return nil
}

type genLogic struct {
	HasSqlc  bool
	HasUtil  bool
	HasModel bool

	Fn string
}

func (g *Generator) genLogicFunction(serviceName, goPackage string, rpc *parser.RPC) (genLogic, error) {
	functions := make([]string, 0)
	gen := genLogic{}
	text, err := pathx.LoadTemplate(category, logicFuncTemplateFileFile, logicFunctionTemplate)
	if err != nil {
		return gen, err
	}
	modelName := ""

	// load curd template
	switch parser.CamelCase(rpc.Name) {
	case fmt.Sprintf("Create%s", parser.CamelCase(rpc.RequestType)):
		text = CreateLogic
		modelName = parser.CamelCase(rpc.RequestType)
		gen.HasSqlc = true
		gen.HasModel = true
	case fmt.Sprintf("Delete%s", parser.CamelCase(rpc.RequestType)):
		text = DeleteLogic
		modelName = parser.CamelCase(rpc.RequestType)
		gen.HasModel = true
	case fmt.Sprintf("Query%sDetail", strings.Replace(parser.CamelCase(rpc.RequestType), "Filter", "", 1)):
		text = QueryDetailLogic
		modelName = strings.Replace(parser.CamelCase(rpc.RequestType), "Filter", "", 1)
		gen.HasModel = true
	case fmt.Sprintf("Query%sList", strings.Replace(parser.CamelCase(rpc.RequestType), "Filter", "", 1)):
		text = QueryLogic
		modelName = strings.Replace(parser.CamelCase(rpc.RequestType), "Filter", "", 1)
		gen.HasUtil = true
		gen.HasModel = true

	case fmt.Sprintf("Update%s", parser.CamelCase(rpc.RequestType)):
		text = UpdateLogic
		gen.HasSqlc = true
		gen.HasUtil = true
		modelName = parser.CamelCase(rpc.RequestType)
		gen.HasModel = true
	}

	logicName := stringx.From(rpc.Name + "_logic").ToCamel()
	comment := parser.GetComment(rpc.Doc())
	streamServer := fmt.Sprintf("%s.%s_%s%s", goPackage, parser.CamelCase(serviceName), parser.CamelCase(rpc.Name), "Server")
	buffer, err := util.With("fun").Parse(text).Execute(map[string]interface{}{
		"logicName":           logicName,
		"method":              parser.CamelCase(rpc.Name),
		"hasReq":              !rpc.StreamsRequest,
		"request":             fmt.Sprintf("*%s.%s", goPackage, parser.CamelCase(rpc.RequestType)),
		"hasReply":            !rpc.StreamsRequest && !rpc.StreamsReturns,
		"response":            fmt.Sprintf("*%s.%s", goPackage, parser.CamelCase(rpc.ReturnsType)),
		"responseType":        fmt.Sprintf("%s.%s", goPackage, parser.CamelCase(rpc.ReturnsType)),
		"stream":              rpc.StreamsRequest || rpc.StreamsReturns,
		"streamBody":          streamServer,
		"hasComment":          len(comment) > 0,
		"comment":             comment,
		"modelName":           modelName,
		"modelNameFirstLower": FirstLower(modelName),
	})
	if err != nil {
		return gen, err
	}

	functions = append(functions, buffer.String())
	gen.Fn = strings.Join(functions, pathx.NL)
	return gen, nil
}

// FirstLower 字符串首字母小写
func FirstLower(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToLower(s[:1]) + s[1:]
}

package generator

import (
	_ "embed"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/emicklei/proto"
	"github.com/zeromicro/go-zero/core/collection"
	conf "github.com/zeromicro/go-zero/tools/goctl/config"
	"github.com/zeromicro/go-zero/tools/goctl/rpc/parser"
	"github.com/zeromicro/go-zero/tools/goctl/util"
	"github.com/zeromicro/go-zero/tools/goctl/util/format"
	"github.com/zeromicro/go-zero/tools/goctl/util/pathx"
	"github.com/zeromicro/go-zero/tools/goctl/util/stringx"
)

const (
	callInterfaceFunctionTemplate = `{{if .hasComment}}{{.comment}}
{{end}}{{.method}}(ctx context.Context{{if .hasReq}}, in *{{.pbRequest}}{{end}}, opts ...grpc.CallOption) ({{if .notStream}}*{{.pbResponse}}, {{else}}{{.streamBody}},{{end}} error)`

	callFunctionTemplate = `
{{if .hasComment}}{{.comment}}{{end}}
func (m *default{{.serviceName}}) {{.method}}(ctx context.Context{{if .hasReq}}, in *{{.pbRequest}}{{end}}, opts ...grpc.CallOption) ({{if .notStream}}*{{.pbResponse}}, {{else}}{{.streamBody}},{{end}} error) {
	client := {{if .isCallPkgSameToGrpcPkg}}{{else}}{{.package}}.{{end}}New{{.rpcServiceName}}Client(m.cli.Conn())
	return client.{{.method}}(ctx{{if .hasReq}}, in{{end}}, opts...)
}
`
)

//go:embed call.tpl
var callTemplateText string

//go:embed client.tpl
var clientTemplateText string

// GenCall generates the rpc client code, which is the entry point for the rpc service call.
// It is a layer of encapsulation for the rpc client and shields the details in the pb.
func (g *Generator) GenCall(ctx DirContext, proto parser.Proto, cfg *conf.Config,
	c *ZRpcContext) error {
	if !c.Multiple {
		return g.genCallInCompatibility(ctx, proto, cfg)
	}

	return g.genCallGroup(ctx, proto, cfg)
}

func (g *Generator) genCallGroup(ctx DirContext, proto parser.Proto, cfg *conf.Config) error {
	dir := ctx.GetCall()
	head := util.GetHead(proto.Name)

	cli := ""
	imports := ""
	newCli := ""
	serviceName := strings.Replace(proto.Name, ".proto", "", -1)
	for _, service := range proto.Service {
		// get xxxCli
		// get xxxImports
		childPkg, err := dir.GetChildPackage(service.Name)
		if err != nil {
			return err
		}

		callFilename, err := format.FileNamingFormat(cfg.NamingFormat, service.Name)
		if err != nil {
			return err
		}

		childDir := filepath.Base(childPkg)
		filename := filepath.Join(dir.Filename, childDir, fmt.Sprintf("%s.go", callFilename))
		isCallPkgSameToPbPkg := childDir == ctx.GetProtoGo().Filename
		isCallPkgSameToGrpcPkg := childDir == ctx.GetProtoGo().Filename

		functions, err := g.genFunction(proto.PbPackage, service, isCallPkgSameToGrpcPkg)
		if err != nil {
			return err
		}

		iFunctions, err := g.getInterfaceFuncs(proto.PbPackage, service, isCallPkgSameToGrpcPkg)
		if err != nil {
			return err
		}

		text, err := pathx.LoadTemplate(category, callTemplateFile, callTemplateText)
		if err != nil {
			return err
		}

		alias := collection.NewSet()
		if !isCallPkgSameToPbPkg {
			for _, item := range proto.Message {
				msgName := getMessageName(*item.Message)
				alias.AddStr(fmt.Sprintf("%s = %s", parser.CamelCase(msgName),
					fmt.Sprintf("%s.%s", "proto", parser.CamelCase(msgName))))
			}
		}

		//pbPackage := fmt.Sprintf(`"%s"`, ctx.GetPb().Package)
		pbPackage := fmt.Sprintf(`%s "proto/%s"`, "proto", proto.PbPackage)
		//protoGoPackage := fmt.Sprintf(`"%s"`, ctx.GetProtoGo().Package)
		protoGoPackage := ""
		if isCallPkgSameToGrpcPkg {
			pbPackage = ""
			protoGoPackage = ""
		}

		aliasKeys := alias.KeysStr()
		sort.Strings(aliasKeys)
		filePackageName := fmt.Sprintf("%sClient", FirstLower(stringx.From(service.Name).ToCamel()))
		if err = util.With("shared").GoFmt(true).Parse(text).SaveTo(map[string]interface{}{
			"name":  callFilename,
			"alias": strings.Join(aliasKeys, pathx.NL),
			"head":  head,
			// "filePackage":    dir.Base,
			"filePackage":    filePackageName,
			"pbPackage":      pbPackage,
			"protoGoPackage": protoGoPackage,
			"serviceName":    stringx.From(service.Name).ToCamel(),
			"functions":      strings.Join(functions, pathx.NL),
			"interface":      strings.Join(iFunctions, pathx.NL),
		}, filename, true); err != nil {
			return err
		}

		name := stringx.From(service.Name).ToCamel()
		importName := strings.ReplaceAll(service.Name, "_", "")
		importName = strings.ToLower(importName)

		cli += fmt.Sprintf("%s  %s.%s%s\n", name, filePackageName, name, "Cli")
		newCli += fmt.Sprintf("%s:%s.New%s(cli),\n", name, filePackageName, name)
		imports += fmt.Sprintf("%s \"%s-service/rpc/client/%s\"\n", filePackageName, serviceName, importName)
	}

	// 生成client
	text, err := pathx.LoadTemplate(category, clientTemplateFile, clientTemplateText)
	if err != nil {
		return err
	}

	filename := filepath.Join(dir.Filename, "client.go")
	serviceName = stringx.From(serviceName).FirstUpper()

	if err = util.With("shared").GoFmt(true).Parse(text).SaveTo(map[string]interface{}{
		"head":    head,
		"imports": imports,
		"cli":     cli,
		"newCli":  newCli,
		"service": serviceName,
	}, filename, true); err != nil {
		return err
	}

	return nil
}

func (g *Generator) genCallInCompatibility(ctx DirContext, proto parser.Proto,
	cfg *conf.Config) error {
	dir := ctx.GetCall()
	service := proto.Service[0]
	head := util.GetHead(proto.Name)
	isCallPkgSameToPbPkg := ctx.GetCall().Filename == ctx.GetPb().Filename
	isCallPkgSameToGrpcPkg := ctx.GetCall().Filename == ctx.GetProtoGo().Filename

	callFilename, err := format.FileNamingFormat(cfg.NamingFormat, service.Name)
	if err != nil {
		return err
	}

	filename := filepath.Join(dir.Filename, fmt.Sprintf("%s.go", callFilename))
	functions, err := g.genFunction(proto.PbPackage, service, isCallPkgSameToGrpcPkg)
	if err != nil {
		return err
	}

	iFunctions, err := g.getInterfaceFuncs(proto.PbPackage, service, isCallPkgSameToGrpcPkg)
	if err != nil {
		return err
	}

	text, err := pathx.LoadTemplate(category, callTemplateFile, callTemplateText)
	if err != nil {
		return err
	}

	alias := collection.NewSet()
	if !isCallPkgSameToPbPkg {
		for _, item := range proto.Message {
			msgName := getMessageName(*item.Message)
			alias.AddStr(fmt.Sprintf("%s = %s", parser.CamelCase(msgName), fmt.Sprintf("%s.%s", "proto", parser.CamelCase(msgName))))
		}
	}

	pbPackage := fmt.Sprintf(`proto "proto/%s"`, proto.PbPackage)
	// protoGoPackage := fmt.Sprintf(`"%s"`, ctx.GetProtoGo().Package)
	protoGoPackage := ""
	// fmt.Printf("head[%s] dir.Base[%s] protoGoPackage[%s] pbPackage[%s]", head, dir.Base, protoGoPackage, pbPackage)
	if isCallPkgSameToGrpcPkg {
		pbPackage = ""
		protoGoPackage = ""
	}
	aliasKeys := alias.KeysStr()
	sort.Strings(aliasKeys)
	return util.With("shared").GoFmt(true).Parse(text).SaveTo(map[string]interface{}{
		"name":           callFilename,
		"alias":          strings.Join(aliasKeys, pathx.NL),
		"head":           head,
		"filePackage":    dir.Base,
		"pbPackage":      pbPackage,
		"protoGoPackage": protoGoPackage,
		"serviceName":    stringx.From(service.Name).ToCamel(),
		"functions":      strings.Join(functions, pathx.NL),
		"interface":      strings.Join(iFunctions, pathx.NL),
	}, filename, true)
}

func getMessageName(msg proto.Message) string {
	list := []string{msg.Name}

	for {
		parent := msg.Parent
		if parent == nil {
			break
		}

		parentMsg, ok := parent.(*proto.Message)
		if !ok {
			break
		}

		tmp := []string{parentMsg.Name}
		list = append(tmp, list...)
		msg = *parentMsg
	}

	return strings.Join(list, "_")
}

func (g *Generator) genFunction(goPackage string, service parser.Service,
	isCallPkgSameToGrpcPkg bool) ([]string, error) {
	functions := make([]string, 0)

	for _, rpc := range service.RPC {
		text, err := pathx.LoadTemplate(category, callFunctionTemplateFile, callFunctionTemplate)
		if err != nil {
			return nil, err
		}

		comment := parser.GetComment(rpc.Doc())
		streamServer := fmt.Sprintf("%s.%s_%s%s", goPackage, parser.CamelCase(service.Name),
			parser.CamelCase(rpc.Name), "Client")
		if isCallPkgSameToGrpcPkg {
			streamServer = fmt.Sprintf("%s_%s%s", parser.CamelCase(service.Name),
				parser.CamelCase(rpc.Name), "Client")
		}
		buffer, err := util.With("sharedFn").Parse(text).Execute(map[string]interface{}{
			"serviceName":            stringx.From(service.Name).ToCamel(),
			"rpcServiceName":         parser.CamelCase(service.Name),
			"method":                 parser.CamelCase(rpc.Name),
			"package":                "proto",
			"pbRequest":              parser.CamelCase(rpc.RequestType),
			"pbResponse":             parser.CamelCase(rpc.ReturnsType),
			"hasComment":             len(comment) > 0,
			"comment":                comment,
			"hasReq":                 !rpc.StreamsRequest,
			"notStream":              !rpc.StreamsRequest && !rpc.StreamsReturns,
			"streamBody":             streamServer,
			"isCallPkgSameToGrpcPkg": isCallPkgSameToGrpcPkg,
		})
		if err != nil {
			return nil, err
		}

		functions = append(functions, buffer.String())
	}

	return functions, nil
}

func (g *Generator) getInterfaceFuncs(goPackage string, service parser.Service,
	isCallPkgSameToGrpcPkg bool) ([]string, error) {
	functions := make([]string, 0)

	for _, rpc := range service.RPC {
		text, err := pathx.LoadTemplate(category, callInterfaceFunctionTemplateFile,
			callInterfaceFunctionTemplate)
		if err != nil {
			return nil, err
		}

		comment := parser.GetComment(rpc.Doc())
		streamServer := fmt.Sprintf("%s.%s_%s%s", goPackage, parser.CamelCase(service.Name),
			parser.CamelCase(rpc.Name), "Client")
		if isCallPkgSameToGrpcPkg {
			streamServer = fmt.Sprintf("%s_%s%s", parser.CamelCase(service.Name),
				parser.CamelCase(rpc.Name), "Client")
		}
		buffer, err := util.With("interfaceFn").Parse(text).Execute(
			map[string]interface{}{
				"hasComment": len(comment) > 0,
				"comment":    comment,
				"method":     parser.CamelCase(rpc.Name),
				"hasReq":     !rpc.StreamsRequest,
				"pbRequest":  parser.CamelCase(rpc.RequestType),
				"notStream":  !rpc.StreamsRequest && !rpc.StreamsReturns,
				"pbResponse": parser.CamelCase(rpc.ReturnsType),
				"streamBody": streamServer,
			})
		if err != nil {
			return nil, err
		}

		functions = append(functions, buffer.String())
	}

	return functions, nil
}

package gogen

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/tools/goctl/api/spec"
	"github.com/zeromicro/go-zero/tools/goctl/config"
	"github.com/zeromicro/go-zero/tools/goctl/util"
	"github.com/zeromicro/go-zero/tools/goctl/util/format"
	"github.com/zeromicro/go-zero/tools/goctl/util/pathx"
	"github.com/zeromicro/go-zero/tools/goctl/vars"
)

const contextFilename = "service_context"

//go:embed svc.tpl
var contextTemplate string

func genServiceContext(dir, rootPkg string, cfg *config.Config, api *spec.ApiSpec) error {
	filename, err := format.FileNamingFormat(cfg.NamingFormat, contextFilename)
	if err != nil {
		return err
	}

	var middlewareStr string
	var middlewareAssignment string
	middlewares := getMiddleware(api)
	types := api.Types
	var need2gen []spec.Type
	//Filter out unnecessary generation types
	for _, tp := range types {
		name := tp.Name()
		if !isStartWith([]string{"Update", "Query", "Create"}, name) {
			need2gen = append(need2gen, tp)
		}
	}
	rpcImport := ""
	rpc := ""
	rpcInit := ""
	if len(need2gen) != 0 {
		rpcImport = genRpcImport(api, need2gen)
		rpc = genRpc(need2gen)
		rpcInit = genRpcInit(need2gen)
	}
	for _, item := range middlewares {
		middlewareStr += fmt.Sprintf("%s rest.Middleware\n", item)
		name := strings.TrimSuffix(item, "Middleware") + "Middleware"
		middlewareAssignment += fmt.Sprintf("%s: %s,\n", item,
			fmt.Sprintf("middleware.New%s().%s", strings.Title(name), "Handle"))
	}

	configImport := "\"" + pathx.JoinPackages(rootPkg, configDir) + "\""
	if len(middlewareStr) > 0 {
		configImport += "\n\t\"" + pathx.JoinPackages(rootPkg, middlewareDir) + "\""
		configImport += fmt.Sprintf("\n\t\"%s/rest\"", vars.ProjectOpenSourceURL)
	}

	return genFile(fileGenConfig{
		dir:             dir,
		subdir:          contextDir,
		filename:        filename + ".go",
		templateName:    "contextTemplate",
		category:        category,
		templateFile:    contextTemplateFile,
		builtinTemplate: contextTemplate,
		data: map[string]string{
			"configImport":         configImport,
			"config":               "config.Config\n",
			"middleware":           middlewareStr,
			"middlewareAssignment": middlewareAssignment,
			"rpcImport":            rpcImport,
			"rpc":                  rpc,
			"rpcInit":              rpcInit,
		},
	})
}

func genRpcImport(api *spec.ApiSpec, types []spec.Type) string {
	var build strings.Builder
	build.WriteString("\"github.com/zeromicro/go-zero/zrpc\"\n")
	for _, tp := range types {
		tableName := util.Title(tp.Name())
		lowerTableName := strings.ToLower(tableName)
		str := fmt.Sprintf("\"go-service/app/%s/rpc/%s\"\n", api.Service.Name, lowerTableName)
		build.WriteString(str)
	}
	return build.String()
}

func genRpc(types []spec.Type) string {
	var build strings.Builder
	for _, tp := range types {
		tableName := util.Title(tp.Name())
		str := fmt.Sprintf("%s %s.%s\n", tableName, strings.ToLower(tableName), tableName)
		build.WriteString(str)
	}
	build.WriteString("")
	return build.String()
}

func genRpcInit(types []spec.Type) string {
	var build strings.Builder
	for _, tp := range types {
		tableName := util.Title(tp.Name())
		str := fmt.Sprintf("%s: %s.New%s(zrpc.MustNewClient(c.%s)),\n", tableName, strings.ToLower(tableName), tableName, tableName)
		build.WriteString(str)
	}
	return build.String()
}

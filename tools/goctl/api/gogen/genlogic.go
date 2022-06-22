package gogen

import (
	_ "embed"
	"fmt"
	"path"
	"strconv"
	"strings"

	"github.com/zeromicro/go-zero/tools/goctl/api/parser/g4/gen/api"
	"github.com/zeromicro/go-zero/tools/goctl/api/spec"
	"github.com/zeromicro/go-zero/tools/goctl/config"
	"github.com/zeromicro/go-zero/tools/goctl/util/format"
	"github.com/zeromicro/go-zero/tools/goctl/util/pathx"
	"github.com/zeromicro/go-zero/tools/goctl/vars"
)

//go:embed logic.tpl
var logicTemplate string

func genLogic(dir, rootPkg string, cfg *config.Config, api *spec.ApiSpec) error {
	for _, g := range api.Service.Groups {
		for _, r := range g.Routes {
			err := genLogicByRoute(dir, rootPkg, cfg, g, r, api)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func genLogicByRoute(dir, rootPkg string, cfg *config.Config, group spec.Group, route spec.Route, api *spec.ApiSpec) error {
	logic := getLogicName(route)
	goFile, err := format.FileNamingFormat(cfg.NamingFormat, logic)
	if err != nil {
		return err
	}

	imports := genLogicImports(route, rootPkg)
	var responseString string
	var returnString string
	var requestString string
	if len(route.ResponseTypeName()) > 0 {
		resp := responseGoTypeName(route, typesPacket)
		responseString = "(resp " + resp + ", err error)"
		returnString = "return"
	} else {
		responseString = "error"
		returnString = "return nil"
	}
	if len(route.RequestTypeName()) > 0 {
		requestString = "req *" + requestGoTypeName(route, typesPacket)
	}
	path := fmt.Sprintf("\"go-service/app/%s/rpc/proto\"", api.Service.Name)
	subDir := getLogicFolderPath(group, route)
	return genFile(fileGenConfig{
		dir:             dir,
		subdir:          subDir,
		filename:        goFile + ".go",
		templateName:    "logicTemplate",
		category:        category,
		templateFile:    logicTemplateFile,
		builtinTemplate: logicTemplate,
		data: map[string]string{
			"pkgName":      subDir[strings.LastIndex(subDir, "/")+1:],
			"imports":      imports,
			"logic":        strings.Title(logic),
			"function":     strings.Title(strings.TrimSuffix(logic, "Logic")),
			"responseType": responseString,
			"returnString": returnString,
			"request":      requestString,
			"context":      genLogicContext(logic),
			"rpcImport":    path,
		},
	})
}

func getLogicFolderPath(group spec.Group, route spec.Route) string {
	folder := route.GetAnnotation(groupProperty)
	if len(folder) == 0 {
		folder = group.GetAnnotation(groupProperty)
		if len(folder) == 0 {
			return logicDir
		}
	}
	folder = strings.TrimPrefix(folder, "/")
	folder = strings.TrimSuffix(folder, "/")
	return path.Join(logicDir, folder)
}

func genLogicImports(route spec.Route, parentPkg string) string {
	var imports []string
	imports = append(imports, `"context"`+"\n")
	imports = append(imports, fmt.Sprintf("\"%s\"", pathx.JoinPackages(parentPkg, contextDir)))
	if shallImportTypesPackage(route) {
		imports = append(imports, fmt.Sprintf("\"%s\"\n", pathx.JoinPackages(parentPkg, typesDir)))
	}
	imports = append(imports, fmt.Sprintf("\"%s/core/logx\"", vars.ProjectOpenSourceURL))
	return strings.Join(imports, "\n\t")
}

func onlyPrimitiveTypes(val string) bool {
	fields := strings.FieldsFunc(val, func(r rune) bool {
		return r == '[' || r == ']' || r == ' '
	})

	for _, field := range fields {
		if field == "map" {
			continue
		}
		// ignore array dimension number, like [5]int
		if _, err := strconv.Atoi(field); err == nil {
			continue
		}
		if !api.IsBasicType(field) {
			return false
		}
	}

	return true
}

func shallImportTypesPackage(route spec.Route) bool {
	if len(route.RequestTypeName()) > 0 {
		return true
	}

	respTypeName := route.ResponseTypeName()
	if len(respTypeName) == 0 {
		return false
	}

	if onlyPrimitiveTypes(respTypeName) {
		return false
	}

	return true
}

func genLogicContext(logic string) string {
	var builder strings.Builder
	title := strings.Title(strings.TrimSuffix(logic, "Logic"))
	if strings.Contains(title, "Create") {
		tableName := title[6:]
		paraName := strings.ToLower(tableName)
		builder.WriteString(fmt.Sprintf("\tvar %s *proto.%s\n", paraName, tableName))
		builder.WriteString(fmt.Sprintf("\treq.Unmarshal(%s)\n", paraName))
		builder.WriteString(fmt.Sprintf("\trpcResp,err := l.svcCtx.%s.%s(l.ctx, %s)\n", tableName, title, paraName))
		builder.WriteString("\tif err != nil {\n")
		builder.WriteString("\t\treturn nil,err\n")
		builder.WriteString("\t}\n")
		builder.WriteString("\tresp.Marshal(rpcResp)\n")
	} else if strings.Contains(title, "Query") {
		tableName := title[5:]
		paraName := strings.ToLower(tableName)
		builder.WriteString(fmt.Sprintf("\tvar %s *proto.%sFilter\n", paraName, tableName))
		builder.WriteString(fmt.Sprintf("\treq.Unmarshal(%s)\n", paraName))
		builder.WriteString(fmt.Sprintf("\trpcResp, err := l.svcCtx.%s.Query%sDetail(l.ctx, %s)\n", tableName, tableName, paraName))
		builder.WriteString("\tif err != nil {\n")
		builder.WriteString("\t\treturn nil,err\n")
		builder.WriteString("\t}\n")
		builder.WriteString("\tresp.Marshal(rpcResp)\n")
	} else if strings.Contains(title, "Update") {
		tableName := title[6:]
		paraName := strings.ToLower(tableName)
		builder.WriteString(fmt.Sprintf("\tvar %s *proto.%s\n", paraName, tableName))
		builder.WriteString(fmt.Sprintf("\treq.Unmarshal(%s)\n", paraName))
		builder.WriteString(fmt.Sprintf("\trpcResp, err := l.svcCtx.%s.Update%s(l.ctx, %s)\n", tableName, tableName, paraName))
		builder.WriteString("\tif err != nil {\n")
		builder.WriteString("\t\treturn nil,err\n")
		builder.WriteString("\t}\n")
		builder.WriteString("\tresp.Marshal(rpcResp)\n")
	} else {
		return ""
	}
	return builder.String()
}

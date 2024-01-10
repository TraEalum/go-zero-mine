package generator

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

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
	var modelDefine, modelInit string
	dir := ctx.GetSvc()
	svcFilename, err := format.FileNamingFormat(cfg.NamingFormat, "service_context")
	if err != nil {
		return err
	}

	// fmt.Println(proto.Tables)
	if len(proto.Tables) != 0 {
		modelDefine, modelInit = genModels(proto.Tables)
	}

	fileName := filepath.Join(dir.Filename, svcFilename+".go")
	text := ""
	if pathx.FileExists(fileName) {
		modelInit, err = dealExistsModelInit(modelInit, fileName, proto.Tables)
		if err != nil {
			return err
		}
		// modify
		text, err = text2Template(fileName)
	} else {
		text, err = pathx.LoadTemplate(category, svcTemplateFile, svcTemplate)
	}
	if err != nil {
		return err
	}

	return util.With("svc").GoFmt(true).Parse(text).SaveTo(map[string]interface{}{
		"imports":     fmt.Sprintf(`"%v"`, ctx.GetConfig().Package),
		"modelDefine": modelDefine,
		"modelInit":   modelInit,
		"serviceName": proto.PbPackage,
	}, fileName, true)
}

func genModels(tables []string) (string, string) {
	modelDefine := ""
	modelInit := ""

	for _, tbl := range tables {
		if tbl == "" {
			continue
		}
		modelDefine += fmt.Sprintf("%sModel model.%sModel\n", tbl, tbl)
		modelInit += fmt.Sprintf("%sModel: model.New%sModel(sqlConn),\n", tbl, tbl)
	}

	return modelDefine, modelInit
}

func dealExistsModelInit(modelInit string, fileName string, tables []string) (string, error) {
	if modelInit == "" {
		return modelInit, nil
	}

	fileBytes, err := os.ReadFile(fileName)
	if err != nil {
		return "", err
	}

	modelInitMap := map[string]string{}
	tmpModelInit := strings.Split(modelInit, ",")
	re := regexp.MustCompile("[1-9|a-zA-z]+Model:")

	for i := 0; i < len(tmpModelInit); i++ {
		match := re.FindAllString(tmpModelInit[i], -1)
		if len(match) == 0 || !strings.Contains(tmpModelInit[i], "Model") {
			continue
		}

		modelInitMap[match[0]] = strings.Trim(tmpModelInit[i], "\r\n")
	}
	//前数据  正则 标签: // <codeGeneratedModelDefine>
	content := string(fileBytes)
	re = regexp.MustCompile("// <codeGeneratedModelInit>([^}]+)")
	tagModelDefineStr := re.FindStringSubmatch(content)

	if len(tagModelDefineStr) == 0 || tagModelDefineStr[0] == "" {
		return modelInit, nil
	}

	tagModelDefineStr[0] = strings.Replace(tagModelDefineStr[0], "// <codeGeneratedModelInit>", "", -1)
	tagModelDefineStr[0] = strings.Replace(tagModelDefineStr[0], "// </codeGeneratedModelInit>", "", -1)
	modelExMap := map[string]string{}
	tmpModelDefineArr := strings.Split(tagModelDefineStr[0], ",")

	re = regexp.MustCompile("[1-9|a-zA-z]+Model:")
	var tmpModelDefineNoMatch []string
	for i := 0; i < len(tmpModelDefineArr); i++ {
		if !strings.Contains(tmpModelDefineArr[i], "Model") {
			continue
		}

		if !strings.Contains(tmpModelDefineArr[i], "(") {
			continue
		}

		if !strings.Contains(tmpModelDefineArr[i], ")") && !strings.Contains(tmpModelDefineArr[i], "//") {
			tmpModelDefineArr[i] += "," + tmpModelDefineArr[i+1]
		}

		match := re.FindAllString(tmpModelDefineArr[i], 1)
		if len(match) == 0 {
			tmpModelDefineNoMatch = append(tmpModelDefineNoMatch, strings.Trim(tmpModelDefineArr[i], "\r\n"))
		} else {
			_, ok := modelExMap[match[0]]
			if ok {
				continue
			}

			modelExMap[match[0]] = strings.Trim(tmpModelDefineArr[i], "\r\n")
		}
	}

	//modelInitMap modelExMap  制作完成
	var newTagInfo []string
	for _, table := range tables {
		tmpName := table + "Model:"
		if _, ok := modelExMap[tmpName]; ok {
			//存在,则沿用之前的
			newTagInfo = append(newTagInfo, modelExMap[tmpName])
		} else {
			if !strings.Contains(modelInitMap[tmpName], "Model") {
				continue
			}

			newTagInfo = append(newTagInfo, modelInitMap[tmpName])
		}
	}

	for _, v := range tmpModelDefineNoMatch {
		newTagInfo = append(newTagInfo, v)
	}

	if len(newTagInfo) > 0 {
		newTagInfo[len(newTagInfo)-1] += ","
	}

	newTagStr := strings.Join(newTagInfo, ",\r\n")
	return newTagStr, nil
}

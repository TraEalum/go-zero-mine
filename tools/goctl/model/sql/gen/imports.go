package gen

import (
	"github.com/zeromicro/go-zero/tools/goctl/model/sql/template"
	"github.com/zeromicro/go-zero/tools/goctl/util"
	"github.com/zeromicro/go-zero/tools/goctl/util/pathx"
)

func genImports(table Table, serviceName string, withCache, timeImport bool) (string, error) {
	upperTblName := table.Name.ToCamel()

	if withCache {
		text, err := pathx.LoadTemplate(category, importsTemplateFile, template.Imports)
		if err != nil {
			return "", err
		}

		buffer, err := util.With("import").Parse(text).Execute(map[string]interface{}{
			"time":        timeImport,
			"data":        table,
			"serviceName": serviceName,
		})
		if err != nil {
			return "", err
		}

		return buffer.String(), nil
	}

	text, err := pathx.LoadTemplate(category, importsWithNoCacheTemplateFile, template.ImportsNoCache)
	if err != nil {
		return "", err
	}

	buffer, err := util.With("import").Parse(text).Execute(map[string]interface{}{
		"time":         timeImport,
		"data":         table,
		"upperTblName": upperTblName,
	})
	if err != nil {
		return "", err
	}

	return buffer.String(), nil
}

package gen

import (
	"github.com/zeromicro/go-zero/tools/goctl/model/sql/template"
	"github.com/zeromicro/go-zero/tools/goctl/util"
	"github.com/zeromicro/go-zero/tools/goctl/util/pathx"
)

func genTrans(table Table) (string, string, error) {
	camel := table.Name.ToCamel()
	text, err := pathx.LoadTemplate(category, transFile, template.Trans)
	if err != nil {
		return "", "", err
	}

	trans, err := util.With("trans").
		Parse(text).
		Execute(map[string]interface{}{
			"upperStartCamelObject": camel,
		})
	if err != nil {
		return "", "", err
	}

	text, err = pathx.LoadTemplate(category, transMethodTemplateFile, template.TransMethod)
	if err != nil {
		return "", "", err
	}

	transMethod, err := util.With("transMethod").
		Parse(text).
		Execute(map[string]interface{}{
			"upperStartCamelObject": camel,
		})
	if err != nil {
		return "", "", err
	}

	return trans.String(), transMethod.String(), nil
}

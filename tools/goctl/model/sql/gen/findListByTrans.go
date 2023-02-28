package gen

import (
	"github.com/zeromicro/go-zero/tools/goctl/model/sql/template"
	"github.com/zeromicro/go-zero/tools/goctl/util"
	"github.com/zeromicro/go-zero/tools/goctl/util/pathx"
)

func genFindListByTrans(table Table) (string, string, error) {
	camel := table.Name.ToCamel()
	text, err := pathx.LoadTemplate(category, findListByTransFile, template.FindListByTrans)
	if err != nil {
		return "", "", err
	}

	findListByTrans, err := util.With("findListByTrans").
		Parse(text).
		Execute(map[string]interface{}{
			"upperStartCamelObject": camel,
		})
	if err != nil {
		return "", "", nil
	}

	text, err = pathx.LoadTemplate(category, findListByTransTemplateFile, template.FindListByTransMethod)
	if err != nil {
		return "", "", err
	}

	findListByTransMethod, err := util.With("findListByTransMethod").
		Parse(text).
		Execute(map[string]interface{}{
			"upperStartCamelObject": camel,
		})
	if err != nil {
		return "", "", err
	}

	return findListByTrans.String(), findListByTransMethod.String(), nil
}

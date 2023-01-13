package gen

import (
	"github.com/zeromicro/go-zero/tools/goctl/model/sql/template"
	"github.com/zeromicro/go-zero/tools/goctl/util"
	"github.com/zeromicro/go-zero/tools/goctl/util/pathx"
)

func genFindOneByCondition(table Table) (string, string, error) {
	camel := table.Name.ToCamel()
	text, err := pathx.LoadTemplate(category, findOneByConditionFile, template.FindOneByCondition)
	if err != nil {
		return "", "", err
	}

	findOneByCondition, err := util.With("findOneByCondition").
		Parse(text).
		Execute(map[string]interface{}{
			"upperStartCamelObject": camel,
		})
	if err != nil {
		return "", "", err
	}

	text, err = pathx.LoadTemplate(category, findOneByFieldMethodTemplateFile, template.FindOneByConditionMethod)
	if err != nil {
		return "", "", err
	}

	findOneByConditionMethod, err := util.With("findOneByConditionMethod").
		Parse(text).
		Execute(map[string]interface{}{
			"upperStartCamelObject": camel,
		})

	return findOneByCondition.String(), findOneByConditionMethod.String(), nil
}

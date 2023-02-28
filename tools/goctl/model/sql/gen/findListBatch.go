package gen

import (
	"fmt"

	"github.com/zeromicro/go-zero/tools/goctl/model/sql/template"
	"github.com/zeromicro/go-zero/tools/goctl/util"
	"github.com/zeromicro/go-zero/tools/goctl/util/pathx"
)

func genFindListBatch(table Table) (string, string, error) {

	camel := table.Name.ToCamel()
	text, err := pathx.LoadTemplate(category, findListBatchFile, template.FindListBatch)
	if err != nil {
		return "", "", err
	}

	maxTag := fmt.Sprintf("`db:%s`", "\"MaxId\"")
	minTag := fmt.Sprintf("`db:%s`", "\"MinId\"")
	countTag := fmt.Sprintf("`db:%s`", "\"count\"")

	findListByTrans, err := util.With("findListBatch").
		Parse(text).
		Execute(map[string]interface{}{
			"upperStartCamelObject": camel,
			"maxTag":                maxTag,
			"minTag":                minTag,
			"countTag":              countTag,
		})
	if err != nil {
		return "", "", nil
	}

	text, err = pathx.LoadTemplate(category, findListBatchTemplateFile, template.FindListBatchMethod)
	if err != nil {
		return "", "", err
	}

	findListByTransMethod, err := util.With("findListBatch").
		Parse(text).
		Execute(map[string]interface{}{
			"upperStartCamelObject": camel,
		})
	if err != nil {
		return "", "", err
	}

	return findListByTrans.String(), findListByTransMethod.String(), nil
}

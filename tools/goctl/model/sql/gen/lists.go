package gen

import (
	"fmt"

	"github.com/zeromicro/go-zero/tools/goctl/model/sql/template"
	"github.com/zeromicro/go-zero/tools/goctl/util"
	"github.com/zeromicro/go-zero/tools/goctl/util/pathx"
	"github.com/zeromicro/go-zero/tools/goctl/util/stringx"
)

func genLists(table Table, withCache, postgreSql bool) (string, string, error) {
	camel := table.Name.ToCamel()
	text, err := pathx.LoadTemplate(category, listsTemplateFile, template.Lists)
	if err != nil {
		return "", "", err
	}

	countTag := fmt.Sprintf("`db:%s`", "\"count\"")

	output, err := util.With("lists").
		Parse(text).
		Execute(map[string]interface{}{
			"primaryKey":                table.PrimaryKey.Name.ToSnake(),
			"withCache":                 withCache,
			"upperStartCamelObject":     camel,
			"lowerStartCamelObject":     stringx.From(camel).Untitle(),
			"originalPrimaryKey":        wrapWithRawString(table.PrimaryKey.Name.Source(), postgreSql),
			"lowerStartCamelPrimaryKey": util.EscapeGolangKeyword(stringx.From(table.PrimaryKey.Name.ToCamel()).Untitle()),
			"dataType":                  table.PrimaryKey.DataType,
			"cacheKey":                  table.PrimaryCacheKey.KeyExpression,
			"cacheKeyVariable":          table.PrimaryCacheKey.KeyLeft,
			"postgreSql":                postgreSql,
			"data":                      table,
			"countTag":                  countTag,
		})
	if err != nil {
		return "", "", err
	}

	text, err = pathx.LoadTemplate(category, listsTemplateMethodFile, template.ListsMethod)
	if err != nil {
		return "", "", err
	}

	listsMethod, err := util.With("listsMethod").
		Parse(text).
		Execute(map[string]interface{}{
			"upperStartCamelObject": camel,
		})
	if err != nil {
		return "", "", err
	}

	return output.String(), listsMethod.String(), nil
}

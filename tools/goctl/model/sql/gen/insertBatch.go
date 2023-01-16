package gen

import (
	"strings"

	"github.com/zeromicro/go-zero/tools/goctl/model/sql/template"
	"github.com/zeromicro/go-zero/tools/goctl/util"
	"github.com/zeromicro/go-zero/tools/goctl/util/pathx"
	"github.com/zeromicro/go-zero/tools/goctl/util/stringx"
)

func genInsertBatch(table Table) (string, string, error) {

	expressions := make([]string, 0)
	expressionValues := make([]string, 0)
	for _, field := range table.Fields {
		camel := util.SafeString(field.Name.ToCamel())
		if camel == "CreateTime" || camel == "UpdateTime" {
			continue
		}

		if field.Name.Source() == table.PrimaryKey.Name.Source() {
			if table.PrimaryKey.AutoIncrement {
				continue
			}
		}

		expressions = append(expressions, "?")
		// modify time default val
		if camel == "CreatedAt" || camel == "UpdatedAt" {
			expressionValues = append(expressionValues, "time.Now().Unix()")
		} else {
			expressionValues = append(expressionValues, "data."+camel)
		}

	}
	text, err := pathx.LoadTemplate(category, insertBatchFile, template.InsertBatch)
	if err != nil {
		return "", "", err
	}

	camel := table.Name.ToCamel()
	insertBatch, err := util.With("insertBatch").
		Parse(text).
		Execute(map[string]interface{}{
			"upperStartCamelObject": camel,
			"lowerStartCamelObject": stringx.From(camel).Untitle(),
			"expression":            strings.Join(expressions, ", "),
			"expressionValues":      strings.Join(expressionValues, ", "),
		})
	if err != nil {
		return "", "", err
	}

	text, err = pathx.LoadTemplate(category, insertBatchFileTemplate, template.InsertBatchMethod)
	if err != nil {
		return "", "", err
	}

	insertBatchMethod, err := util.With("insertBatchMethod").
		Parse(text).
		Execute(map[string]interface{}{
			"upperStartCamelObject": camel,
		})

	return insertBatch.String(), insertBatchMethod.String(), nil
}

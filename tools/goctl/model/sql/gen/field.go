package gen

import (
	"strings"

	"github.com/zeromicro/go-zero/tools/goctl/model/sql/parser"
	"github.com/zeromicro/go-zero/tools/goctl/model/sql/template"
	"github.com/zeromicro/go-zero/tools/goctl/util"
	"github.com/zeromicro/go-zero/tools/goctl/util/pathx"
)

func genFields(table Table, fields []*parser.Field) (string, error) {
	var list []string

	for _, field := range fields {
		result, err := genField(table, field)
		if err != nil {
			return "", err
		}

		list = append(list, result)
	}

	return strings.Join(list, "\n"), nil
}

func genFieldParser(table Table, fields []*parser.Field) (string, string, error) {
	var marshalList []string
	var unmarshalList []string

	for _, field := range fields {
		marshalRes, err := genMarshalFields(table, field)
		if err != nil {
			return "", "", err
		}

		unmarshalRes, err := genUnmarshalFields(table, field)
		if err != nil {
			return "", "", err
		}

		marshalList = append(marshalList, marshalRes)
		unmarshalList = append(unmarshalList, unmarshalRes)
	}

	return strings.Join(marshalList, "\n"), strings.Join(unmarshalList, "\n"), nil
}

func genField(table Table, field *parser.Field) (string, error) {
	tag, err := genTag(table, field.NameOriginal)
	if err != nil {
		return "", err
	}

	text, err := pathx.LoadTemplate(category, fieldTemplateFile, template.Field)
	if err != nil {
		return "", err
	}

	output, err := util.With("types").
		Parse(text).
		Execute(map[string]interface{}{
			"name":       util.SafeString(field.Name.ToCamel()),
			"type":       field.DataType,
			"tag":        tag,
			"hasComment": field.Comment != "",
			"comment":    field.Comment,
			"data":       table,
		})
	if err != nil {
		return "", err
	}

	return output.String(), nil
}

func genMarshalFields(table Table, field *parser.Field) (string, error) {
	text, err := pathx.LoadTemplate(category, fieldTemplateFile, template.MarshalFields)
	if err != nil {
		return "", err
	}

	output, err := util.With("types").
		Parse(text).
		Execute(map[string]interface{}{
			"name": util.SafeString(field.Name.ToCamel()),
		})
	if err != nil {
		return "", err
	}

	return output.String(), nil
}

func genUnmarshalFields(table Table, field *parser.Field) (string, error) {
	text, err := pathx.LoadTemplate(category, fieldTemplateFile, template.UnmarshalFields)
	if err != nil {
		return "", err
	}

	output, err := util.With("types").
		Parse(text).
		Execute(map[string]interface{}{
			"name": util.SafeString(field.Name.ToCamel()),
		})
	if err != nil {
		return "", err
	}

	return output.String(), nil
}

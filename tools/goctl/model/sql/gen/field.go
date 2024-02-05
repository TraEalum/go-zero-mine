package gen

import (
	"fmt"
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

func genFieldParser(table Table, fields []*parser.Field, skipFields map[string]struct{}) (string, string, error) {
	var marshalList []string
	var unmarshalList []string

	for _, field := range fields {
		marshalRes, err := genMarshalFields(table, field, skipFields)
		if err != nil {
			return "", "", err
		}

		unmarshalRes, err := genUnmarshalFields(table, field, skipFields)
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

func genMarshalFields(table Table, field *parser.Field, skipFields map[string]struct{}) (string, error) {
	text, err := pathx.LoadTemplate(category, fieldTemplateFile, template.MarshalFields)
	if err != nil {
		return "", err
	}

	name := "m." + util.SafeString(field.Name.ToCamel()) // 左侧model的名称
	_, ok := skipFields[field.Name.Source()]
	if ok {
		name = "// " + name
	}

	protoConvertType := getMarshalProtoDataType(field)

	output, err := util.With("types").
		Parse(text).
		Execute(map[string]interface{}{
			"name":      name,
			"protoName": protoConvertType,
		})
	if err != nil {
		return "", err
	}

	return output.String(), nil
}

func genUnmarshalFields(table Table, field *parser.Field, skipFields map[string]struct{}) (string, error) {
	text, err := pathx.LoadTemplate(category, fieldTemplateFile, template.UnmarshalFields)
	if err != nil {
		return "", err
	}

	name := "p." + util.SafeString(field.Name.ToCamel())
	_, ok := skipFields[field.Name.Source()]
	if ok {
		name = "// " + name
	}

	modelName := getUnmarshalModelDataType(field)

	output, err := util.With("types").
		Parse(text).
		Execute(map[string]interface{}{
			"name":      name,
			"modelName": modelName,
		})
	if err != nil {
		return "", err
	}

	return output.String(), nil
}

func getMarshalProtoDataType(field *parser.Field) string {
	fileName := field.Name.ToCamel()

	switch field.DataType {
	case "sql.NullString":
		{
			return fmt.Sprintf("&sql.NullString{String:p.%s, Valid: true}", fileName)
		}
	case "sql.NullInt16":
		{
			return fmt.Sprintf("&sql.NullInt16{Int16:p.%s, Valid: true}", fileName)
		}
	case "sql.NullInt32":
		{
			return fmt.Sprintf("&sql.NullInt32{Int32:p.%s, Valid: true}", fileName)
		}
	case "sql.NullInt64":
		{
			return fmt.Sprintf("&sql.NullInt64{Int64: int64(p.%s), Valid: true}", fileName)
		}
	case "sql.NullBool":
		{
			return fmt.Sprintf("&sql.NullBool{Bool:p.%s, Valid: true}", fileName)
		}
	case "sql.NullByte":
		{
			return fmt.Sprintf("&sql.NullByte{Byte:p.%s, Valid: true}", fileName)
		}
	case "sql.NullFloat64":
		{
			return fmt.Sprintf("&sql.NullFloat64{Float64:p.%s, Valid: true}", fileName)
		}
	default:
		{
			return fmt.Sprintf("&p.%s", fileName)
		}
	}
}

func getUnmarshalModelDataType(field *parser.Field) string {
	fileName := field.Name.ToCamel()

	switch field.DataType {
	case "sql.NullString":
		{
			return fmt.Sprintf("m.%s.String", fileName)
		}
	case "sql.NullInt16":
		{
			return fmt.Sprintf("m.%s.Int16}", fileName)
		}
	case "sql.NullInt32":
		{
			return fmt.Sprintf("m.%s.Int32", fileName)
		}
	case "sql.NullInt64":
		{
			return fmt.Sprintf("uint64(m.%s.Int64)", fileName)
		}
	case "sql.NullBool":
		{
			return fmt.Sprintf("m.%s.Bool", fileName)
		}
	case "sql.NullByte":
		{
			return fmt.Sprintf("m.%s.Byte", fileName)
		}
	case "sql.NullFloat64":
		{
			return fmt.Sprintf("m.%s.Float64", fileName)
		}
	default:
		{
			return fmt.Sprintf("*m.%s", fileName)
		}
	}
}

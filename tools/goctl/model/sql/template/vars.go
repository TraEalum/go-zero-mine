package template

import "fmt"

// Vars defines a template for var block in model
var Vars = fmt.Sprintf(
	`
var (
	{{.lowerStartCamelObject}}FieldNames          = builder.RawFieldNames(&{{.upperStartCamelObject}}{}{{if .postgreSql}},true{{end}})
	{{.lowerStartCamelObject}}Rows                = strings.Join({{.lowerStartCamelObject}}FieldNames, ",")
	{{.lowerStartCamelObject}}RowsExpectAutoSet   = {{if .postgreSql}}strings.Join(stringx.Remove({{.lowerStartCamelObject}}FieldNames, {{if .autoIncrement}}"{{.originalPrimaryKey}}"{{end}}), ","){{else}}strings.Join(stringx.Remove({{.lowerStartCamelObject}}FieldNames, {{if .autoIncrement}}"{{.originalPrimaryKey}}"{{end}}), ","){{end}}
	{{.lowerStartCamelObject}}RowsWithPlaceHolder = {{if .postgreSql}}builder.PostgreSqlJoin(stringx.Remove({{.lowerStartCamelObject}}FieldNames, "{{.originalPrimaryKey}}")){{else}}strings.Join(stringx.Remove({{.lowerStartCamelObject}}FieldNames, "{{.originalPrimaryKey}}"), "=?,") + "=?"{{end}}

	{{if .withCache}}{{.cacheKeys}}{{end}}
)
`)

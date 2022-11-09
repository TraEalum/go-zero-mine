package template

// Field defines a filed template for types
const Field = `{{.name}} *{{.type}} {{.tag}} {{if .hasComment}}//  {{.comment}} // src:tools/goctl/model/sql/template/field.go {{end}}`

const MarshalFields = `m.{{.name}} = &p.{{.name}} // src:tools/goctl/model/sql/template/field.go`

const UnmarshalFields = `p.{{.name}} = *m.{{.name}} // src:tools/goctl/model/sql/template/field.go`

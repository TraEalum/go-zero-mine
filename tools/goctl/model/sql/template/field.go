package template

// Field defines a filed template for types
const Field = `{{.name}} {{.type}} {{.tag}} {{if .hasComment}}// {{.comment}}  // src:tools/goctl/model/sql/template/field.go@Field {{end}}`
const FieldPtr = `{{.name}} *{{.type}} {{.tag}} {{if .hasComment}}//  {{.comment}} // src:tools/goctl/model/sql/template/field.go@FieldPtr {{end}}`

const MarshalFields = `m.{{.name}} = p.{{.name}} // src:tools/goctl/model/sql/template/field.go@MarshalFields`

const MarshalFieldsUpdate = `m.{{.name}} = *p.{{.name}} // src:tools/goctl/model/sql/template/field.go@MarshalFieldsUpdate`

const UnmarshalFields = `p.{{.name}} = m.{{.name}} // src:tools/goctl/model/sql/template/field.go@UnmarshalFields`

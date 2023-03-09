package template

// Field defines a filed template for types
const Field = `{{.name}} *{{.type}} {{.tag}} {{if .hasComment}}//  {{.comment}} {{end}}`

const MarshalFields = `m.{{.name}} = {{.protoName}} `

const UnmarshalFields = `p.{{.name}} = {{.modelName}} `

package template

// Field defines a filed template for types
const Field = `{{.name}} *{{.type}} {{.tag}} {{if .hasComment}}//  {{.comment}} {{end}}`

const MarshalFields = `{{.name}} = {{.protoName}} `

const UnmarshalFields = `{{.name}} = {{.modelName}} `

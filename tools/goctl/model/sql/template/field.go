package template

// Field defines a filed template for types
const Field = `{{.name}} {{.type}} {{.tag}} {{if .hasComment}}// {{.comment}}{{end}}`

const MarshalFields = `m.{{.name}} = p.{{.name}}`

const UnmarshalFields = `p.{{.name}} = m.{{.name}}`

package template

import (
	"fmt"

	"github.com/zeromicro/go-zero/tools/goctl/util"
)

// ModelCustom defines a template for extension
const ModelCustom = `package {{.pkg}}
{{if .withCache}}
import (
	proto "proto/{{.serviceName}}"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)
{{else}}
import (
	proto "proto/{{.serviceName}}"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)
{{end}}
var _ {{.upperStartCamelObject}}Model = (*custom{{.upperStartCamelObject}}Model)(nil)

type (
	// {{.upperStartCamelObject}}Model is an interface to be customized, add more methods here,
	// and implement the added methods in custom{{.upperStartCamelObject}}Model.
	{{.upperStartCamelObject}}Model interface {
		{{.lowerStartCamelObject}}Model
	}

	custom{{.upperStartCamelObject}}Model struct {
		*default{{.upperStartCamelObject}}Model
	}
)

// New{{.upperStartCamelObject}}Model returns a model for the database table.
func New{{.upperStartCamelObject}}Model(conn sqlx.SqlConn{{if .withCache}}, c cache.CacheConf{{end}}) {{.upperStartCamelObject}}Model {
	return &custom{{.upperStartCamelObject}}Model{
		default{{.upperStartCamelObject}}Model: new{{.upperStartCamelObject}}Model(conn{{if .withCache}}, c{{end}}),
	}
}

func (m *{{.upperStartCamelObject}}) marshal(p *proto.{{.upperStartCamelObject}}) error {
	m.marshal(p)

	return nil
}

func (m *{{.upperStartCamelObject}}) unmarshal(p *proto.{{.upperStartCamelObject}}) error {
	m.unmarshal(p)
	
	return nil
}
`

// ModelCustomSubTable defines a template for extension
const ModelCustomSubTable = `package {{.pkg}}
{{if .withCache}}
import (
    "fmt"
	proto "proto/{{.serviceName}}"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)
{{else}}
import (
	proto "proto/{{.serviceName}}"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)
{{end}}
var _ {{.upperStartCamelObject}}Model = (*custom{{.upperStartCamelObject}}Model)(nil)

type (
	// {{.upperStartCamelObject}}Model is an interface to be customized, add more methods here,
	// and implement the added methods in custom{{.upperStartCamelObject}}Model.
	{{.upperStartCamelObject}}Model interface {
		{{.lowerStartCamelObject}}Model
	}

	custom{{.upperStartCamelObject}}Model struct {
		*default{{.upperStartCamelObject}}Model
	}
)

// New{{.upperStartCamelObject}}Model returns a model for the database table.
func New{{.upperStartCamelObject}}Model(conn sqlx.SqlConn{{if .withCache}}, c cache.CacheConf{{end}}) {{.upperStartCamelObject}}Model {
	return &custom{{.upperStartCamelObject}}Model{
		default{{.upperStartCamelObject}}Model: new{{.upperStartCamelObject}}Model(conn{{if .withCache}}, c{{end}}),
	}
}

func (m *{{.upperStartCamelObject}}) TableName() string {
	return fmt.Sprintf("{{.fmtSubTableName}}", m.{{.upperSubTableKey}}%{{.subTableNumber}})
}
`

// ModelGen defines a template for model
var ModelGen = fmt.Sprintf(`%s

package {{.pkg}}
{{.imports}}
{{.vars}}
{{.types}}
{{.new}}

func (m *{{.upperStartCamelObject}}) marshal(p *proto.{{.upperStartCamelObject}}) error {
	{{.marshalFields}}

	return nil
}

func (m *{{.upperStartCamelObject}}) unmarshal(p *proto.{{.upperStartCamelObject}}) error {
	{{.unmarshallFields}}
	
	return nil
}

func Marshal{{.upperStartCamelObject}}Lst(lst *[]{{.upperStartCamelObject}}, protoLst []*proto.{{.upperStartCamelObject}}) {
	for _, v := range protoLst {
		var tmp {{.upperStartCamelObject}}
		tmp.Marshal(v)
		*lst = append(*lst, tmp)
	}
}

func Unmarshal{{.upperStartCamelObject}}Lst(protoLst *[]*proto.{{.upperStartCamelObject}}, lst []{{.upperStartCamelObject}}) {
	for _, v := range lst {
		var tmp proto.{{.upperStartCamelObject}}
		v.Unmarshal(&tmp)
		*protoLst = append(*protoLst, &tmp)
	}
}

func (m *{{.upperStartCamelObject}}) TableName() string {
	return "{{.table}}"
}

{{.insert}}
{{.find}}
{{.update}}
{{.delete}}
{{.findlist}}
{{.extraMethod}}
{{.tableName}}
`, util.DoNotEditHead)

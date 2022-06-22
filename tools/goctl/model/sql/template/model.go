package template

import (
	"fmt"

	"github.com/zeromicro/go-zero/tools/goctl/util"
)

// ModelCustom defines a template for extension
const ModelCustom = `package {{.pkg}}
{{if .withCache}}
import (
	"go-service/app/{{.serviceName}}/rpc/proto"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)
{{else}}
import (
	"go-service/app/{{.serviceName}}/rpc/proto"

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


func (m *{{.upperStartCamelObject}}) Marshal(p *proto.{{.upperStartCamelObject}}) error {
	{{.marshalFields}}

	return nil
}

func (m *{{.upperStartCamelObject}}) Unmarshal(p *proto.{{.upperStartCamelObject}}) error {
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
`

// ModelGen defines a template for model
var ModelGen = fmt.Sprintf(`%s

package {{.pkg}}
{{.imports}}
{{.vars}}
{{.types}}
{{.new}}
{{.insert}}
{{.find}}
{{.update}}
{{.delete}}
{{.findlist}}
{{.extraMethod}}
{{.tableName}}
`, util.DoNotEditHead)

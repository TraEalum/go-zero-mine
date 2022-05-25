package template

// Error defines an error template
const Error = `package {{.pkg}}

import (
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"go-service/app/{{.serviceName}}/rpc/proto"
)


var ErrNotFound = sqlx.ErrNotFound

func (m *default{{.upperStartCamelObject}}Model) Marshal(o *proto.{{.upperStartCamelObject}}) error {
}
`
